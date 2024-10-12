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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/service/obcluster"
	"github.com/oceanbase/obshell/param"
)

var clusterService = obcluster.ObclusterService{}

func InitRestoreRoutes(r *gin.RouterGroup, isLocalRoute bool) {
	restoreGroup := r.Group(constant.URI_RESTORE)
	tenantGroup := r.Group(constant.URI_TENANT_GROUP)
	if !isLocalRoute {
		tenantGroup.Use(common.Verify())
		restoreGroup.Use(common.Verify())
	}

	restoreGroup.GET(constant.URI_WINDOWS, getRestoreWindowsHandler)

	tenantGroup.POST(constant.URI_RESTORE, tenantRestoreHandler)
	tenantGroup.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_RESTORE, cancelRestoreTaskHandler)
	tenantGroup.GET(constant.URI_PATH_PARAM_NAME+constant.URI_RESTORE+constant.URI_OVERVIEW, getRestoreOverviewHandler)

}

// @ID			tenantRestore
// @Summary	Restore tenant
// @Tags		Restore
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Header	header	string				true	"Authorization"
// @Param		body			body	param.RestoreParam	true	"Restore tenant"
// @Success	200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure	400				object	http.OcsAgentResponse
// @Failure	401				object	http.OcsAgentResponse
// @Failure	500				object	http.OcsAgentResponse
// @Router		/api/v1/tenant/restore [post]
func tenantRestoreHandler(c *gin.Context) {
	var p param.RestoreParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if err := checkRestoreParam(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dag, err := ob.TenantRestore(&p)
	common.SendResponse(c, dag, err)
}

func checkRestoreParam(p *param.RestoreParam) *errors.OcsAgentError {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return errors.Occurf(errors.ErrKnown, "agent identity is '%v', cannot restore", meta.OCS_AGENT.GetIdentity())
	}

	if p.TenantName == constant.TENANT_SYS {
		return errors.Occurf(errors.ErrIllegalArgument, "'%s' is system tenant, cannot restore", p.TenantName)
	}

	log.Infof("check tenant %s", p.TenantName)
	tenant, err := tenantService.GetTenantByName(p.TenantName)
	if err != nil {
		return errors.Occur(errors.ErrUnexpected, err)
	}
	if tenant != nil {
		return errors.Occurf(errors.ErrIllegalArgument, "tenant '%s' already exists", p.TenantName)
	}

	if len(p.ZoneList) == 0 {
		return errors.Occurf(errors.ErrIllegalArgument, "zone_list is empty")
	}

	return nil
}

// @ID			cancelRestoreTask
// @Summary	Get restore task id
// @Tags		Restore
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Header	header	string	true	"Authorization"
// @Param		tenantName		path	string	true	"Tenant name"
// @Success	200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure	400				object	http.OcsAgentResponse
// @Failure	401				object	http.OcsAgentResponse
// @Failure	500				object	http.OcsAgentResponse
// @Router		/api/v1/tenant/:tenantName/restore [delete]
func cancelRestoreTaskHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "agent identity is %s.", meta.OCS_AGENT.GetIdentity()))
		return
	}

	tenantName := c.Param(constant.URI_PARAM_NAME)
	if tenantName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "tenant name is empty"))
		return
	}
	if tenantName == constant.TENANT_SYS {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrIllegalArgument, "'%s' is system tenant, cannot restore", tenantName))
		return
	}

	dag, err := ob.CancelRestoreTaskForTenant(tenantName)
	if err == nil && dag == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err)
	}
}

// @ID			getRestoreOverview
// @Summary	Get restore overview
// @Tags		Restore
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Header	header	string	true	"Authorization"
// @Param		tenantName		path	string	true	"Tenant name"
// @Success	200				object	http.OcsAgentResponse{data=param.RestoreOverview}
// @Failure	400				object	http.OcsAgentResponse
// @Failure	401				object	http.OcsAgentResponse
// @Failure	500				object	http.OcsAgentResponse
// @Router		/api/v1/tenant/:tenantName/restore/overview [get]
func getRestoreOverviewHandler(c *gin.Context) {
	tenant, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	overview, err := ob.GetRestoreOverview(tenant.TenantName)
	common.SendResponse(c, overview, err)
}

// @ID			getRestoreWindows
// @Summary	Get restore windows
// @Tags		Restore
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Header	header	string	true	"Authorization"
// @Success	200				object	http.OcsAgentResponse{data=param.RestoreWindowsParam}
// @Failure	400				object	http.OcsAgentResponse
// @Failure	401				object	http.OcsAgentResponse
// @Failure	500				object	http.OcsAgentResponse
// @Router		/api/v1/restore/windows [get]
func getRestoreWindowsHandler(c *gin.Context) {
	var p param.RestoreWindowsParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	windows, err := ob.GetRestoreWindows(&p)
	common.SendResponse(c, windows, err)
}
