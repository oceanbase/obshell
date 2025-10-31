/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agent

import (
	"time"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
)

func (s *AgentService) GetAgentInstanceByIpAndRpcPortFromOB(ip string, mysqlPort int) (agent *meta.AgentInstance, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	hasTable := oceanbaseDb.Migrator().HasTable(&oceanbase.AllAgent{})
	if !hasTable {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Where("ip=? and mysql_port=?", ip, mysqlPort).Scan(&agent).Error
	return
}

func (s *AgentService) UpdateAgentPublicKey(publicKey string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Where("ip=? and port=?", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()).Update(constant.AGENT_PUBLIC_KEY, publicKey).Error
	return
}

func (s *AgentService) TakeOver() (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	agent := oceanbase.AllAgent{
		Ip:           meta.OCS_AGENT.GetIp(),
		Port:         meta.OCS_AGENT.GetPort(),
		Os:           global.Os,
		Architecture: global.Architecture,
		HomePath:     global.HomePath,
		MysqlPort:    meta.MYSQL_PORT,
		Version:      meta.OCS_AGENT.GetVersion(),
		PublicKey:    secure.Public(),
		Identity:     string(meta.CLUSTER_AGENT),
	}
	err = oceanbaseDb.Transaction(func(oceanbaseTx *gorm.DB) error {
		return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
			// clear all_agent
			if err = oceanbaseTx.Model(&oceanbase.AllAgent{}).Delete(&oceanbase.AllAgent{}, "1=1").Error; err != nil {
				return err
			}
			if err = oceanbaseTx.Model(&oceanbase.AllAgent{}).Create(&agent).Error; err != nil {
				return err
			}
			return s.TakeOverOrRebuild(sqliteTx, oceanbaseTx)
		})
	})
	return
}

func (s *AgentService) Rebuild(agentInstance *meta.AgentInstance) (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Transaction(func(oceanbaseTx *gorm.DB) error {
		return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
			return s.TakeOverOrRebuild(sqliteTx, oceanbaseTx)
		})
	})
}

func (s *AgentService) UpdateAgentVersion() (err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return db.Model(&oceanbase.AllAgent{}).Where("ip=? and port=?", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()).Update("version", meta.OCS_AGENT.GetVersion()).Error
}

func (s *AgentService) TakeOverOrRebuild(sqliteTx *gorm.DB, oceanbaseTx *gorm.DB) error {
	var err error
	observerUser, err := process.GetObserverUser()
	if err != nil {
		return err
	}
	if err = s.updateIdentity(sqliteTx, meta.CLUSTER_AGENT); err != nil {
		return err
	}
	if err := s.updateObserverUser(sqliteTx, observerUser); err != nil {
		return err
	}
	// update obshell type
	if err := s.updateOBConfig(sqliteTx, constant.CONFIG_OBSHELL_TYPE, "seekdb"); err != nil {
		return err
	}
	// update data dir
	dataDir, err := observerService.GetOBStringParatemerByName(constant.CONFIG_DATA_DIR)
	if err != nil {
		return err
	}
	if err := s.updateObserverDataDir(sqliteTx, dataDir); err != nil {
		return err
	}
	// update redo dir
	redoDir, err := observerService.GetOBStringParatemerByName(constant.CONFIG_REDO_DIR)
	if err != nil {
		return err
	}
	if err := s.updateObserverRedoDir(sqliteTx, redoDir); err != nil {
		return err
	}
	// update redo dir
	if err := s.initializeClusterStatus(oceanbaseTx); err != nil {
		return err
	}
	// update obversion
	obVersion, err := obclusterService.GetObVersion()
	if err != nil {
		return err
	}
	// update ob create time
	observer, err := obclusterService.GetOBServer()
	if err != nil {
		return err
	}
	if err := s.updateOBConfig(sqliteTx, constant.CONFIG_CREATED_TIME, observer.CreateTime.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := s.updateObserverObVersion(sqliteTx, obVersion); err != nil {
		return err
	}
	err = sqliteTx.Exec("delete from " + constant.TABLE_OB_SYS_PARAMETER).Error
	if err != nil {
		return err
	}
	var obSysParameter []sqlite.ObSysParameter
	err = oceanbaseTx.Raw("select * from oceanbase.GV$OB_PARAMETERS").Find(&obSysParameter).Error
	if err != nil {
		return err
	}
	for _, parameter := range obSysParameter {
		err = sqliteTx.Model(&sqlite.ObSysParameter{}).Create(&parameter).Error
		if err != nil {
			return err
		}
	}
	return nil
}
