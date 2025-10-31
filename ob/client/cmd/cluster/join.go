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
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/ob"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/param"
	"github.com/oceanbase/obshell/ob/utils"
)

type AgentJoinFlags struct {
	server  string
	verbose bool
	ObserverConfigFlags
}

func newJoinCmd() *cobra.Command {
	opts := &AgentJoinFlags{}
	joinCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_JOIN,
		Short:   "Join the cluster by specifying the target node before cluster has been initialized.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			return agentJoin(cmd, opts)
		}),
		Example: joinCmdExample(),
	})

	joinCmd.Flags().SortFlags = false
	// Setup of required flags for 'join' command.
	joinCmd.VarsPs(&opts.server, []string{FLAG_SERVER_SH, FLAG_SERVER}, "", "The target server you intend to join. If the port is unspecified, it will be 2886.", true)
	joinCmd.VarsPs(&opts.zone, []string{FLAG_ZONE_SH, FLAG_ZONE}, "", "The zone in which you are located.", true)

	// Configuration of optional flags for more detailed setup.
	joinCmd.VarsPs(&opts.mysqlPort, []string{FLAG_MYSQL_PORT, FLAG_MYSQL_PORT_SH}, 0, "The SQL service port for the current node.", false)
	joinCmd.VarsPs(&opts.rpcPort, []string{FLAG_RPC_PORT, FLAG_RPC_PORT_SH}, 0, "The remote access port for intra-cluster communication.", false)
	joinCmd.VarsPs(&opts.dataDir, []string{FLAG_DATA_DIR, FLAG_DATA_DIR_SH}, "", "The directory for storing the observer's data.", false)
	joinCmd.VarsPs(&opts.redoDir, []string{FLAG_REDO_DIR, FLAG_REDO_DIR_SH}, "", "The directory for storing the observer's clogs.", false)
	joinCmd.VarsPs(&opts.logLevel, []string{FLAG_LOG_LEVEL, FLAG_LOG_LEVEL_SH}, "", "The log print level for the observer.", false)
	joinCmd.VarsPs(&opts.optStr, []string{FLAG_OPT_STR, FLAG_OPT_STR_SH}, "", "Additional parameters for the observer, use the format key=value for each configuration, separated by commas.", false)
	joinCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return joinCmd.Command
}

func agentJoin(cmd *cobra.Command, flags *AgentJoinFlags) error {
	if err := parseObserverConfigFlags(cmd, &flags.ObserverConfigFlags); err != nil {
		return err
	}
	// check status
	stdio.StartLoading("Check agent status for agent join")
	if err := checkStatus(); err != nil {
		return err
	}
	stdio.StopLoading()

	targetAgent, err := meta.ConvertAddressToAgentInfo(flags.server)
	if err != nil {
		return err
	}

	// Initiate the join process by calling the join API endpoint with the necessary parameters.
	joinParam := &param.JoinApiParam{
		AgentInfo: *targetAgent,
		ZoneName:  flags.zone,
	}
	dag, err := api.CallApiAndPrintStage(constant.URI_AGENT_API_PREFIX+constant.URI_JOIN, joinParam)
	if err != nil {
		return err
	}
	log.Infof("[join] Join cluster with dag: %+v", dag)

	// config server
	params, err := buildObServerConfigParams(flags.parsedConfig)
	if err != nil {
		return err
	}

	if len(params.ObServerConfig) > 0 {
		dag, err = api.CallApiAndPrintStage(constant.URI_API_V1+constant.URI_OBSERVER_GROUP+constant.URI_CONFIG, params)
		if err != nil {
			return err
		}
		log.Infof("[join] Set observer config with dag: %+v", dag)
	}

	return nil
}

func buildObServerConfigParams(configs map[string]string) (obParams param.ObServerConfigParams, err error) {
	stdio.Verbose("Build observer config params")
	// Remove any configurations from flags.parsedConfig that are explicitly denied.
	for _, key := range ob.DeniedConfig {
		delete(configs, key)
	}
	obParams.ObServerConfig = configs
	agentInfo, err := api.GetMyAgentInfo()
	if err != nil {
		return
	}
	stdio.Verbosef("My agent is %s", agentInfo.String())
	obParams.Scope = param.Scope{
		Type:   ob.SCOPE_SERVER,
		Target: []string{agentInfo.String()},
	}
	return
}

func checkStatus() error {
	stdio.Verbose("Get my agent status")
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}
	stdio.Verbosef("My agent is %s", agentStatus.Agent.GetIdentity())
	if !agentStatus.Agent.IsSingleAgent() {
		return errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, agentStatus.Agent.String(), agentStatus.Agent.GetIdentity(), meta.SINGLE)
	}
	stdio.Verbosef("My agent is under maintenance: %v", agentStatus.UnderMaintenance)
	if agentStatus.UnderMaintenance {
		return errors.Occur(errors.ErrAgentUnderMaintenance, agentStatus.Agent.String())
	}
	return nil
}

func joinCmdExample() string {
	return `  obshell cluster join -z zone1 -s 192.168.1.1
  obshell cluster join -z zone1 -s 192.168.1.1:2886`
}

func isValidRsList(rsList string) bool {
	stdio.Verbose("Check rs_list format is valid or not.")
	servers := strings.Split(rsList, ";")
	for _, server := range servers {
		if server != "" {
			if _, err := meta.ConvertAddressToAgentInfo(server); err != nil {
				return false
			}
		}
	}
	return true
}

