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
	"os"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
)

type RemoveUpgradeCheckDirTask struct {
	task.Task
}

func newRemoveUpgradeCheckDirTask() *RemoveUpgradeCheckDirTask {
	newTask := &RemoveUpgradeCheckDirTask{
		Task: *task.NewSubTask(TASK_REMOVE_UPGRADE_CHECK_TASK_DIR),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *RemoveUpgradeCheckDirTask) Execute() (err error) {
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

	t.ExecuteLog("remove upgrade check dir")
	if err = t.removeUpgradeCheckDir(); err != nil {
		return
	}
	t.ExecuteLog("remove upgrade check dir finished")
	return nil
}

func (t *RemoveUpgradeCheckDirTask) removeUpgradeCheckDir() (err error) {
	upgradeCheckTaskDir, ok := t.GetLocalData(PARAM_UPGRADE_CHECK_TASK_DIR).(string)
	if !ok {
		return errors.New("get upgrade check task dir failed")
	}
	return os.RemoveAll(upgradeCheckTaskDir)
}
