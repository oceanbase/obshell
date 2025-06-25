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
)

type CheckAllRequiredPkgsTask struct {
	task.Task
}

func newCheckAllRequiredPkgsTask() *CheckAllRequiredPkgsTask {
	newTask := &CheckAllRequiredPkgsTask{
		Task: *task.NewSubTask(TASK_CHECK_ALL_REQUIRED_PKGS),
	}
	newTask.
		SetCanContinue().
		SetCanRollback().
		SetCanRetry().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *CheckAllRequiredPkgsTask) Execute() (err error) {
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

	if err = t.checkAllRequiredPkgs(); err != nil {
		return
	}
	return nil
}

func (t *CheckAllRequiredPkgsTask) checkAllRequiredPkgs() (err error) {
	keys, err := getKeyForPkgInfoMap(t.GetContext())
	if err != nil {
		return err
	}
	for _, key := range keys {
		t.ExecuteLogf("Check the package for version %s", key)
		var rpmPkgInfo rpmPacakgeInstallInfo
		if err = t.GetLocalDataWithValue(key, &rpmPkgInfo); err != nil {
			t.ExecuteErrorLogf("get local data failed, key: %s, err: %s", key, err.Error())
			return err
		}
		t.ExecuteLogf("  package name is %s", rpmPkgInfo.RpmName)
		t.ExecuteLogf("  package build version is %s", rpmPkgInfo.RpmBuildVersion)
		t.ExecuteLogf("  package path is %s", rpmPkgInfo.RpmPkgPath)
		t.ExecuteLogf("  package extract path is %s", rpmPkgInfo.RpmPkgExtractPath)
		t.ExecuteLogf("  package home path is %s", rpmPkgInfo.RpmPkgHomepath)
		if err = t.checkUpgradePkgFromDb(rpmPkgInfo.RpmPkgPath); err != nil {
			t.ExecuteErrorLogf("  check failed: %v", err)
			return err
		}
		t.ExecuteLog("The package check is complete.")
	}

	t.ExecuteInfoLog("All packages are checked successfully.")
	return nil
}

func (t *CheckAllRequiredPkgsTask) checkUpgradePkgFromDb(filePath string) (err error) {
	input, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer input.Close()
	r := &upgradeRpmPkgInfo{
		rpmFile: input,
	}

	if err = r.CheckUpgradePkg(false); err != nil {
		return err
	}
	return nil
}
