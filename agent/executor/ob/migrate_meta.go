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

package ob

import (
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

type MigrateTableTask struct {
	task.Task
}

func newMigrateTableTask() *MigrateTableTask {
	newTask := &MigrateTableTask{
		Task: *task.NewSubTask(TASK_NAME_MIGRATE_TABLE),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *MigrateTableTask) Execute() (err error) {
	if err = loadOceanbaseInstanceWithoutDBName(t); err != nil {
		return err
	}
	if err = oceanbase.CreateDataBase(constant.DB_OCS); err != nil {
		return errors.Wrap(err, "create database failed")
	}
	if err = oceanbase.LoadOceanbaseInstance(config.NewObMysqlDataSourceConfig()); err != nil {
		return errors.Wrap(err, "connect ocs database failed")
	}
	if err = oceanbase.AutoMigrateObTables(false); err != nil {
		return errors.Wrap(err, "register ob tables failed")
	}
	return nil
}

type MigrateDataTask struct {
	task.Task
}

func newMigrateDataTask() *MigrateDataTask {
	newTask := &MigrateDataTask{
		Task: *task.NewSubTask(TASK_NAME_MIGRATE_DATA),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *MigrateDataTask) Execute() error {
	return t.migrateAllAgents()
}

func (t *MigrateDataTask) migrateAllAgents() error {
	t.ExecuteLog("migrate all agents")
	return obclusterService.MigrateAllAgentToOb()
}
