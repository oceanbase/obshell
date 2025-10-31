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
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
	cmdlib "github.com/oceanbase/obshell/seekdb/client/lib/cmd"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	"github.com/oceanbase/obshell/seekdb/client/utils/api"
	"github.com/oceanbase/obshell/seekdb/client/utils/printer"
)

type TaskWatchFlags struct {
	id string
}

func newWatchCmd() *cobra.Command {
	opts := &TaskWatchFlags{}
	watchCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_WATCH,
		Short:   "Watch OceanBase SeekDB task running info.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetSilenceMode(false)
			return taskWatch(opts)
		}),
		Example: `  obshell task watch -i 11 --seekdb
  obshell task watch -i 11 --port 2886 --seekdb`,
	})

	watchCmd.Flags().SortFlags = false
	watchCmd.VarsPs(&opts.id, []string{clientconst.FLAG_ID, clientconst.FLAG_ID_SH}, "", "Task ID.", true)

	return watchCmd.Command
}

func taskWatch(flags *TaskWatchFlags) (err error) {
	stdio.StartLoadingf("Get task %s detail", flags.id)
	dag, err := api.GetDagDetail(flags.id)
	if err != nil {
		stdio.LoadErrorf("Failed to get task %s detail", flags.id)
		return err
	}
	stdio.StopLoading()
	if !dag.IsRunning() {
		printer.PrintDagStruct(dag, false)
		return nil
	}
	// Watch task detail.
	return api.NewDagHandler(dag).PrintDagStage()
}
