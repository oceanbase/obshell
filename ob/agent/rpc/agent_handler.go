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

package rpc

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/engine/coordinator"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/agent"
	"github.com/oceanbase/obshell/ob/agent/executor/ob"
	"github.com/oceanbase/obshell/ob/agent/meta"
	agentservice "github.com/oceanbase/obshell/ob/agent/service/agent"
	"github.com/oceanbase/obshell/ob/param"
)

func agentAddTokenHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsMasterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.MASTER))
		return
	}
	var param param.AddTokenParam
	ip := c.RemoteIP()
	if err := c.Bind(&param); err != nil {
		return
	}
	if param.AgentInfo.Ip == "" {
		param.AgentInfo.Ip = ip
	}

	agentService := agentservice.AgentService{}
	agentInstance, err := agentService.FindAgentInstance(&param.AgentInfo)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if agentInstance != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentAlreadyExists, agentInstance.String()))
		return
	}

	common.SendResponse(c, nil, agent.AddSingleToken(param))
}

func agentJoinHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsMasterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.MASTER))
		return
	}

	ip := c.RemoteIP()
	var param param.JoinMasterParam
	if err := c.Bind(&param); err != nil {
		return
	}
	if param.JoinApiParam.AgentInfo.Ip == "" {
		param.JoinApiParam.AgentInfo.Ip = ip
	}

	agentService := agentservice.AgentService{}
	agentInstance, err := agentService.FindAgentInstance(&param.JoinApiParam.AgentInfo)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if agentInstance != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentAlreadyExists, agentInstance.String()))
		return
	} else {
		if err := agent.AddFollowerAgent(param); err != nil {
			common.SendResponse(c, nil, err)
			return
		}
	}
	selfAgent := meta.NewAgentInstanceByAgent(meta.OCS_AGENT)
	common.SendResponse(c, selfAgent, nil)
}

func agentRemoveHandler(c *gin.Context) {
	if meta.OCS_AGENT.IsMasterAgent() {
		masterRemoveFollower(c)
	} else if meta.OCS_AGENT.IsFollowerAgent() {
		followerRemoveSelf(c)
	} else if meta.OCS_AGENT.IsSingleAgent() {
		common.SendResponse(c, task.DagDetailDTO{}, nil)
	} else {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), strings.Join([]string{(string)(meta.MASTER), (string)(meta.FOLLOWER), (string)(meta.SINGLE)}, " or ")))
	}
}

func masterRemoveFollower(c *gin.Context) {
	var param meta.AgentInfo
	if err := c.Bind(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	agentInstance, err := agent.GetFollowerAgent(&param)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if agentInstance == nil {
		// Agent not exists, return success.
		common.SendResponse(c, task.DagDetailDTO{}, nil)
	} else {
		dag, err := agent.CreateRemoveFollowerAgentDag(param, false)
		if err != nil {
			common.SendResponse(c, nil, err)
		} else {
			common.SendResponse(c, task.NewDagDetailDTO(dag), nil)
		}
	}
}

func followerRemoveSelf(c *gin.Context) {
	dag, err := agent.CreateToSingleDag()
	if err != nil {
		common.SendResponse(c, nil, err)
	} else {
		common.SendResponse(c, task.NewDagDetailDTO(dag), nil)
	}
}

func getMaintainerHandler(c *gin.Context) {
	maintainer, err := coordinator.GetMaintainer()
	common.SendResponse(c, maintainer, err)
}

func updateAllAgentsHandler(c *gin.Context) {
	var allAgentsSyncData param.AllAgentsSyncData
	if err := c.BindJSON(&allAgentsSyncData); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if coordinator.OCS_AGENT_SYNCHRONIZER == nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentSynchronizerNotInitialized))
		return
	}
	coordinator.OCS_AGENT_SYNCHRONIZER.Update(coordinator.ConvertToAllAgentsSyncData(allAgentsSyncData))
	common.SendResponse(c, nil, nil)
}

func obServerDeployHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsFollowerAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.FOLLOWER))
		return
	}
	var dirs param.DeployTaskParams
	if err := c.BindJSON(&dirs); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dag, err := ob.CreateDeploySelfDag(dirs.Dirs)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, dag, nil)
}

func obServerDestroyHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsFollowerAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.FOLLOWER))
		return
	}
	dag, err := ob.CreateDestroyDag()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, dag, nil)
}

func obStartHandler(c *gin.Context) {
	if meta.OCS_AGENT.IsClusterAgent() {
		var param ob.CreateSubDagParam
		if err := c.Bind(&param); err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		data, err := ob.CreateStartDag(param)
		common.SendResponse(c, data, err)
	} else {
		var config param.StartTaskParams
		if err := c.BindJSON(&config); err != nil {
			common.SendResponse(c, nil, err)
			return
		}

		dag, err := ob.CreateStartSelfDag(config.Config, config.HealthCheck)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		common.SendResponse(c, dag, nil)
	}
}

func obStopHandler(c *gin.Context) {
	if meta.OCS_AGENT.IsClusterAgent() {
		var param ob.CreateSubDagParam
		if err := c.Bind(&param); err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		data, err := ob.CreateStopDag(param)
		common.SendResponse(c, data, err)
	} else {
		dag, err := ob.CreateStopSelfDag()
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		common.SendResponse(c, dag, nil)
	}
}

func obLocalScaleOutHandler(c *gin.Context) {
	var param param.LocalScaleOutParam
	if err := c.Bind(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	data, err := ob.HandleLocalScaleOut(param)
	common.SendResponse(c, data, err)
}

func agentUpdateHandler(c *gin.Context) {
	var params param.SyncAgentParams
	if err := c.BindJSON(&params); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.CreateAgentSyncDag(params.Password)
	common.SendResponse(c, dag, err)
}

func takeOverAgentUpdateBinaryHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsTakeover() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), strings.Join([]string{(string)(meta.TAKE_OVER_MASTER), (string)(meta.TAKE_OVER_FOLLOWER)}, " or ")))
		return
	}

	if dag, err := ob.TakeOverUpdateAgentVersion(); err != nil {
		common.SendResponse(c, nil, err)
		return
	} else if dag == nil {
		common.SendNoContentResponse(c, err)
		return
	} else {
		common.SendResponse(c, dag, nil)
	}
}

// killObserverHandler only used for delete server
func killObserverHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsScalingInAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.SCALING_IN))
		return
	}
	dag, err := ob.CreateKillObserverDag()
	common.SendResponse(c, dag, err)
}

func startObserverHandler(c *gin.Context) {
	dag, err := ob.CreateStartObserverDag()
	if dag == nil && err == nil {
		common.SendNoContentResponse(c, err)
		return
	}
	common.SendResponse(c, dag, err)
}