func isValidLogLevel(level string) bool {
	if level == "" {
		return true
	}
	for _, v := range LOGLEVEL {
		if v == level {
			return true
		}
	}
	return false
}

func checkServerConfigFlags(config map[string]string) error {
	// Validate the MySQL port and RPC port.
	stdio.Verbose("Check whether the configs is valid")
	if mysqlPort, ok := config[constant.CONFIG_MYSQL_PORT]; ok {
		stdio.Verbosef("Check mysql port: %s", mysqlPort)
		if !utils.IsValidPort(mysqlPort) {
			return errors.Occur(errors.ErrCommonInvalidPort, mysqlPort)
		}
	}

	if rpcPort, ok := config[constant.CONFIG_RPC_PORT]; ok {
		stdio.Verbosef("Check rpc port: %s", rpcPort)
		if !utils.IsValidPort(rpcPort) {
			return errors.Occur(errors.ErrCommonInvalidPort, rpcPort)
		}
	}

	// Standardize and validate the log level.
	if logLevel, ok := config[constant.CONFIG_LOG_LEVEL]; ok {
		stdio.Verbosef("Check log level: %s", logLevel)
		config[constant.CONFIG_LOG_LEVEL] = strings.ToUpper(logLevel)
		if !isValidLogLevel(config[constant.CONFIG_LOG_LEVEL]) {
			return errors.Occurf(errors.ErrCliUsageError, "Invalid log level: %s. (support: %v)", logLevel, LOGLEVEL)
		}
	}

	// If provided, validate the format of the rs_list.
	if rsList, ok := config[constant.CONFIG_RS_LIST]; ok {
		stdio.Verbose("Check rs_list is valid or not")
		if !isValidRsList(rsList) {
			return errors.Occur(errors.ErrObParameterRsListInvalid, rsList, "Please use the format `--rs 'ip:rpc_port:mysql_port;ip:rpc_port:mysql_port'")
		}
	}

	// Check the validity of the data directory path and redo log directory path.
	if dataDir, ok := config[constant.CONFIG_DATA_DIR]; ok {
		stdio.Verbosef("Check data directory: %s", dataDir)
		if err := utils.CheckPathValid(dataDir); err != nil {
			return errors.Wrap(err, "Invalid data directory")
		}
	}

	if redoDir, ok := config[constant.CONFIG_REDO_DIR]; ok {
		stdio.Verbosef("Check redo directory: %s", redoDir)
		if err := utils.CheckPathValid(redoDir); err != nil {
			return errors.Wrap(err, "Invalid redo directory")
		}
	}
	return nil
}

func parseObserverConfigFlags(cmd *cobra.Command, flags *ObserverConfigFlags) error {
	stdio.Verbose("Parse observer config flags")
	config := stringToMap(flags.optStr)

	// Check if both mysql_porth and mysqlPort are set.
	for k, v := range constant.OB_CONFIG_COMPATIBLE_MAP {
		if val, ok := config[k]; ok {
			if val2, ok2 := config[v]; ok2 && val != val2 {
				return errors.Occurf(errors.ErrCliUsageError, "You cannot set both %s and %s in '%s', use %s instead.", k, v, flags.optStr, k)
			}
			delete(config, v)
		} else if val, ok := config[v]; ok {
			config[k] = val
			delete(config, v)
		}
	}

	flagConfigs := map[string]string{
		constant.CONFIG_DATA_DIR:     flags.dataDir,
		constant.CONFIG_REDO_DIR:     flags.redoDir,
		constant.CONFIG_LOG_LEVEL:    flags.logLevel,
		constant.CONFIG_CLUSTER_NAME: flags.clusterName,
		constant.CONFIG_RS_LIST:      flags.rsList,
		constant.CONFIG_ZONE:         flags.zone,
	}

	if cmd.Flags().Changed(FLAG_MYSQL_PORT) {
		flagConfigs[constant.CONFIG_MYSQL_PORT] = strconv.Itoa(flags.mysqlPort)
	}
	if cmd.Flags().Changed(FLAG_RPC_PORT) {
		flagConfigs[constant.CONFIG_RPC_PORT] = strconv.Itoa(flags.rpcPort)
	}
	if cmd.Flags().Changed(FLAG_CLUSTER_ID) {
		flagConfigs[constant.CONFIG_CLUSTER_ID] = strconv.Itoa(flags.clusterId)
	}

	for k, v := range flagConfigs {
		if v != "" {
			if val, ok := config[k]; ok && v != val {
				return errors.Occurf(errors.ErrCliUsageError, "Duplicate observer config: %s", k)
			} else {
				config[k] = strings.TrimSpace(v)
			}
		}
	}

	// Perform validation checks on the flags to ensure all configurations are valid.
	if err := checkServerConfigFlags(config); err != nil {
		return err
	}
	flags.parsedConfig = config

	stdio.Verbosef("Observer config: %v\n", config)
	return nil
}

func stringToMap(str string) map[string]string {
	m := make(map[string]string)
	if str == "" {
		return m
	}
	for _, kv := range strings.Split(str, ",") {
		kvPair := strings.Split(kv, "=")
		if len(kvPair) != 2 {
			// Warn about invalid key-value pairs and ignore them because observer will ignore them
			stdio.Warnf("Invalid observer config: %s", kv)
			continue
		}
		m[kvPair[0]] = kvPair[1]
	}
	return m
}
