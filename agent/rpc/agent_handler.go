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
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/engine/coordinator"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/agent"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/meta"
	agentservice "github.com/oceanbase/obshell/agent/service/agent"
	"github.com/oceanbase/obshell/param"
)

func agentJoinHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsMasterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s:%d is not master", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()))
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
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s:%d already exists", agentInstance.Ip, agentInstance.Port))
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
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s:%d is %s", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort(), meta.OCS_AGENT.GetIdentity()))
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
			common.SendResponse(c, nil, errors.Occur(errors.ErrTaskCreateFailed, err))
		} else {
			common.SendResponse(c, task.NewDagDetailDTO(dag), nil)
		}
	}
}

func followerRemoveSelf(c *gin.Context) {
	dag, err := agent.CreateToSingleDag()
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskCreateFailed, err))
	} else {
		common.SendResponse(c, task.NewDagDetailDTO(dag), nil)
	}
}

func getMaintainerHandler(c *gin.Context) {
	maintainer, err := coordinator.GetMaintainer()
	common.SendResponse(c, maintainer, err)
}

func obServerDeployHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsFollowerAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s:%d is not follower agent", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()))
		return
	}
	var dirs param.DeployTaskParams
	if err := c.BindJSON(&dirs); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
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
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s:%d is not follower agent", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()))
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
			common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
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

func createStopDagHandler(c *gin.Context) {

}

func agentUpdateHandler(c *gin.Context) {
	var params param.SyncAgentParams
	if err := c.BindJSON(&params); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}
	dag, err := ob.CreateAgentSyncDag(params.Password)
	common.SendResponse(c, dag, err)
}
