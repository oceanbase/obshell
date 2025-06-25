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
	"github.com/oceanbase/obshell/agent/executor/obproxy"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

func InitObproxyRoutes(r *gin.RouterGroup, isLocalRoute bool) {
	obproxy := r.Group(constant.URI_OBPROXY_GROUP)
	if !isLocalRoute {
		obproxy.Use(common.Verify(secure.ROUTE_OBPROXY))
	}

	// obproxy routes
	obproxy.POST("", obproxyAddHandler)
	obproxy.DELETE("", obproxyDeleteHandler)
	obproxy.POST(constant.URI_START, obproxyStartHandler)
	obproxy.POST(constant.URI_STOP, obproxyStopHandler)
	obproxy.POST(constant.URI_PACKAGE, obproxyPkgUploadHandler)
	obproxy.POST(constant.URI_UPGRADE, obproxyUpgradeHandler)
}

// @ID			obproxyAdd
// @Summary	Add obproxy
// @Tags		Obproxy
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Agent-Header	header	string					true	"Authorization"
// @Param		body				body	param.AddObproxyParam	true	"Add obproxy"
// @Success	200					object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure	400					object	http.OcsAgentResponse
// @Failure	401					object	http.OcsAgentResponse
// @Failure	500					object	http.OcsAgentResponse
// @Router		/api/v1/obproxy [post]
func obproxyAddHandler(c *gin.Context) {
	var param param.AddObproxyParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := obproxy.AddObproxy(param)
	common.SendResponse(c, dag, err)
}

// @ID			obproxyStop
// @Summary	Stop obproxy
// @Tags		Obproxy
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Agent-Header	header	string	true	"Authorization"
// @Success	200					object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure	400					object	http.OcsAgentResponse
// @Failure	401					object	http.OcsAgentResponse
// @Failure	500					object	http.OcsAgentResponse
// @Router		/api/v1/obproxy/stop [post]
func obproxyStopHandler(c *gin.Context) {
	dag, err := obproxy.StopObproxy()
	common.SendResponse(c, dag, err)
}

// @ID			obproxyStart
// @Summary	Start obproxy
// @Tags		Obproxy
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Agent-Header	header	string	true	"Authorization"
// @Success	200					object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure	400					object	http.OcsAgentResponse
// @Failure	401					object	http.OcsAgentResponse
// @Failure	500					object	http.OcsAgentResponse
// @Router		/api/v1/obproxy/start [post]
func obproxyStartHandler(c *gin.Context) {
	dag, err := obproxy.StartObproxy()
	common.SendResponse(c, dag, err)
}

// @ID			obproxyDelete
// @Summary	Delete obproxy
// @Tags		Obproxy
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Agent-Header	header	string	true	"Authorization"
// @Success	204					object	http.OcsAgentResponse
// @Failure	400					object	http.OcsAgentResponse
// @Failure	401					object	http.OcsAgentResponse
// @Failure	500					object	http.OcsAgentResponse
// @Router		/api/v1/obproxy [delete]
func obproxyDeleteHandler(c *gin.Context) {
	dag, err := obproxy.DeleteObproxy()
	if dag == nil && err == nil {
		common.SendNoContentResponse(c, nil)
	}
	common.SendResponse(c, dag, err)
}

// @ID			obproxyUpgrade
// @Summary	Upgrade obproxy
// @Tags		Obproxy
// @Accept		application/json
// @Produce	application/json
// @Param		X-OCS-Agent-Header	header	string						true	"Authorization"
// @Param		body				body	param.UpgradeObproxyParam	true	"Upgrade obproxy"
// @Success	200					object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure	400					object	http.OcsAgentResponse
// @Failure	401					object	http.OcsAgentResponse
// @Failure	500					object	http.OcsAgentResponse
// @Router		/api/v1/obproxy/upgrade [post]
func obproxyUpgradeHandler(c *gin.Context) {
	var param param.UpgradeObproxyParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dag, err := obproxy.UpgradeObproxy(param)
	common.SendResponse(c, dag, err)
}

// @ID			obproxyPkgUpload
// @Summary	Upload obproxy package
// @Tags		Obproxy
// @Accept		multipart/form-data
// @Produce	application/json
// @Param		X-OCS-Agent-Header	header		string	true	"Authorization"
// @Param		file				formData	file	true	"Obproxy package"
// @Success	200					object		http.OcsAgentResponse{data=sqlite.UpgradePkgInfo}
// @Failure	400					object		http.OcsAgentResponse
// @Failure	401					object		http.OcsAgentResponse
// @Failure	500					object		http.OcsAgentResponse
// @Router		/api/v1/obproxy/package [post]
func obproxyPkgUploadHandler(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrRequestFileMissing, "file", err.Error()))
		return
	}
	defer file.Close()
	data, agentErr := obproxy.UpgradePkgUpload(file)
	common.SendResponse(c, &data, agentErr)
}
