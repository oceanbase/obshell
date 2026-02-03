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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/tenant"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/param"
)

func InitTenantRoutes(v1 *gin.RouterGroup, isLocalRoute bool) {
	tenant := v1.Group(constant.URI_TENANT_GROUP)
	tenants := v1.Group(constant.URI_TENANTS_GROUP)
	if !isLocalRoute {
		tenant.Use(common.Verify())
		tenants.Use(common.Verify())
	}
	tenant.POST("", tenantCreateHandler)
	tenant.DELETE(constant.URI_PATH_PARAM_NAME, tenantDropHandler)
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_NAME, tenantExistHandlerWrapper(tenantRenameHandler))
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_LOCK, tenantExistHandlerWrapper(tenantLockHandler))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_LOCK, tenantExistHandlerWrapper(tenantUnlockHandler))
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_REPLICAS, tenantStatusHandlerWrapper(tenantAddReplicasHandler))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_REPLICAS, tenantStatusHandlerWrapper(tenantRemoveReplicasHandler))
	tenant.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_REPLICAS, tenantStatusHandlerWrapper(tenantModifyReplicasHandler))

	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_PRIMARYZONE, tenantStatusHandlerWrapper(tenantModifyPrimaryZoneHandler))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_ROOTPASSWORD, tenantStatusHandlerWrapper(tenantModifyPasswordHandler))
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_ROOTPASSWORD+constant.URI_PERSIST, tenantExistHandlerWrapper(common.AutoForwardToMaintainerWrapper(persistTenantRootPassword)))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_WHITELIST, tenantStatusHandlerWrapper(tenantModifyWhitelistHandler))

	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_PARAMETERS, tenantStatusHandlerWrapper(tenantSetParametersHandler))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_VARIABLES, tenantStatusHandlerWrapper(tenantSetVariableHandler))
	tenant.GET(constant.URI_PATH_PARAM_NAME, getTenantInfo)
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_PRECHECK, tenantHandlerWrapper(tenantPrecheck))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_PARAMETER+constant.URI_PATH_PARAM_PARA, tenantStatusHandlerWrapper(getTenantParameter))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_VARIABLE+constant.URI_PATH_PARAM_VAR, tenantStatusHandlerWrapper(getTenantVariable))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_PARAMETERS, tenantStatusHandlerWrapper(getTenantParameters))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_VARIABLES, tenantStatusHandlerWrapper(getTenantVariables))

	tenant.GET(constant.URI_SUPPORT_TEMPLATES, listParameterTemplatesHandler)

	// for user
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_USER, tenantHandlerWrapper(createUserHandler))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER, tenantHandlerWrapper(dropUserHandler))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_USER, tenantHandlerWrapper(listUsers))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_USERS, tenantHandlerWrapper(listUsers))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_OBJECTS, tenantHandlerWrapper(listObjects))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER, tenantHandlerWrapper(getUser))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_DB_PRIVILEGE, tenantHandlerWrapper(modifyDbPrivilege))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_DB_PRIVILEGES, tenantHandlerWrapper(modifyDbPrivilege))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_GLOBAL_PRIVILEGE, tenantHandlerWrapper(modifyGlobalPrivilege))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_GLOBAL_PRIVILEGES, tenantHandlerWrapper(modifyGlobalPrivilege))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(modifyUserObjectPrivilege, constant.ORACLE_MODE))
	tenant.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(patchUserObjectPrivilege, constant.ORACLE_MODE))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(revokeUserObjectPrivilege, constant.ORACLE_MODE))
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(grantUserObjectPrivilege, constant.ORACLE_MODE))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_ROLES, tenantHandlerWrapper(modifyUserRoles, constant.ORACLE_MODE))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_PASSWORD, tenantHandlerWrapper(changePassword))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_LOCK, tenantHandlerWrapper(lockUser))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_STATS, tenantExistHandlerWrapper(getUserStats))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER+constant.URI_LOCK, tenantHandlerWrapper(unlockUser))

	// for database
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_DATABASES, tenantHandlerWrapper(createDatabase, constant.MYSQL_MODE))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_DATABASES, tenantHandlerWrapper(listDatabases, constant.MYSQL_MODE))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_DATABASES+constant.URI_PATH_PARAM_DATABASE, tenantHandlerWrapper(updateDatabase, constant.MYSQL_MODE))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_DATABASES+constant.URI_PATH_PARAM_DATABASE, tenantHandlerWrapper(getDatabase, constant.MYSQL_MODE))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_DATABASES+constant.URI_PATH_PARAM_DATABASE, tenantHandlerWrapper(deleteDatabase, constant.MYSQL_MODE))

	// for role
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE, tenantHandlerWrapper(createRole, constant.ORACLE_MODE))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE, tenantHandlerWrapper(dropRole, constant.ORACLE_MODE))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE, tenantHandlerWrapper(getRole, constant.ORACLE_MODE))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_ROLES, tenantHandlerWrapper(listRoles, constant.ORACLE_MODE))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE+constant.URI_ROLES, tenantHandlerWrapper(modifyRole, constant.ORACLE_MODE))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE+constant.URI_GLOBAL_PRIVILEGES, tenantHandlerWrapper(modifyRoleGlobalPrivilege, constant.ORACLE_MODE))
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(modifyRoleObjectPrivilege, constant.ORACLE_MODE))
	tenant.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(patchRoleObjectPrivilege, constant.ORACLE_MODE))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(revokeRoleObjectPrivilege, constant.ORACLE_MODE))
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_ROLE+constant.URI_PATH_PARAM_ROLE+constant.URI_OBJECT_PRIVILEGES, tenantHandlerWrapper(grantRoleObjectPrivilege, constant.ORACLE_MODE))

	// for compaction
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_COMPACTION, tenantExistHandlerWrapper(getTenantCompactionHandler))
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_COMPACT, tenantExistHandlerWrapper(tenantMajorCompactionHandler))
	tenant.GET(constant.URI_TOP_COMPACTIONS, getTenantTopCompactionsHandler)
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_COMPACTION_ERROR, tenantExistHandlerWrapper(clearTenantCompactionErrorHandler))

	// for slow sql
	tenant.GET(constant.URI_TOP_SLOW_SQLS, getTenantTopSlowSqlRankHandler)

	// for session management
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_SESSIONS, tenantHandlerWrapper(getTenantSessions))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_SESSIONS+constant.URI_PATH_PARAM_SESSION_ID, tenantHandlerWrapper(getTenantSession))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_SESSIONS+constant.URI_STATS, tenantHandlerWrapper(getTenantSessionsStats))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_SESSIONS, tenantHandlerWrapper(killTenantSessionsHandler))
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_SESSIONS+constant.URI_QUERIES, tenantHandlerWrapper(killTenantSessionQueryHandler))
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_DEADLOCKS, tenantHandlerWrapper(ListTenantDeadlocksHandler))

	tenants.GET(constant.URI_OVERVIEW, findAvailableClusterAgentIfNeedWrapper(getTenantOverView))
}

