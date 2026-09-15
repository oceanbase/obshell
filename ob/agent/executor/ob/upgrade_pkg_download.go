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
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	agentpath "github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
)

const diskSpaceSafetyPercent uint64 = 10

type GetAllRequiredPkgsTask struct {
	task.Task
	obType              modelob.OBType
	needPkgNameList     []string
	upgradeDir          string
	upgradeRoute        []RouteNode
	targetBuildNumber   string
	targetVersion       string
	distribution        string
	upgradePkgInfo      []oceanbase.UpgradePkgInfo
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

	if err = t.getAllRequiredPkgs(); err != nil {
		return
	}
	return nil
}

func (t *GetAllRequiredPkgsTask) getParams() (err error) {
	onlyForAgent := t.GetContext().GetParam(PARAM_ONLY_FOR_AGENT)
	if onlyForAgent != nil {
		t.needPkgNameList = []string{constant.PKG_OBSHELL}
	} else {
		obType := t.GetContext().GetParam(PARAM_OB_TYPE) // compatible with previous version
		t.obType = modelob.OBTypeCommunity
		if obType != nil {
			if obType, ok := obType.(string); ok {
				t.obType = modelob.OBType(obType)
			}
		}
		t.needPkgNameList = constant.REQUIRE_UPGRADE_PKG_NAMES_MAP[t.obType]
	}

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

	t.upgradeRoute, err = getUpgradeRouteForTask(t.GetContext())
	if err != nil {
		return err
	}

	t.upgradePkgInfo = make([]oceanbase.UpgradePkgInfo, 0)
	t.ExecuteLogf("The required upgrade package is %v", t.needPkgNameList)
	for i, node := range t.upgradeRoute {
		t.ExecuteLogf("The %dth version in the upgrade route is: %s", i+1, node.BuildVersion)
	}

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
	for _, needPkgName := range t.needPkgNameList {
		for _, node := range t.upgradeRoute {
			log.Infof("get pkg '%s' info '%v'", needPkgName, node)
			var pkgInfo oceanbase.UpgradePkgInfo
			arch := global.Architecture
			if node.Release == RELEASE_NULL {
				pkgInfo, err = obclusterService.GetUpgradePkgInfoByVersion(needPkgName, node.Version, t.distribution, arch, node.DeprecatedInfo)
			} else {
				pkgInfo, err = obclusterService.GetUpgradePkgInfoByVersionAndRelease(needPkgName, node.Version, node.Release, t.distribution, arch)
			}
			if err != nil {
				return err
			}
			if pkgInfo.Name != needPkgName {
				return errors.Occur(errors.ErrPackageNameMismatch, pkgInfo.Name, needPkgName)
			}
			t.upgradePkgInfo = append(t.upgradePkgInfo, pkgInfo)
		}
	}

	if err = t.CheckDiskFreeSpace(); err != nil {
		return
	}

	return t.downloadAllRequiredPkgs()
}

