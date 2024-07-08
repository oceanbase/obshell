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
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)
			if err := taskRollback(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
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

	dag, err := api.GetDagDetail(id)
	if err != nil {
		return err
	}
	printer.PrintDagStruct(dag, false)

	dagHandler := api.NewDagHandler(dag)
	err = dagHandler.Rollback()
	if err != nil {
		stdio.Failedf("Sorry! The rollback of task (ID: %s) has failed: %v", id, err)
		log.Errorf("task operation failed, err: %s", err)
	} else {
		stdio.Successf("Congratulations! Task with ID %s has been successfully rolled back.", id)
	}
	return nil
}

func rollbackCmdExample() string {
	return `  obshell task rollback -i 11`
}
