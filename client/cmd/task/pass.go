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

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
)

type TaskPassFlags struct {
	id          string
	skipConfirm bool
}

func newPassCmd() *cobra.Command {
	opts := &TaskPassFlags{}
	passCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_PASS,
		Aliases: []string{CMD_SKIP},
		Short:   "Pass a failed task.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)
			return taskPass(opts)
		}),
		Example: passCmdExample(),
	})

	passCmd.Flags().SortFlags = false
	passCmd.VarsPs(&opts.id, []string{clientconst.FLAG_ID, clientconst.FLAG_ID_SH}, "", "Task ID.", true)
	passCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt.", false)

	return passCmd.Command
}

func taskPass(flags *TaskPassFlags) error {
	id := flags.id
	if err := askConfirmForTaskOperation(id, task.PASS_STR); err != nil {
		return err
	}

	stdio.StartLoadingf("Get task %s detail", id)
	dag, err := api.GetDagDetail(id)
	if err != nil {
		stdio.LoadFailedf("Sorry! The get of task (ID: %s) detail has failed: %v", id, err)
		return err
	}
	stdio.StopLoading()
	printer.PrintDagStruct(dag, false)

	dagHandler := api.NewDagHandler(dag)
	stdio.StartLoadingf("Trying to pass task %s", id)
	err = dagHandler.PassDag()
	if err != nil {
		stdio.LoadFailedf("Sorry! The pass of task (ID: %s) has failed: %v", id, err)
		log.Errorf("task operation failed, err: %s", err)
	} else {
		stdio.LoadSuccessf("Congratulations! Task with ID %s has been successfully passed.", id)
	}
	return nil
}

func passCmdExample() string {
	return `  obshell task pass -i 11 `
}
