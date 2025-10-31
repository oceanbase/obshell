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

package obproxy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/ulikunitz/xz"

	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
	log "github.com/sirupsen/logrus"
)

var (
	confficient = 1.1
)

type GetObproxyPkgTask struct {
	task.Task
	targetBuildNumber string
	targetVersion     string
	distribution      string
	upgradeDir        string

	upgradePkgInfo sqlite.UpgradePkgInfo
}

func newGetObproxyPkgTask() *GetObproxyPkgTask {
	newTask := &GetObproxyPkgTask{
		Task: *task.NewSubTask(TASK_DOWNLOAD_RPM_FROM_SQLITE),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *GetObproxyPkgTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
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
	return nil
}

func (t *GetObproxyPkgTask) Execute() (err error) {
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

func (t *GetObproxyPkgTask) getAllRequiredPkgs() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	t.ExecuteLogf("The directory for this upgrade check task is %s", t.upgradeDir)
	if err = os.MkdirAll(t.upgradeDir, 0755); err != nil {
		return err
	}

	t.ExecuteLog("Confirm that all the required packages have been uploaded.")

	if t.upgradePkgInfo, err = agentService.GetUpgradePkgInfoByVersionAndRelease(constant.PKG_OBPROXY_CE, t.targetVersion, t.targetBuildNumber, t.distribution, global.Architecture); err != nil {
		return err
	}

	if err = t.CheckDiskFreeSpace(); err != nil {
		return
	}

	return t.downloadAllRequiredPkgs()
}

func (t *GetObproxyPkgTask) CheckDiskFreeSpace() error {
	t.ExecuteLog("Check the remaining disk space.")
	t.ExecuteLogf("The directory being checked is %s", t.upgradeDir)
	expectedSize := (t.upgradePkgInfo.Size + t.upgradePkgInfo.PayloadSize) * uint64(confficient)
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

func (t *GetObproxyPkgTask) downloadAllRequiredPkgs() (err error) {
	t.ExecuteLogf("Download all packages to %s", t.upgradeDir)
	pkgInfo := t.upgradePkgInfo
	rpmDir := GenerateUpgradeRpmDir(t.upgradeDir, pkgInfo.Version, pkgInfo.Architecture)
	if err := os.MkdirAll(rpmDir, 0755); err != nil {
		return err
	}
	rpmPkgPath := GenerateRpmPkgPath(rpmDir, pkgInfo.Name)
	if err = agentService.DownloadUpgradePkgChunkInBatch(rpmPkgPath, pkgInfo.PkgId, pkgInfo.ChunkCount); err != nil {
		return err
	}
	t.GetContext().SetParam(PARAM_OBPROXY_RPM_PKG_PATH, rpmPkgPath)
	t.ExecuteLogf("Downloaded pkg '%s' to '%s'", pkgInfo.Name, rpmPkgPath)
	return nil
}

func (t *GetObproxyPkgTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if err = t.deleteAllRequiredPkgs(); err != nil {
		return
	}
	t.ExecuteLog("Successfully deleted.")
	return nil
}

func (t *GetObproxyPkgTask) deleteAllRequiredPkgs() (err error) {
	t.ExecuteLog("Delete all previously downloaded packages.")
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return err
	}
	return os.RemoveAll(t.upgradeDir)

}

func GenerateUpgradeRpmDir(upgradeDir, version, arch string) string {
	return filepath.Join(upgradeDir, arch, version)
}

func GenerateRpmPkgPath(rpmDir, rpmName string) string {
	return fmt.Sprintf("%s/%s.rpm", rpmDir, rpmName)
}

type CheckObproxyPkgTask struct {
	task.Task
	pkgPath string
}

func newCheckObproxyPkgTask() *CheckObproxyPkgTask {
	newTask := &CheckObproxyPkgTask{
		Task: *task.NewSubTask(TASK_CHECK_OBPROXY_PKG),
	}
	newTask.
		SetCanContinue().
		SetCanRollback().
		SetCanRetry().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *CheckObproxyPkgTask) Execute() (err error) {
	if t.GetContext().GetParamWithValue(PARAM_OBPROXY_RPM_PKG_PATH, &t.pkgPath); err != nil {
		return err
	}
	if err = t.checkRequiredPkgs(); err != nil {
		return
	}
	return nil
}

func (t *CheckObproxyPkgTask) checkRequiredPkgs() (err error) {
	if err = t.checkUpgradePkgFromDb(t.pkgPath); err != nil {
		return err
	}
	t.ExecuteInfoLog("obproxy-ce package is checked successfully.")
	return nil
}

func (t *CheckObproxyPkgTask) checkUpgradePkgFromDb(filePath string) (err error) {
	input, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer input.Close()
	r := &upgradeRpmPkgInfo{
		rpmFile: input,
	}

	if err = r.CheckUpgradePkg(); err != nil {
		return err
	}
	return nil
}

type ReinstallObproxyBinTask struct {
	task.Task
	rpmPkgPath string
}

func newReinstallObproxyBinTask() *ReinstallObproxyBinTask {
	newTask := &ReinstallObproxyBinTask{
		Task: *task.NewSubTask(TASK_REINSTALL_OBPROXY_BIN),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanRollback().
		SetCanCancel().
		SetCanPass()
	return newTask
}

func (t *ReinstallObproxyBinTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_RPM_PKG_PATH, &t.rpmPkgPath); err != nil {
		return err
	}
	if err := t.installRpmPkgInPlace(t.rpmPkgPath); err != nil {
		return err
	}
	t.ExecuteLogf("Successfully installed %s", t.rpmPkgPath)
	return nil
}

func (t *ReinstallObproxyBinTask) installRpmPkgInPlace(rpmPkgPath string) (err error) {
	log.Infof("InstallRpmPkg: %s", rpmPkgPath)
	f, err := os.Open(rpmPkgPath)
	if err != nil {
		return
	}
	defer f.Close()

	rpmPkg, err := rpm.Read(f)
	if err != nil {
		return
	}

	if err = pkg.CheckCompressAndFormat(rpmPkg); err != nil {
		return
	}

	xzReader, err := xz.NewReader(f)
	if err != nil {
		return
	}
	cpioReader := cpio.NewReader(xzReader)

	for {
		hdr, err := cpioReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		m := hdr.Mode
		if m.IsRegular() && hdr.FileInfo().Name() == "obproxy" {
			// Remove obproxy
			if err := os.RemoveAll(path.ObproxyBinPath()); err != nil {
				return err
			}
			outFile, err := os.OpenFile(path.ObproxyBinPath(), os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return err
			}
			defer outFile.Close()
			log.Infof("Extracting %s", hdr.Name)
			if _, err := io.Copy(outFile, cpioReader); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *ReinstallObproxyBinTask) Rollback() (err error) {
	t.ExecuteLog("uninstall new obproxy")
	var upgradeDir string
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &upgradeDir); err != nil {
		return err
	}

	backupDir := filepath.Join(upgradeDir, "backup")

	dest := path.ObproxyBinPath()
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return system.CopyFile(fmt.Sprintf("%s/%s", backupDir, constant.PROC_OBPROXY), dest)
}

type BackupObproxyForUpgradeTask struct {
	task.Task
	upgradeDir string
	backupDir  string
}

func newBackupObproxyForUpgradeTask() *BackupObproxyForUpgradeTask {
	newTask := &BackupObproxyForUpgradeTask{
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

func (t *BackupObproxyForUpgradeTask) Execute() (err error) {
	if t.IsContinue() {
		t.ExecuteLog("The task is continuing.")
		if err = t.Rollback(); err != nil {
			return err
		}
	}

	if err = t.BackupObproxyForUpgrade(); err != nil {
		return
	}
	return nil
}

func (t *BackupObproxyForUpgradeTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return err
	}

	t.backupDir = filepath.Join(t.upgradeDir, "backup")
	return nil
}

func (t *BackupObproxyForUpgradeTask) BackupObproxyForUpgrade() error {
	t.ExecuteLog("Backup important files.")
	if err := t.getParams(); err != nil {
		return err
	}

	t.ExecuteLogf("The directory for backup is %s", t.backupDir)
	t.ExecuteLogf("Backup the bin directory %s", path.BinDir())
	if err := system.CopyDirs(path.ObproxyBinDir(), t.backupDir); err != nil {
		return err
	}
	return nil
}

func (t *BackupObproxyForUpgradeTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if err = t.deleteBackupDir(); err != nil {
		return err
	}
	t.ExecuteLog("Successfully deleted")
	return nil
}

func (t *BackupObproxyForUpgradeTask) deleteBackupDir() (err error) {
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
