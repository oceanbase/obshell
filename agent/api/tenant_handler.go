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
	"github.com/oceanbase/obshell/agent/executor/tenant"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
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
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_NAME, tenantRenameHandler)
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_LOCK, tenantLockHandler)
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_LOCK, tenantUnlockHandler)
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_REPLICAS, tenantAddReplicasHandler)
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_REPLICAS, tenantRemoveReplicasHandler)
	tenant.PATCH(constant.URI_PATH_PARAM_NAME+constant.URI_REPLICAS, tenantModifyReplicasHandler)

	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_PRIMARYZONE, tenantModifyPrimaryZoneHandler)
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_ROOTPASSWORD, tenantModifyPasswordHandler)
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_WHITELIST, tenantModifyWhitelistHandler)

	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_PARAMETERS, tenantSetParametersHandler)
	tenant.PUT(constant.URI_PATH_PARAM_NAME+constant.URI_VARIABLES, tenantSetVariableHandler)
	tenant.GET(constant.URI_PATH_PARAM_NAME, getTenantInfo)
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_PARAMETER+constant.URI_PATH_PARAM_PARA, getTenantParameter)
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_VARIABLE+constant.URI_PATH_PARAM_VAR, getTenantVariable)
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_PARAMETERS, getTenantParameters)
	tenant.GET(constant.URI_PATH_PARAM_NAME+constant.URI_VARIABLES, getTenantVariables)
	tenant.POST(constant.URI_PATH_PARAM_NAME+constant.URI_USER, createUserHandler)
	tenant.DELETE(constant.URI_PATH_PARAM_NAME+constant.URI_USER+constant.URI_PATH_PARAM_USER, dropUserHandler)

	tenants.GET(constant.URI_OVERVIEW, getTenantOverView)
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
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "Tenant name is empty."))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	dag, err := tenant.CreateTenant(&param)
	common.SendResponse(c, dag, err)
}

func tenantCheckWithName(c *gin.Context) (string, error) {
	name := c.Param(constant.URI_PARAM_NAME)
	if name == "" {
		return "", errors.Occur(errors.ErrIllegalArgument, "Tenant name is empty.")
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		return "", errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String())
	}
	return name, nil
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	param.Name = name
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	var param param.ModifyTenantWhitelistParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
// @Router /api/v1/tenant/{name}/rootpassword [put]
func tenantModifyPasswordHandler(c *gin.Context) {
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
// @Router /api/v1/tenant/{name}/primaryzone [put]
func tenantModifyPrimaryZoneHandler(c *gin.Context) {
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	parameterName := c.Param(constant.URI_PARAM_PARA)
	if parameterName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "Parameter name is empty."))
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	variableName := c.Param(constant.URI_PARAM_VAR)
	if variableName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "Variable name is empty."))
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
// @Router /api/v1/tenant/{name}/variable/{} [get]
func getTenantVariables(c *gin.Context) {
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
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
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.DbaObTenant}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenants/overview [get]
func getTenantOverView(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
	}
	tenants, err := tenant.GetTenantsOverView()
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	var param param.CreateUserParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, tenant.CreateUser(name, param))
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
	name, err := tenantCheckWithName(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	user := c.Param(constant.URI_PARAM_USER)
	if user == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "User name is empty."))
		return
	}

	var param param.DropUserParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	common.SendResponse(c, nil, tenant.DropUser(name, user, param.RootPassword))
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
}

// @ID getDatabase
// @Summary get database
// @Description get database from a tenant
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param databaseName path string true "database name"
// @Success 200 object http.OcsAgentResponse{data=bo.Database}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases/{databaseName} [GET]
func getDatabase(c *gin.Context) {
	common.SendResponse(c, nil, nil)
}

// @ID deleteDatabase
// @Summary delete database
// @Description delete database
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param databaseName path string true "database name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases/{databaseName} [DELETE]
func deleteDatabase(c *gin.Context) {
	common.SendResponse(c, nil, nil)
}

// @ID updateDatabase
// @Summary update database
// @Description update database
// @Tags tenant
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param name path string true "tenant name"
// @Param databaseName path string true "database name"
// @Param body body param.ModifyDatabaseParam true "modify database param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/tenant/{name}/databases/{databaseName} [PUT]
func updateDatabase(c *gin.Context) {
	common.SendResponse(c, nil, nil)
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
	common.SendResponse(c, nil, nil)
}
