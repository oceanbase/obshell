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
	"github.com/oceanbase/obshell/agent/engine/task"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

type SubTaskLogService struct {
	taskService
}

func (s *SubTaskLogService) InsertRemote(subTaskLog task.TaskExecuteLogDTO) (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Create(&oceanbase.SubTaskLog{
		SubTaskId:    subTaskLog.TaskId,
		ExecuteTimes: subTaskLog.ExecuteTimes,
		LogContent:   subTaskLog.LogContent,
	}).Error
}

func (s *SubTaskLogService) InsertLocal(subTaskLog task.TaskExecuteLogDTO) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Create(&sqlite.SubTaskLog{
		SubTaskId:    subTaskLog.TaskId,
		ExecuteTimes: subTaskLog.ExecuteTimes,
		LogContent:   subTaskLog.LogContent,
		IsSync:       subTaskLog.IsSync,
	}).Error
}

func (s *SubTaskLogService) InsertLocalToRemote(subTaskLog task.TaskExecuteLogDTO) error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	var taskMap sqlite.TaskMapping
	if err = sqliteDb.Model(&taskMap).Where("local_task_id=?", subTaskLog.TaskId).First(&taskMap).Error; err != nil {
		return err
	}

	return oceanbaseDb.Create(&oceanbase.SubTaskLog{
		SubTaskId:    taskMap.RemoteTaskId,
		ExecuteTimes: subTaskLog.ExecuteTimes,
		LogContent:   subTaskLog.LogContent,
	}).Error
}

func (s *SubTaskLogService) SetLocalIsSync(subTaskLog *sqlite.SubTaskLog) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Model(subTaskLog).Update("is_sync", 1).Error
}

func (s *taskService) GetSubTaskLogsByTaskID(taskID int64) (subTaskLogs []string, err error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	err = db.Model(s.getSubTaskLogModel()).Select("log_content").Where("sub_task_id=?", taskID).Find(&subTaskLogs).Error
	return
}

func (s *taskService) GetFullSubTaskLogsByTaskID(taskID int64) (subTaskLogs []*bo.SubTaskLog, err error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	dest := s.getSubTaskLogModelSlice()
	if err = db.Model(s.getSubTaskLogModel()).Where("sub_task_id = ?", taskID).Find(dest).Error; err != nil {
		return nil, err
	}
	subTaskLogsBO := s.convertSubTaskLogBOSlice(dest)
	if err != nil {
		return nil, err
	}
	return subTaskLogsBO, nil
}

func (s *SubTaskLogService) GetUnSyncSubTaskLogById(id int64, limit int) (subTaskLogs []sqlite.SubTaskLog, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}

	columns := "sub_task_log.id as id, task_mapping.local_task_id as sub_task_id, " +
		"sub_task_log.execute_times as execute_times, log_content, sub_task_log.is_sync, " +
		"sub_task_log.create_time as create_time, sub_task_log.update_time as update_time"

	err = sqliteDb.Model(&subTaskLogs).Select(columns).
		Joins("join task_mapping on sub_task_log.sub_task_id = task_mapping.local_task_id").
		Where("sub_task_log.id > ? and sub_task_log.is_sync = 0", id).Limit(limit).Find(&subTaskLogs).Error
	return
}
