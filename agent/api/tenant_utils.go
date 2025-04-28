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
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	tenantservice "github.com/oceanbase/obshell/agent/service/tenant"
	"github.com/oceanbase/obshell/param"
)

// Reentrant
func getRootPasswordFromBody(c *gin.Context) (*param.TenantRootPasswordParam, error) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.Errorf("read request body failed: %s", err.Error())
	}
	bodyInterface := make(map[string]interface{})
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &bodyInterface); err != nil {
		return nil, errors.Errorf("unmarshal request body failed: %s", err.Error())
	}

	var param param.TenantRootPasswordParam
	if password, ok := bodyInterface["root_password"]; ok {
		passwordStr := fmt.Sprintf("%v", password)
		param.RootPassword = &passwordStr
	} else {
		param.RootPassword = nil
	}

	return &param, nil
}

// Non-reentrant
func getBodyFromContext(c *gin.Context) (map[string]interface{}, error) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.Errorf("read request body failed: %s", err.Error())
	}
	bodyInterface := make(map[string]interface{})
	if err := json.Unmarshal(bodyBytes, &bodyInterface); err != nil {
		return nil, errors.Errorf("unmarshal request body failed: %s", err.Error())
	}
	return bodyInterface, nil
}

func tenantHandlerWrapper(f func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		// prev check
		err := checkClusterAgent()
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		tenantName := c.Param(constant.URI_PARAM_NAME)
		if tenantName == "" {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "tenant name is empty"))
			return
		}
		if exist, err := tenantService.IsTenantExist(tenantName); err != nil {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrUnexpected, "check tenant '%s' exist failed", tenantName))
			return
		} else if !exist {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "Tenant '%s' not exists.", tenantName))
			return
		}

		if tenantName == constant.TENANT_SYS {
			f(c)
			return
		}

		param, err := getRootPasswordFromBody(c)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}

		if param.RootPassword != nil {
			// attention: please ensure that the body of all API requests forwarded to the execute agent always contains the "password" field; otherwise, it may cause an infinite loop.
			ForwardToActiveAgentWrapper(f)(c)
		} else {
			common.AutoForwardToMaintainerWrapper(ForwardToActiveAgentWrapper(f))(c)
		}
	}
}

func resetTenantRootPassword(c *gin.Context, password *string) error {
	body, err := getBodyFromContext(c)
	if err != nil {
		return err
	}
	body["root_password"] = password
	modifiedBodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
	return nil
}

func ForwardToActiveAgentWrapper(f func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		tenantName := c.Param(constant.URI_PARAM_NAME)
		param, err := getRootPasswordFromBody(c)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		if param.RootPassword == nil {
			// Only maintainers will execute this logic.
			passwordMap := tenantservice.GetPasswordMap()
			password, _ := passwordMap.Get(tenantName)
			param.RootPassword = &password
			if err := resetTenantRootPassword(c, param.RootPassword); err != nil {
				log.WithContext(c).Errorf("reset tenant root password failed: %s", err.Error())
				common.SendResponse(c, nil, err)
				return
			}
		}

		executeAgent, err := GetExecuteAgentForTenant(tenantName)
		if err != nil {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrUnexpected, "get execute agent failed: %s", err.Error()))
			return
		}
		if meta.OCS_AGENT.Equal(executeAgent) {
			f(c)
		} else {
			bodyInterface, err := getBodyFromContext(c)
			if err != nil {
				log.WithContext(c).Errorf("get body from context failed: %s", err.Error())
				common.SendResponse(c, nil, err)
				return
			}
			common.ForwardRequest(c, executeAgent, bodyInterface)
		}
	}
}

func GetExecuteAgentForTenant(tenantName string) (meta.AgentInfoInterface, error) {
	isTenantOn, err := tenantService.IsTenantActiveAgent(tenantName, meta.OCS_AGENT.GetIp(), meta.RPC_PORT)
	if err != nil {
		return nil, err
	}
	if isTenantOn {
		return meta.OCS_AGENT, nil
	}
	executeAgent, err := tenantService.GetTenantActiveAgent(tenantName)
	if err != nil {
		return nil, err
	}
	if executeAgent == nil {
		return executeAgent, errors.New("tenant is not active")
	}
	return executeAgent, err
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
