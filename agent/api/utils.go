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
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
)

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
		return errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}
	return nil
}
