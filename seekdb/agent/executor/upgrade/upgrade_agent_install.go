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

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/lib/system"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
)

type InstallNewAgentTask struct {
	task.Task
	rpmPkgInfo rpmPacakgeInstallInfo
}

func newInstallNewAgentTask() *InstallNewAgentTask {
	newTask := &InstallNewAgentTask{
		Task: *task.NewSubTask(TASK_INSTALL_NEW_OBSHELL),
	}
	newTask.
		SetCanRetry().
		SetCanRollback().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *InstallNewAgentTask) getParams() (err error) {
	if err = t.GetContext().GetDataWithValue(PARAM_UPGRADE_PKG_INSTALL_INFO, &t.rpmPkgInfo); err != nil {
		return err
	}
	return nil
}

func (t *InstallNewAgentTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	if err = t.installNewAgent(); err != nil {
		return
	}
	return nil
}

func (t *InstallNewAgentTask) installNewAgent() error {
	t.ExecuteLogf("Install new obshell '%s'", t.rpmPkgInfo.RpmPkgHomepath)
	if err := os.RemoveAll(path.ObshellBinPath()); err != nil {
		return err
	}
	src := filepath.Join(t.rpmPkgInfo.RpmPkgHomepath, constant.DIR_BIN, constant.PROC_OBSHELL)
	return system.CopyFile(src, path.ObshellBinPath())
}

func (t *InstallNewAgentTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	t.ExecuteLog("uninstall new obshell")
	var backupDir string
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(meta.OCS_AGENT.String(), DATA_BACKUP_DIR, &backupDir); err != nil {
		return err
	}

	dest := path.ObshellBinPath()
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return system.CopyFile(fmt.Sprintf("%s/%s", backupDir, constant.PROC_OBSHELL), dest)
}
