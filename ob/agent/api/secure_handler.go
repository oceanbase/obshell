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

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/executor/session"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
)

// @ID getSecret
// @Summary get secret
// @Description get secret
// @Tags v1
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=meta.AgentSecret}
// @Router /api/v1/secret [get]
func secretHandler(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	data := secure.GetSecret(ctx)
	common.SendResponse(c, data, nil)
}

// @ID login
// @Summary login
// @Description login
// @Tags v1
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/login [post]
func loginHandler(c *gin.Context) {
	data, err := session.Login(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, data, nil)
}

// @ID logout
// @Summary logout
// @Description logout
// @Tags v1
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Param body body param.LogoutSessionParam true "logout session param"
// @Router /api/v1/logout [post]
func logoutHandler(c *gin.Context) {
	var param param.LogoutSessionParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := session.Logout(c, param.SessionID)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, nil)
}
