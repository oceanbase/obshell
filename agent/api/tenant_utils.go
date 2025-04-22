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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/tenant"
	"github.com/oceanbase/obshell/agent/meta"
)

func tenantHandlerWrapper(f func(*gin.Context)) func(*gin.Context) {
	return checkClusterAgentWrapper(common.AutoForwardToMaintainerWrapper(checkTenantRootpasswordValidWrapper(f)))
}

func checkTenantRootpasswordValidWrapper(f func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		name := c.Param(constant.URI_PARAM_NAME)
		if exist, err := tenantService.IsTenantExist(name); err != nil {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrUnexpected, "check tenant '%s' exist failed", name))
		} else if !exist {
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "Tenant '%s' not exists.", name))
		}
		db, err := tenant.GetConnection(name)
		defer tenant.CloseDbConnection(db)
		if err != nil {
			if strings.Contains(err.Error(), "Access denied") {
				common.SendResponse(c, nil, errors.Occurf(errors.ErrTenantNotConnectable, "Tenant '%s' password is incorrect.", name))
			} else {
				common.SendResponse(c, nil, errors.Occurf(errors.ErrUnexpected, "Tenant %s is not connectable", name))
			}
		} else {
			f(c)
		}
	}
}

func checkClusterAgentWrapper(f func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := checkClusterAgent()
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		f(c)
	}
}

func checkClusterAgent() error {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String())
	}
	return nil
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
