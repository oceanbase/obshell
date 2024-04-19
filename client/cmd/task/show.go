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

	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
	"github.com/oceanbase/obshell/agent/config"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
)

type TaskShowFlags struct {
	id      string
	detail  bool
	verbose bool
}

func newShowCmd() *cobra.Command {
	opts := &TaskShowFlags{}
	showCmd := &cobra.Command{
		Use:     CMD_SHOW,
		Short:   "Show OceanBase task info.",
		PreRunE: cmdlib.ValidateArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			if err := taskShow(opts); err != nil {
				stdio.Error(err.Error())
			}
		},
		Example: showCmdExample(),
	}

	showCmd.Flags().SortFlags = false
	showCmd.Flags().StringVarP(&opts.id, clientconst.FLAG_ID, clientconst.FLAG_ID_SH, "", "Task ID.")
	showCmd.Flags().BoolVarP(&opts.detail, clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH, false, "Show detailed information about the task.")
	showCmd.Flags().BoolVarP(&opts.verbose, clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH, false, "Activate verbose output.")

	showCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printer.PrintHelpFunc(cmd, []string{})
	})
	return showCmd
}

func taskShow(flags *TaskShowFlags) (err error) {
	id := strings.TrimSpace(flags.id)
	if flags.detail && id == "" {
		stdio.Error("Please specify the ID with '-i' to view detailed information.")
		return nil
	}

	if id != "" {
		dag, err := api.GetDagDetail(id)
		if err != nil {
			return err
		}
		printer.PrintDagStruct(dag, flags.detail)
		return nil
	}

	// Query all unfinished tasks and display them.
	dags, err := api.GetAllUnfinishedDags()
	if err != nil {
		return err
	}
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
