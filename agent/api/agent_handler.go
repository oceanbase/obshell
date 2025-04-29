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
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/agent"
	"github.com/oceanbase/obshell/agent/executor/host"
	"github.com/oceanbase/obshell/agent/lib/binary"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
	agentservice "github.com/oceanbase/obshell/agent/service/agent"
	"github.com/oceanbase/obshell/param"
)

var (
	agentService = agentservice.AgentService{}
)

// @Summary join the specified agent
// @Description join the specified agent
// @Tags agent
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.JoinApiParam true "agent info with zone name"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/agent [post]
// @Router /api/v1/agent/join [post]
func agentJoinHandler(c *gin.Context) {
	var param param.JoinApiParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if !meta.OCS_AGENT.IsSingleAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s is not single agent", meta.OCS_AGENT.String()))
		return
	}

	var dag *task.Dag
	var err error
	if meta.OCS_AGENT.Equal(&param.AgentInfo) {
		dag, err = agent.CreateJoinSelfDag(param.ZoneName)
	} else {
		var agentStatus meta.AgentStatus
		if err = http.SendGetRequest(&param.AgentInfo, constant.URI_API_V1+constant.URI_INFO, nil, &agentStatus); err != nil {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "get agent info failed: %s", err.Error()))
			return
		} else if !agentStatus.AgentInstance.IsMasterAgent() {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not master agent", param.AgentInfo.String()))
			return
		}

		// check version consistent
		if agentStatus.Version != constant.VERSION_RELEASE {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "obshell version is not consistent between %s(%s) and %s(%s)",
				param.AgentInfo.String(), agentStatus.Version, meta.OCS_AGENT.String(), constant.VERSION_RELEASE))
			return
		}
		if obVersion, _, err := binary.GetMyOBVersion(); err != nil {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrUnexpected, "get ob version failed: %s", err.Error()))
			return
		} else if obVersion != agentStatus.OBVersion {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "ob version is not consistent between %s(%s) and %s(%s)",
				param.AgentInfo.String(), agentStatus.OBVersion, meta.OCS_AGENT.String(), obVersion))
			return
		}
		// send token to master early.
		if err = agent.SendTokenToMaster(param.AgentInfo, param.MasterPassword); err != nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrTaskCreateFailed, err))
			return
		}

		dag, err = agent.CreateJoinMasterDag(param.AgentInfo, param.ZoneName, param.MasterPassword)
	}

	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskCreateFailed, err))
		return
	}
	common.SendResponse(c, task.NewDagDetailDTO(dag), nil)
}

// @Summary remove the specified agent
// @Description remove the specified agent
// @Tags agent
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body meta.AgentInfo true "agent info"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/agent [delete]
// @Router /api/v1/agent/remove [post]
func agentRemoveHandler(c *gin.Context) {
	var param meta.AgentInfo
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	var dag *task.Dag
	var err error
	switch meta.OCS_AGENT.GetIdentity() {
	case meta.FOLLOWER:
		master := agentService.GetMasterAgentInfo()
		if master == nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Master Agent is not found"))
			return
		}
		var agentStatus http.AgentStatus
		if err = http.SendGetRequest(master, constant.URI_API_V1+constant.URI_STATUS, nil, &agentStatus); err == nil {
			if agentStatus.UnderMaintenance {
				common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Master Agent is under maintenance"))
				return
			}
		}
		if !meta.OCS_AGENT.Equal(&param) {
			// If the current agent is not the target, forward the request to the master agent.
			common.ForwardRequest(c, master, param)
			return
		}
		dag, err = agent.CreaetFollowerRemoveSelfDag()
	case meta.MASTER:
		if isRunning, err := localTaskService.IsRunning(); err != nil {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "get local task status failed: %s", err.Error()))
			return
		} else if !isRunning {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Master Agent is under maintenance"))
			return
		}

		if meta.OCS_AGENT.Equal(&param) {
			dag, err = agent.CreateRemoveAllAgentsDag()
		} else {
			targetAgent, err := agentService.FindAgentInstance(&param)
			if err != nil {
				common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "get agent instance failed: %s", err.Error()))
				return
			}
			if targetAgent == nil {
				common.SendNoContentResponse(c, nil)
				return
			} else {
				dag, err = agent.CreateRemoveFollowerAgentDag(param, true)
			}
		}
	case meta.SINGLE:
		if meta.OCS_AGENT.Equal(&param) {
			common.SendNoContentResponse(c, nil)
			return
		}
		fallthrough
	default:
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s is not master or follower agent", meta.OCS_AGENT.String()))
		return
	}

	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskCreateFailed, err))
		return
	}
	common.SendResponse(c, task.NewDagDetailDTO(dag), nil)
}

func agentSetPasswordHandler(c *gin.Context) {
	var param param.SetAgentPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	common.SendResponse(c, nil, agentService.SetAgentPassword(param.Password))
}

// GitHostInfo returns the host info
//
// @ID GetHostInfo
// @Summary get host info
// @Description get host info
// @Tags agent
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=bo.HostInfo}
// @Router /api/v1/agent/host-info [GET]
func GetHostInfo(c *gin.Context) {
	common.SendResponse(c, host.GetInfo(), nil)
}
