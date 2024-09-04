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

package api

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/binary"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/agent/service/task"
)

var localTaskService = task.NewLocalTaskService()

// TimeHandler returns the current time
//
//	@ID				getTime
//	@Summary		get current time
//	@Description	get current time
//	@Tags			v1
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	object	http.OcsAgentResponse{data=time.Time}
//	@Router			/api/v1/time [get]
func TimeHandler(c *gin.Context) {
	common.SendResponse(c, time.Now(), nil)
}

// InfoHandler returns the agent info
//
//	@ID				getAgentInfo
//	@Summary		get agent info
//	@Description	get agent info
//	@Tags			v1
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	object	http.OcsAgentResponse{data=meta.AgentStatus}
//	@Router			/api/v1/info [get]
func InfoHandler(s *http.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		obVersion, err := binary.GetMyOBVersion()
		agentStatus := meta.NewAgentStatus(meta.OCS_AGENT, global.Pid, s.GetState(), global.StartAt, global.HomePath, obVersion)
		common.SendResponse(c, agentStatus, err)
	}
}

// GitInfoHandler returns the agent git info
//
//	@ID				getGitInfo
//	@Summary		get git info
//	@Description	get git info
//	@Tags			v1
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	object	http.OcsAgentResponse{data=config.GitInfo}
//	@Router			/api/v1/git-info [get]
func GitInfoHandler(c *gin.Context) {
	common.SendResponse(c, config.GetGitInfo(), nil)
}

func GetAgentStatus(s *http.State) (http.AgentStatus, error) {
	isrunning, err := localTaskService.IsRunning()
	var status = http.AgentStatus{
		State:            s.GetState(),
		Pid:              global.Pid,
		StartAt:          global.StartAt,
		Version:          constant.VERSION_RELEASE,
		UnderMaintenance: !isrunning,
	}
	if meta.OCS_AGENT != nil {
		status.Agent.AgentInfo = meta.OCS_AGENT.GetAgentInfo()
		status.Agent.Identity = meta.OCS_AGENT.GetIdentity()
		status.Port = meta.OCS_AGENT.GetPort()
	}
	status.OBState = oceanbase.GetState()
	return status, err
}

// StatusHandler returns the agent status
//
//	@ID				getStatus
//	@Summary		get agent status
//	@Description	get agent status
//	@Tags			v1
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	object	http.OcsAgentResponse{data=http.AgentStatus}
//	@Router			/api/v1/status [get]
func StatusHandler(s *http.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := GetAgentStatus(s)
		common.SendResponse(c, status, err)
	}
}

// StatusHandler returns the agent status
//
//	@ID				getAllAgentsStatus
//	@Summary		get all agent status
//	@Description	get all agent status
//	@Tags			v1
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	object	http.OcsAgentResponse{dat=map[string]http.AgentStatus}
//	@Router			/api/v1/agents/status [get]
func GetAllAgentStatus(s *http.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentsStatus := make(map[string]http.AgentStatus)
		status, err := GetAgentStatus(s)
		if err == nil {
			agentsStatus[meta.OCS_AGENT.String()] = status
		}

		agents, err := agentService.GetAllAgents()
		if err == nil {
			uri := constant.URI_API_V1 + constant.URI_STATUS
			for _, agent := range agents {
				if agent.Equal(meta.OCS_AGENT) {
					continue
				}

				status := http.AgentStatus{}
				err := secure.SendGetRequest(&agent, uri, nil, &status)
				if err == nil {
					agentsStatus[agent.String()] = status
				} else {
					log.WithContext(c).Warnf("Failed to get status of agent %s: %s", agent.String(), err.Error())
				}
			}
		}
		common.SendResponse(c, agentsStatus, nil)
	}
}
