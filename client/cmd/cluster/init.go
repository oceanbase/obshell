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

	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/param"
)

type ClusterInitFlags struct {
	password string
	verbose  bool
	ObserverConfigFlags
}

func newInitCmd() *cobra.Command {
	opts := &ClusterInitFlags{}
	initCmd := &cobra.Command{
		Use:   CMD_INIT,
		Short: "Perform one-time-only initialization of a OceanBase cluster.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdlib.ValidateArgs(cmd, args); err != nil {
				return err
			}
			if opts.password == "" {
				return errors.New("password is required")
			}
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			if err := clusterInit(opts); err != nil {
				stdio.Error(err.Error())
			}
		},
		Example: initCmdExample(),
	}

	initCmd.Flags().SortFlags = false
	// Setup of required flags for 'init' command.
	initCmd.Flags().StringVarP(&opts.clusterName, FLAG_CLUSTER_NAME, FLAG_CLUSTER_NAME_SH, "", "Set a name to verify the identity of OceanBase cluster.")
	initCmd.Flags().StringVar(&opts.password, FLAG_PASSWORD, "", "Password for OceanBase root@sys user.")
	initCmd.Flags().StringVar(&opts.password, FLAG_PASSWORD_ALIAS, "", "")
	initCmd.Flags().Lookup(FLAG_PASSWORD).Annotations = map[string][]string{
		printer.ANNOTATIONS_ALIAS: {FLAG_PASSWORD_ALIAS},
	}
	initCmd.Flags().MarkHidden(FLAG_PASSWORD_ALIAS)
	initCmd.MarkFlagRequired(FLAG_CLUSTER_NAME)

	// Configuration of optional flags for more detailed setup.
	initCmd.Flags().StringVarP(&opts.clusterId, FLAG_CLUSTER_ID, FLAG_CLUSTER_ID_SH, "", "Set a id to verify the identity of OceanBase cluster.")
	initCmd.Flags().StringVarP(&opts.mysqlPort, FLAG_MYSQL_PORT, FLAG_MYSQL_PORT_SH, "", "The SQL service port for the current node.")
	initCmd.Flags().StringVarP(&opts.rpcPort, FLAG_RPC_PORT, FLAG_RPC_PORT_SH, "", "The remote access port for intra-cluster communication.")
	initCmd.Flags().StringVarP(&opts.dataDir, FLAG_DATA_DIR, FLAG_DATA_DIR_SH, "", "The directory for storing the observer's data.")
	initCmd.Flags().StringVarP(&opts.redoDir, FLAG_REDO_DIR, FLAG_REDO_DIR_SH, "", "The directory for storing the observer's clogs.")
	initCmd.Flags().StringVarP(&opts.logLevel, FLAG_LOG_LEVEL, FLAG_LOG_LEVEL_SH, "", "The log print level for the observer.")
	initCmd.Flags().StringVarP(&opts.optStr, FLAG_OPT_STR, FLAG_OPT_STR_SH, "", "Additional parameters for the observer, use the format key=value for each configuration, separated by commas.")
	initCmd.Flags().StringVarP(&opts.password, FLAG_RS_LIST, "", "", "Root service list.")
	initCmd.Flags().StringVarP(&opts.password, FLAG_RS_LIST_ALIAS, "", "", "")
	// Configures root service list flag and its alias, hiding the alias from help display.
	initCmd.Flags().MarkHidden(FLAG_RS_LIST_ALIAS)
	initCmd.Flags().Lookup(FLAG_RS_LIST).Annotations = map[string][]string{
		printer.ANNOTATIONS_ALIAS: {FLAG_RS_LIST_ALIAS},
	}
	initCmd.Flags().BoolVarP(&opts.verbose, clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH, false, "Activate verbose output")

	initCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		printer.PrintUsageFunc(cmd)
		return nil
	})

	return initCmd
}

func clusterInit(flags *ClusterInitFlags) error {
	if err := checkFlagsForInitCmd(flags); err != nil {
		return err
	}
	if err := checkInitStatus(); err != nil {
		return err
	}
	if err := callObclusterConfig(flags); err != nil {
		return errors.Wrap(err, "set obcluster config failed")
	}
	if err := callObserverConfig(flags); err != nil {
		return errors.Wrap(err, "set observer config failed")
	}
	if err := callInit(); err != nil {
		return errors.Wrap(err, "init failed")
	}
	return nil
}

// buildObclusterConfigParams constructs an ObClusterConfigParams struct using provided flag values.
func buildObclusterConfigParams(flags *ClusterInitFlags) *param.ObClusterConfigParams {
	obclusterConfigParams := param.ObClusterConfigParams{
		ClusterName: &flags.clusterName,
		RootPwd:     &flags.password,
		RsList:      &flags.rsList,
	}
	if flags.clusterId != "" {
		// Convert cluster ID from string to integer safely, as it has been pre-validated.
		id, _ := strconv.Atoi(flags.clusterId)
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

func callObserverConfig(flags *ClusterInitFlags) error {
	obGlobalConfigParams := buildObGlobalConfigParams(flags)

	if obGlobalConfigParams.ObServerConfig != nil && len(obGlobalConfigParams.ObServerConfig) > 0 {
		dag, err := api.CallApiAndPrintStage(constant.URI_API_V1+constant.URI_OBSERVER_GROUP+constant.URI_CONFIG, obGlobalConfigParams)
		if err != nil {
			return err
		}
		stdio.Verbosef("Set observer global config with DAG %s", dag.GenericID)
	}
	return nil
}

func buildObGlobalConfigParams(flags *ClusterInitFlags) param.ObServerConfigParams {
	// Remove any configurations from flags.parsedConfig that are explicitly denied.
	for _, key := range ob.DeniedConfig {
		delete(flags.parsedConfig, key)
	}

	return param.ObServerConfigParams{
		ObServerConfig: flags.parsedConfig,
		Scope: param.Scope{
			Type: ob.SCOPE_GLOBAL,
		},
	}
}

func callInit() error {
	dag, err := api.CallApiAndPrintStage(constant.URI_OB_API_PREFIX+constant.URI_INIT, nil)
	if err != nil {
		return err
	}
	log.Infof("[init] Init with DAG: %v", dag)
	return nil
}

func checkFlagsForInitCmd(flags *ClusterInitFlags) error {
	return parseConfig(&flags.ObserverConfigFlags)
}

func checkInitStatus() error {
	stdio.Verbose("Check status for init")
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
