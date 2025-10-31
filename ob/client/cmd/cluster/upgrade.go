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

package cluster

import (
	"fmt"
	"strings"

	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/ob"
	"github.com/oceanbase/obshell/ob/agent/lib/binary"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/client/utils/printer"
	rpmutil "github.com/oceanbase/obshell/ob/client/utils/rpm"
	"github.com/oceanbase/obshell/ob/param"
	"github.com/oceanbase/obshell/ob/utils"
)

var upgradeFlagUsage = fmt.Sprintf("Cluster upgrade mode: '%s' or '%s'", ob.PARAM_ROLLING_UPGRADE, ob.PARAM_STOP_SERVICE_UPGRADE)

type clusterUpgradeFlags struct {
	pkgDir      string
	version     string
	mode        string
	upgradeDir  string
	skipConfirm bool
	verbose     bool
}

func newUpgradeCmd() *cobra.Command {
	opts := &clusterUpgradeFlags{}
	upgradeCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_UPGRADE,
		Short:   "Upgrade the OceanBase cluster to the specified version.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return clusterUpgrade(opts)
		}),
		Example: upgradeCmdExample(),
	})

	upgradeCmd.Flags().SortFlags = false
	upgradeCmd.VarsPs(&opts.pkgDir, []string{FLAG_PKG_DIR, FLAG_PKG_DIR_SH}, "", "The directory where the package is located.", true)

	upgradeCmd.VarsPs(&opts.version, []string{FLAG_VERSION, FLAG_VERSION_SH}, "", "Target build version for the OceanBase upgrade", false)
	upgradeCmd.VarsPs(&opts.mode, []string{FLAG_MODE, FLAG_MODE_SH}, ob.PARAM_ROLLING_UPGRADE, upgradeFlagUsage, false)
	upgradeCmd.VarsPs(&opts.upgradeDir, []string{FLAG_UPGRADE_DIR, FLAG_UPGRADE_DIR_SH}, "", "Temporary directory used by upgrade tasks", false)
	upgradeCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	upgradeCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return upgradeCmd.Command
}

func CheckIdentityForUpgrade() error {
	stdio.Verbose("Checking my agent identity")
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return errors.Wrap(err, "failed to get my agent status")
	}
	stdio.Verbosef("My agent is %s", agentStatus.Agent.GetIdentity())
	if !agentStatus.Agent.IsClusterAgent() {
		return errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, agentStatus.Agent.String(), agentStatus.Agent.GetIdentity(), meta.CLUSTER_AGENT)
	}

	return nil
}

func clusterUpgrade(opts *clusterUpgradeFlags) (err error) {
	if err := CheckIdentityForUpgrade(); err != nil {
		return err
	}

	stdio.Verbose("Upgrading the OceanBase cluster to the specified version")
	stdio.Verbosef("The specified params is %#+v", opts)

	// check if the cluster is under maintenance
	isRunning, err := api.CheckOBMaintenance()
	if err != nil {
		return err
	}
	if !isRunning {
		return errors.Occur(errors.ErrObClusterUnderMaintenance)
	}

	if err := checkFlagsForUpgrade(opts); err != nil {
		return err
	}

	pkgs, err := getAllOBRpmsInDir(opts.pkgDir)
	if err != nil {
		return err
	}

	params, err := getUpgradeParams(opts, pkgs)
	if err != nil {
		return err
	}

	if err := uploadPkgsByNameInDir(params, opts.pkgDir, pkgs); err != nil {
		return err
	}

	if err := obUpgrade(params); err != nil {
		return err
	}

	return nil
}

