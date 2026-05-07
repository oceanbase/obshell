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
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

type SubTaskLogService struct {
	taskService
}

func (s *SubTaskLogService) InsertRemote(subTaskLog task.TaskExecuteLogDTO) (err error) {
	return nil
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
	return nil
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
