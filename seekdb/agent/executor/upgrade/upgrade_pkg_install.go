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
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/pkg"
	"github.com/oceanbase/obshell/seekdb/agent/lib/system"
)

type InstallAllRequiredPkgsTask struct {
	task.Task
	targetBuildNumber  string
	targetVersion      string
	targetBuildVersion string
}

func newInstallAllRequiredPkgsTask() *InstallAllRequiredPkgsTask {
	newTask := &InstallAllRequiredPkgsTask{
		Task: *task.NewSubTask(TASK_INSTALL_ALL_REQUIRED_PKGS),
	}
	newTask.
		SetCanContinue().
		SetCanRollback().
		SetCanRetry().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *InstallAllRequiredPkgsTask) Execute() (err error) {
	if t.IsContinue() {
		t.ExecuteLog("The task is continuing")
		if err = t.Rollback(); err != nil {
			return err
		}
	}

	if err = t.installAllRequiredPkgs(); err != nil {
		return
	}
	t.ExecuteLog("install all required pkgs success")
	return nil
}

func (t *InstallAllRequiredPkgsTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_BUILD_NUMBER, &t.targetBuildNumber); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_VERSION, &t.targetVersion); err != nil {
		return err
	}
	t.targetBuildVersion = fmt.Sprintf("%s-%s", t.targetVersion, t.targetBuildNumber)
	log.Infof("target version is %s", t.targetBuildVersion)
	return nil
}

func (t *InstallAllRequiredPkgsTask) installAllRequiredPkgs() (err error) {
	t.ExecuteLog("Unpack and check all packages")
	// Get the target version and build number.
	if err := t.getParams(); err != nil {
		return err
	}

	success := true
	var rpmPkgInfo rpmPacakgeInstallInfo
	if err = t.GetContext().GetDataWithValue(PARAM_UPGRADE_PKG_INSTALL_INFO, &rpmPkgInfo); err != nil {
		t.ExecuteErrorLogf("get data failed, key: %s, err: %s", PARAM_UPGRADE_PKG_INSTALL_INFO, err.Error())
		success = false
	}
	t.ExecuteLogf("Unpack '%s'", rpmPkgInfo.RpmPkgPath)
	if err = pkg.InstallRpmPkgInPlace(rpmPkgInfo.RpmPkgPath); err != nil {
		success = false
		t.ExecuteErrorLog(err)
	}
	t.ExecuteLogf("Successfully installed %s", rpmPkgInfo.RpmPkgPath)
	if !success {
		return errors.Wrap(err, "failed to unpack and check all required pkgs")
	}
	return nil
}

func (t *InstallAllRequiredPkgsTask) getAgentVersion(rpmPkgInfo *rpmPacakgeInstallInfo) (err error) {
	// Because the obshell is unpacked, so need to join the path.
	obshellBinPath := path.Join(rpmPkgInfo.RpmPkgHomepath, constant.DIR_BIN, constant.PKG_OBSHELL)

	// Set the permission of the obshell.
	if err = os.Chmod(obshellBinPath, 0755); err != nil {
		return errors.Wrapf(err, "chmod failed %s", obshellBinPath)
	}

	// Get the version of the target obshell.
	buildVersion, err := system.GetBinaryVersion(obshellBinPath)
	if err != nil {
		return errors.Wrapf(err, "get binary version failed %s", obshellBinPath)
	}

	t.GetContext().SetParam(PARAM_TARGET_AGENT_BUILD_VERSION, buildVersion)
	t.ExecuteLogf("target obshell version is %s", buildVersion)
	return nil
}

func (t *InstallAllRequiredPkgsTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if err = t.uninstallAllRequiredPkgs(); err != nil {
		return
	}
	t.ExecuteLog("Successfully deleted.")
	return nil
}

func (t *InstallAllRequiredPkgsTask) uninstallAllRequiredPkgs() (err error) {
	t.ExecuteLog("Delete all previously installed package files.")
	success := true

	var rpmPkgInfo rpmPacakgeInstallInfo
	if err = t.GetContext().GetDataWithValue(PARAM_UPGRADE_PKG_INSTALL_INFO, &rpmPkgInfo); err != nil {
		t.ExecuteErrorLogf("get data failed, key: %s, err: %s", PARAM_UPGRADE_PKG_INSTALL_INFO, err.Error())
		success = false
	}
	if rpmPkgInfo.RpmPkgExtractPath != "" {
		if err = os.RemoveAll(rpmPkgInfo.RpmPkgExtractPath); err != nil {
			success = false
		}
	}
	if !success {
		return errors.Wrap(err, "uninstall all required pkgs failed")
	}
	return nil
}
