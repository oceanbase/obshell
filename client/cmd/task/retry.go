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
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
)

type TaskRetryFlags struct {
	id          string
	skipConfirm bool
}

func newRetryCmd() *cobra.Command {
	opts := &TaskRetryFlags{}
	requiredFlags := []string{clientconst.FLAG_ID}
	retryCmd := &cobra.Command{
		Use:     CMD_RETRY,
		Aliases: []string{CMD_RERUN},
		Short:   "Retry a failed task.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)
			if err := taskRetry(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: retryCmdExample(),
	}

	retryCmd.Flags().SortFlags = false
	retryCmd.Flags().StringVarP(&opts.id, clientconst.FLAG_ID, clientconst.FLAG_ID_SH, "", "Task ID.")
	retryCmd.Flags().BoolVarP(&opts.skipConfirm, clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH, false, "Skip the confirmation prompt.")

	retryCmd.MarkFlagRequired(clientconst.FLAG_ID)

	retryCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printer.PrintHelpFunc(cmd, requiredFlags)
	})
	return retryCmd
}

func taskRetry(flags *TaskRetryFlags) error {
	id := flags.id
	if err := askConfirmForTaskOperation(id, task.RETRY_STR); err != nil {
		return err
	}

	dag, err := api.GetDagDetail(id)
	if err != nil {
		return err
	}
	printer.PrintDagStruct(dag, false)

	dagHandler := api.NewDagHandler(dag)
	err = dagHandler.Retry()
	if err != nil {
		stdio.Failedf("Sorry! The retry of task (ID: %s) has failed: %v", id, err)
		log.Errorf("task operation failed, err: %s", err)
	} else {
		stdio.Successf("Congratulations! Task with ID %s has been successfully retried.", id)
	}
	return nil
}

func retryCmdExample() string {
	return `  obshell task retry -i 11`
}
