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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

type InstallNewAgentTask struct {
	task.Task
	realExecAgent meta.AgentInfo
	upgradeRoute  []RouteNode
	rpmPkgInfoKey string
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
	t.rpmPkgInfoKey = targetBuildVersion
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(t.realExecAgent.String(), targetBuildVersion, &t.rpmPkgInfo); err != nil {
		return err
	}
	if t.rpmPkgInfo.RpmName == constant.PKG_OBSHELL && t.rpmPkgInfo.ObshellSHA256 == "" {
		t.ExecuteLog("Re-verify the OBShell RPM to restore verification metadata from an older upgrade task")
		if err = verifyAndExtractObshellRpm(t.GetContext(), &t.rpmPkgInfo); err != nil {
			return err
		}
		t.GetContext().SetAgentDataByAgentKey(t.realExecAgent.String(), t.rpmPkgInfoKey, t.rpmPkgInfo)
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
	src := filepath.Join(t.rpmPkgInfo.RpmPkgHomepath, constant.DIR_BIN, constant.PROC_OBSHELL)
	if t.rpmPkgInfo.RpmName == constant.PKG_OBSHELL {
		return installVerifiedObshellBinary(src, path.ObshellBinPath(), t.rpmPkgInfo.ObshellSHA256)
	}
	if err := os.RemoveAll(path.ObshellBinPath()); err != nil {
		return err
	}
	return system.CopyFile(src, path.ObshellBinPath())
}

func installVerifiedObshellBinary(src, dest, expectedSHA256 string) (err error) {
	if expectedSHA256 == "" {
		return errors.Occur(errors.ErrObPackageCorrupted, constant.PKG_OBSHELL, "missing signed binary digest")
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "close staged OBShell binary")
		}
	}()

	temp, err := os.CreateTemp(filepath.Dir(dest), ".obshell-upgrade-*")
	if err != nil {
		return errors.Wrap(err, "create verified OBShell binary")
	}
	tempPath := temp.Name()
	defer func() {
		if removeErr := os.Remove(tempPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.WithError(removeErr).Warn("remove temporary OBShell binary")
		}
	}()

	digest := sha256.New()
	if _, err := io.Copy(io.MultiWriter(temp, digest), source); err != nil {
		if closeErr := temp.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close incomplete OBShell binary")
		}
		return errors.Wrap(err, "copy staged OBShell binary")
	}
	actualSHA256 := hex.EncodeToString(digest.Sum(nil))
	if actualSHA256 != expectedSHA256 {
		if closeErr := temp.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close rejected OBShell binary")
		}
		return errors.Occur(errors.ErrObPackageCorrupted, constant.PKG_OBSHELL, "staged binary digest does not match signed RPM")
	}
	if err := temp.Chmod(0755); err != nil {
		if closeErr := temp.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell binary after chmod failure")
		}
		return errors.Wrap(err, "chmod verified OBShell binary")
	}
	if err := temp.Sync(); err != nil {
		if closeErr := temp.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell binary after sync failure")
		}
		return errors.Wrap(err, "sync verified OBShell binary")
	}
	if err := temp.Close(); err != nil {
		return errors.Wrap(err, "close verified OBShell binary")
	}
	if err := os.Rename(tempPath, dest); err != nil {
		return errors.Wrap(err, "replace OBShell binary")
	}
	return nil
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