// @ID tenantCreate
// @Summary create tenant
// @Description create tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.CreateTenantParam true "create tenant params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant [post]
func tenantCreateHandler(c *gin.Context) {
	var param param.CreateTenantParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if *param.Name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantNameEmpty))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	dag, err := tenant.CreateTenant(&param)
	common.SendResponse(c, dag, err)
}

// @ID tenantDrop
// @Summary drop tenant
// @Description drop tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.DropTenantParam true "drop tenant params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name} [delete]
func tenantDropHandler(c *gin.Context) {
	var param param.DropTenantParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantNameEmpty))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	param.Name = name
	if dag, err := tenant.DropTenant(&param); err == nil && dag == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err)
	}
}

// @ID tenantRename
// @Summary rename tenant
// @Description rename tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.RenameTenantParam true "rename tenant params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name} [put]
func tenantRenameHandler(c *gin.Context) {
	var param param.RenameTenantParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	param.Name = c.Param(constant.URI_PARAM_NAME)
	common.SendResponse(c, nil, tenant.RenameTenant(param))
}

// @ID tenantLock
// @Summary lock tenant
// @Description lock tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/lock [post]
func tenantLockHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	common.SendResponse(c, nil, tenant.LockTenant(name))
}

// @ID tenantUnlock
// @Summary unlock tenant
// @Description unlock tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/lock [delete]
func tenantUnlockHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	common.SendResponse(c, nil, tenant.UnlockTenant(name))
}

