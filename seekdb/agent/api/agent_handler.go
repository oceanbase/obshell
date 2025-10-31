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
	"github.com/oceanbase/obshell/seekdb/agent/api/common"
	"github.com/oceanbase/obshell/seekdb/agent/executor/upgrade"
	"github.com/oceanbase/obshell/seekdb/param"
)

// @ID agentUpgradeCheck
// @Summary check agent upgrade
// @Description check agent upgrade
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.UpgradeCheckParam true "agent upgrade check params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/agent/upgrade/check [post]
func agentUpgradeCheckHandler(c *gin.Context) {
	var param param.UpgradeCheckParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	task, err := upgrade.AgentUpgradeCheck(param)
	common.SendResponse(c, task, err)
}

// @ID agentUpgrade
// @Summary upgrade agent
// @Description upgrade agent
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.UpgradeCheckParam true "agent upgrade check params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/agent/upgrade [post]
func agentUpgradeHandler(c *gin.Context) {
	var param param.UpgradeCheckParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := upgrade.AgentUpgrade(param)
	common.SendResponse(c, dag, err)
}