func (t *GetAllRequiredPkgsTask) CheckDiskFreeSpace() error {
	t.ExecuteLog("Check the remaining disk space.")

	if !containsObshellPackage(t.upgradePkgInfo) {
		// Keep the existing OceanBase/standalone upgrade behavior unchanged.
		// The additional snapshot and destination-file reservations below are
		// required only by the verified standalone OBShell upgrade path.
		expectedSize := calculateLegacyDiskSpaceRequirement(t.upgradePkgInfo)
		t.ExecuteLogf("The directory being checked is %s", t.upgradeDir)
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

	_, upgradeFsID, err := system.GetFsId(t.upgradeDir)
	if err != nil {
		return errors.Wrap(err, "failed to get upgrade directory file system")
	}
	_, obshellBinFsID, err := system.GetFsId(agentpath.BinDir())
	if err != nil {
		return errors.Wrap(err, "failed to get OBShell bin directory file system")
	}
	obshellSharesUpgradeFileSystem := upgradeFsID == obshellBinFsID

	expectedUpgradeSize, expectedObshellBinSize := calculateDiskSpaceRequirements(t.upgradePkgInfo, obshellSharesUpgradeFileSystem)
	expectedUpgradeSize = addDiskSpaceSafetyMargin(expectedUpgradeSize)
	expectedObshellBinSize = addDiskSpaceSafetyMargin(expectedObshellBinSize)

	t.ExecuteLogf("The upgrade directory being checked is %s", t.upgradeDir)
	t.ExecuteLogf("The required upgrade directory disk size is %d", expectedUpgradeSize)
	diskInfo, err := system.GetDiskInfo(t.upgradeDir)
	if err != nil {
		return errors.Wrap(err, "failed to get disk info")
	}
	t.ExecuteLogf("The remaining upgrade directory disk size is %d", diskInfo.FreeSizeBytes)
	if diskInfo.FreeSizeBytes < expectedUpgradeSize {
		return errors.Occur(errors.ErrEnvironmentDiskSpaceNotEnough, diskInfo.FreeSizeBytes, expectedUpgradeSize)
	}

	if expectedObshellBinSize == 0 {
		return nil
	}
	t.ExecuteLogf("The OBShell bin directory being checked is %s", agentpath.BinDir())
	t.ExecuteLogf("The required OBShell bin directory disk size is %d", expectedObshellBinSize)
	obshellBinDiskInfo, err := system.GetDiskInfo(agentpath.BinDir())
	if err != nil {
		return errors.Wrap(err, "failed to get OBShell bin directory disk info")
	}
	t.ExecuteLogf("The remaining OBShell bin directory disk size is %d", obshellBinDiskInfo.FreeSizeBytes)
	if obshellBinDiskInfo.FreeSizeBytes < expectedObshellBinSize {
		return errors.Occur(errors.ErrEnvironmentDiskSpaceNotEnough, obshellBinDiskInfo.FreeSizeBytes, expectedObshellBinSize)
	}
	return nil
}

func calculateLegacyDiskSpaceRequirement(upgradePkgInfo []oceanbase.UpgradePkgInfo) (expectedSize uint64) {
	for _, info := range upgradePkgInfo {
		expectedSize += info.Size + info.PayloadSize
	}
	return expectedSize
}

func containsObshellPackage(upgradePkgInfo []oceanbase.UpgradePkgInfo) bool {
	for _, info := range upgradePkgInfo {
		if info.Name == constant.PKG_OBSHELL {
			return true
		}
	}
	return false
}

func calculateDiskSpaceRequirements(upgradePkgInfo []oceanbase.UpgradePkgInfo, obshellSharesUpgradeFileSystem bool) (upgradeDirSize, obshellBinDirSize uint64) {
	var obshellSnapshotSize uint64
	for _, info := range upgradePkgInfo {
		rpmDownloadSize := rpmDownloadSizeUpperBound(info)
		// All downloaded RPMs and their extracted files remain in upgradeDir.
		upgradeDirSize = saturatingAdd(upgradeDirSize, saturatingAdd(rpmDownloadSize, info.Size))
		if info.Name != constant.PKG_OBSHELL {
			continue
		}
		// Only one private RPM snapshot and one final OBShell temp binary exist
		// at a time. Size is the installed footprint, so it safely bounds the
		// OBShell binary that is copied beside the current executable.
		if rpmDownloadSize > obshellSnapshotSize {
			obshellSnapshotSize = rpmDownloadSize
		}
		if info.Size > obshellBinDirSize {
			obshellBinDirSize = info.Size
		}
	}

	if obshellSharesUpgradeFileSystem {
		// Snapshot creation and final binary replacement are different phases,
		// so the shared file system needs the larger peak, not their sum.
		if obshellBinDirSize > obshellSnapshotSize {
			obshellSnapshotSize = obshellBinDirSize
		}
		upgradeDirSize = saturatingAdd(upgradeDirSize, obshellSnapshotSize)
		return upgradeDirSize, 0
	}

	upgradeDirSize = saturatingAdd(upgradeDirSize, obshellSnapshotSize)
	return upgradeDirSize, obshellBinDirSize
}

func rpmDownloadSizeUpperBound(info oceanbase.UpgradePkgInfo) uint64 {
	if info.ChunkCount <= 0 {
		return info.PayloadSize
	}
	chunkCount := uint64(info.ChunkCount)
	maxUint64 := ^uint64(0)
	if chunkCount > maxUint64/constant.CHUNK_SIZE {
		return maxUint64
	}
	// DownloadUpgradePkgChunkInBatch writes at most ChunkCount chunks. This
	// bounds the complete RPM, including its lead and signature header, unlike
	// RPMSIGTAG_SIZE (PayloadSize), which covers only the bytes after them.
	downloadSize := chunkCount * constant.CHUNK_SIZE
	if downloadSize < info.PayloadSize {
		return info.PayloadSize
	}
	return downloadSize
}

func addDiskSpaceSafetyMargin(size uint64) uint64 {
	margin := size / 100 * diskSpaceSafetyPercent
	if remainder := size % 100; remainder != 0 {
		// Divide before multiplying to avoid overflow, then round the fractional
		// safety margin upward instead of silently truncating it.
		margin = saturatingAdd(margin, (remainder*diskSpaceSafetyPercent+99)/100)
	}
	return saturatingAdd(size, margin)
}

func saturatingAdd(left, right uint64) uint64 {
	maxUint64 := ^uint64(0)
	if left > maxUint64-right {
		return maxUint64
	}
	return left + right
}

type rpmPacakgeInstallInfo struct {
	RpmName           string
	RpmVersion        string
	RpmRelease        string
	RpmArchitecture   string
	RpmBuildVersion   string
	RpmDir            string
	RpmPkgPath        string
	RpmPkgExtractPath string
	RpmPkgHomepath    string
	ObshellSHA256     string
}

func (t *GetAllRequiredPkgsTask) downloadAllRequiredPkgs() (err error) {
	t.ExecuteLogf("Download all packages to %s", t.upgradeCheckTaskDir)
	for _, pkgInfo := range t.upgradePkgInfo {
		rpmDir := GenerateUpgradeRpmDir(t.upgradeCheckTaskDir, pkgInfo.Version, pkgInfo.Architecture)
		if err := os.MkdirAll(rpmDir, 0755); err != nil {
			return err
		}
		rpmPkgPath := GenerateRpmPkgPath(rpmDir, pkgInfo.Name)
		rpmPkgExtractPath := GenerateRpmPkgExtractPath(rpmDir)
		rpmPkgPkgHomepath := GenerateRpmPkgHomepath(rpmDir)
		buildVersion := fmt.Sprintf("%s-%s", pkgInfo.Version, pkgInfo.Release)
		version := getVersionInUpgradeRoute(buildVersion, t.upgradeRoute)
		rpmPkgInfo := rpmPacakgeInstallInfo{
			RpmName:           pkgInfo.Name,
			RpmVersion:        pkgInfo.Version,
			RpmRelease:        pkgInfo.ReleaseDistribution,
			RpmArchitecture:   pkgInfo.Architecture,
			RpmBuildVersion:   buildVersion,
			RpmDir:            rpmDir,
			RpmPkgPath:        rpmPkgPath,
			RpmPkgExtractPath: rpmPkgExtractPath,
			RpmPkgHomepath:    rpmPkgPkgHomepath,
		}

		if pkgInfo.Name == constant.PKG_OCEANBASE_CE_LIBS {
			t.SetLocalData(GenerateLibsBuildVersion(version), rpmPkgInfo)
		} else {
			t.SetLocalData(version, rpmPkgInfo)
		}
		if err = obclusterService.DownloadUpgradePkgChunkInBatch(rpmPkgPath, pkgInfo.PkgId, pkgInfo.ChunkCount); err != nil {
			return err
		}
		t.ExecuteLogf("Downloaded pkg '%s' to '%s'", pkgInfo.Name, rpmPkgPath)
	}
	return nil
}

func getVersionInUpgradeRoute(buildversion string, upgradeRoute []RouteNode) (version string) {
	for _, v := range upgradeRoute {
		if v.BuildVersion == buildversion {
			return buildversion
		}
	}
	return strings.Split(buildversion, "-")[0]
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
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

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
