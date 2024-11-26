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
	group := r.Group("/task")
	group.POST(constant.URI_SUB_TASK, StartTask)
	group.PATCH(constant.URI_SUB_TASK, UpdateTask)
	group.DELETE(constant.URI_SUB_TASK, CancelTask)
	group.POST(constant.URI_LOG, SyncLog)
}

// StartTask will start remote subtask and create local subtask instance by remote subtask if not exist.
// Cluster task scheduler use this rpc to send task to agent.
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

// UpdateTask used to update the task in cluster.
// If the local task finished and the agent can't commite the task to cluster, the agent will update by this rpc.
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

// CancelTask used to cancel the task in local.
func CancelTask(c *gin.Context) {
	var remoteTask task.RemoteTask
	if err := c.ShouldBind(&remoteTask); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	localTask, err := localTaskService.GetLocalTaskInstanceByRemoteTaskId(remoteTask.TaskID)
	if err != nil {
		log.WithError(err).Warnf("get sub task by task id %d error", remoteTask.TaskID)
		common.SendResponse(c, nil, err)
		return
	} else if localTask == nil {
		log.Warnf("remote task %d not exist in local", remoteTask.TaskID)
		common.SendResponse(c, nil, nil)
		return
	}

	if remoteTask.ExecuteTimes <= localTask.ExecuteTimes {
		log.Warnf("remote task %d execute times %d <= local task %d execute times %d, reject it", remoteTask.TaskID, remoteTask.ExecuteTimes, localTask.Id, localTask.ExecuteTimes)
		common.SendResponse(c, nil, nil)
		return
	}

	go executor.OCS_EXECUTOR_POOL.CancelTask(localTask.Id)
	common.SendResponse(c, nil, nil)
}

// SyncLog used to sync log from local to remote.
// If the agent can't commit the log to cluster, the agent will commit the log by this rpc.
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
