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
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

const obSysParameterQuery = `
	SELECT NAME, DATA_TYPE, VALUE, INFO, SECTION, EDIT_LEVEL, DEFAULT_VALUE,
	       CASE WHEN ISDEFAULT = 'YES' THEN 1 ELSE 0 END AS IS_DEFAULT
	FROM oceanbase.V$OB_PARAMETERS
`

func queryObSysParameters(db *gorm.DB) ([]sqlite.ObSysParameter, error) {
	var parameters []sqlite.ObSysParameter
	err := db.Raw(obSysParameterQuery).Find(&parameters).Error
	return parameters, err
}

func (s *AgentService) TakeOver() (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	agent := sqlite.AllAgent{
		Ip:           meta.OCS_AGENT.GetIp(),
		Port:         meta.OCS_AGENT.GetPort(),
		Os:           global.Os,
		Architecture: global.Architecture,
		HomePath:     global.HomePath,
		MysqlPort:    meta.MYSQL_PORT,
		Version:      meta.OCS_AGENT.GetVersion(),
		Identity:     string(meta.CLUSTER_AGENT),
	}
	return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
		if err = sqliteTx.Model(&sqlite.AllAgent{}).Delete(&sqlite.AllAgent{}, "1=1").Error; err != nil {
			return err
		}
		if err = sqliteTx.Model(&sqlite.AllAgent{}).Create(&agent).Error; err != nil {
			return err
		}
		return s.TakeOverOrRebuild(sqliteTx)
	})
}

func (s *AgentService) Rebuild(agentInstance *meta.AgentInstance) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
		return s.TakeOverOrRebuild(sqliteTx)
	})
}

func (s *AgentService) UpdateAgentVersion() (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Model(&sqlite.AllAgent{}).Where("ip=? and port=?", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()).Update("version", meta.OCS_AGENT.GetVersion()).Error
}

func (s *AgentService) TakeOverOrRebuild(sqliteTx *gorm.DB) error {
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
	if err := s.initializeClusterStatus(sqliteTx); err != nil {
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
	if err := s.updateOBConfig(sqliteTx, constant.CONFIG_CREATED_TIME, observer.CreateTime); err != nil {
		return err
	}
	if err := s.updateObserverObVersion(sqliteTx, obVersion); err != nil {
		return err
	}
	err = sqliteTx.Exec("delete from " + constant.TABLE_OB_SYS_PARAMETER).Error
	if err != nil {
		return err
	}
	obConn, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	obSysParameter, err := queryObSysParameters(obConn)
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
