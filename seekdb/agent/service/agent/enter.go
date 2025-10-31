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
	"strconv"

	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/service/obcluster"
)

var (
	ErrOtherAgentUpgrading = errors.Occur(errors.ErrCommonUnexpected, "other agents are upgrading the binary")
	observerService        = obcluster.ObserverService{}
	obclusterService       = obcluster.ObclusterService{}
)

type AgentService struct{}

type Agent struct {
	meta.AgentInstance
}

var ocsAgent *Agent

func (s *AgentService) InitAgent() error {
	if meta.OCS_AGENT != nil {
		return errors.Occur(errors.ErrAgentAlreadyInitialized)
	}

	agentInstance, err := s.getAgentInfo()
	if err != nil {
		return err
	}

	ocsAgent = &Agent{
		AgentInstance: agentInstance,
	}
	switch agentInstance.Identity {
	case meta.SINGLE:
	case meta.MASTER:
		fallthrough
	case meta.TAKE_OVER_MASTER:
		fallthrough
	case meta.CLUSTER_AGENT:
		if err := s.initOBPort(); err != nil {
			return err
		}
	default:
	}

	meta.OCS_AGENT = ocsAgent
	return nil
}

func (s *AgentService) initOBPort() error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		defer func() {
			if err != nil {
				meta.MYSQL_PORT = 0
			}
		}()

		return s.getOBConifg(tx, constant.CONFIG_MYSQL_PORT, &meta.MYSQL_PORT)
	})
}

func (agentService *AgentService) InitializeAgentStatus() (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	if err = db.Create(&sqlite.OcsInfo{Name: constant.OCS_INFO_STATUS, Value: strconv.Itoa(task.NOT_UNDER_MAINTENANCE)}).Error; err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if !ok || sqliteErr.Code != sqlite3.ErrConstraint {
			return
		}
	}
	return nil
}

func (s *AgentService) getAgentInfo() (agentInfo meta.AgentInstance, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}

	var data []sqlite.OcsInfo
	if err = sqliteDb.Model(&sqlite.OcsInfo{}).Scan(&data).Error; err != nil {
		return
	}

	agentInfo.Identity = meta.UNIDENTIFIED
	agentInfo.Version = constant.VERSION_RELEASE
	for _, info := range data {
		switch info.Name {
		case constant.OCS_INFO_IDENTITY:
			agentInfo.Identity = meta.AgentIdentity(info.Value)
		case constant.OCS_INFO_IP:
			agentInfo.Ip = info.Value
		case constant.OCS_INFO_PORT:
			if port, err := strconv.Atoi(info.Value); err != nil {
				return agentInfo, err
			} else {
				agentInfo.Port = port
			}
		}
	}
	return
}
