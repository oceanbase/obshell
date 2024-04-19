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

package secure

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
)

type GetPasswordResp struct {
	Password string `json:"password"`
}

func GetSecret(ctx context.Context) *meta.AgentSecret {
	return meta.NewAgentSecretByAgentInfo(meta.OCS_AGENT, Public())
}

// FetchPasswordFromOtherAgents fetch password from other agents.
// This action is required if failed to check password in sqlite after restart
func FetchPasswordFromOtherAgents() error {
	ocsAgents, err := getAllAgentsInfo()
	if err != nil {
		return err
	}
	param := meta.NewAgentSecretByAgentInfo(meta.OCS_AGENT, Public())
	defer dumpPassword()
	for _, ocsAgent := range ocsAgents {
		if ocsAgent.Ip == meta.OCS_AGENT.GetIp() && ocsAgent.Port == meta.OCS_AGENT.GetPort() {
			continue
		}
		respPwd := sendGetPasswordRPC(param, ocsAgent)
		if respPwd == "" {
			continue
		}
		password, err := Crypter.Decrypt(respPwd)
		if err != nil {
			log.WithError(err).Error("decrypt password failed")
			continue
		}
		if err = VerifyOceanbasePassword(string(password)); err != nil {
			log.WithError(err).Error("invalid password")
			continue
		}
		// Response password is correct.
		return nil
	}
	return fmt.Errorf("fetch password from %v agents failed", len(ocsAgents)-1)
}

func sendGetSecretApi(agentInfo meta.AgentInfoInterface) *meta.AgentSecret {
	log.Infof("Send get secret request from '%s:%d'", agentInfo.GetIp(), agentInfo.GetPort())
	ret := &meta.AgentSecret{}
	err := http.SendGetRequest(agentInfo, "/api/v1/secret", nil, ret)
	if err != nil {
		log.WithError(err).Error("Get secret failed")
	}
	return ret
}

func sendGetPasswordRPC(param *meta.AgentSecret, targetAgent meta.AgentInfo) string {
	log.Infof("send get password rpc request from '%s:%d' to '%v' ", param.Ip, param.Port, targetAgent)
	ret := GetPasswordResp{}
	err := http.SendGetRequest(&targetAgent, "/rpc/v1/password", param, &ret)
	if err != nil {
		log.WithError(err).Error("Get password by rpc failed")
		return ""
	}
	return ret.Password
}
