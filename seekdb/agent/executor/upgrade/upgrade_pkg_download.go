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

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/system"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
)

var (
	confficient = 1.1
)

const (
	OCEANBASE_HOMEPATH = "/home/admin/oceanbase"
	OCEANBASE_HOME     = "/home"
)

type GetAllRequiredPkgsTask struct {
	task.Task
	upgradeDir          string
	targetBuildNumber   string
	targetVersion       string
	distribution        string
	upgradePkgInfo      oceanbase.UpgradePkgInfo
	upgradeCheckTaskDir string
}

func newGetAllRequiredPkgsTask() *GetAllRequiredPkgsTask {
	newTask := &GetAllRequiredPkgsTask{
		Task: *task.NewSubTask(TASK_GET_ALL_REQUIRED_PKGS),
	}
	newTask.
		SetCanContinue().
		SetCanRollback().
		SetCanRetry().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *GetAllRequiredPkgsTask) Execute() (err error) {
	if t.IsContinue() {
		t.ExecuteLog("The task is continuing.")
		if err = t.Rollback(); err != nil {
			return err
		}
	}

	if err = t.getAllRequiredPkgs(); err != nil {
		return
	}
	return nil
}

func (t *GetAllRequiredPkgsTask) getParams() (err error) {
	if err = t.GetLocalDataWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return err
	}
	if err = t.GetLocalDataWithValue(PARAM_UPGRADE_CHECK_TASK_DIR, &t.upgradeCheckTaskDir); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_BUILD_NUMBER, &t.targetBuildNumber); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_VERSION, &t.targetVersion); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_DISTRIBUTION, &t.distribution); err != nil {
		return err
	}
	t.ExecuteLogf("The required upgrade package is %v", fmt.Sprintf("%s-%s", t.targetVersion, t.targetBuildNumber))

	return nil
}

func (t *GetAllRequiredPkgsTask) getAllRequiredPkgs() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	t.ExecuteLogf("The directory for this upgrade check task is %s", t.upgradeCheckTaskDir)
	if err = os.MkdirAll(t.upgradeCheckTaskDir, 0755); err != nil {
		return err
	}

	t.ExecuteLog("Confirm that all the required packages have been uploaded.")
	var pkgInfo oceanbase.UpgradePkgInfo
	arch := global.Architecture

	pkgInfo, err = obclusterService.GetUpgradePkgInfoByVersionAndRelease(constant.PKG_OBSHELL, t.targetVersion, t.targetBuildNumber, t.distribution, arch)
	if err != nil {
		return err
	}
	t.upgradePkgInfo = pkgInfo

	if err = t.CheckDiskFreeSpace(); err != nil {
		return
	}

	return t.downloadAllRequiredPkgs()
}

func (t *GetAllRequiredPkgsTask) CheckDiskFreeSpace() error {
	t.ExecuteLog("Check the remaining disk space.")
	t.ExecuteLogf("The directory being checked is %s", t.upgradeDir)
	var expectedSize uint64
	expectedSize += (t.upgradePkgInfo.Size + t.upgradePkgInfo.PayloadSize)
	expectedSize = (expectedSize) * uint64(confficient)
	t.ExecuteLogf("The required disk size is %d", expectedSize)
	diskInfo, err := system.GetDiskInfo(t.upgradeDir)
	if err != nil {
		return errors.Wrap(err, "failed to get disk info")
	}
	t.ExecuteLogf("The remaining disk size is %d", diskInfo.FreeSizeBytes)
	if diskInfo.FreeSizeBytes < expectedSize {
		return errors.Occur(errors.ErrEnvironmentDiskSpaceNotEnough, diskInfo.FreeSizeBytes, expectedSize)
	}
	return nil
}

type rpmPacakgeInstallInfo struct {
	RpmName           string
	RpmBuildVersion   string
	RpmDir            string
	RpmPkgPath        string
	RpmPkgExtractPath string
	RpmPkgHomepath    string
}

func (t *GetAllRequiredPkgsTask) downloadAllRequiredPkgs() (err error) {
	t.ExecuteLogf("Download all packages to %s", t.upgradeCheckTaskDir)
	pkgInfo := t.upgradePkgInfo
	rpmDir := GenerateUpgradeRpmDir(t.upgradeCheckTaskDir, pkgInfo.Version, pkgInfo.Architecture)
	if err := os.MkdirAll(rpmDir, 0755); err != nil {
		return err
	}
	rpmPkgPath := GenerateRpmPkgPath(rpmDir, pkgInfo.Name)
	rpmPkgExtractPath := GenerateRpmPkgExtractPath(rpmDir)
	rpmPkgPkgHomepath := GenerateRpmPkgHomepath(rpmDir)
	buildVersion := fmt.Sprintf("%s-%s", pkgInfo.Version, pkgInfo.Release)
	rpmPkgInfo := rpmPacakgeInstallInfo{
		RpmName:           pkgInfo.Name,
		RpmBuildVersion:   buildVersion,
		RpmDir:            rpmDir,
		RpmPkgPath:        rpmPkgPath,
		RpmPkgExtractPath: rpmPkgExtractPath,
		RpmPkgHomepath:    rpmPkgPkgHomepath,
	}
	t.GetContext().SetData(PARAM_UPGRADE_PKG_INSTALL_INFO, rpmPkgInfo)
	if err = obclusterService.DownloadUpgradePkgChunkInBatch(rpmPkgPath, pkgInfo.PkgId, pkgInfo.ChunkCount); err != nil {
		return err
	}
	t.ExecuteLogf("Downloaded pkg '%s' to '%s'", pkgInfo.Name, rpmPkgPath)

	return nil
}

func (t *GetAllRequiredPkgsTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if err = t.deleteAllRequiredPkgs(); err != nil {
		return
	}
	t.ExecuteLog("Successfully deleted.")
	return nil
}

func (t *GetAllRequiredPkgsTask) deleteAllRequiredPkgs() (err error) {
	t.ExecuteLog("Delete all previously downloaded packages.")
	if err = t.GetLocalDataWithValue(PARAM_UPGRADE_CHECK_TASK_DIR, &t.upgradeCheckTaskDir); err != nil {
		return err
	}
	return os.RemoveAll(t.upgradeCheckTaskDir)

}

func GenerateUpgradeRpmDir(upgradeCheckTaskDir, version, arch string) string {
	return path.Join(upgradeCheckTaskDir, arch, version)
}

func GenerateRpmPkgPath(rpmDir, rpmName string) string {
	return fmt.Sprintf("%s/%s.rpm", rpmDir, rpmName)
}

func GenerateRpmPkgHomepath(rpmDir string) string {
	return path.Join(rpmDir, OCEANBASE_HOMEPATH)
}

func GenerateRpmPkgExtractPath(rpmDir string) string {
	return path.Join(rpmDir, OCEANBASE_HOME)
}
