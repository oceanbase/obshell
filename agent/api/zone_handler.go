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
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/meta"
)

//	@ID				DeleteZone
//
//	@Summary		delete zone
//	@Description	delete zone
//	@Tags			ob
//	@Accept			application/json
//	@Produce		application/json
//	@Param			X-OCS-Header	header	string	true	"Authorization"
//	@Param			zoneName		path	string	true	"zone name"
//	@Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
//	@Success		204				object	http.OcsAgentResponse
//	@Failure		401				object	http.OcsAgentResponse
//	@Failure		500				object	http.OcsAgentResponse
//	@Router			/api/v1/zone/{zoneName} [delete]
func zoneDeleteHandler(c *gin.Context) {
	zoneName := c.Param(constant.URI_PARAM_NAME)
	if zoneName == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "zone name is empty"))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	dag, err := ob.DeleteZone(zoneName)
	if dag == nil && err == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err)
	}
}
