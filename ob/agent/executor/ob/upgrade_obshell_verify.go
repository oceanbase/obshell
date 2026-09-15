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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
)

// verifyAndExtractObshellRpm rebuilds the verification metadata introduced for
// signed OBShell upgrades and then verifies the RPM before extracting it. The
// rebuild keeps upgrade tasks created by an older OBShell resumable without
// trusting either the RPM header or a previously extracted binary.
func verifyAndExtractObshellRpm(ctx *task.TaskContext, rpmPkgInfo *rpmPacakgeInstallInfo) error {
	expectedIdentity, err := restoreObshellRpmIdentity(ctx, rpmPkgInfo)
	if err != nil {
		return err
	}

	digest, err := pkg.InstallVerifiedObshellRpmInPlace(rpmPkgInfo.RpmPkgPath, expectedIdentity)
	if err != nil {
		return err
	}
	rpmPkgInfo.ObshellSHA256 = digest
	return nil
}

func restoreObshellRpmIdentity(ctx *task.TaskContext, rpmPkgInfo *rpmPacakgeInstallInfo) (pkg.RpmIdentity, error) {
	if rpmPkgInfo.RpmName != constant.PKG_OBSHELL {
		return pkg.RpmIdentity{}, errors.Occur(errors.ErrPackageNameMismatch, rpmPkgInfo.RpmName, constant.PKG_OBSHELL)
	}

	var version string
	if err := ctx.GetParamWithValue(PARAM_VERSION, &version); err != nil {
		return pkg.RpmIdentity{}, err
	}
	var releaseDistribution string
	if err := ctx.GetParamWithValue(PARAM_RELEASE_DISTRIBUTION, &releaseDistribution); err != nil {
		return pkg.RpmIdentity{}, err
	}
	if _, _, err := pkg.SplitRelease(releaseDistribution); err != nil {
		return pkg.RpmIdentity{}, err
	}

	expectedIdentity := pkg.RpmIdentity{
		Name:         constant.PKG_OBSHELL,
		Version:      version,
		Release:      releaseDistribution,
		Architecture: global.Architecture,
	}
	if err := matchOrRestoreObshellIdentityField("version", &rpmPkgInfo.RpmVersion, expectedIdentity.Version); err != nil {
		return pkg.RpmIdentity{}, err
	}
	if err := matchOrRestoreObshellIdentityField("release", &rpmPkgInfo.RpmRelease, expectedIdentity.Release); err != nil {
		return pkg.RpmIdentity{}, err
	}
	if err := matchOrRestoreObshellIdentityField("architecture", &rpmPkgInfo.RpmArchitecture, expectedIdentity.Architecture); err != nil {
		return pkg.RpmIdentity{}, err
	}
	return expectedIdentity, nil
}

func matchOrRestoreObshellIdentityField(fieldName string, persisted *string, expected string) error {
	if *persisted != "" && *persisted != expected {
		return errors.Occur(
			errors.ErrObPackageCorrupted,
			constant.PKG_OBSHELL,
			fmt.Sprintf("persisted RPM %s %q does not match upgrade request %q", fieldName, *persisted, expected),
		)
	}
	*persisted = expected
	return nil
}