func checkFlagsForUpgrade(opts *clusterUpgradeFlags) (err error) {
	stdio.Verbosef("Checking if %s is a valid directory.", opts.pkgDir)
	if err = utils.CheckPathExistAndValid(opts.pkgDir); err != nil {
		return err
	}

	if opts.upgradeDir != "" {
		stdio.Verbosef("Checking if %s is a valid directory.", opts.upgradeDir)
		if err = utils.CheckPathValid(opts.upgradeDir); err != nil {
			return err
		}
	}

	stdio.Verbosef("Checking if %s is a valid mode.", opts.mode)
	mode := strings.ToUpper(opts.mode)
	switch mode {
	case ob.PARAM_ROLLING_UPGRADE:
		stdio.Verbose("Checking if the number of zones is greater than 3.")
		obInfo, err := api.GetObInfo()
		if err != nil {
			return err
		}
		if len(obInfo.Config.ZoneConfig) < 3 {
			return errors.Occur(errors.ErrObUpgradeUnableToRollingUpgrade)
		}
	case ob.PARAM_STOP_SERVICE_UPGRADE:
	default:
		return errors.Occur(errors.ErrObUpgradeModeNotSupported, opts.mode)
	}
	return nil
}

func obUpgrade(params *param.ObUpgradeParam) (err error) {
	// This is a two-step process: upgrade check and upgrade
	uri := constant.URI_OB_API_PREFIX + constant.URI_UPGRADE + constant.URI_CHECK
	upgradeCheckParam := &param.UpgradeCheckParam{
		Version:    params.Version,
		Release:    params.Release,
		UpgradeDir: params.UpgradeDir,
	}
	dag, err := api.CallApiAndPrintStage(uri, upgradeCheckParam)
	if err != nil {
		return err
	}
	log.Info("upgrade check dag: ", dag)

	// This will call the upgrade API.
	uri = constant.URI_OB_API_PREFIX + constant.URI_UPGRADE
	dag, err = api.CallApi(uri, params)
	if err != nil {
		return err
	}
	dagHandler := api.NewDagHandler(dag)
	dagHandler.SetRetryTimes(600)
	dagHandler.SetForUpgrade()
	if err = dagHandler.PrintDagStage(); err != nil {
		return err
	}
	log.Info("upgrade dag: ", dag)
	return nil
}

func uploadPkgsByNameInDir(params *param.ObUpgradeParam, pkgDir string, pkgs map[string]*rpm.Package) (err error) {
	stdio.Verbose("Uploading OceanBase packages to the cluster")
	myOBVersion, _, _ := binary.GetMyOBVersion()
	return UploadPkgsByNameAndVersionInDir(pkgDir, pkgs, myOBVersion, params.Version, params.Release, false)
}

func UploadPkgsByNameAndVersionInDir(pkgDir string, pkgs map[string]*rpm.Package, myVersion, targetVersion, targetRelease string, onlyTarget bool) (err error) {
	for fileName, p := range pkgs {
		stdio.Verbosef("Checking if %s needs to be uploaded.", fileName)
		items := strings.Split(p.Release(), ".")
		currBV := fmt.Sprintf("%s-%s", p.Version(), items[0])

		items = strings.Split(targetRelease, ".")
		targetBV := fmt.Sprintf("%s-%s", targetVersion, items[0])

		if (onlyTarget && currBV == targetBV) ||
			(!onlyTarget && pkg.CompareVersion(currBV, targetBV) <= 0 && pkg.CompareVersion(currBV, myVersion) > 0) {
			if err := rpmutil.CallUploadPkgAndPrint(pkgDir, fileName); err != nil {
				return err
			}
			continue
		}

		stdio.Verbosef("%s does not need to be uploaded.", fileName)
	}
	return nil
}

func getUpgradeParams(opts *clusterUpgradeFlags, pkgs map[string]*rpm.Package) (params *param.ObUpgradeParam, err error) {
	targetBV, err := getTargetVersion(opts, pkgs)
	if err != nil {
		return nil, err
	}

	items := strings.Split(targetBV, "-")
	stdio.Verbosef("My dist is %s", constant.DIST)
	params = &param.ObUpgradeParam{
		UpgradeCheckParam: param.UpgradeCheckParam{
			Version:    items[0],
			Release:    fmt.Sprintf("%s%s", items[1], constant.DIST),
			UpgradeDir: opts.upgradeDir,
		},
		Mode: opts.mode,
	}
	log.Infof("upgrade params: %#+v", params)
	return params, nil
}

