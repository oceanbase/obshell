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
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
)

type UpgradePostTableMaintainTask struct {
	task.Task
}

func newUpgradePostTableMaintainTask() *UpgradePostTableMaintainTask {
	newTask := &UpgradePostTableMaintainTask{
		Task: *task.NewSubTask(TASK_UPGRADE_POST_TABLE_MAINTAIN),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanCancel()
	return newTask
}

func (t *UpgradePostTableMaintainTask) Execute() (err error) {
	t.ExecuteLog("Start to upgrade post table maintain")
	if err := oceanbase.AutoMigrateObTables(true); err != nil {
		return err
	}
	return nil
}
