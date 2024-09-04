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
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
)

type AgentJoinFlags struct {
	server  string
	zone    string
	verbose bool
	ObserverConfigFlags
}

func newJoinCmd() *cobra.Command {
	opts := &AgentJoinFlags{}
	joinCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_JOIN,
		Short:   "Join the cluster by specifying the target node before cluster has been initialized.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			if err := agentJoin(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: joinCmdExample(),
	})

	joinCmd.Flags().SortFlags = false
	// Setup of required flags for 'join' command.
	joinCmd.VarsPs(&opts.server, []string{FLAG_SERVER_SH, FLAG_SERVER}, "", "The target server you intend to join. If the port is unspecified, it will be 2886.", true)
	joinCmd.VarsPs(&opts.zone, []string{FLAG_ZONE_SH, FLAG_ZONE}, "", "The zone in which you are located.", true)

	// Configuration of optional flags for more detailed setup.
	joinCmd.VarsPs(&opts.mysqlPort, []string{FLAG_MYSQL_PORT, FLAG_MYSQL_PORT_SH}, "", "The SQL service port for the current node.", false)
	joinCmd.VarsPs(&opts.rpcPort, []string{FLAG_RPC_PORT, FLAG_RPC_PORT_SH}, "", "The remote access port for intra-cluster communication.", false)
	joinCmd.VarsPs(&opts.dataDir, []string{FLAG_DATA_DIR, FLAG_DATA_DIR_SH}, "", "The directory for storing the observer's data.", false)
	joinCmd.VarsPs(&opts.redoDir, []string{FLAG_REDO_DIR, FLAG_REDO_DIR_SH}, "", "The directory for storing the observer's clogs.", false)
	joinCmd.VarsPs(&opts.logLevel, []string{FLAG_LOG_LEVEL, FLAG_LOG_LEVEL_SH}, "", "The log print level for the observer.", false)
	joinCmd.VarsPs(&opts.optStr, []string{FLAG_OPT_STR, FLAG_OPT_STR_SH}, "", "Additional parameters for the observer, use the format key=value for each configuration, separated by commas.", false)
	joinCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return joinCmd.Command
}

func agentJoin(flags *AgentJoinFlags) error {
	if err := checkFlagsForJoinCmd(flags); err != nil {
		return err
	}
	if err := checkStatus(); err != nil {
		return err
	}

	targetAgent, err := NewAgentByString(flags.server)
	if err != nil {
		return err
	}

	// config server
	params, err := buildObServerConfigParams(flags)
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

	if params.ObServerConfig != nil && len(params.ObServerConfig) > 0 {
		dag, err = api.CallApiAndPrintStage(constant.URI_API_V1+constant.URI_OBSERVER_GROUP+constant.URI_CONFIG, params)
		if err != nil {
			return err
		}
		log.Infof("[join] Set observer config with dag: %+v", dag)
	}

	return nil
}

func buildObServerConfigParams(flags *AgentJoinFlags) (obParams param.ObServerConfigParams, err error) {
	stdio.Verbose("Build observer config params")
	obParams.ObServerConfig = flags.parsedConfig
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
	stdio.Verbose("Check status for join")
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}
	stdio.Verbosef("My agent is %s", agentStatus.Agent.GetIdentity())
	if !agentStatus.Agent.IsSingleAgent() {
		return errors.New("The current node is already in a cluster and cannot join another cluster.")
	}
	stdio.Verbosef("My agent is under maintenance: %v", agentStatus.UnderMaintenance)
	if agentStatus.UnderMaintenance {
		return errors.New("The current node is under maintenance and cannot join the cluster.")
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
			arr := strings.Split(server, ":")
			if len(arr) != 3 {
				return false
			}
			if !isValidIp(arr[0]) || !isValidPort(arr[1]) || !isValidPort(arr[2]) {
				return false
			}
		}
	}
	return true
}

