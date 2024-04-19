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
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/system"
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
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

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
	keys, err := getKeyForPkgInfoMap(t.GetContext())
	if err != nil {
		return err
	}
	for _, key := range keys {
		var rpmPkgInfo rpmPacakgeInstallInfo
		if err = t.GetLocalDataWithValue(key, &rpmPkgInfo); err != nil {
			t.ExecuteErrorLogf("get local data failed, key: %s, err: %s", key, err.Error())
			success = false
			continue
		}
		t.ExecuteLogf("Unpack '%s'", rpmPkgInfo.RpmPkgPath)
		if err = installRpmPkgInPlace(rpmPkgInfo.RpmPkgPath); err != nil {
			success = false
			continue
		}
		t.ExecuteLogf("Successfully installed %s", rpmPkgInfo.RpmPkgPath)

		// Only check the observer bin when the package is oceanbase-ce and not only for agent.
		if rpmPkgInfo.RpmName == constant.PKG_OCEANBASE_CE && t.GetContext().GetParam(PARAM_ONLY_FOR_AGENT) == nil {
			if err = t.checkObserverBinAvailable(rpmPkgInfo); err != nil {
				t.ExecuteErrorLogf("check observer bin failed, err: %s", err.Error())
				success = false
			}
			// If current pkg version is the target version, then get the agent version as the agent's target version.
			if rpmPkgInfo.RpmBuildVersion == t.targetBuildVersion {
				if err = t.getAgentVersion(&rpmPkgInfo); err != nil {
					t.ExecuteErrorLogf("get agent version failed, err: %s", err.Error())
					success = false
				}
			}
		}
	}
	if !success {
		return errors.New("failed to unpack and check all required pkgs")
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

func installRpmPkgInPlace(path string) (err error) {
	log.Infof("InstallRpmPkg: %s", path)
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	pkg, err := rpm.Read(f)
	if err != nil {
		return
	}
	if err = checkCompressAndFormat(pkg); err != nil {
		return
	}
	xzReader, err := xz.NewReader(f)
	if err != nil {
		return
	}
	installPath := filepath.Dir(path)
	cpioReader := cpio.NewReader(xzReader)
	for {
		hdr, err := cpioReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !hdr.Mode.IsRegular() {
			continue
		}
		if dirName := filepath.Dir(hdr.Name); dirName != "" {
			dest := filepath.Join(installPath, dirName)
			if err := os.MkdirAll(dest, 0755); err != nil {
				log.WithError(err).Error("mkdir failed")
				return err
			}
		}
		dest := filepath.Join(installPath, hdr.Name)
		if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			log.WithError(err).Error("mkdir failed")
			return err
		}
		outFile, err := os.Create(dest)
		if err != nil {
			return err
		}
		log.Infof("Extracting %s", hdr.Name)
		if _, err := io.Copy(outFile, cpioReader); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}
	return nil
}

func (t *InstallAllRequiredPkgsTask) checkObserverBinAvailable(pkgInfo rpmPacakgeInstallInfo) (err error) {
	t.ExecuteLog("Check if the observer binary is available.")
	observerBinPath := path.Join(pkgInfo.RpmPkgHomepath, constant.DIR_BIN, constant.PROC_OBSERVER)
	if err = os.Chmod(observerBinPath, 0755); err != nil {
		return
	}
	bash := fmt.Sprintf("export LD_LIBRARY_PATH='%s/lib'; %s -V", pkgInfo.RpmPkgHomepath, observerBinPath)
	t.ExecuteLogf("The test command is %s", bash)
	cmd := exec.Command("/bin/bash", "-c", bash)
	if stderr, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, string(stderr))
	}
	t.ExecuteLog("Successfully checked.")
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
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

	t.ExecuteLog("Delete all previously installed package files.")
	keys, err := getKeyForPkgInfoMap(t.GetContext())
	if err != nil {
		return err
	}
	success := true
	for _, key := range keys {
		var rpmPkgInfo rpmPacakgeInstallInfo
		if err = t.GetLocalDataWithValue(key, &rpmPkgInfo); err != nil {
			t.ExecuteErrorLogf("get local data failed, key: %s, err: %s", key, err.Error())
			success = false
			continue
		}
		if rpmPkgInfo.RpmPkgExtractPath != "" {
			if err = os.RemoveAll(rpmPkgInfo.RpmPkgExtractPath); err != nil {
				success = false
				continue
			}
		}
	}
	if !success {
		return errors.New("uninstall all required pkgs failed")
	}
	return nil
}
