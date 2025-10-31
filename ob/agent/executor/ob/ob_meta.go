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

package ob

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/service/agent"
)

func HandleOBMeta() (err error) {
	log.Info("start to handle ob meta")
	for {
		if oceanbase.IsConnecting() {
			// If the connection is being established, wait for it to complete
		} else if _, err = oceanbase.GetOcsInstance(); err == nil {
			break
		} else {
			// Because initConnection() only returns when there's a password error or on success
			// and the password must be correct at this point, so we can wait here until the connection is successful
			log.WithError(err).Error("get ocs db connection failed")
		}
		time.Sleep(constant.GET_INSTANCE_RETRY_INTERVAL * time.Second)
	}

	log.Info("try to start migrate table")
	if err = oceanbase.AutoMigrateObTables(true); err != nil {
		log.WithError(err).Error("auto migrate ob tables failed")
		return
	}

	log.Info("try to update agent version")
	if err = agentService.UpdateAgentVersion(); err != nil {
		log.WithError(err).Error("failed to update agent version in ob meta")
		return
	}

	log.Info("try to check agent binary synced")
	if synced, err := agentService.IsBinarySynced(); err != nil {
		log.WithError(err).Error("failed to check agent version synced in ob meta")
		return err
	} else if synced {
		log.Info("agent binary is synced")
		return nil
	}

	return syncAgentBinary()
}

func syncAgentBinary() (err error) {
	log.Info("try to upgrade binary")
	for {
		err = agentService.UpgradeBinary()
		if errors.Is(err, agent.ErrOtherAgentUpgrading) {
			log.Info("other agent is upgrading, wait for a while")
			time.Sleep(constant.UPGRADE_BINARY_RETRY_INTERVAL * time.Second)
		} else {
			break
		}
	}

	if err == nil {
		log.Info("upgrade binary success")
		if err = agentService.SetBinarySynced(true); err != nil {
			log.WithError(err).Error("set binary synced failed")
		}
	} else {
		log.WithError(err).Error("upgrade binary failed")
	}
	return
}
