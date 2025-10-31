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
	"path/filepath"

	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
)

type BackupAgentForUpgradeTask struct {
	task.Task
	upgradeCheckTaskDir string
	backupDir           string
}

func newBackupAgentForUpgradeTask() *BackupAgentForUpgradeTask {
	newTask := &BackupAgentForUpgradeTask{
		Task: *task.NewSubTask(TASK_BACKUP_FOR_UPGRADE),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanRollback().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *BackupAgentForUpgradeTask) Execute() (err error) {
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

	if t.IsContinue() {
		t.ExecuteLog("The task is continuing.")
		if err = t.Rollback(); err != nil {
			return err
		}
	}

	if err = t.BackupAgentForUpgrade(); err != nil {
		return
	}
	return nil
}

func (t *BackupAgentForUpgradeTask) getParams() (err error) {
	if err = t.GetLocalDataWithValue(PARAM_UPGRADE_CHECK_TASK_DIR, &t.upgradeCheckTaskDir); err != nil {
		return err
	}

	t.backupDir = filepath.Join(t.upgradeCheckTaskDir, "backup")
	return nil
}

func (t *BackupAgentForUpgradeTask) BackupAgentForUpgrade() error {
	t.ExecuteLog("Backup important files.")
	if err := t.getParams(); err != nil {
		return err
	}

	t.SetLocalData(DATA_BACKUP_DIR, t.backupDir)
	t.ExecuteLogf("The directory for backup is %s", t.backupDir)
	t.ExecuteLogf("Backup the bin directory %s", path.BinDir())
	if err := system.CopyDirs(path.BinDir(), t.backupDir); err != nil {
		return err
	}
	return nil
}

func (t *BackupAgentForUpgradeTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if err = t.deleteBackupDir(); err != nil {
		return err
	}
	t.ExecuteLog("Successfully deleted")
	return nil
}

func (t *BackupAgentForUpgradeTask) deleteBackupDir() (err error) {
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

	if err = t.getParams(); err != nil {
		return err
	}
	if t.backupDir != "" {
		t.ExecuteLog("Delete " + t.backupDir)
		if err := os.RemoveAll(t.backupDir); err != nil {
			return err
		}
	}
	return nil
}
