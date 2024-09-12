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

package task

import (
	"github.com/gin-gonic/gin"

	agentservice "github.com/oceanbase/obshell/agent/service/agent"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
	"github.com/oceanbase/obshell/param"
)

var (
	localTaskService   = taskservice.NewLocalTaskService()
	clusterTaskService = taskservice.NewClusterTaskService()
	agentService       = agentservice.AgentService{}
)

var (
	// for local scale out dag
	PARAM_COORDINATE_DAG_ID = "coordinateDagId"
	PARAM_COORDINATE_AGENT  = "coordinateAgent"
)

func getTaskQueryParams(c *gin.Context) *param.TaskQueryParams {
	var params param.TaskQueryParams
	c.ShouldBindQuery(&params)
	return &params
}
