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
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/unit"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
)

// @ID unitConfigCreate
// @Summary create resource unit config
// @Description create resource unit config
// @Tags unit
// @Accept application/json
// @Produce application/json
// @Param body body param.CreateResourceUnitConfigParams true "Resource unit config"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/unit/config [post]
func unitConfigCreateHandler(c *gin.Context) {
	var param param.CreateResourceUnitConfigParams
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	if *param.Name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "name", "Resource unit name is empty."))
		return
	}
	err := unit.CreateUnitConfig(param)
	common.SendResponse(c, nil, err)
}

// @ID unitConfigDrop
// @Summary drop resource unit config
// @Description drop resource unit config
// @Tags unit
// @Accept application/json
// @Produce application/json
// @Param name path string true "resource unit name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/unit/config/{name} [delete]
func unitConfigDropHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {

		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	err := unit.DropUnitConfig(name)
	common.SendResponse(c, nil, err)
}

// @ID unitConfigList
// @Summary get all resource unit configs
// @Description get all resource unit configs in the cluster
// @Tags unit
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.DbaObUnitConfig}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/units/config [get]
func unitConfigListHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	units, err := unit.GetAllUnitConfig()
	common.SendResponse(c, units, err)
}

// @ID unitConfigGet
// @Summary get resource unit config
// @Description get resource unit config
// @Tags unit
// @Accept application/json
// @Produce application/json
// @Param name path string true "resource unit name"
// @Success 200 object http.OcsAgentResponse{data=oceanbase.DbaObUnitConfig}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/unit/config/{name} [get]
func unitConfigGetHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObResourceUnitConfigNameEmpty))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	unit, err := unit.GetUnitConfig(name)
	common.SendResponse(c, unit, err)
}
