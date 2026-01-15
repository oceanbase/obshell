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
	"github.com/oceanbase/obshell/seekdb/agent/executor/observer"

	"github.com/oceanbase/obshell/seekdb/param"
)

func InitDatabaseRoutes(v1 *gin.RouterGroup, isLocalRoute bool) {
	database := v1.Group(constant.URI_DATABASE_GROUP)
	databases := v1.Group(constant.URI_DATABASES_GROUP)

	if !isLocalRoute {
		database.Use(common.Verify())
		databases.Use(common.Verify())
	}

	// for database
	database.POST("", createDatabase)
	databases.GET("", listDatabases)
	database.PUT(constant.URI_PATH_PARAM_DATABASE, updateDatabase)
	database.GET(constant.URI_PATH_PARAM_DATABASE, getDatabase)
	database.DELETE(constant.URI_PATH_PARAM_DATABASE, deleteDatabase)
}

// @ID listDatabases
// @Summary list databases
// @Description list databases
// @Tags database
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]bo.Database}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/databases [GET]
func listDatabases(c *gin.Context) {
	databases, err := observer.ListDatabases()
	common.SendResponse(c, databases, err)
}

// @ID getDatabase
// @Summary get database
// @Description get database
// @Tags database
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param database path string true "database name"
// @Success 200 object http.OcsAgentResponse{data=bo.Database}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/database/{database} [GET]
func getDatabase(c *gin.Context) {
	databaseName := c.Param(constant.URI_PARAM_DATABASE)
	database, err := observer.GetDatabase(databaseName)
	common.SendResponse(c, database, err)
}

// @ID deleteDatabase
// @Summary delete database
// @Description delete database
// @Tags database
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param database path string true "database name"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/database/{database} [DELETE]
func deleteDatabase(c *gin.Context) {
	databaseName := c.Param(constant.URI_PARAM_DATABASE)
	err := observer.DeleteDatabase(databaseName)
	common.SendResponse(c, nil, err)
}

// @ID updateDatabase
// @Summary update database
// @Description update database
// @Tags database
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param database path string true "database name"
// @Param body body param.ModifyDatabaseParam true "modify database param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/database/{database} [PUT]
func updateDatabase(c *gin.Context) {
	databaseName := c.Param(constant.URI_PARAM_DATABASE)
	modifyDatabaseParam := param.ModifyDatabaseParam{}
	err := c.BindJSON(&modifyDatabaseParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = observer.AlterDatabase(databaseName, &modifyDatabaseParam)
	common.SendResponse(c, nil, err)
}

// @ID createDatabase
// @Summary create database
// @Description create database
// @Tags database
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.CreateDatabaseParam true "create database param"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/database [POST]
func createDatabase(c *gin.Context) {
	createDatabaseParam := param.CreateDatabaseParam{}
	err := c.BindJSON(&createDatabaseParam)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err = observer.CreateDatabase(&createDatabaseParam)
	common.SendResponse(c, nil, err)
}
