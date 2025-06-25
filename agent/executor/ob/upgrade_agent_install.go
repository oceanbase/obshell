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
	"fmt"
	"os"
	"path/filepath"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/meta"
)

type InstallNewAgentTask struct {
	task.Task
	realExecAgent meta.AgentInfo
	upgradeRoute  []RouteNode
	rpmPkgInfo    rpmPacakgeInstallInfo
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

func (t *InstallNewAgentTask) getExecAgent() (err error) {
	_, t.realExecAgent, err = isRealExecuteAgent(t)
	if err != nil {
		return err
	}
	return nil
}

func isRealExecuteAgent(t task.ExecutableTask) (res bool, realExecuteAgent meta.AgentInfo, err error) {
	localAgent := t.GetExecuteAgent()
	var allExecAgents []meta.AgentInfo
	if err = t.GetContext().GetParamWithValue(PARAM_ALL_AGENTS, &allExecAgents); err != nil {
		return
	}
	for _, agent := range allExecAgents {
		if agent.Equal(&localAgent) {
			return true, agent, nil
		}
		if agent.Ip == localAgent.Ip {
			t.ExecuteLog("Due to multiple obshell being on this machine, only one needs to perform this sub task.")
			t.ExecuteLogf("The actual obshell server executing the task is %s", agent.String())
			return false, agent, nil
		}
	}
	return false, realExecuteAgent, errors.Occur(errors.ErrCommonUnexpected, "get real execute agent failed")
}

func (t *InstallNewAgentTask) getParams() (err error) {
	if t.upgradeRoute, err = getUpgradeRouteForTask(t.GetContext()); err != nil {
		return err
	}
	targetBuildVersion := t.upgradeRoute[len(t.upgradeRoute)-1].BuildVersion
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(t.realExecAgent.String(), targetBuildVersion, &t.rpmPkgInfo); err != nil {
		return err
	}
	return nil
}

func (t *InstallNewAgentTask) Execute() (err error) {
	t.ExecuteLog("get real execute agent")
	if err = t.getExecAgent(); err != nil {
		return err
	}
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
	if err = t.getExecAgent(); err != nil {
		return err
	}

	t.ExecuteLog("uninstall new obshell")
	var backupDir string
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(t.realExecAgent.String(), DATA_BACKUP_DIR, &backupDir); err != nil {
		return err
	}

	dest := path.ObshellBinPath()
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return system.CopyFile(fmt.Sprintf("%s/%s", backupDir, constant.PROC_OBSHELL), dest)
}