func isValidIp(ip string) bool {
	ipRegexp := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)$`)
	return ipRegexp.MatchString(ip)
}

func isValidPort(port string) bool {
	if port == "" {
		return true
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return p > 1024 && p < 65536
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

func checkFlagsForJoinCmd(flags *AgentJoinFlags) error {
	stdio.Verbose("Check flags for join command")
	return parseConfig(&flags.ObserverConfigFlags)
}

func checkServerConfigFlags(flags *ObserverConfigFlags) error {
	// Validate the MySQL port and RPC port.
	stdio.Verbose("Check observer config flags")
	stdio.Verbosef("Check mysql port: %s", flags.mysqlPort)
	if !isValidPort(flags.mysqlPort) {
		return errors.Errorf("Invalid port: %s. Port number should be in the range [1024, 65535].", flags.mysqlPort)
	}
	stdio.Verbosef("Check rpc port: %s", flags.rpcPort)
	if !isValidPort(flags.rpcPort) {
		return errors.Errorf("Invalid port: %s. Port number should be in the range [1024, 65535].", flags.rpcPort)
	}

	// Standardize and validate the log level.
	flags.logLevel = strings.ToUpper(flags.logLevel)
	stdio.Verbosef("Check log level: %s", flags.logLevel)
	if !isValidLogLevel(flags.logLevel) {
		return errors.Errorf("Invalid log level: %s. (support: %v)", flags.logLevel, LOGLEVEL)
	}

	// If provided, validate the format of the rs_list.
	if flags.rsList != "" {
		stdio.Verbose("Check rs_list is valid or not")
		if !isValidRsList(flags.rsList) {
			return errors.Errorf("Invaild rs_list format '%s'. Please use the format `--rs 'ip:rpc_port:mysql_port;ip:rpc_port:mysql_port'`", flags.rsList)
		}
	}

	// If provided, validate the cluster ID.
	if flags.clusterId != "" {
		stdio.Verbose("Check cluster id is valid or not")
		if _, err := strconv.Atoi(flags.clusterId); err != nil {
			return errors.Errorf("Invalid cluster id: %s", flags.clusterId)
		}
	}

	// Check the validity of the data directory path and redo log directory path.
	stdio.Verbosef("Check data directory: %s", flags.dataDir)
	if flags.dataDir != "" && utils.CheckPathValid(flags.dataDir) != nil {
		return errors.Errorf("Invalid data directory: %s", flags.dataDir)
	}
	stdio.Verbosef("Check redo directory: %s", flags.redoDir)
	if flags.redoDir != "" && utils.CheckPathValid(flags.redoDir) != nil {
		return errors.Errorf("Invalid redo directory: %s", flags.redoDir)
	}
	return nil
}

func parseConfig(flags *ObserverConfigFlags) error {
	stdio.Verbose("Parse observer config")
	config := stringToMap(flags.optStr)

	// Check if both mysql_porth and mysqlPort are set.
	for k, v := range constant.OB_CONFIG_COMPATIBLE_MAP {
		if val, ok := config[k]; ok {
			if val2, ok2 := config[v]; ok2 && val != val2 {
				return errors.Errorf("You cannot set both %s and %s, use %s instead.", k, v, k)
			}
			delete(config, v)
		} else if val, ok := config[v]; ok {
			config[k] = val
			delete(config, v)
		}
	}

	flagConfigs := map[string]string{
		constant.CONFIG_MYSQL_PORT:   flags.mysqlPort,
		constant.CONFIG_RPC_PORT:     flags.rpcPort,
		constant.CONFIG_DATA_DIR:     flags.dataDir,
		constant.CONFIG_REDO_DIR:     flags.redoDir,
		constant.CONFIG_LOG_LEVEL:    flags.logLevel,
		constant.CONFIG_CLUSTER_NAME: flags.clusterName,
		constant.CONFIG_RS_LIST:      flags.rsList,
		constant.CONFIG_CLUSTER_ID:   flags.clusterId,
		constant.CONFIG_ZONE:         flags.zone,
	}
	for k, v := range flagConfigs {
		if v != "" {
			if val, ok := config[k]; ok && v != val {
				return errors.Errorf("Duplicate observer config: %s", k)
			} else {
				config[k] = strings.TrimSpace(v)
			}
		}
	}

	// Perform validation checks on the flags to ensure all configurations are valid.
	if err := checkServerConfigFlags(flags); err != nil {
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
