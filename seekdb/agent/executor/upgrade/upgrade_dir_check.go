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

package upgrade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/system"
)

type CreateUpgradeDirTask struct {
	task.Task
	taskTime            string
	upgradeDir          string
	upgradeCheckTaskDir string
}

func newCreateUpgradeDirTask() *CreateUpgradeDirTask {
	newTask := &CreateUpgradeDirTask{
		Task: *task.NewSubTask(TASK_CREATE_UPGRADE_DIR),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanRollback().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *CreateUpgradeDirTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TASK_TIME, &t.taskTime); err != nil {
		return err
	}

	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return err
	}
	if t.upgradeDir == "" {
		t.ExecuteLog("Since the upgrade directory is not specified, the default upgrade directory is used")
		t.upgradeDir = filepath.Join(global.HomePath, "upgrade")
	}
	t.ExecuteLogf("Upgrade dir is %s", t.upgradeDir)
	return nil
}

func (t *CreateUpgradeDirTask) Execute() (err error) {
	if err = t.checkUpgradeDir(); err != nil {
		return err
	}
	return nil
}

func (t *CreateUpgradeDirTask) checkUpgradeDir() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	t.SetLocalData(PARAM_UPGRADE_DIR, t.upgradeDir)
	t.ExecuteLogf("Mkdir %s ", t.upgradeDir)
	if err = os.MkdirAll(t.upgradeDir, 0755); err != nil {
		return err
	}

	t.upgradeCheckTaskDir = t.generateUpgradeCheckTaskDir()
	t.ExecuteLogf("The temporary directory for this task is %s ", t.upgradeCheckTaskDir)
	if _, err = os.Stat(t.upgradeCheckTaskDir); err != nil && !os.IsNotExist(err) {
		return err
	}
	isDirEmpty, err := system.IsDirEmpty(t.upgradeCheckTaskDir)
	if err != nil {
		return err
	}
	if !isDirEmpty {
		return errors.Occur(errors.ErrCommonDirNotEmpty, t.upgradeCheckTaskDir)
	}
	t.SetLocalData(PARAM_UPGRADE_CHECK_TASK_DIR, t.upgradeCheckTaskDir)
	return nil
}

func (t *CreateUpgradeDirTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if dirParam := t.GetLocalData(PARAM_UPGRADE_DIR); dirParam == nil {
		return nil
	}
	if err = t.GetLocalDataWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return err
	}
	if t.upgradeDir == "" {
		return nil
	}
	t.ExecuteLog("Remove " + t.upgradeDir)
	return os.RemoveAll(t.upgradeDir)
}

func (t *CreateUpgradeDirTask) generateUpgradeCheckTaskDir() string {
	return fmt.Sprintf("%s/rpm-%s", t.upgradeDir, t.taskTime)
}