func getTargetVersion(opts *clusterUpgradeFlags, pkgs map[string]*rpm.Package) (targetBuildVersion string, err error) {
	stdio.Verbose("Getting target build version")
	targetBuildVersion = opts.version
	if opts.version == "" {
		targetBuildVersion, err = GetTargetBuildVersion(pkgs)
	} else if !strings.Contains(opts.version, "-") {
		targetBuildVersion, err = GetTargetBuildVersionByVersion(opts.version, pkgs)
	}
	if err != nil {
		return "", err
	}
	stdio.Verbosef("The target version is %s", targetBuildVersion)

	myOBVersion, _, err := binary.GetMyOBVersion()
	if err != nil {
		return "", err
	}
	stdio.Verbosef("My OceanBase version is %s", myOBVersion)

	msg := fmt.Sprintf("Please confirm if you need to upgrade cluster from to %s to %s", myOBVersion, targetBuildVersion)
	res, err := stdio.Confirm(msg)
	if err != nil {
		return "", errors.Wrap(err, "ask for upgrade confirmation failed")
	}
	if !res {
		return "", errors.Occur(errors.ErrCliOperationCancelled)
	}
	return targetBuildVersion, nil
}

func GetTargetBuildVersionByVersion(version string, pkgs map[string]*rpm.Package) (targetBuildVersion string, err error) {
	stdio.Verbosef("Getting target build version by '%s'", version)
	var release string
	for name, p := range pkgs {
		if p.Version() == version {
			items := strings.Split(p.Release(), ".")
			if pkg.CompareVersion(items[0], release) > 0 {
				release = items[0]
			}
			stdio.Verbosef("%s version is %s-%s", name, version, items[0])
		}
	}
	if release == "" {
		return "", errors.Occur(errors.ErrCliUpgradeNoValidTargetBuildVersionFound, version)
	}
	return fmt.Sprintf("%s-%s", version, release), nil
}

func GetTargetBuildVersion(pkgs map[string]*rpm.Package) (targetBuildVersion string, err error) {
	for name, p := range pkgs {
		items := strings.Split(p.Release(), ".")
		currentBV := fmt.Sprintf("%s-%s", p.Version(), items[0])
		if targetBuildVersion == "" {
			targetBuildVersion = currentBV
		} else if pkg.CompareVersion(targetBuildVersion, currentBV) < 0 {
			targetBuildVersion = currentBV
		}
		stdio.Verbosef("%s version is %s", name, currentBV)
	}
	if targetBuildVersion == "" {
		return "", errors.Occur(errors.ErrCommonUnexpected, "no valid version found") // should not happen
	}
	return targetBuildVersion, nil
}

var pkgNames = []string{constant.PKG_OCEANBASE_CE, constant.PKG_OCEANBASE_CE_LIBS}

func getAllOBRpmsInDir(pkgDir string) (rpmPkgs map[string]*rpm.Package, err error) {
	stdio.Printf("Getting all rpm packages in %s", pkgDir)
	clusterBasicInfo, err := api.GetObclusterSummary()
	if err != nil {
		return nil, err
	}

	if clusterBasicInfo.IsCommunityEdition {
		pkgNames = []string{constant.PKG_OCEANBASE_CE, constant.PKG_OCEANBASE_CE_LIBS}
	} else if clusterBasicInfo.IsStandalone {
		pkgNames = []string{constant.PKG_OCEANBASE_STANDALONE}
	} else {
		pkgNames = []string{constant.PKG_OCEANBASE}
	}

	rpmPkgs, err = rpmutil.GetAllRpmsInDirByNames(pkgDir, pkgNames)
	if err != nil {
		return nil, err
	}
	if len(rpmPkgs) == 0 {
		return nil, errors.Occur(errors.ErrCliUpgradePackageNotFoundInPath, strings.Join(pkgNames, ","), pkgDir)
	}
	printer.PrintPkgsTable(rpmPkgs)
	return rpmPkgs, nil
}

func upgradeCmdExample() string {
	return `  obshell cluster upgrade -d /home/oceanbase/upgrade/
  obshell cluster upgrade -d /home/oceanbase/upgrade/ -V 4.2.1.0-20231224224959 -m stopService`
}