// @ID tenantAddReplicas
// @Summary add replicas to tenant
// @Description add replicas to tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.ScaleOutTenantReplicasParam true "add tenant replicas params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/replicas [post]
func tenantAddReplicasHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.ScaleOutTenantReplicasParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := tenant.ScaleOutTenantReplicas(name, &param)
	common.SendResponse(c, dag, err)
}

// @ID tenantRemoveReplicas
// @Summary remove replicas from tenant
// @Description remove replicas from tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.ScaleInTenantReplicasParam true "remove tenant replicas params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/replicas [delete]
func tenantRemoveReplicasHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.ScaleInTenantReplicasParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if dag, err := tenant.ScaleInTenantReplicas(name, &param); err == nil && dag == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err)
	}
}

// @ID tenantModifyReplicas
// @Summary modify tenant replicas
// @Description modify tenant replicas
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.ModifyReplicasParam true "modify tenant replicas params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/replicas [patch]
func tenantModifyReplicasHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.ModifyReplicasParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err1 := tenant.ModifyTenantReplica(name, &param)
	if err1 == nil && dag == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err1)
	}
}

// @ID tenantModifyWhitelist
// @Summary modify tenant whitelist
// @Description modify tenant whitelist
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.ModifyTenantWhitelistParam true "modify whitelist params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/whitelist [put]
func tenantModifyWhitelistHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.ModifyTenantWhitelistParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	var err error
	if param.Whitelist == nil {
		err = tenant.ModifyTenantWhitelist(name, "")
	} else {
		err = tenant.ModifyTenantWhitelist(name, *param.Whitelist)
	}
	common.SendResponse(c, nil, err)
}

// @ID tenantModifyPassword
// @Summary modify tenant root password
// @Description modify tenant root password
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.ModifyTenantRootPasswordParam true "modify tenant root password params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/password [put]
func tenantModifyPasswordHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.ModifyTenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err, isForwarded := tenant.ModifyTenantRootPassword(c, name, param)
	if isForwarded {
		return
	}
	common.SendResponse(c, nil, err)
}

// @ID persistTenantRootPassword
// @Summary persist tenant root password
// @Description persist tenant root password
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.PersistTenantRootPasswordParam true "persist tenant root password param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/password/persist [POST]
func persistTenantRootPassword(c *gin.Context) {
	//all checks are done in the wrapper, just save the password
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.PersistTenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := tenant.PersistTenantRootPassword(c, name, param.Password)
	common.SendResponse(c, nil, err)
}

// @ID tenantModifyPrimaryZone
// @Summary modify tenant primary zone
// @Description modify tenant primary zone
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.ModifyTenantPrimaryZoneParam true "modify tenant primary zone params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/primary-zone [put]
func tenantModifyPrimaryZoneHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.ModifyTenantPrimaryZoneParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := tenant.ModifyTenantPrimaryZone(name, &param)
	common.SendResponse(c, dag, err)
}

// @ID tenantSetParameters
// @Summary set tenant parameters
// @Description set tenant parameters
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.SetTenantParametersParam true "set tenant parameters params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/parameters [put]
func tenantSetParametersHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.SetTenantParametersParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, tenant.SetTenantParameters(name, param.Parameters))
}

// @ID tenantSetVariable
// @Summary set tenant variables
// @Description set tenant variables
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.SetTenantVariablesParam true "set tenant global variables params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/variables [put]
func tenantSetVariableHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.SetTenantVariablesParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, tenant.SetTenantVariables(c, name, param))
}

// @ID getTenantInfo
// @Summary get tenant info
// @Description get tenant info
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse{data=bo.TenantInfo}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name} [get]
func getTenantInfo(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	tenantInfo, err := tenant.GetTenantInfo(name)
	common.SendResponse(c, tenantInfo, err)
}

// @ID getTenantParameter
// @Summary get tenant parameter
// @Description get tenant parameter
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param para path string true "parameter name"
// @Success 200 object http.OcsAgentResponse{data=oceanbase.GvObParameter}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/parameter/{para} [get]
func getTenantParameter(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	parameterName := c.Param(constant.URI_PARAM_PARA)
	if parameterName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantParameterNameEmpty))
		return
	}

	parameter, err := tenant.GetTenantParameter(name, parameterName)
	common.SendResponse(c, parameter, err)
}

