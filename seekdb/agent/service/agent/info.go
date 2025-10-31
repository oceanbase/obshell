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
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

func (s *AgentService) updateInfo(db *gorm.DB, info *sqlite.OcsInfo) error {
	return db.Model(info).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(info).Error
}

func (s *AgentService) updateIdentity(db *gorm.DB, identity meta.AgentIdentity) error {
	info := sqlite.OcsInfo{
		Name:  constant.OCS_INFO_IDENTITY,
		Value: string(identity),
	}
	if err := s.updateInfo(db, &info); err != nil {
		return err
	}
	ocsAgent.Identity = identity
	return nil
}

func (s *AgentService) updateAgentInfo(db *gorm.DB, agentInfo meta.AgentInfoInterface) (err error) {

	infos := []*sqlite.OcsInfo{
		{Name: constant.OCS_INFO_IP, Value: agentInfo.GetIp()},
		{Name: constant.OCS_INFO_PORT, Value: fmt.Sprintf("%d", agentInfo.GetPort())},
	}

	for _, info := range infos {
		if err = s.updateInfo(db, info); err != nil {
			return
		}
	}
	ocsAgent.Ip = agentInfo.GetIp()
	ocsAgent.Port = agentInfo.GetPort()
	return nil
}

func (s *AgentService) UpdateAgentIP(ip string) error {
	if ocsAgent == nil {
		return errors.Occur(errors.ErrAgentNotInitialized)
	}
	if ocsAgent.GetIp() != ip {
		if !ocsAgent.IsSingleAgent() && !ocsAgent.IsUnidentified() {
			// IP recorded in meta is inconsistent with IP recorded in ob config.bin
			return errors.Occur(errors.ErrAgentIpInconsistentWithOBServer)
		}
		ocsAgent.Ip = ip
	}
	return nil
}

func (s *AgentService) UpdateAgentInfo(agentInfo meta.AgentInfoInterface) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// Agent is not under maintenance.
		var status int
		if err = tx.Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", constant.OCS_INFO_STATUS).Scan(&status).Error; err != nil {
			return err
		}
		if status == task.GLOBAL_MAINTENANCE {
			return errors.Occur(errors.ErrAgentUnderMaintenance, agentInfo.String())
		}
		return s.updateAgentInfo(tx, agentInfo)
	})
}

func (s *AgentService) UpdateBaseInfo() error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) (err error) {
		ocsInfoList := []*sqlite.OcsInfo{
			{Name: constant.OCS_INFO_OS, Value: global.Os},
			{Name: constant.OCS_INFO_ARCHITECTURE, Value: global.Architecture},
		}

		var curVersion string
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", constant.OCS_INFO_VERSION).Find(&curVersion).Error; err != nil {
			return err
		}

		if curVersion != constant.VERSION {
			ocsInfoList = append(ocsInfoList, &sqlite.OcsInfo{Name: constant.OCS_INFO_VERSION, Value: constant.VERSION})
		}

		for _, info := range ocsInfoList {
			if err := s.updateInfo(tx, info); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AgentService) GetIP() (ip string, err error) {
	err = getOCSInfo(constant.OCS_INFO_IP, &ip)
	return
}

func getOCSInfo(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", key).First(value).Error
	return
}
