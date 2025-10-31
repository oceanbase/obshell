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

package instance

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
	cmdlib "github.com/oceanbase/obshell/seekdb/client/lib/cmd"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	"github.com/oceanbase/obshell/seekdb/client/utils/api"
	"github.com/oceanbase/obshell/seekdb/param"
)

type StopObserverFlags struct {
	stopBehaviorFlags
	id          string
	verbose     bool
	skipConfirm bool
}

type stopBehaviorFlags struct {
	terminate bool
}

func newStopCmd() *cobra.Command {
	opts := &StopObserverFlags{}
	stopCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_STOP,
		Short:   "Stop observer.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return observerStop(opts)
		}),
		Example: stopCmdExample(),
	})

	stopCmd.Flags().SortFlags = false

	stopCmd.VarsPs(&opts.terminate, []string{FLAG_TERMINATE, FLAG_TERMINATE_SH}, false, "Trigger a 'MINOR FREEZE' command before forcefully killing the observer with 'kill -9'.", false)
	stopCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	stopCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of stop operation", false)

	return stopCmd.Command
}

func observerStop(flags *StopObserverFlags) (err error) {
	if err = confirmStop(); err != nil {
		return
	}

	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return
	}

	if agentStatus.OBState != oceanbase.STATE_CONNECTION_AVAILABLE && flags.terminate {
		return errors.Occur(errors.ErrCliUsageError, "The current observer is not available, please don't use '-t'.")
	}

	if agentStatus.UnderMaintenance {
		if err = CheckAgentMaintenance(); err != nil {
			return err
		}
	}

	stdio.Verbosef("stop seekdb in %s", path.AgentDir())
	if err = callStopApi(flags); err != nil {
		return
	}
	return nil
}

func confirmStop() error {
	msg := "Please confirm if you need to stop observer."
	res, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrap(err, "ask for stop confirmation failed")
	}
	if !res {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}
	return nil
}

func callStopApi(flags *StopObserverFlags) (err error) {
	stdio.Verbosef("Calling stop API with flags: %+v", flags)

	param := &param.ObStopParam{
		Terminate: flags.terminate,
	}
	uri := constant.URI_OBSERVER_API_PREFIX + constant.URI_STOP
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
	if dag == nil {
		if uri == constant.URI_OBSERVER_API_PREFIX+constant.URI_STOP {
			stdio.Print("seekdb is already stopped.")
		}
		return nil
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
	msg := fmt.Sprintf("Sorry, task '%s' has failed. Due to the failed task, the observer is currently under maintenance.", dag.Name)
	stdio.Warn(msg)
	stdio.StartOrUpdateLoading("Please do not perform any actions. Attempting to automatically release the Maintenance state")

	maintainDag, err := api.GetAgentLastMaintenanceDag()
	if err != nil {
		return err
	}
	if maintainDag == nil {
		return nil
	}

	if err = autoFinishMaintainDag(maintainDag); err != nil {
		stdio.LoadFailed("Failed to automatically release the Maintenance state.")
		return err
	}
	stdio.LoadSuccess("Maintenance state released successfully.")
	return nil
}

func stopCmdExample() string {
	return `  obshell seekdb stop -t
  obshell seekdb stop -t --port 2886`
}
