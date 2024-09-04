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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

var LOGLEVEL = []string{"DEBUG", "TRACE", "WDIAG", "EDIAG", "INFO", "WARN", "ERROR"}

type ObserverConfigFlags struct {
	mysqlPort    string
	rpcPort      string
	dataDir      string
	redoDir      string
	logLevel     string
	clusterName  string
	clusterId    string
	rsList       string
	zone         string
	optStr       string
	parsedConfig map[string]string
}

type ClusterScaleOutFlags struct {
	agent       string // tht address of any agent in the target cluster
	password    string
	skipConfirm bool
	verbose     bool
	ObserverConfigFlags
}

func NewScaleOutCmd() *cobra.Command {
	opts := &ClusterScaleOutFlags{}
	scaleOutCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SCALE_OUT,
		Short: "Add new observer to scale-out OceanBase cluster to improve performance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			ocsagentlog.SetDBLoggerLevel(ocsagentlog.Silent)
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			if err := clusterScaleOut(opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: scaleOutCmdExample(),
	})

	scaleOutCmd.Flags().SortFlags = false
	// Setup of required flags for 'scale-out' command.
	scaleOutCmd.VarsPs(&opts.agent, []string{FLAG_SERVER_SH, FLAG_SERVER}, "", "Any server in the cluster. If the port is unspecified, it will be 2886.", true)
	scaleOutCmd.VarsPs(&opts.zone, []string{FLAG_ZONE_SH, FLAG_ZONE}, "", "The zone in which you are located", true)

	// Configuration of optional flags for more detailed setup.
	scaleOutCmd.VarsPs(&opts.mysqlPort, []string{FLAG_MYSQL_PORT_SH, FLAG_MYSQL_PORT}, "", "The SQL service port for the current node.", false)
	scaleOutCmd.VarsPs(&opts.rpcPort, []string{FLAG_RPC_PORT_SH, FLAG_RPC_PORT}, "", "The remote access port for intra-cluster communication.", false)
	scaleOutCmd.VarsPs(&opts.dataDir, []string{FLAG_DATA_DIR_SH, FLAG_DATA_DIR}, "", "The directory for storing the observer's data.", false)
	scaleOutCmd.VarsPs(&opts.redoDir, []string{FLAG_REDO_DIR_SH, FLAG_REDO_DIR}, "", "The directory for storing the observer's clogs.", false)
	scaleOutCmd.VarsPs(&opts.logLevel, []string{FLAG_LOG_LEVEL_SH, FLAG_LOG_LEVEL}, "", "The log print level for the observer.", false)
	scaleOutCmd.VarsPs(&opts.optStr, []string{FLAG_OPT_STR_SH, FLAG_OPT_STR}, "", "Additional parameters for the observer, use the format key=value for each configuration, separated by commas.", false)
	scaleOutCmd.VarsPs(&opts.password, []string{FLAG_PASSWORD, FLAG_PASSWORD_ALIAS}, "", "Password for OceanBase root@sys user.", false)
	scaleOutCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	scaleOutCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return scaleOutCmd.Command
}

func clusterScaleOut(flags *ClusterScaleOutFlags) (err error) {
	if err := parseConfig(&flags.ObserverConfigFlags); err != nil {
		return err
	}

	targetAgentInfo, err := NewAgentByString(flags.agent)
	if err != nil {
		return err
	}

	pass, err := stdio.Confirm(fmt.Sprintf("Please confirm if you need to scale out current node into the cluster via %s.", flags.agent))
	if err != nil {
		return errors.New("ask for scale-out confirmation failed")
	}
	if !pass {
		return nil
	}
	meta.SetOceanbasePwd(flags.password)
	scaleOutReq, err := buildScaleOutParam(flags)
	if err != nil {
		return err
	}
	dag, err := callScaleOutApi(targetAgentInfo, scaleOutReq)
	if err != nil {
		return err
	}
	log.Infof("Scale out with dag: %+v", dag)

	return
}

func callScaleOutApi(agent meta.AgentInfoInterface, param interface{}) (*task.DagDetailDTO, error) {
	dag, err := api.CallApiViaTCP(agent, constant.URI_OB_API_PREFIX+constant.URI_SCALE_OUT, param)
	if err != nil {
		return nil, err
	}

	dagHandler := api.NewDagHandlerWithAgent(dag, agent)
	if err = dagHandler.PrintDagStage(); err != nil {
		return nil, err
	}
	return dag, nil
}

func buildScaleOutParam(flags *ClusterScaleOutFlags) (*param.ScaleOutParam, error) {
	myAgent, err := api.GetMyAgentInfo()
	if err != nil {
		return nil, err
	}

	stdio.Verbosef("My agent is %s", myAgent.GetIdentity())
	if !myAgent.IsSingleAgent() {
		return nil, errors.New("The current agent is not a single agent, please use the single agent to scale out")
	}

	stdio.Printf("Start to scale out observer with agent: %v", myAgent.AgentInfo.String())
	scaleOutReq := &param.ScaleOutParam{
		AgentInfo: myAgent.AgentInfo,
		Zone:      flags.zone,
		ObConfigs: flags.parsedConfig,
	}

	return scaleOutReq, nil
}

func scaleOutCmdExample() string {
	return `  obshell cluster scale-out -s 192.168.1.1:2886 -z zone1 --rp ****`
}

func NewAgentByString(str string) (*meta.AgentInfo, error) {
	stdio.Verbosef("Parse target agent info from string: %s", str)
	info := strings.Split(str, ":")
	if !isValidIp(info[0]) {
		return nil, errors.Errorf("Invalid ip address: %s", info[0])
	}
	//If the observer provides a port number, use the port number,
	//otherwise use the default port number 2886
	agent := &meta.AgentInfo{
		Ip:   info[0],
		Port: constant.DEFAULT_AGENT_PORT,
	}
	if len(info) > 1 {
		if info[1] == "" {
			return nil, errors.Errorf("Invalid server format: '%s:'", info[0])
		}
		port, err := strconv.Atoi(info[1])
		if err != nil || !isValidPort(info[1]) {
			return nil, errors.Errorf("Invalid port: %s. Port number should be in the range [1024, 65535].", info[1])
		}
		agent.Port = port
	}
	stdio.Verbosef("Parsed target agent info: %v", agent)
	return agent, nil
}
