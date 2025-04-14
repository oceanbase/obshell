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
	"github.com/oceanbase/obshell/agent/executor/recyclebin"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
)

//@ID recyclebinTenantPurge
//@Summary purge recyclebin tenant
//@Description purge tenant in recyclebin
//@Tags recyclebin
//@Accept application/json
//@Produce application/json
//@Param X-OCS-Header header string true "Authorization"
//@Param name path string true "original tenant name or object name in recyclebin"
//@Success 200 object http.OcsAgentResponse
//@Failure 400 object http.OcsAgentResponse
//@Failure 401 object http.OcsAgentResponse
//@Failure 500 object http.OcsAgentResponse
//@Router /api/v1/recyclebin/bin/:name [delete]
func recyclebinPurgeTenantHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "Tenant name or object name is empty."))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	if dag, err := recyclebin.PurgeRecyclebinTenant(name); err == nil && dag == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err)
	}
}

//@ID recyclebinTenantList
//@Summary list all tenants in recyclebin
//@Description list all tenants in recyclebin
//@Tags recyclebin
//@Accept application/json
//@Produce application/json
//@Param X-OCS-Header header string true "Authorization"
//@Success 200 object http.OcsAgentResponse{data=[]oceanbase.DbaRecyclebin}
//@Failure 400 object http.OcsAgentResponse
//@Failure 401 object http.OcsAgentResponse
//@Failure 500 object http.OcsAgentResponse
//@Router /api/v1/recyclebin/tenants [get]
func recyclebinListTenantHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	tenants, err := recyclebin.ListRecyclebinTenant()
	common.SendResponse(c, tenants, err)
}

//@ID recyclebinFlashbackTenant
//@Summary flashback tenant from recyclebin
//@Description flashback tenant from recyclebin
//@Tags recyclebin
//@Accept application/json
//@Produce application/json
//@Param X-OCS-Header header string true "Authorization"
//@Param name path string true "original tenant name or object name in recyclebin"
//@Param body body param.FlashBackTenantParam true "Flashback tenant param"
//@Success 200 object http.OcsAgentResponse
//@Failure 400 object http.OcsAgentResponse
//@Failure 401 object http.OcsAgentResponse
//@Failure 500 object http.OcsAgentResponse
//@Router /api/v1/recyclebin/flashback/{name} [post]
func recyclebinFlashbackTenantHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "Tenant name or object name is empty."))
		return
	}
	var param param.FlashBackTenantParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "Flashback tenant param is invalid."))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	if param.NewName != nil && *param.NewName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "New name can not be empty."))
		return
	}
	err := recyclebin.FlashbackTenant(name, param.NewName)
	common.SendResponse(c, nil, err)
}
