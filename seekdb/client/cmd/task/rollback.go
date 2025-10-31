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

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
	cmdlib "github.com/oceanbase/obshell/seekdb/client/lib/cmd"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	"github.com/oceanbase/obshell/seekdb/client/utils/api"
	"github.com/oceanbase/obshell/seekdb/client/utils/printer"
)

type TaskRollbackFlags struct {
	id          string
	skipConfirm bool
}

func newRollbackCmd() *cobra.Command {
	opts := &TaskRollbackFlags{}
	rollbackCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_ROLLBACK,
		Short:   "Rollback a failed task.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)
			return taskRollback(opts)
		}),
		Example: rollbackCmdExample(),
	})

	rollbackCmd.Flags().SortFlags = false
	rollbackCmd.VarsPs(&opts.id, []string{clientconst.FLAG_ID, clientconst.FLAG_ID_SH}, "", "Task ID.", true)
	rollbackCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt.", false)

	return rollbackCmd.Command
}

func taskRollback(flags *TaskRollbackFlags) error {
	id := flags.id
	if err := askConfirmForTaskOperation(id, task.ROLLBACK_STR); err != nil {
		return err
	}

	stdio.StartLoadingf("Get task %s detail", id)
	dag, err := api.GetDagDetail(id)
	if err != nil {
		stdio.LoadFailedf("Sorry! The task (ID: %s) does not exist.", id)
		return err
	}
	printer.PrintDagStruct(dag, false)

	dagHandler := api.NewDagHandler(dag)
	stdio.StartLoadingf("Try to rollback task %s", id)
	err = dagHandler.Rollback()
	if err != nil {
		stdio.LoadFailedf("Sorry! The rollback of task (ID: %s) has failed: %v", id, err)
		log.Errorf("task operation failed, err: %s", err)
	} else {
		stdio.LoadSuccessf("Congratulations! Task with ID %s has been successfully rolled back.", id)
	}
	return nil
}

func rollbackCmdExample() string {
	return `  obshell task rollback -i 11 --seekdb
  obshell task rollback -i 11 --port 2886 --seekdb`
}
