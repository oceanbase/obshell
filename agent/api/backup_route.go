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
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/service/tenant"
	"github.com/oceanbase/obshell/param"
)

var (
	tenantService = tenant.TenantService{}
)

func InitBackupRoutes(v1 *gin.RouterGroup, isLocalRoute bool) {
	tenantGroup := v1.Group(constant.URI_TENANT_GROUP)
	obclusterGroup := v1.Group(constant.URI_OBCLUSTER_GROUP)

	groups := make([]*gin.RouterGroup, 0)
	groups = append(groups, tenantGroup, obclusterGroup)
	if !isLocalRoute {
		for _, group := range groups {
			group.Use(common.Verify())
		}
	}

	tenantGroup.POST(constant.URI_PATH_PARAM_NAME+constant.URI_BACKUP+constant.URI_CONFIG, tenantBackupConfigHandler)
	tenantGroup.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_BACKUP+constant.URI_CONFIG, patchTenantBackupConfigHandler)
	tenantGroup.POST(constant.URI_PATH_PARAM_NAME+constant.URI_BACKUP, tenantStartBackupHandler)
	tenantGroup.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_BACKUP, patchTenantBackupHandler)
	tenantGroup.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_BACKUP+constant.URI_ARCHIVE, patchTenantArchiveLogHandler)
	tenantGroup.GET(constant.URI_PATH_PARAM_NAME+constant.URI_BACKUP+constant.URI_OVERVIEW, tenantBackupOverviewHandler)

	obclusterGroup.POST(constant.URI_BACKUP+constant.URI_CONFIG, obclusterBackupConfigHandler)
	obclusterGroup.PATCH(constant.URI_BACKUP+constant.URI_CONFIG, patchObclusterBackupConfigHandler)
	obclusterGroup.POST(constant.URI_BACKUP, obclusterStartBackupHandler)
	obclusterGroup.PATCH(constant.URI_BACKUP, patchObclusterBackupHandler)
	obclusterGroup.PATCH(constant.URI_BACKUP+constant.URI_ARCHIVE, patchObclusterArchiveLogHandler)
	obclusterGroup.GET(constant.URI_BACKUP+constant.URI_OVERVIEW, obclusterBackupOverviewHandler)
}

// @ID				obclusterBackupConfig
// @Summary		Set backup config for all tenants
// @Description	Set backup config for all tenants
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string							true	"Authorization"
// @Param			body			body	param.ClusterBackupConfigParam	true	"Backup config"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/backup/config [post]
func obclusterBackupConfigHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	var param param.ClusterBackupConfigParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.PostObclusterBackupConfig(&param)
	common.SendResponse(c, dag, err)
}

// @ID				patchObclusterBackupConfig
// @Summary		Patch backup config for all tenants
// @Description	Patch backup config for all tenants
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string							true	"Authorization"
// @Param			body			body	param.ClusterBackupConfigParam	true	"Backup config"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/backup/config [patch]
func patchObclusterBackupConfigHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	var param param.ClusterBackupConfigParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.PatchObclusterBackupConfig(&param)
	common.SendResponse(c, dag, err)
}

// @ID				tenantBackupConfig
// @Summary		Set backup config for tenant
// @Description	Set backup config for tenant
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string							true	"Authorization"
// @Param			name			path	string							true	"Tenant name"
// @Param			body			body	param.TenantBackupConfigParam	true	"Backup config"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/backup/config [post]
func tenantBackupConfigHandler(c *gin.Context) {
	tenant, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	var param param.TenantBackupConfigParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.PostTenantBackupConfig(tenant.TenantName, &param)
	common.SendResponse(c, dag, err)
}

// @ID				patchTenantBackupConfig
// @Summary		Patch backup config for tenant
// @Description	Patch backup config for tenant
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string							true	"Authorization"
// @Param			name			path	string							true	"Tenant name"
// @Param			body			body	param.TenantBackupConfigParam	true	"Backup config"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
func patchTenantBackupConfigHandler(c *gin.Context) {
	tenantName, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	var param param.TenantBackupConfigParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.PatchTenantBackupConfig(tenantName, &param)
	common.SendResponse(c, dag, err)
}

