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

package task

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/engine/task"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
)

type TaskCancelFlags struct {
	id          string
	skipConfirm bool
}

func newCancelCmd() *cobra.Command {
	opts := &TaskCancelFlags{}
	cancelCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_CANCEL,
		Aliases: []string{CMD_STOP},
		Short:   "Cancel a task.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)
			if err := taskCancel(opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: cancelCmdExample(),
	})

	cancelCmd.Flags().SortFlags = false
	cancelCmd.VarsPs(&opts.id, []string{clientconst.FLAG_ID, clientconst.FLAG_ID_SH}, "", "Task ID.", true)
	cancelCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt.", false)

	return cancelCmd.Command
}

func taskCancel(flags *TaskCancelFlags) error {
	id := flags.id
	if err := askConfirmForTaskOperation(id, task.CANCEL_STR); err != nil {
		return err
	}

	stdio.StartLoadingf("Get task %s detail", id)
	dag, err := api.GetDagDetail(id)
	if err != nil {
		stdio.LoadFailedf("Sorry! Failed to get the task (ID: %s) detail: %v", id, err)
		return err
	}
	stdio.StopLoading()
	printer.PrintDagStruct(dag, false)

	dagHandler := api.NewDagHandler(dag)
	stdio.StartLoadingf("Trying to cancel the task %s", id)
	err = dagHandler.CancelDag()
	if err != nil {
		stdio.LoadFailedf("Sorry! The cancellation of task (ID: %s) has failed: %v", id, err)
		log.Errorf("task operation failed, err: %s", err)
	} else {
		stdio.LoadSuccessf("Congratulations! Task with ID %s has been successfully cancelled.", id)
	}
	return nil
}

func cancelCmdExample() string {
	return `  obshell task cancel -i 11 `
}