// @ID getTenantParameters
// @Summary get tenant parameters
// @Description get tenant parameters
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param filter query string false "filter format"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.GvObParameter}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/parameters [get]
func getTenantParameters(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	format := c.Query("filter")
	parameters, err := tenant.GetTenantParameters(name, format)
	common.SendResponse(c, parameters, err)
}

// @ID getTenantVariable
// @Summary get tenant variable
// @Description get tenant variable
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param var path string true "variable name"
// @Success 200 object http.OcsAgentResponse{data=oceanbase.CdbObSysVariable}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/variable/{var} [get]
func getTenantVariable(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	variableName := c.Param(constant.URI_PARAM_VAR)
	if variableName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantVariableNameEmpty))
		return
	}

	variable, err := tenant.GetTenantVariable(name, variableName)
	common.SendResponse(c, variable, err)
}

// @ID getTenantVariables
// @Summary get tenant variables
// @Description get tenant variables
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param filter query string false "filter format"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.CdbObSysVariable}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/variables [get]
func getTenantVariables(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	format := c.Query("filter")
	variables, err := tenant.GetTenantVariables(name, format)
	common.SendResponse(c, variables, err)
}

// @ID getTenantOverView
// @Summary get tenant overview
// @Description get tenant overview
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param mode query string false "tenant compitable mode: MYSQL or ORACLE"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.DbaObTenant}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenants/overview [get]
func getTenantOverView(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	mode := c.Query("mode")
	tenants, err := tenant.GetTenantsOverView(mode)
	common.SendResponse(c, tenants, err)
}

// @ID createUser
// @Summary create user
// @Description create user
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.CreateUserParam true "create user params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user [post]
func createUserHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.CreateUserParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, tenant.CreateUser(name, &param))
}

// @ID dropUser
// @Summary drop user
// @Description drop user
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.DropUserParam true "drop user params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user} [delete]
func dropUserHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	if user == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObUserNameEmpty))
		return
	}

	var param param.DropUserParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	common.SendResponse(c, nil, tenant.DropUser(name, user, &param))
}

// @ID listUsers
// @Summary list users
// @Description list users from a tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse{data=[]bo.ObUser}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user [GET]
func listUsers(c *gin.Context) {
	var p param.TenantRootPasswordParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	queryParam := &param.ListUsersQueryParam{}
	if err := c.BindQuery(queryParam); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	queryParam.Format()

	name := c.Param(constant.URI_PARAM_NAME)
	obusers, err := tenant.ListUsers(name, p.RootPassword, queryParam)
	common.SendResponse(c, obusers, err)
}

// @ID getUser
// @Summary get user
// @Description get user from a tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse{data=bo.ObUser}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user} [GET]
func getUser(c *gin.Context) {
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	obuser, err := tenant.GetUser(name, user, param.RootPassword)
	common.SendResponse(c, obuser, err)
}

