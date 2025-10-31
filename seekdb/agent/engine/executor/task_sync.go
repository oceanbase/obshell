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

package executor

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

type taskSynchronizer struct {
	lastSyncTaskTime time.Time
	unSyncTaskList   []int64 // GmtModify < lastSyncTaskTime && unsync, remote task id
}

func newTaskSynchronizer() *taskSynchronizer {
	return &taskSynchronizer{
		lastSyncTaskTime: constant.ZERO_TIME,
		unSyncTaskList:   make([]int64, 0, constant.SYNC_TASK_BUFFER_SIZE),
	}
}

func (synchronizer *taskSynchronizer) sync() {
	newUnSyncTaskList := make([]int64, 0)
	for _, taskId := range synchronizer.unSyncTaskList {
		taskMap, err := localTaskService.GetTaskMappingByRemoteTaskId(taskId)
		if err != nil {
			log.WithError(err).Warnf("get task mapping by remote task id %d failed", taskId)
			continue
		}
		if err = synchronizer.syncTask(*taskMap); err != nil {
			log.WithError(err).Warnf("sync task %d failed", taskMap.RemoteTaskId)
			newUnSyncTaskList = append(newUnSyncTaskList, taskMap.RemoteTaskId)
		}
	}

	if len(newUnSyncTaskList) == 0 {
		taskMapList, err := localTaskService.GetUnSyncTaskMappingByTime(synchronizer.lastSyncTaskTime, constant.SYNC_TASK_BUFFER_SIZE)
		if err != nil {
			log.WithError(err).Warn("get unsync task mapping failed")
			return
		}

		for _, taskMap := range taskMapList {
			if taskMap.GmtModify.After(synchronizer.lastSyncTaskTime) {
				synchronizer.lastSyncTaskTime = taskMap.GmtModify
			}
			if err = synchronizer.syncTask(taskMap); err != nil {
				log.WithError(err).Warnf("sync task %d failed", taskMap.RemoteTaskId)
				newUnSyncTaskList = append(newUnSyncTaskList, taskMap.RemoteTaskId)
			}
		}
	}
	synchronizer.unSyncTaskList = newUnSyncTaskList
	if len(synchronizer.unSyncTaskList) > 0 {
		log.Infof("finish sync task , last sync time %s, unsync task count %d", synchronizer.lastSyncTaskTime, len(synchronizer.unSyncTaskList))
	}
}

func (synchronizer *taskSynchronizer) syncTask(taskMap sqlite.TaskMapping) error {
	subtask, err := localTaskService.GetSubTaskByTaskID(taskMap.LocalTaskId)
	if err != nil {
		return err
	}
	if err = finishRemoteTaskByService(taskMap.RemoteTaskId, subtask); err != nil {
		log.WithError(err).Warn("finish remote task by service failed")
		return err
	}

	// Finish remote task success, set task mapping sync
	if err = localTaskService.SetTaskMappingSync(taskMap.RemoteTaskId, taskMap.ExecuteTimes); err != nil {
		return err
	}
	return nil
}

func finishRemoteTaskByService(remoteTaskId int64, subTask task.ExecutableTask) (err error) {
	// Finish task in remote, try to get remote task from ob
	remoteSubTask, err := clusterTaskService.GetSubTaskByTaskID(remoteTaskId)
	if err != nil {
		return err
	}
	// Only finish remote task when execute times is equal
	if remoteSubTask.GetExecuteTimes() == subTask.GetExecuteTimes() {
		if remoteSubTask.IsFinished() {
			log.Debugf("remote task %d is finished, execute times %d", remoteTaskId, remoteSubTask.GetExecuteTimes())
		} else {
			// Try to finish remote task in ob
			remoteSubTask.SetContext(subTask.GetContext())
			if err = clusterTaskService.FinishSubTask(remoteSubTask, subTask.GetState()); err != nil {
				return err
			}
		}
		return nil
	} else {
		log.Warnf("remote task %d execute times %d != local task %d execute times %d", remoteTaskId, remoteSubTask.GetExecuteTimes(), subTask.GetID(), subTask.GetExecuteTimes())
		return nil
	}

}
