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
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type ClusterInitFlags struct {
	password     string
	verbose      bool
	importScript bool
	ObserverConfigFlags
}

func newInitCmd() *cobra.Command {
	opts := &ClusterInitFlags{}
	initCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_INIT,
		Short: "Perform one-time-only initialization of a OceanBase cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			if err := clusterInit(opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: initCmdExample(),
	})

	initCmd.Flags().SortFlags = false
	// Setup of required flags for 'init' command.
	initCmd.VarsPs(&opts.clusterName, []string{FLAG_CLUSTER_NAME, FLAG_CLUSTER_NAME_SH}, "", "Set a name to verify the identity of OceanBase cluster.", true)
	initCmd.VarsPs(&opts.password, []string{FLAG_PASSWORD, FLAG_PASSWORD_ALIAS}, "", "Password for OceanBase root@sys user.", true)

	// Configuration of optional flags for more detailed setup.
	initCmd.VarsPs(&opts.clusterId, []string{FLAG_CLUSTER_ID, FLAG_CLUSTER_ID_SH}, "", "Set a id to verify the identity of OceanBase cluster.", false)
	initCmd.VarsPs(&opts.mysqlPort, []string{FLAG_MYSQL_PORT, FLAG_MYSQL_PORT_SH}, "", "The SQL service port for the current node.", false)
	initCmd.VarsPs(&opts.rpcPort, []string{FLAG_RPC_PORT, FLAG_RPC_PORT_SH}, "", "The remote access port for intra-cluster communication.", false)
	initCmd.VarsPs(&opts.dataDir, []string{FLAG_DATA_DIR, FLAG_DATA_DIR_SH}, "", "The directory for storing the observer's data.", false)
	initCmd.VarsPs(&opts.redoDir, []string{FLAG_REDO_DIR, FLAG_REDO_DIR_SH}, "", "The directory for storing the observer's clogs.", false)
	initCmd.VarsPs(&opts.logLevel, []string{FLAG_LOG_LEVEL, FLAG_LOG_LEVEL_SH}, "", "The log print level for the observer.", false)
	initCmd.VarsPs(&opts.optStr, []string{FLAG_OPT_STR, FLAG_OPT_STR_SH}, "", "Additional parameters for the observer, use the format key=value for each configuration, separated by commas.", false)
	initCmd.VarsPs(&opts.rsList, []string{FLAG_RS_LIST, FLAG_RS_LIST_ALIAS}, "", "Root service list", false)
	initCmd.VarsPs(&opts.importScript, []string{FLAG_IMPORT_SCRIPT}, false, "Import the observer's scripts for sys tenant.", false)

	initCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return initCmd.Command
}

func clusterInit(flags *ClusterInitFlags) error {
	if err := parseObserverConfigFlags(&flags.ObserverConfigFlags); err != nil {
		return err
	}
	// check status
	stdio.StartLoading("Check agent status for init")
	if err := checkInitStatus(); err != nil {
		return err
	}
	stdio.StopLoading()
	// config obcluster
	if err := callObclusterConfig(flags); err != nil {
		return errors.Wrap(err, "set obcluster config failed")
	}
	if err := callObserverConfig(flags.parsedConfig); err != nil {
		return errors.Wrap(err, "set observer config failed")
	}
	if err := callInit(buildInitParams(flags)); err != nil {
		return errors.Wrap(err, "init failed")
	}
	return nil
}

func buildInitParams(flags *ClusterInitFlags) *param.ObInitParam {
	return &param.ObInitParam{
		ImportScript: flags.importScript,
	}
}

// buildObclusterConfigParams constructs an ObClusterConfigParams struct using provided flag values.
func buildObclusterConfigParams(flags *ClusterInitFlags) *param.ObClusterConfigParams {
	obclusterConfigParams := param.ObClusterConfigParams{
		ClusterName: &flags.clusterName,
		RootPwd:     &flags.password,
	}

	if rsList, ok := flags.parsedConfig[FLAG_RS_LIST]; ok {
		obclusterConfigParams.RsList = &rsList
	}

	if id, ok := flags.parsedConfig[FLAG_CLUSTER_ID]; ok {
		// Convert cluster ID from string to integer safely, as it has been pre-validated.
		id, _ := strconv.Atoi(id)
		obclusterConfigParams.ClusterId = &id
	} else {
		ts := int(time.Now().Unix())
		obclusterConfigParams.ClusterId = &ts
	}
	return &obclusterConfigParams
}

func callObclusterConfig(flags *ClusterInitFlags) error {
	obclusterConfigParams := buildObclusterConfigParams(flags)
	dag, err := api.CallApiAndPrintStage(constant.URI_API_V1+constant.URI_OBCLUSTER_GROUP+constant.URI_CONFIG, obclusterConfigParams)
	if err != nil {
		return err
	}
	stdio.Verbosef("Set obcluster config with DAG %s", dag.GenericID)
	return nil
}

func callObserverConfig(configs map[string]string) error {
	obGlobalConfigParams := buildObGlobalConfigParams(configs)

	if obGlobalConfigParams.ObServerConfig != nil && len(obGlobalConfigParams.ObServerConfig) > 0 {
		dag, err := api.CallApiAndPrintStage(constant.URI_API_V1+constant.URI_OBSERVER_GROUP+constant.URI_CONFIG, obGlobalConfigParams)
		if err != nil {
			return err
		}
		stdio.Verbosef("Set observer global config with DAG %s", dag.GenericID)
	}
	return nil
}

func buildObGlobalConfigParams(configs map[string]string) param.ObServerConfigParams {
	// Remove any configurations from flags.parsedConfig that are explicitly denied.
	for _, key := range ob.DeniedConfig {
		delete(configs, key)
	}

	return param.ObServerConfigParams{
		ObServerConfig: configs,
		Scope: param.Scope{
			Type: ob.SCOPE_GLOBAL,
		},
	}
}

func callInit(param *param.ObInitParam) error {
	dag, err := api.CallApiAndPrintStage(constant.URI_OB_API_PREFIX+constant.URI_INIT, param)
	if err != nil {
		return err
	}
	log.Infof("[init] Init with DAG: %v", dag)
	return nil
}

func checkInitStatus() error {
	stdio.Verbose("Get my agent status")
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}

	stdio.Verbosef("My agent is %s", agentStatus.Agent.GetIdentity())
	if !agentStatus.Agent.IsFollowerAgent() && !agentStatus.Agent.IsMasterAgent() {
		return errors.Errorf("%s can not init.", string(agentStatus.Agent.Identity))
	}

	stdio.Verbosef("My agent is under maintenance: %v", agentStatus.UnderMaintenance)
	if agentStatus.UnderMaintenance {
		return errors.New("The current node is under maintenance.")
	}
	return nil
}

func initCmdExample() string {
	return `  obshell cluster init -n ob-test --rp ****`
}
