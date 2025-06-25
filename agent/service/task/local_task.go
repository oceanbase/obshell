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
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/errors"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

type LocalTaskService struct {
	taskService
}

func NewLocalTaskService() *LocalTaskService {
	return &LocalTaskService{
		taskService: taskService{
			StatusMaintainerInterface: &agentStatusMaintainer{},
			isLocal:                   true,
		},
	}
}

func (s *LocalTaskService) DeleteRemoteTask() error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("node_id = 0").Delete(&sqlite.SubtaskInstance{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&sqlite.TaskMapping{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&sqlite.SubTaskLog{}).Update("is_sync", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *LocalTaskService) GetNodeOperatorBySubTaskId(taskID int64) (int, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return 0, err
	}

	var subTaskInstance sqlite.SubtaskInstance
	if err := db.Model(&subTaskInstance).Where("id=?", taskID).First(&subTaskInstance).Error; err != nil {
		return 0, err
	}

	if subTaskInstance.NodeId == 0 {
		return 0, errors.Occurf(errors.ErrCommonUnexpected, "task %d is a remote task", taskID)
	}

	var node sqlite.NodeInstance
	if err := db.Where("id = ?", subTaskInstance.NodeId).First(&node).Error; err != nil {
		return 0, err
	}

	return node.Operator, nil
}
