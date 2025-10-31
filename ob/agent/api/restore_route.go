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

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/ob"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/service/obcluster"
	"github.com/oceanbase/obshell/ob/param"
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
	restoreGroup.POST(constant.URI_SOURCE_INFO, getRestoreTenantInfoHandler)
	restoreGroup.GET(constant.URI_TASKS, listRestoreTasksHandler)

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

func checkRestoreParam(p *param.RestoreParam) error {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}

	if p.TenantName == constant.TENANT_SYS {
		return errors.Occur(errors.ErrObTenantSysOperationNotAllowed)
	}

	log.Infof("check tenant %s", p.TenantName)
	tenant, err := tenantService.GetTenantByName(p.TenantName)
	if err != nil {
		return err
	}
	if tenant != nil {
		return errors.Occur(errors.ErrObTenantExisted, p.TenantName)
	}

	if len(p.ZoneList) == 0 {
		return errors.Occur(errors.ErrObTenantZoneListEmpty)
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
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	tenantName := c.Param(constant.URI_PARAM_NAME)
	if tenantName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantNameEmpty))
		return
	}
	if tenantName == constant.TENANT_SYS {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantSysOperationNotAllowed))
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

// @ID			listRestoreTasks
// @Summary	List restore tasks
// @Tags		Restore
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Header	header	string	true	"Authorization"
// @Param		tenantName		path	string	true	"Tenant name"
// @Success	200				object	http.OcsAgentResponse{data=bo.PaginatedRestoreTaskResponse}
// @Failure	400				object	http.OcsAgentResponse
// @Failure	401				object	http.OcsAgentResponse
// @Failure	500				object	http.OcsAgentResponse
// @Router		/api/v1/restore/tasks [get]
func listRestoreTasksHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	p := &param.QueryRestoreTasksParam{}
	if err := c.BindQuery(p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	p.Format()
	tasks, err := ob.GetAllRestoreTasks(p)
	common.SendResponse(c, tasks, err)
}

// @ID			getRestoreSourceTenantInfo
// @Summary	Get original restore tenant info by ob_admin
// @Tags		Restore
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Header	header	string	true	"Authorization"
// @Param		body			body	param.RestoreStorageParam	true	"the storage uri of data backup and archive log"
// @Success	200				object	http.OcsAgentResponse{data=system.RestoreTenantInfo}
// @Failure	400				object	http.OcsAgentResponse
// @Failure	401				object	http.OcsAgentResponse
// @Failure	500				object	http.OcsAgentResponse
// @Router		/api/v1/restore/source-tenant-info [post]
func getRestoreTenantInfoHandler(c *gin.Context) {
	var p param.RestoreStorageParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	info, err := ob.GetRestoreSourceTenantInfo(&p)
	common.SendResponse(c, info, err)
}
