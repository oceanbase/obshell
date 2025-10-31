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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/coordinator"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

type taskLogSynchronizer struct {
	coordinator   *coordinator.Coordinator
	lastSyncLogId int64
	unSyncLogList []*sqlite.SubTaskLog
}

func newTaskLogSynchronizer(coordinator *coordinator.Coordinator) *taskLogSynchronizer {
	return &taskLogSynchronizer{
		coordinator:   coordinator,
		lastSyncLogId: 0,
		unSyncLogList: make([]*sqlite.SubTaskLog, 0, constant.SYNC_TASK_LOG_BUFFER_SIZE),
	}
}

func (synchronizer *taskLogSynchronizer) sync() {
	newUnSyncLogList := make([]*sqlite.SubTaskLog, 0)
	for idx := range synchronizer.unSyncLogList {
		taskLog := synchronizer.unSyncLogList[idx]
		if err := synchronizer.syncTaskLog(taskLog); err != nil {
			log.WithError(err).Warnf("sync task log %d failed", taskLog.Id)
			newUnSyncLogList = append(newUnSyncLogList, taskLog)
		}
	}

	if len(newUnSyncLogList) == 0 {
		taskLogs, err := subTaskLogService.GetUnSyncSubTaskLogById(synchronizer.lastSyncLogId, constant.SYNC_TASK_LOG_BUFFER_SIZE)
		if err != nil {
			log.WithError(err).Warn("get unsync task log failed")
			return
		}
		for idx := range taskLogs {
			taskLog := taskLogs[idx]
			if taskLog.Id > synchronizer.lastSyncLogId {
				synchronizer.lastSyncLogId = taskLog.Id
			}
			if err = synchronizer.syncTaskLog(&taskLog); err != nil {
				log.WithError(err).Warnf("sync task log %d failed", taskLog.Id)
				newUnSyncLogList = append(newUnSyncLogList, &taskLog)
			}
		}
	}
	synchronizer.unSyncLogList = newUnSyncLogList
	if len(synchronizer.unSyncLogList) > 0 {
		log.Infof("finish sync task log, last sync log id %d, unsync log count %d", synchronizer.lastSyncLogId, len(synchronizer.unSyncLogList))
	}
}

func (synchronizer *taskLogSynchronizer) syncTaskLog(taskLog *sqlite.SubTaskLog) error {
	remoteTaskId, err := localTaskService.GetRemoteTaskIdByLocalTaskId(taskLog.SubTaskId)
	if err != nil {
		return errors.Wrapf(err, "get remote task id by local task id %d failed", taskLog.SubTaskId)
	}
	tasklogDTO := task.TaskExecuteLogDTO{
		TaskId:       remoteTaskId,
		ExecuteTimes: taskLog.ExecuteTimes,
		LogContent:   taskLog.LogContent,
	}
	// Try to sync task log by service.
	if err := subTaskLogService.InsertLocalToRemote(tasklogDTO); err != nil {
		log.WithError(err).Warnf("sync task log %d by service failed", taskLog.Id)
		// Sync task log by rpc.
		if err := postTaskLogToRemote(tasklogDTO); err != nil {
			return errors.Wrapf(err, "sync task log %d by rpc failed", taskLog.Id)
		}
	}

	// Set task log sync status.
	taskLog.IsSync = true
	if err := subTaskLogService.SetLocalIsSync(taskLog); err != nil {
		return errors.Wrapf(err, "set task log %d sync status failed", taskLog.Id)
	}
	log.Infof("sync task log %d success", taskLog.Id)
	return nil
}

func postTaskLogToRemote(taskLog task.TaskExecuteLogDTO) error {
	maintainerAgent := coordinator.OCS_COORDINATOR.Maintainer
	log.Infof("send task log to %s", maintainerAgent.String())
	return secure.SendPostRequest(maintainerAgent, constant.URI_TASK_RPC_PREFIX+constant.URI_LOG, taskLog, nil)
}
