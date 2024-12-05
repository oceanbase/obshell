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
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

type AgentRemoveFlags struct {
	server      string
	skipConfirm bool
	verbose     bool
}

func newRemoveCmd() *cobra.Command {
	opts := &AgentRemoveFlags{}
	removeCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_REMOVE,
		Short:   "Remove the specified the target node from cluster before cluster has been initialized.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			if err := agentRemove(opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: removeCmdExample(),
	})

	removeCmd.Flags().SortFlags = false
	// Setup of required flags for 'remove' command.
	removeCmd.VarsPs(&opts.server, []string{FLAG_SERVER, FLAG_SERVER_SH}, "", "The target server you intend to remove. If the port is unspecified, it will be 2886.", true)

	removeCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of removing", false)
	removeCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return removeCmd.Command
}

func agentRemove(flags *AgentRemoveFlags) error {
	targetAgent, err := NewAgentByString(flags.server)
	if err != nil {
		return err
	}

	pass, err := stdio.Confirmf("Please confirm if you need to remove %s", targetAgent.String())
	if err != nil {
		return errors.New("ask for remove confirmation failed")
	}
	if !pass {
		return nil
	}

	stdio.StartLoading("Check agent status for agent remove")
	if err := checkRemoveStatus(); err != nil {
		return err
	}
	stdio.StopLoading()

	dag := task.DagDetailDTO{}
	if err := api.CallApiWithMethod(http.POST, constant.URI_AGENT_API_PREFIX+constant.URI_REMOVE, targetAgent, &dag); err != nil {
		return err
	}
	if dag.GenericDTO == nil {
		stdio.Infof("%s is not in cluster", targetAgent.String())
		return nil
	}
	return api.NewDagHandler(&dag).PrintDagStage()
}

func checkRemoveStatus() error {
	stdio.Verbose("Get my agent status")
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}
	stdio.Verbosef("My agent is %s", agentStatus.Agent.GetIdentity())
	if !agentStatus.Agent.IsFollowerAgent() && !agentStatus.Agent.IsMasterAgent() {
		return errors.Errorf("%s cannot remove agent from the cluster.", string(agentStatus.Agent.Identity))
	}
	stdio.Verbosef("My agent is under maintenance %v", agentStatus.UnderMaintenance)
	if agentStatus.UnderMaintenance {
		return errors.New("The current node is under maintenance and cannot remove from the cluster.")
	}
	return nil
}

func removeCmdExample() string {
	return `  obshell cluster remove -s 192.168.1.1
  obshell cluster remove -s 192.168.1.1:2886`
}
