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
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

// return the agent with available OB connection
// if no agent with available OB connection, return the OCS_AGENT
func getAgentWithAvailableOB() (agent meta.AgentInfoInterface) {
	err := oceanbase.QuickHealthCheck()
	if err != nil {
		agents, err := agentService.GetAllAgents()
		if err == nil {
			uri := constant.URI_API_V1 + constant.URI_STATUS + "?ob_query_timeout=" + oceanbase.DEFAULT_OCEANBASE_QUERY_TIMEOUT
			for _, agent := range agents {
				if agent.Equal(meta.OCS_AGENT) {
					continue
				}

				status := http.AgentStatus{}
				err := secure.SendGetRequest(&agent, uri, nil, &status)
				if err == nil {
					if status.OBState == oceanbase.STATE_CONNECTION_AVAILABLE {
						return &agent
					}
				}
			}
		}
	}
	return meta.OCS_AGENT
}