// @ID modifyDbPrivilege
// @Summary modify db privilege of a user
// @Description modify db privilege of a user
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.ModifyUserDbPrivilegeParam true "modify db privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/db-privilege [PUT]
func modifyDbPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	modifyUserDbPrivilegeParam := param.ModifyUserDbPrivilegeParam{}
	err := c.BindJSON(&modifyUserDbPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyUserDbPrivilege(name, user, &modifyUserDbPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID modifyUserObjectPrivilege
// @Summary modify user object privilege
// @Description modify user object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.ModifyObjectPrivilegeParam true "modify user object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/object-privileges [PUT]
func modifyUserObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	modifyUserObjectPrivilegeParam := param.ModifyObjectPrivilegeParam{}
	err := c.BindJSON(&modifyUserObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyUserObjectPrivilege(name, user, &modifyUserObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID patchUserObjectPrivilege
// @Summary patch user object privilege
// @Description patch user object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.ModifyObjectPrivilegeParam true "patch user object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/object-privileges [PATCH]
func patchUserObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	patchUserObjectPrivilegeParam := param.ModifyObjectPrivilegeParam{}
	err := c.BindJSON(&patchUserObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.PatchUserObjectPrivilege(name, user, &patchUserObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID listObjects
// @Summary list objects
// @Description list objects from a tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.TenantRootPasswordParam true "tenant root password"
// @Success 200 object http.OcsAgentResponse{data=[]bo.DbaObjectBo}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/objects [GET]
func listObjects(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	objects, err := tenant.ListObjects(name, param.RootPassword)
	common.SendResponse(c, objects, err)
}

// @ID revokeUserObjectPrivilege
// @Summary revoke user object privilege
// @Description revoke user object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.RevokeObjectPrivilegeParam true "revoke user object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/object-privileges [DELETE]
func revokeUserObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	revokeUserObjectPrivilegeParam := param.RevokeObjectPrivilegeParam{}
	err := c.BindJSON(&revokeUserObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.RevokeUserObjectPrivilege(name, user, &revokeUserObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID grantUserObjectPrivilege
// @Summary grant user object privilege
// @Description grant user object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.GrantObjectPrivilegeParam true "grant user object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/object-privileges [POST]
func grantUserObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	grantUserObjectPrivilegeParam := param.GrantObjectPrivilegeParam{}
	err := c.BindJSON(&grantUserObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.GrantUserObjectPrivilege(name, user, &grantUserObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID modifyUserRoles
// @Summary modify user role
// @Description modify user role
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.ModifyRoleParam true "modify user role param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/roles [PUT]
func modifyUserRoles(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	modifyUserRoleParam := param.ModifyRoleParam{}
	err := c.BindJSON(&modifyUserRoleParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyUserRole(name, user, &modifyUserRoleParam)
	common.SendResponse(c, nil, err)
}

// @ID getStats
// @Summary get user stats
// @Description get user stats
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse{data=bo.ObUserStats}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/stats [GET]
func getUserStats(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	user := c.Param(constant.URI_PARAM_USER)
	userStats, err := tenant.GetUserStats(name, user)
	common.SendResponse(c, userStats, err)
}

// @ID tenantPreCheck
// @Summary check tenant accessibility
// @Description check tenant accessibility
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse{data=bo.ObTenantPreCheckResult}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/precheck [GET]
func tenantPrecheck(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObTenantNameEmpty))
		return
	}
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	preCheckResult, err := tenant.TenantPreCheck(name, param.RootPassword)
	common.SendResponse(c, preCheckResult, err)
}

// @ID modifyGlobalPrivilege
// @Summary modify global privilege of a user
// @Description modify global privilege of a user
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.ModifyUserGlobalPrivilegeParam true "modify global privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/global-privilege [PUT]
func modifyGlobalPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	modifyUserGlobalPrivilegeParam := param.ModifyUserGlobalPrivilegeParam{}
	err := c.BindJSON(&modifyUserGlobalPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyUserGlobalPrivilege(name, user, &modifyUserGlobalPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID changePassword
// @Summary change user password
// @Description change user password
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Param body body param.ChangeUserPasswordParam true "change password param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/password [PUT]
func changePassword(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	changeUserPasswordParam := param.ChangeUserPasswordParam{}
	err := c.BindJSON(&changeUserPasswordParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ChangeUserPassword(name, user, &changeUserPasswordParam)
	common.SendResponse(c, nil, err)
}

// @ID lockUser
// @Summary lock user
// @Description lock user
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/lock [PUT]
func lockUser(c *gin.Context) {
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	err := tenant.LockUser(name, user, param.RootPassword)
	common.SendResponse(c, nil, err)
}

// @ID unlockUser
// @Summary unlock user
// @Description unlock user
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/user/{user}/lock [DELETE]
func unlockUser(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	user := c.Param(constant.URI_PARAM_USER)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := tenant.UnlockUser(name, user, param.RootPassword)
	common.SendResponse(c, nil, err)
}

// @ID listDatabases
// @Summary list databases
// @Description list databases from a tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse{data=[]bo.Database}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases [GET]
func listDatabases(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	databases, err := tenant.ListDatabases(name, param.RootPassword)
	common.SendResponse(c, databases, err)
}

// @ID getDatabase
// @Summary get database
// @Description get database from a tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param database path string true "database name"
// @Success 200 object http.OcsAgentResponse{data=bo.Database}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases/{database} [GET]
func getDatabase(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	databaseName := c.Param(constant.URI_PARAM_DATABASE)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	database, err := tenant.GetDatabase(name, databaseName, param.RootPassword)
	common.SendResponse(c, database, err)
}

// @ID deleteDatabase
// @Summary delete database
// @Description delete database
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param database path string true "database name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases/{database} [DELETE]
func deleteDatabase(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	databaseName := c.Param(constant.URI_PARAM_DATABASE)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := tenant.DeleteDatabase(name, databaseName, param.RootPassword)
	common.SendResponse(c, nil, err)
}

// @ID updateDatabase
// @Summary update database
// @Description update database
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param database path string true "database name"
// @Param body body param.ModifyDatabaseParam true "modify database param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases/{database} [PUT]
func updateDatabase(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	databaseName := c.Param(constant.URI_PARAM_DATABASE)
	modifyDatabaseParam := param.ModifyDatabaseParam{}
	err := c.BindJSON(&modifyDatabaseParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.AlterDatabase(name, databaseName, &modifyDatabaseParam)
	common.SendResponse(c, nil, err)
}

// @ID createDatabase
// @Summary create database
// @Description create database
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.CreateDatabaseParam true "create database param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases [POST]
func createDatabase(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	createDatabaseParam := param.CreateDatabaseParam{}
	err := c.BindJSON(&createDatabaseParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.CreateDatabase(name, &createDatabaseParam)
	common.SendResponse(c, nil, err)
}

// @ID listRoles
// @Summary list roles
// @Description list roles only for oracle tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse{data=[]bo.ObRole}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/roles [GET]
func listRoles(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	roles, err := tenant.ListRoles(name, param.RootPassword)
	common.SendResponse(c, roles, err)
}

// @ID getRole
// @Summary get role
// @Description get role
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Success 200 object http.OcsAgentResponse{data=bo.ObRole}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role} [GET]
func getRole(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	var param param.TenantRootPasswordParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	roleInfo, err := tenant.GetRole(name, role, param.RootPassword)
	common.SendResponse(c, roleInfo, err)
}

// @ID dropRole
// @Summary drop role
// @Description drop role
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.DropRoleParam true "drop role param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role} [DELETE]
func dropRole(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	var param param.DropRoleParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := tenant.DropRole(name, role, &param)
	common.SendResponse(c, nil, err)
}

// @ID createRole
// @Summary create role
// @Description create role
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.CreateRoleParam true "create role param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role [POST]
func createRole(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	createRoleParam := param.CreateRoleParam{}
	err := c.BindJSON(&createRoleParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.CreateRole(name, &createRoleParam)
	common.SendResponse(c, nil, err)
}

// @ID modifyRole
// @Summary modify role
// @Description modify role
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.ModifyRoleParam true "modify role param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role}/roles [PUT]
func modifyRole(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	modifyRoleParam := param.ModifyRoleParam{}
	err := c.BindJSON(&modifyRoleParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyRole(name, role, &modifyRoleParam)
	common.SendResponse(c, nil, err)
}

// @ID modifyRoleGlobalPrivilege
// @Summary modify role global privilege
// @Description modify role global privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.ModifyRoleGlobalPrivilegeParam true "modify role global privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role}/global-privileges [PUT]
func modifyRoleGlobalPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	modifyRoleGlobalPrivilegeParam := param.ModifyRoleGlobalPrivilegeParam{}
	err := c.BindJSON(&modifyRoleGlobalPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyRoleGlobalPrivilege(name, role, &modifyRoleGlobalPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID modifyRoleObjectPrivilege
// @Summary modify role object privilege
// @Description modify role object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.ModifyObjectPrivilegeParam true "modify role object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role}/object-privileges [PUT]
func modifyRoleObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	modifyRoleObjectPrivilegeParam := param.ModifyObjectPrivilegeParam{}
	err := c.BindJSON(&modifyRoleObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.ModifyRoleObjectPrivilege(name, role, &modifyRoleObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID patchRoleObjectPrivilege
// @Summary patch role object privilege
// @Description patch role object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.ModifyObjectPrivilegeParam true "patch role object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role}/object-privileges [PATCH]
func patchRoleObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	modifyRoleObjectPrivilegeParam := param.ModifyObjectPrivilegeParam{}
	err := c.BindJSON(&modifyRoleObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.PatchRoleObjectPrivilege(name, role, &modifyRoleObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID revokeRoleObjectPrivilege
// @Summary revoke role object privilege
// @Description revoke role object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.RevokeObjectPrivilegeParam true "revoke role object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role}/object-privileges [DELETE]
func revokeRoleObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	revokeRoleObjectPrivilegeParam := param.RevokeObjectPrivilegeParam{}
	err := c.BindJSON(&revokeRoleObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.RevokeRoleObjectPrivilege(name, role, &revokeRoleObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID grantRoleObjectPrivilege
// @Summary grant role object privilege
// @Description grant role object privilege
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param role path string true "role name"
// @Param body body param.GrantObjectPrivilegeParam true "grant role object privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/role/{role}/object-privileges [POST]
func grantRoleObjectPrivilege(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	role := c.Param(constant.URI_PARAM_ROLE)
	grantRoleObjectPrivilegeParam := param.GrantObjectPrivilegeParam{}
	err := c.BindJSON(&grantRoleObjectPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.GrantRoleObjectPrivilege(name, role, &grantRoleObjectPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID				getTenantCompaction
// @Summary		get tenant major compaction info
// @Description	get tenant major compaction info
// @Tags			tenant
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=bo.TenantCompaction}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/compaction [get]
func getTenantCompactionHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	compaction, err := tenant.GetTenantCompaction(name)
	common.SendResponse(c, compaction, err)
}

// @ID				tenantMajorCompaction
// @Summary		trigger tenant major compaction
// @Description	trigger tenant major compaction
// @Tags			tenant
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/compact [post]
func tenantMajorCompactionHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	common.SendResponse(c, nil, tenant.TenantMajorCompaction(name))
}

// @ID				tenantClearCompactionError
// @Summary		clear tenant major compaction error
// @Description	clear tenant major compaction error
// @Tags			tenant
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/{name}/compaction-error [delete]
func clearTenantCompactionErrorHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	common.SendResponse(c, nil, tenant.ClearTenantCompactionError(name))
}

// @ID				getTenantTopCompaction
// @Summary		query tenant information ranked by the cost of major compaction.
// @Description	query tenant information ranked by the cost of major compaction, limited to the top n.
// @Tags			tenant
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Param			limit				query	string	false	"top n"
// @Success		200				object	http.OcsAgentResponse{data=[]bo.TenantCompactionHistory}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/top-compactions [get]
func getTenantTopCompactionsHandler(c *gin.Context) {
	topStr := c.Query("limit")
	top := 3
	if topStr != "" && topStr != "0" {
		parsedTop, err := strconv.Atoi(topStr)
		if err != nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrRequestQueryParamIllegal, "limit"))
			return
		}
		top = parsedTop
	}

	compaction, err := tenant.GetTopCompactions(top)
	common.SendResponse(c, compaction, err)
}

// @ID				getTenantTopSlowSqlRank
// @Summary		query tenant information ranked by the number of slow SQL statements.
// @Description	query tenant information ranked by the number of slow SQL statements, limited to the top n.
// @Tags			tenant
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Param			start_time		query	string	true	"start time"
// @Param			end_time		query	string	true	"end time"
// @Param			limit				query	string	false	"top n"
// @Success		200				object	http.OcsAgentResponse{data=[]bo.TenantSlowSqlCount}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/tenant/top-slow-sqls [get]
func getTenantTopSlowSqlRankHandler(c *gin.Context) {
	// Require the SQL processing end time to be between start_time and end_time.
	start_time := c.Query("start_time")
	end_time := c.Query("end_time")
	top := c.Query("limit")
	var param param.QuerySlowSqlRankParam
	if top == "" {
		param.Top = 3
	} else {
		parsedTop, err := strconv.Atoi(top)
		if err != nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrRequestQueryParamIllegal, "limit"))
			return
		}
		param.Top = parsedTop
	}
	if start_time != "" {
		parsedTime, err := time.Parse(time.RFC3339, start_time)
		if err != nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrRequestQueryParamIllegal, "start_time"))
			return
		}
		param.StartTime = parsedTime
	} else {
		common.SendResponse(c, nil, errors.Occur(errors.ErrRequestQueryParamEmpty, "start_time"))
		return
	}
	if end_time != "" {
		parsedTime, err := time.Parse(time.RFC3339, end_time)
		if err != nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrRequestQueryParamIllegal, "end_time"))
			return
		}
		param.EndTime = parsedTime
	} else {
		common.SendResponse(c, nil, errors.Occur(errors.ErrRequestQueryParamEmpty, "end_time"))
		return
	}

	res, err := tenantService.GetSlowSqlRank(param.Top, param.StartTime.UnixMicro(), param.EndTime.UnixMicro())
	common.SendResponse(c, res, err)
}

// @ID listSupportParameterTemplates
// @Summary list support parameter templates
// @Description list support parameter templates
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]tenant.ParameterTemplate}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/support-templates [get]
func listParameterTemplatesHandler(c *gin.Context) {
	language := c.GetHeader(constant.ACCEPT_LANGUAGE)
	templates := tenant.GetAllSupportedScenarios(language)
	common.SendResponse(c, templates, nil)
}

// @ID getTenantSessions
// @Summary get tenant sessions
// @Description get tenant sessions
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param page query uint64 false "page"
// @Param size query uint64 false "size"
// @Param user query string false "db user"
// @Param db query string false "db name"
// @Param client_ip query string false "client ip"
// @Param id query string false "session id"
// @Param active_only query boolean false "active only"
// @Param svr_ip query string false "server ip"
// @Param sort query string false "sort"
// @Success 200 object http.OcsAgentResponse{data=bo.PaginatedTenantSessions}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/sessions [get]
func getTenantSessions(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	p := &param.QueryTenantSessionParam{}
	if err := c.BindQuery(p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	p.Format()
	sessions, err := tenant.GetTenantSessions(name, p)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	} else {
		common.SendResponse(c, sessions, nil)
	}
}

// @ID getTenantSession
// @Summary get tenant session
// @Description get tenant session
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param sessionId path string true "session id"
// @Success 200 object http.OcsAgentResponse{data=bo.TenantSession}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/sessions/{sessionId} [get]
func getTenantSession(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	sessionId := c.Param(constant.URI_PARAM_SESSION_ID)
	session, err := tenant.GetTenantSession(name, sessionId)
	common.SendResponse(c, session, err)
}

// @ID getTenantSessionsStats
// @Summary get tenant sessions stats
// @Description get tenant sessions stats
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Success 200 object http.OcsAgentResponse{data=bo.TenantSessionStats}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/sessions/stats [get]
func getTenantSessionsStats(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	stats, err := tenant.GetTenantSessionStats(name)
	common.SendResponse(c, stats, err)
}

// @ID killTenantSessions
// @Summary kill tenant sessions
// @Description kill tenant sessions
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.KillTenantSessionsParam true "kill tenant sessions param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/sessions [delete]
func killTenantSessionsHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	p := param.KillTenantSessionsParam{}
	err := c.BindJSON(&p)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.KillTenantSessions(name, p.SessionIds)
	common.SendResponse(c, nil, err)
}

// @ID killTenantSessionQuery
// @Summary kill tenant session query
// @Description kill tenant session query
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param body body param.KillTenantSessionQueryParam true "kill tenant session query param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/sessions/queries [delete]
func killTenantSessionQueryHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	p := param.KillTenantSessionsParam{}
	err := c.BindJSON(&p)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = tenant.KillTenantSessionQueries(name, p.SessionIds)
	common.SendResponse(c, nil, err)
}

// @ID listTenantDeadlocks
// @Summary list tenant deadlocks
// @Description list tenant deadlocks
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param page query uint64 false "page"
// @Param size query uint64 false "size"
// @Success 200 object http.OcsAgentResponse{data=bo.PaginatedDeadLocks}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/deadlocks [get]
func ListTenantDeadlocksHandler(c *gin.Context) {
	name := c.Param(constant.URI_PARAM_NAME)
	p := &param.QueryTenantDeadLocksParam{}
	err := c.BindQuery(p)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	p.Format()
	deadlocks, err := tenant.ListTenantDeadLocks(name, p)
	common.SendResponse(c, deadlocks, err)
}