func checkTenantAndGetName(c *gin.Context) (*oceanbase.DbaObTenant, error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}

	tenantName := c.Param(constant.URI_PARAM_NAME)
	if tenantName == "" {
		return nil, errors.Occur(errors.ErrObTenantNameEmpty)
	}

	if tenantName == constant.TENANT_SYS {
		return nil, errors.Occur(errors.ErrObTenantSysOperationNotAllowed)
	}

	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.Occur(errors.ErrObTenantNotExist, tenantName)
	}
	return tenant, nil
}

// @ID				obclusterStartBackup
// @Summary		Start backup for all tenants
// @Description	Start backup for all tenants
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string				true	"Authorization"
// @Param			body			body	param.BackupParam	true	"Backup param"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/backup [post]
func obclusterStartBackupHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	var p param.BackupParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	log.Infof("%#+v", p)
	dag, err := ob.ObclusterStartBackup(&p)
	common.SendResponse(c, dag, err)
}

// @ID				tenantStartBackup
// @Summary		Start backup for tenant
// @Description	Start backup for tenant
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string				true	"Authorization"
// @Param			name			path	string				true	"Tenant name"
// @Param			body			body	param.BackupParam	true	"Backup param"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/backup [post]
func tenantStartBackupHandler(c *gin.Context) {
	tenant, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	var p param.BackupParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dag, err := ob.TenantStartBackup(tenant, &p)
	common.SendResponse(c, dag, err)
}

// @ID				patchObclusterBackup
// @Summary		Patch backup status for all tenants
// @Description	Patch backup status for all tenants
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string					true	"Authorization"
// @Param			body			body	param.BackupStatusParam	true	"Backup status"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/backup [patch]
func patchObclusterBackupHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	var p param.BackupStatusParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	err := ob.PatchObclusterBackup(&p)
	common.SendResponse(c, nil, err)
}

// @ID				patchTenantBackup
// @Summary		Patch backup status for tenant
// @Description	Patch backup status for tenant
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string					true	"Authorization"
// @Param			name			path	string					true	"Tenant name"
// @Param			body			body	param.BackupStatusParam	true	"Backup status"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/backup [patch]
func patchTenantBackupHandler(c *gin.Context) {
	tenant, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	var p param.BackupStatusParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	err = ob.PatchTenantBackup(tenant.TenantName, &p)
	common.SendResponse(c, nil, err)
}

// @ID				patchObclusterArchiveLog
// @Summary		Patch archive log status for all tenants
// @Description	Patch archive log status for all tenants
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string						true	"Authorization"
// @Param			body			body	param.ArchiveLogStatusParam	true	"Archive log status"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/backup/log [patch]
func patchObclusterArchiveLogHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	var p param.ArchiveLogStatusParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	err := ob.PatchObclusterArchiveLog(&p)
	common.SendResponse(c, nil, err)
}

// @ID				patchTenantArchiveLog
// @Summary		Patch archive log status for tenant
// @Description	Patch archive log status for tenant
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string						true	"Authorization"
// @Param			name			path	string						true	"Tenant name"
// @Param			body			body	param.ArchiveLogStatusParam	true	"Archive log status"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/backup/log [patch]
func patchTenantArchiveLogHandler(c *gin.Context) {
	tenant, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	var p param.ArchiveLogStatusParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	err = ob.PatchTenantArchiveLog(tenant.TenantName, &p)
	common.SendResponse(c, nil, err)
}

// @ID				obclusterBackupOverview
// @Summary		Get backup overview for all tenants
// @Description	Get backup overview for all tenants
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=param.BackupOverview}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/backup/overview [get]
func obclusterBackupOverviewHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	overview, err := ob.GetObclusterBackupOverview()
	common.SendResponse(c, overview, err)
}

// @ID				tenantBackupOverview
// @Summary		Get backup overview for tenant
// @Description	Get backup overview for tenant
// @Tags			Backup
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Param			name			path	string	true	"Tenant name"
// @Success		200				object	http.OcsAgentResponse{data=param.BackupOverview}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/backup/overview [get]
func tenantBackupOverviewHandler(c *gin.Context) {
	tenant, err := checkTenantAndGetName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	overview, err := ob.GetTenantBackupOverview(tenant.TenantName)
	common.SendResponse(c, overview, err)
}
