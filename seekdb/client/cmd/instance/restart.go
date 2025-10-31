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
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
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

type RestartObserverFlags struct {
	stopBehaviorFlags
	verbose     bool
	skipConfirm bool
}

func newRestartCmd() *cobra.Command {
	opts := &RestartObserverFlags{}
	restartCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_RESTART,
		Short:   "Restart observer.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return observerRestart(opts)
		}),
		Example: restartCmdExample(),
	})

	restartCmd.Flags().SortFlags = false

	restartCmd.VarsPs(&opts.terminate, []string{FLAG_TERMINATE, FLAG_TERMINATE_SH}, false, "Trigger a 'MINOR FREEZE' command before restart observer.", false)
	restartCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	restartCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of stop operation", false)

	return restartCmd.Command
}

func observerRestart(flags *RestartObserverFlags) (err error) {
	if err = confirmRestart(); err != nil {
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

	stdio.Verbosef("restart seekdb in %s", path.AgentDir())
	if err = callRestartApi(flags); err != nil {
		return
	}
	return nil
}

func confirmRestart() error {
	msg := "Please confirm if you need to restart observer."
	res, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrap(err, "ask for restart confirmation failed")
	}
	if !res {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}
	return nil
}

func callRestartApi(flags *RestartObserverFlags) (err error) {
	stdio.Verbosef("Calling stop API with flags: %+v", flags)

	param := &param.ObRestartParam{
		Terminate: flags.terminate,
	}
	uri := constant.URI_OBSERVER_API_PREFIX + constant.URI_RESTART
	if err = callEmerTypeApi(uri, param); err != nil {
		return
	}
	return nil

}

func restartCmdExample() string {
	return `  obshell seekdb restart -t
  obshell seekdb restart -t --port 2886`
}
