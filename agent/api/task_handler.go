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
	"github.com/oceanbase/obshell/agent/executor/task"
	"github.com/oceanbase/obshell/agent/secure"
)

func InitTaskRoutes(r *gin.RouterGroup, isLocalRoute bool) {
	group := r.Group(constant.URI_TASK_GROUP)
	if !isLocalRoute {
		group.Use(common.Verify(secure.ROUTE_TASK))
	}
	group.GET(constant.URI_SUB_TASK+"/:id", task.GetSubTaskDetail)
	group.GET(constant.URI_NODE+"/:id", task.GetNodeDetail)
	group.GET(constant.URI_DAG+"/:id", task.GetDagDetail)
	group.POST(constant.URI_DAG+"/:id", task.DagHandler)
	group.GET(constant.URI_DAG+constant.URI_MAINTAIN+constant.URI_OB_GROUP, task.GetObLastMaintenanceDag)
	group.GET(constant.URI_DAG+constant.URI_MAINTAIN+constant.URI_AGENT_GROUP, task.GetAgentLastMaintenanceDag)
	group.GET(constant.URI_DAG+constant.URI_MAINTAIN+constant.URI_AGENTS_GROUP, task.GetAllAgentLastMaintenanceDag)
	group.GET(constant.URI_DAG+constant.URI_UNFINISH, task.GetUnfinishedDags)
	group.GET(constant.URI_DAG+constant.URI_OB_GROUP+constant.URI_UNFINISH, task.GetClusterUnfinishDags)
	group.GET(constant.URI_DAG+constant.URI_AGENT_GROUP+constant.URI_UNFINISH, task.GetAgentUnfinishDags)

}
