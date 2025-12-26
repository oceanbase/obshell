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
package inspection

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"

	obconstant "github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/inspection/constant"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/lib/sshutil"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
)

func checkObdiagAvailability() (obdiagBinPath string, needInstall bool, useWorkPath bool, err error) {
	workDirPath := path.ObdiagBinPath()
	available, _, err := checkObdiagVersion(workDirPath)
	if err != nil {
		return "", false, false, err
	}
	if available {
		return workDirPath, false, true, nil
	}

	systemPath, err := exec.LookPath(constant.BINARY_OBDIAG)
	if err == nil {
		available, _, err := checkObdiagVersion(systemPath)
		if err != nil {
			return "", false, false, err
		}
		if available {
			return systemPath, false, false, nil
		}
	}

	pkgInfo, err := obclusterService.GetLatestUpgradePkgInfo(obconstant.PKG_OCEANBASE_DIAGNOSTIC_TOOL, global.Architecture, obconstant.DIST)
	if err != nil {
		return "", false, false, errors.Occur(errors.ErrEnvironmentObdiagNotAvailable, constant.OBDIAG_VERSION_MIN)
	}
	if pkgInfo.Version < constant.OBDIAG_VERSION_MIN {
		return "", false, false, errors.Occur(errors.ErrObClusterInspectionObdiagVersionNotSupported, pkgInfo.Version, constant.OBDIAG_VERSION_MIN)
	}

	return "", true, true, nil
}

func checkSSHAccess() (usePasswordlessSSH bool, err error) {
	agents, err := agentService.GetAllAgentsDOFromOB()
	if err != nil {
		return false, err
	}

	if len(agents) <= 1 {
		return true, nil
	}

	currentUser, _ := user.Current()
	sshUser := "root"
	if currentUser != nil && currentUser.Username != "" {
		sshUser = currentUser.Username
	}

	// Check each agent: either passwordless SSH or credential is required
	allPasswordless := true
	for _, agent := range agents {
		if agent.Ip == meta.OCS_AGENT.GetIp() {
			continue
		}

		cmd := exec.Command("ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5", fmt.Sprintf("%s@%s", sshUser, agent.Ip), "echo", "test")
		if err := cmd.Run(); err != nil {
			// Passwordless SSH failed, check if credential exists
			allPasswordless = false
			if err := validateHostCredential(agent.Ip); err != nil {
				return false, err
			}
		}
	}
	// Return true only if all agents support passwordless SSH
	return allPasswordless, nil
}

func checkObdiagVersion(obdiagPath string) (bool, string, error) {
	if _, err := os.Stat(obdiagPath); err != nil {
		return false, "", nil
	}

	output, err := exec.Command(obdiagPath, "--version").Output()
	if err != nil {
		return false, "", errors.Wrap(err, "failed to execute obdiag --version")
	}

	versionStr := strings.TrimSpace(string(output))
	regex := regexp.MustCompile(`OceanBase Diagnostic Tool:\s*([\d\.]+)`)
	match := regex.FindStringSubmatch(versionStr)
	if len(match) == 2 {
		versionStr = match[1]
	}

	if pkg.CompareVersion(versionStr, constant.OBDIAG_VERSION_MIN) < 0 {
		return false, versionStr, nil
	}

	return true, versionStr, nil
}

func validateHostCredential(ip string) error {
	cred, err := credentialService.GetByHost(ip)
	if err != nil {
		return errors.Wrap(err, "get host credential failed")
	}
	if cred == nil {
		return errors.Occur(errors.ErrObClusterInspectionHostCredentialNotFound, ip)
	}

	var secret modelob.CredentialSecretData
	if err := secret.ParseFrom(cred.Secret); err != nil {
		return errors.WrapRetain(errors.ErrCredentialSecretFormatInvalid, err)
	}
	// Find matching target in Targets array
	matchedIP, port, err := findTargetForIP(secret.Targets, ip)
	if err != nil {
		return errors.WrapRetain(errors.ErrObClusterInspectionHostCredentialNotFound, err)
	}
	if matchedIP != ip {
		return errors.Occur(errors.ErrObClusterInspectionHostCredentialNotFound, ip)
	}

	passphrase, err := secure.DecryptCredentialPassphrase(secret.Passphrase)
	if err != nil {
		return errors.WrapRetain(errors.ErrCredentialDecryptFailed, err)
	}

	if err := sshutil.ValidateSSHConnection(ip, port, secret.Username, passphrase); err != nil {
		return errors.WrapRetain(errors.ErrCredentialSSHValidationFailed, err)
	}
	return nil
}
