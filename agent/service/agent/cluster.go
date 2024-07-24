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

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func (s *AgentService) ConvertToClusterAgent(agent meta.Agent) error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
		return oceanbaseDb.Transaction(func(oceanbaseTx *gorm.DB) error {
			isMaster := agent.IsMasterAgent() || agent.IsTakeOverMasterAgent()
			agentDO := oceanbase.AllAgent{
				Ip:       agent.GetIp(),
				Port:     agent.GetPort(),
				Identity: string(meta.CLUSTER_AGENT),
			}
			if err := oceanbaseTx.Updates(&agentDO).Error; err != nil {
				return err
			}

			if ocsAgent.Equal(agent) {
				if err = s.updateIdentity(sqliteTx, meta.CLUSTER_AGENT); err != nil {
					return err
				}
			}

			if isMaster {
				if err := s.initializeMaintainer(oceanbaseTx); err != nil {
					return err
				}
				return s.initializeClusterStatus(oceanbaseTx)
			}
			return nil
		})
	})
}

func (s *AgentService) initializeMaintainer(tx *gorm.DB) error {
	var maintainer oceanbase.TaskMaintainer
	if err := tx.Model(&oceanbase.TaskMaintainer{}).Where("id = 1").Scan(&maintainer).Error; err != nil {
		return err
	}
	if maintainer.Id == 0 {
		return tx.Model(&oceanbase.TaskMaintainer{}).Create(&oceanbase.TaskMaintainer{
			Id:        1,
			AgentIp:   meta.OCS_AGENT.GetIp(),
			AgentPort: meta.OCS_AGENT.GetPort(),
			AgentTime: time.Now().Unix(),
		}).Error
	}
	return nil
}

func (s *AgentService) initializeClusterStatus(tx *gorm.DB) error {
	var clusterStatus oceanbase.ClusterStatus
	if err := tx.Model(&oceanbase.ClusterStatus{}).Where("id = 1").Scan(&clusterStatus).Error; err != nil {
		return err
	}
	if clusterStatus.Id == 0 {
		return tx.Model(&oceanbase.ClusterStatus{}).Create(&oceanbase.ClusterStatus{
			Id:     1,
			Status: task.NOT_UNDER_MAINTENANCE,
		}).Error
	}
	return nil
}
