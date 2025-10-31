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
	"strings"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/client/utils/printer"
)

type TaskShowFlags struct {
	id      string
	detail  bool
	verbose bool
}

func newShowCmd() *cobra.Command {
	opts := &TaskShowFlags{}
	showCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_SHOW,
		Short:   "Show OceanBase task info.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return taskShow(opts)
		}),
		Example: showCmdExample(),
	})

	showCmd.Flags().SortFlags = false
	showCmd.VarsPs(&opts.id, []string{clientconst.FLAG_ID, clientconst.FLAG_ID_SH}, "", "Task ID.", false)
	showCmd.VarsPs(&opts.detail, []string{clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH}, false, "Show detailed information about the task.", false)
	showCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output.", false)

	return showCmd.Command
}

func taskShow(flags *TaskShowFlags) (err error) {
	id := strings.TrimSpace(flags.id)
	if flags.detail && id == "" {
		return errors.Occur(errors.ErrCliUsageError, "Please specify the ID with '-i' to view detailed information.")
	}

	if id != "" {
		stdio.StartLoadingf("Get task %s detail", id)
		dag, err := api.GetDagDetail(id)
		if err != nil {
			stdio.LoadFailedf("Failed to get task %s detail", id)
			return err
		}
		stdio.StopLoading()
		printer.PrintDagStruct(dag, flags.detail)
		return nil
	}

	// Query all unfinished tasks and display them.
	stdio.StartLoading("Get all unfinished tasks")
	dags, err := api.GetAllUnfinishedDags()
	if err != nil {
		return err
	}
	stdio.StopLoading()
	if len(dags) == 0 {
		stdio.Info("No unfinished task found. If you want to show a specific task, please use '-i'")
		return nil
	}

	for _, dag := range dags {
		printer.PrintDagStruct(dag, false)
		stdio.Print("")
	}

	return nil
}

func showCmdExample() string {
	return `  obshell task show
  obshell task show -i 11 -d`
}
