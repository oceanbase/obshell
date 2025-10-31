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
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/executor/observer"

	"github.com/oceanbase/obshell/seekdb/param"
)

func InitUserRoutes(v1 *gin.RouterGroup, isLocalRoute bool) {
	user := v1.Group(constant.URI_USER_GROUP)
	users := v1.Group(constant.URI_USERS_GROUP)
	if !isLocalRoute {
		user.Use(common.Verify())
	}

	user.POST("", createUserHandler)
	user.DELETE(constant.URI_PATH_PARAM_USER, dropUserHandler)
	users.GET("", listUsers)
	user.GET(constant.URI_PATH_PARAM_USER, getUser)
	user.PUT(constant.URI_PATH_PARAM_USER+constant.URI_DB_PRIVILEGES, modifyDbPrivilege)
	user.PUT(constant.URI_PATH_PARAM_USER+constant.URI_GLOBAL_PRIVILEGES, modifyGlobalPrivilege)
	user.PUT(constant.URI_PATH_PARAM_USER+constant.URI_PASSWORD, changePassword)
	user.GET(constant.URI_PATH_PARAM_USER+constant.URI_STATS, getUserStats)
	user.POST(constant.URI_PATH_PARAM_USER+constant.URI_LOCK, lockUser)
	user.DELETE(constant.URI_PATH_PARAM_USER+constant.URI_LOCK, unlockUser)
}

// @ID createUser
// @Summary create user
// @Description create user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.CreateUserParam true "create user params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user [post]
func createUserHandler(c *gin.Context) {
	var param param.CreateUserParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, observer.CreateUser(&param))
}

// @ID dropUser
// @Summary drop user
// @Description drop user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user} [delete]
func dropUserHandler(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	if user == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObUserNameEmpty))
		return
	}

	common.SendResponse(c, nil, observer.DropUser(user))
}

// @ID listUsers
// @Summary list users
// @Description list users
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]bo.ObUser}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/users [GET]
func listUsers(c *gin.Context) {
	queryParam := &param.ListUsersQueryParam{}
	if err := c.BindQuery(queryParam); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	queryParam.Format()

	obusers, err := observer.ListUsers(queryParam)
	common.SendResponse(c, obusers, err)
}

// @ID getUser
// @Summary get user
// @Description get user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse{data=bo.ObUser}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user} [GET]
func getUser(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	obuser, err := observer.GetUser(user)
	common.SendResponse(c, obuser, err)
}

// @ID modifyDbPrivilege
// @Summary modify db privilege of a user
// @Description modify db privilege of a user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Param body body param.ModifyUserDbPrivilegeParam true "modify db privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user}/db-privileges [PUT]
func modifyDbPrivilege(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	modifyUserDbPrivilegeParam := param.ModifyUserDbPrivilegeParam{}
	err := c.BindJSON(&modifyUserDbPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = observer.ModifyUserDbPrivilege(user, &modifyUserDbPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID getStats
// @Summary get user stats
// @Description get user stats
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse{data=bo.ObUserStats}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user}/stats [GET]
func getUserStats(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	userStats, err := observer.GetUserStats(user)
	common.SendResponse(c, userStats, err)
}

// @ID modifyGlobalPrivilege
// @Summary modify global privilege of a user
// @Description modify global privilege of a user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Param body body param.ModifyUserGlobalPrivilegeParam true "modify global privilege param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user}/global-privileges [PUT]
func modifyGlobalPrivilege(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	modifyUserGlobalPrivilegeParam := param.ModifyUserGlobalPrivilegeParam{}
	err := c.BindJSON(&modifyUserGlobalPrivilegeParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = observer.ModifyUserGlobalPrivilege(user, &modifyUserGlobalPrivilegeParam)
	common.SendResponse(c, nil, err)
}

// @ID changePassword
// @Summary change user password
// @Description change user password
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Param body body param.ChangeUserPasswordParam true "change password param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user}/password [PUT]
func changePassword(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	changeUserPasswordParam := param.ChangeUserPasswordParam{}
	err := c.BindJSON(&changeUserPasswordParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = observer.ChangeUserPassword(user, &changeUserPasswordParam)
	common.SendResponse(c, nil, err)
}

// @ID lockUser
// @Summary lock user
// @Description lock user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user}/lock [POST]
func lockUser(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	err := observer.LockUser(user)
	common.SendResponse(c, nil, err)
}

// @ID unlockUser
// @Summary unlock user
// @Description unlock user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param user path string true "user name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/user/{user}/lock [DELETE]
func unlockUser(c *gin.Context) {
	user := c.Param(constant.URI_PARAM_USER)
	err := observer.UnlockUser(user)
	common.SendResponse(c, nil, err)
}
