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

package rpc

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/executor"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
)

var (
	logService         = taskservice.SubTaskLogService{}
	localTaskService   = taskservice.NewLocalTaskService()
	clusterTaskService = taskservice.NewClusterTaskService()
)

func InitTaskRoutes(r *gin.RouterGroup) {
	task := r.Group("/task")
	task.POST(constant.URI_SUB_TASK, StartTask)
	task.PATCH(constant.URI_SUB_TASK, UpdateTask)
	task.POST(constant.URI_LOG, SyncLog)
}

// StartTask will start remote subtask and create local subtask instance by remote subtask if not exist.
func StartTask(c *gin.Context) {
	var remoteTask task.RemoteTask
	if err := c.ShouldBind(&remoteTask); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	localTask, err := localTaskService.GetLocalTaskInstanceByRemoteTaskId(remoteTask.TaskID)
	if err != nil {
		log.WithError(err).Warnf("get local task instance by remote task %d error", remoteTask.TaskID)
		common.SendResponse(c, nil, err)
		return
	}
	if localTask == nil {
		localTaskId, err := localTaskService.CreateLocalTaskInstanceByRemoteTask(&remoteTask)
		if err != nil {
			log.WithError(err).Warnf("create local task instance by remote task %d error", remoteTask.TaskID)
			common.SendResponse(c, nil, err)
			return
		}
		executor.OCS_EXECUTOR_POOL.AddTask(localTaskId)
		common.SendResponse(c, task.TaskDetail{TaskID: localTaskId}, nil)
		return
	} else {
		if remoteTask.ExecuteTimes <= localTask.ExecuteTimes {
			common.SendResponse(c, task.TaskDetail{TaskID: localTask.Id}, nil)
			return
		} else {
			err := localTaskService.UpdateLocalTaskInstanceByRemoteTask(&remoteTask)
			if err != nil {
				log.WithError(err).Warnf("update local task %d instance by remote task %d error", localTask.Id, remoteTask.TaskID)
				common.SendResponse(c, nil, err)
				return
			}
			executor.OCS_EXECUTOR_POOL.AddTask(localTask.Id)
			common.SendResponse(c, task.TaskDetail{TaskID: localTask.Id}, nil)
			return
		}
	}
}

func UpdateTask(c *gin.Context) {
	var remoteTask task.RemoteTask
	if err := c.ShouldBind(&remoteTask); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	taskInstance, err := clusterTaskService.GetSubTaskByTaskID(remoteTask.TaskID)
	if err != nil {
		log.WithError(err).Warnf("get sub task by task id %d error", remoteTask.TaskID)
		common.SendResponse(c, nil, err)
		return
	}

	if taskInstance.GetExecuteTimes() != remoteTask.ExecuteTimes {
		err := errors.Occur(errors.ErrUnexpected, "execute times not match")
		log.Warnf("execute times not match, local task %d, remote task %d", taskInstance.GetExecuteTimes(), remoteTask.ExecuteTimes)
		common.SendResponse(c, nil, err)
		return
	}

	if taskInstance.IsPending() || taskInstance.IsFinished() {
		err := errors.Occur(errors.ErrUnexpected, "task not running")
		log.Warnf("task not running, local task %d, remote task %d", taskInstance.GetExecuteTimes(), remoteTask.ExecuteTimes)
		common.SendResponse(c, nil, err)
		return
	}

	// Start remote task.
	if taskInstance.IsReady() {
		if remoteTask.State != task.RUNNING {
			err := errors.Occur(errors.ErrUnexpected, "task not running")
			log.Warnf("task not running, local task %d, remote task %d", taskInstance.GetExecuteTimes(), remoteTask.ExecuteTimes)
			common.SendResponse(c, nil, err)
			return
		}
		if err := clusterTaskService.StartSubTask(taskInstance); err != nil {
			log.WithError(err).Warnf("start sub task %d error", taskInstance.GetExecuteTimes())
			common.SendResponse(c, nil, err)
			return
		}
		common.SendResponse(c, nil, nil)
		return
	}

	// Finish remote task.
	if taskInstance.IsRunning() {
		if remoteTask.State != task.FAILED && remoteTask.State != task.SUCCEED {
			err := errors.Occur(errors.ErrUnexpected, "task is running")
			common.SendResponse(c, nil, err)
			return
		}
		taskInstance.SetContext(&remoteTask.Context)
		taskInstance.SetState(remoteTask.State)
		if err := clusterTaskService.FinishSubTask(taskInstance, taskInstance.GetState()); err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		common.SendResponse(c, nil, nil)
		return
	}

}

func SyncLog(c *gin.Context) {
	var taskLogDTIO task.TaskExecuteLogDTO
	if err := c.ShouldBind(&taskLogDTIO); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	subTask, err := clusterTaskService.GetSubTaskByTaskID(taskLogDTIO.TaskId)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if subTask.GetExecuteTimes() > taskLogDTIO.ExecuteTimes {
		log.Infof("task %d execute times %d > log execute times %d", subTask.GetID(), subTask.GetExecuteTimes(), taskLogDTIO.ExecuteTimes)
		common.SendResponse(c, nil, nil)
		return
	}

	if subTask.IsPending() || subTask.IsReady() {
		log.Infof("task %d state is %s", subTask.GetID(), task.STATE_MAP[subTask.GetState()])
		common.SendResponse(c, nil, nil)
		return
	}

	if err := logService.InsertRemote(taskLogDTIO); err != nil {
		log.WithError(err).Warn("insert remote log error")
		common.SendResponse(c, nil, err)
		return
	}

	common.SendResponse(c, nil, nil)
}
