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

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
)

// get sub_task detail by id
//
//	@ID				getSubTaskDetail
//	@Summary		get sub_task detail by sub_task_id
//	@Description	get sub_task detail by sub_task_id
//	@Tags			task
//	@Accept			application/json
//	@Produce		application/json
//	@Param			X-OCS-Header	header	string					true	"Authorization"
//	@Param			id				path	string					true	"id"
//	@Param			showDetails		query	param.TaskQueryParams	true	"show details"
//	@Success		200				object	http.OcsAgentResponse{data=task.TaskDetailDTO}
//	@Failure		400				object	http.OcsAgentResponse
//	@Failure		404				object	http.OcsAgentResponse
//	@Failure		500				object	http.OcsAgentResponse
//	@Router			/api/v1/task/sub_task/{id} [get]
func GetSubTaskDetail(c *gin.Context) {
	var taskDTOParam task.TaskDetailDTO
	var taskDetailDTO *task.TaskDetailDTO
	var service taskservice.TaskServiceInterface

	if err := c.BindUri(&taskDTOParam); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	taskID, agent, err := task.ConvertGenericID(taskDTOParam.GenericID)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		if meta.OCS_AGENT.IsFollowerAgent() {
			// forward request to master
			master := agentService.GetMasterAgentInfo()
			if master == nil {
				common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Master Agent is not found"))
				return
			}
			common.ForwardRequest(c, master, nil)
		} else {
			common.ForwardRequest(c, agent, nil)
		}
		return
	}

	if agent == nil {
		service = clusterTaskService
	} else {
		service = localTaskService
	}

	subTask, err := service.GetSubTaskByTaskID(taskID)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound, err))
		return
	}

	taskDetailDTO, err = getSubTaskDetail(service, subTask)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrUnexpected, err))
		return
	}
	taskDetailDTO.SetVisible(true)
	common.SendResponse(c, taskDetailDTO, nil)
}

func getSubTaskDetail(service taskservice.TaskServiceInterface, subTask task.ExecutableTask) (taskDetailDTO *task.TaskDetailDTO, err error) {
	taskDetailDTO = task.NewTaskDetailDTO(subTask)
	if subTask.IsRunning() || subTask.IsFinished() {
		taskDetailDTO.TaskLogs, err = service.GetSubTaskLogsByTaskID(subTask.GetID())
	}
	return
}
