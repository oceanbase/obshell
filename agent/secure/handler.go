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

func sendGetSecretApi(agentInfo meta.AgentInfoInterface) *meta.AgentSecret {
	log.Infof("Send get secret request from '%s'", agentInfo.String())
	ret := &meta.AgentSecret{}
	err := http.SendGetRequest(agentInfo, "/api/v1/secret", nil, ret)
	if err != nil {
		log.WithError(err).Error("Get secret failed")
	}
	return ret
}
