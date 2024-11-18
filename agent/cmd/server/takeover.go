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

package server

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

func (a *Agent) handleTakeOverOrRebuild() {
	if err := a.takeOverOrRebuild(); err != nil {
		log.WithError(err).Error("take over or rebuild failed")
		process.ExitWithFailure(constant.EXIT_CODE_ERROR_TAKE_OVER_FAILED, fmt.Sprintf("take over or rebuild failed: %v", err))
	}
	a.Server.startChan <- true
}

func (a *Agent) takeOverOrRebuild() (err error) {
	log.Info("start to take over or rebuild")
	for {
		if _, err = oceanbase.GetAvailableInstance(); err == nil {
			break
		} else if errors.Is(err, oceanbase.ERR_OBSERVER_NOT_EXIST) {
			return err
		} else {
			if !oceanbase.IsConnecting() {
				if _, err = oceanbase.GetAvailableInstance(); err != nil {
					return err
				}
			}
			log.WithError(err).Warn("get ob connection failed")
		}
		time.Sleep(constant.GET_INSTANCE_RETRY_INTERVAL * time.Second)
	}

	if err = oceanbase.CreateDataBase(constant.DB_OCS); err != nil {
		return err
	}

	for {
		if _, err = oceanbase.GetOcsInstance(); err == nil {
			break
		} else {
			// Because initConnection() only returns when there's a password error or on success
			// and the password must be correct at this point, so we can wait here until the connection is successful
			log.WithError(err).Error("get ocs db connection failed")
		}
		time.Sleep(constant.GET_INSTANCE_RETRY_INTERVAL * time.Second)
	}

	if err = oceanbase.AutoMigrateObTables(true); err != nil {
		return err
	}

	var agentInstance *meta.AgentInstance
	if agentInstance, err = agentService.GetAgentInstanceByIpAndRpcPortFromOB(meta.OCS_AGENT.GetIp(), meta.RPC_PORT); err != nil {
		log.WithError(err).Errorf("get ocs agent by ip and rpc port failed, ip: %s, rpc port: %d", meta.OCS_AGENT.GetIp(), meta.RPC_PORT)
		return err
	}
	if agentInstance == nil {
		log.Infof("agent with ip %s and rpc port %d not found, need to take over", meta.OCS_AGENT.GetIp(), meta.RPC_PORT)
		if err = ob.TakeOver(); err != nil {
			log.WithError(err).Error("take over failed")
			return err
		} else {
			return nil
		}
	}
	if err = ob.Rebuild(agentInstance); err != nil {
		log.WithError(err).Error("rebuild failed")
		return err
	}
	return nil
}
