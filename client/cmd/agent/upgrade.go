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

package agent

import (
	"fmt"
	"strings"

	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/path"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
	rpmutil "github.com/oceanbase/obshell/client/utils/rpm"
	"github.com/oceanbase/obshell/param"
)

type agentUpgradeFlags struct {
	pkgDir      string
	version     string
	upgradeDir  string
	skipConfirm bool
	verbose     bool
}

func newUpgradeCmd() *cobra.Command {
	opts := &agentUpgradeFlags{}
	requiredFlags := []string{FLAG_PKG_DIR}
	upgradeCmd := &cobra.Command{
		Use:     CMD_UPGRADE,
		Short:   "Upgrade the OceanBase cluster to the specified version.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			global.InitGlobalVariable()
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			if err := cluster.CheckAndStartDaemon(true); err != nil {
				stdio.StopLoading()
				stdio.Error(err.Error())
				return err
			}
			if err := agentUpgrade(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: upgradeCmdExample(),
	}

	upgradeCmd.Flags().SortFlags = false
	upgradeCmd.Flags().StringVarP(&opts.pkgDir, FLAG_PKG_DIR, FLAG_PKG_DIR_SH, "", "The directory where the package is located")
	upgradeCmd.MarkFlagRequired(FLAG_PKG_DIR)

	upgradeCmd.Flags().StringVarP(&opts.version, FLAG_VERSION, FLAG_VERSION_SH, "", "Target build version for the obshell upgrade")
	upgradeCmd.Flags().StringVarP(&opts.upgradeDir, FLAG_UPGRADE_DIR, FLAG_UPGRADE_DIR_SH, "", "Temporary directory used by upgrade tasks")
	upgradeCmd.Flags().BoolVarP(&opts.skipConfirm, clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH, false, "Skip the confirmation prompt")
	upgradeCmd.Flags().BoolVarP(&opts.verbose, clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH, false, "Activate verbose output")

	upgradeCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printer.PrintHelpFunc(cmd, requiredFlags)
	})
	return upgradeCmd
}

func agentUpgrade(opts *agentUpgradeFlags) (err error) {
	if err := cluster.CheckIdentityForUpgrade(); err != nil {
		return err
	}
	stdio.Verbose("Upgrading the OBShell to the specified version")
	stdio.Verbosef("The specified params is %#+v", opts)

	stdio.Verbosef("Checking if %s is a valid directory.", opts.pkgDir)
	if err = path.CheckPathExistAndVaild(opts.pkgDir); err != nil {
		return err
	}

	// check if the cluster is under maintenance
	isRunning, err := api.CheckOBMaintenance()
	if err != nil {
		return err
	}
	if !isRunning {
		return errors.New("The cluster is under maintenance, unable to upgrade")
	}

	// get all the rpm packages in the specified directory
	pkgs, err := getAllObshellRpmsInDir(opts.pkgDir)
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

	if err := upgrade(params); err != nil {
		return err
	}
	return nil
}

func upgrade(params *param.UpgradeCheckParam) (err error) {
	// Perform the upgrade check by making an API call and printing its stages.
	uri := constant.URI_AGENT_API_PREFIX + constant.URI_UPGRADE + constant.URI_CHECK
	dag, err := api.CallApiAndPrintStage(uri, params)
	if err != nil {
		return err
	}
	log.Info("upgrade check dag: ", dag)

	// Proceed to the actual upgrade by making another API call.
	uri = constant.URI_AGENT_API_PREFIX + constant.URI_UPGRADE
	dag, err = api.CallApi(uri, params)
	if err != nil {
		return err
	}
	dagHandler := api.NewDagHandler(dag)
	dagHandler.SetRetryTimes(60)
	dagHandler.SetForUpgrade()
	if err = dagHandler.PrintDagStage(); err != nil {
		return err
	}
	return nil
}

func uploadPkgsByNameInDir(params *param.UpgradeCheckParam, pkgDir string, pkgs map[string]*rpm.Package) (err error) {
	stdio.Verbose("Uploading obshell packages to the cluster")
	return cluster.UploadPkgsByNameAndVersionInDir(pkgDir, pkgs, constant.VERSION_RELEASE, params.Version, params.Release, true)
}

func getUpgradeParams(opts *agentUpgradeFlags, pkgs map[string]*rpm.Package) (params *param.UpgradeCheckParam, err error) {
	// getUpgradeParams constructs and returns upgrade parameters based on the specified options
	// and the set of OBShell RPM packages available.
	targetBV, err := getTargetVersion(opts, pkgs)
	if err != nil {
		return nil, err
	}

	items := strings.Split(targetBV, "-")
	stdio.Verbosef("My dist is %s", constant.DIST)
	params = &param.UpgradeCheckParam{
		Version:    items[0],
		Release:    fmt.Sprintf("%s%s", items[1], constant.DIST),
		UpgradeDir: opts.upgradeDir,
	}
	log.Infof("upgrade params are %#+v", params)
	return params, nil
}

func getTargetVersion(opts *agentUpgradeFlags, pkgs map[string]*rpm.Package) (targetBuildVersion string, err error) {
	stdio.Verbose("Getting target build version")
	targetBuildVersion = opts.version
	if opts.version == "" {
		targetBuildVersion, err = cluster.GetTargetBuildVersion(pkgs)
	} else if !strings.Contains(opts.version, "-") {
		targetBuildVersion, err = cluster.GetTargetBuildVersionByVersion(opts.version, pkgs)
	}
	if err != nil {
		return "", err
	}
	stdio.Verbosef("The target version is %s", targetBuildVersion)
	stdio.Verbosef("My OBShell version is %s", constant.VERSION_RELEASE)

	msg := fmt.Sprintf("Please confirm if you need to upgrade OBShell from to %s to %s ", constant.VERSION_RELEASE, targetBuildVersion)
	res, err := stdio.Confirm(msg)
	if err != nil {
		return "", errors.Wrap(err, "ask for upgrade confirmation failed")
	}
	if !res {
		return "", errors.New("upgrade cancelled")
	}
	return targetBuildVersion, nil
}

var pkgNames = []string{constant.PKG_OBSHELL}

// getAllObshellRpmsInDir retrieves a map of OBShell rpm packages found within the specified directory.
// If no valid OBShell RPM packages are found, the function returns a descriptive error.
func getAllObshellRpmsInDir(pkgDir string) (rpmPkgs map[string]*rpm.Package, err error) {
	stdio.Printf("Getting all rpm packages in %s", pkgDir)
	rpmPkgs, err = rpmutil.GetAllRpmsInDirByNames(pkgDir, pkgNames)
	if err != nil {
		return nil, err
	}
	if len(rpmPkgs) == 0 {
		return nil, fmt.Errorf("no valid obshell package found in %s", pkgDir)
	}
	printer.PrintPkgsTable(rpmPkgs)
	return rpmPkgs, nil
}

func upgradeCmdExample() string {
	return `  obshell agent upgrade -d /home/oceanbase/upgrade/  
  obshell agent upgrade -d /home/oceanbase/upgrade/ -V 4.2.2.0-20231224224959`
}
