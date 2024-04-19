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

	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
)

type TaskCancelFlags struct {
	id          string
	skipConfirm bool
}

func newCancelCmd() *cobra.Command {
	opts := &TaskCancelFlags{}
	requiredFlags := []string{clientconst.FLAG_ID}
	cancelCmd := &cobra.Command{
		Use:     CMD_CANCEL,
		Aliases: []string{CMD_STOP},
		Short:   "Cancel a task.",
		PreRunE: cmdlib.ValidateArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)
			if err := taskCancel(opts); err != nil {
				stdio.Error(err.Error())
			}
		},
		Example: cancelCmdExample(),
	}

	cancelCmd.Flags().SortFlags = false
	cancelCmd.Flags().StringVarP(&opts.id, clientconst.FLAG_ID, clientconst.FLAG_ID_SH, "", "Task ID.")
	cancelCmd.Flags().BoolVarP(&opts.skipConfirm, clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH, false, "Skip the confirmation prompt.")

	cancelCmd.MarkFlagRequired(clientconst.FLAG_ID)

	cancelCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printer.PrintHelpFunc(cmd, requiredFlags)
	})
	return cancelCmd
}

func taskCancel(flags *TaskCancelFlags) error {
	id := flags.id
	if err := askConfirmForTaskOperation(id, task.CANCEL_STR); err != nil {
		return err
	}

	dag, err := api.GetDagDetail(id)
	if err != nil {
		return errors.Wrapf(err, "alling dag detail API failed (ID: %s)", id)
	}
	printer.PrintDagStruct(dag, false)

	dagHandler := api.NewDagHandler(dag)
	err = dagHandler.CancelDag()
	if err != nil {
		stdio.Failedf("Sorry! The cancellation of task (ID: %s) has failed: %v", id, err)
		log.Errorf("task operation failed, err: %s", err)
	} else {
		stdio.Successf("Congratulations! Task with ID %s has been successfully cancelled.", id)
	}
	return nil
}

func cancelCmdExample() string {
	return `  obshell task cancel -i 11 `
}
