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
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/ob"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/param"
)

type ClusterStopFlags struct {
	scopeFlags
	stopBehaviorFlags
	id          string
	verbose     bool
	skipConfirm bool
}

type stopBehaviorFlags struct {
	force     bool
	terminate bool
	immediate bool
}

func newStopCmd() *cobra.Command {
	opts := &ClusterStopFlags{}
	stopCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_STOP,
		Short:   "Stop observers within the specified range.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return clusterStop(opts)
		}),
		Example: stopCmdExample(),
	})

	stopCmd.Flags().SortFlags = false

	stopCmd.VarsPs(&opts.server, []string{FLAG_SERVER, FLAG_SERVER_SH}, "", "The operations address of the target server to stop.", false)
	stopCmd.VarsPs(&opts.zone, []string{FLAG_ZONE, FLAG_ZONE_SH}, "", "Stop all servers within the specified zone.", false)
	stopCmd.VarsPs(&opts.global, []string{FLAG_ALL, FLAG_ALL_SH}, false, "Stop all servers within the cluster.", false)

	stopCmd.VarsPs(&opts.force, []string{FLAG_FORCE, FLAG_FORCE_SH}, false, "Forcefully kill the observer using 'kill -9'", false)
	stopCmd.VarsPs(&opts.terminate, []string{FLAG_TERMINATE, FLAG_TERMINATE_SH}, false, "Trigger a 'MINOR FREEZE' command before forcefully killing the observer with 'kill -9'.", false)
	stopCmd.VarsPs(&opts.immediate, []string{FLAG_IMMEDIATE, FLAG_IMMEDIATE_SH}, false, "Trigger a 'STOP SERVER' command and will not forcefully kill the observer.", false)

	stopCmd.VarsPs(&opts.id, []string{FLAG_ID, FLAG_ID_SH}, "", "ID of the previous start/stop task. Separated by commas if multiple tasks are specified", false)
	stopCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	stopCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of stop operation", false)

	return stopCmd.Command
}

func clusterStop(flags *ClusterStopFlags) (err error) {
	if err = validateScopeFlags(&flags.scopeFlags); err != nil {
		return
	}

	if err = vaildStopFlags(&flags.stopBehaviorFlags); err != nil {
		return
	}

	if err = confirmStop(); err != nil {
		return
	}

	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return
	}

	if agentStatus.OBState != oceanbase.STATE_CONNECTION_AVAILABLE && !flags.force {
		return errors.Occur(errors.ErrCliUsageError, "The current observer is not available, please use '-f'.")
	}

	if flags.server == "" && flags.zone == "" && !flags.global {
		flags.server = agentStatus.Agent.String()
	}

	if err = CheckAllAgentMaintenance(); err != nil {
		return err
	}

	if err = callStopApi(flags); err != nil {
		return
	}
	return nil
}

func confirmStop() error {
	msg := "Please confirm if you need to stop servers, as it will leave the cluster in an unsafe state."
	res, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrap(err, "ask for stop confirmation failed")
	}
	if !res {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}
	return nil
}

func callStopApi(flags *ClusterStopFlags) (err error) {
	stdio.Verbosef("Calling stop API with flags: %+v", flags)

	param := &param.ObStopParam{
		Scope:             newScopeParam(&flags.scopeFlags),
		Force:             flags.force,
		Terminate:         flags.terminate || (!flags.force && !flags.immediate),
		ForcePassDagParam: *newForcePassIdParam(flags.id),
	}
	uri := constant.URI_OB_API_PREFIX + constant.URI_STOP
	if err = callEmerTypeApi(uri, param); err != nil {
		return
	}
	return nil

}

func callEmerTypeApi(uri string, param interface{}) (err error) {
	dag, err := api.CallApi(uri, param)
	if err != nil {
		return err
	}
	dagHandler := api.NewDagHandler(dag)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		stdio.Printf("\nReceived signal: %v", sig)
		stdio.Info("try to cancel the task, please wait...")
		if err := dagHandler.CancelDag(); err != nil {
			stdio.Warnf("Failed to cancel the task: %s", err.Error())
			os.Exit(1)
		}
	}()

	if err = dagHandler.PrintDagStage(); err != nil {
		return err
	}

	if dagHandler.Dag.IsSucceed() {
		return nil
	}

	return handleDagFailed(dagHandler.Dag)
}

func handleDagFailed(dag *task.DagDetailDTO) (err error) {
	msg := fmt.Sprintf("Sorry, task '%s' has failed. Due to the failed task, the cluster is currently under maintenance.", dag.Name)
	stdio.Warn(msg)
	stdio.StartOrUpdateLoading("Please do not perform any actions. Attempting to automatically release the Maintenance state")

	mainDags, _, err := api.GetAllMainAndMaintainDag()
	if err != nil {
		return err
	}

	if err = autoFinishMainDag(mainDags); err != nil {
		stdio.LoadFailed("Failed to automatically release the Maintenance state.")
		return err
	}
	stdio.LoadSuccess("Maintenance state released successfully.")
	return nil
}

func newScopeParam(flags *scopeFlags) param.Scope {
	stdio.Verbosef("Creating scope param with flags: %+v", flags)
	scopeParam := param.Scope{}
	switch getScopeType(flags) {
	case ob.SCOPE_SERVER:
		servers := strings.Split(strings.TrimSpace(flags.server), ",")
		scopeParam.Type = ob.SCOPE_SERVER
		scopeParam.Target = servers
	case ob.SCOPE_ZONE:
		zones := strings.Split(strings.TrimSpace(flags.zone), ",")
		scopeParam.Type = ob.SCOPE_ZONE
		scopeParam.Target = zones
	case ob.SCOPE_GLOBAL:
		scopeParam.Type = ob.SCOPE_GLOBAL
	}
	stdio.Verbosef("Scope param created: %#+v", scopeParam)
	return scopeParam
}

func vaildStopFlags(flags *stopBehaviorFlags) (err error) {
	stdio.Verbosef("Validating stop flags: %+v", flags)
	if flags.force && flags.terminate && flags.immediate {
		return errors.Occur(errors.ErrCliUsageError, "Only one of the flags -f, -t, -I can be specified")
	}
	if flags.force && flags.terminate {
		return errors.Occur(errors.ErrCliUsageError, "Only one of the flags -f, -t can be specified")
	}
	if flags.force && flags.immediate {
		return errors.Occur(errors.ErrCliUsageError, "Only one of the flags -f, -I can be specified")
	}
	if flags.terminate && flags.immediate {
		return errors.Occur(errors.ErrCliUsageError, "Only one of the flags -t, -I can be specified")
	}
	return nil
}

func stopCmdExample() string {
	return `  obshell cluster stop -s 192.168.1.1:2886 -f
  obshell cluster stop -z zone1,zone2 -t
  obshell cluster stop -z zone1 -I
  obshell cluster stop -a -f`
}
