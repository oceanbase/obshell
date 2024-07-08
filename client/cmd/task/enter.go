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

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

const (
	// show command
	CMD_SHOW = "show"

	// cancel command
	CMD_CANCEL = "cancel"
	CMD_STOP   = "stop"

	// rollback command
	CMD_ROLLBACK = "rollback"

	// retry command
	CMD_RETRY = "retry"
	CMD_RERUN = "rerun"

	// pass command
	CMD_PASS = "pass"
	CMD_SKIP = "skip"
)

func NewTaskCmd() *cobra.Command {
	taskCmd := command.NewCommand(&cobra.Command{
		Use:     clientconst.CMD_TASK,
		Aliases: []string{clientconst.CMD_UTIL},
		Short:   "Display and manage the OBShell task.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			defer stdio.StopLoading()
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			global.InitGlobalVariable()
			if err := cluster.CheckAndStartDaemon(); err != nil {
				stdio.StopLoading()
				stdio.Error(err.Error())
				return nil
			}
			return nil
		},
	})
	taskCmd.AddCommand(newShowCmd())
	taskCmd.AddCommand(newCancelCmd())
	taskCmd.AddCommand(newRollbackCmd())
	taskCmd.AddCommand(newRetryCmd())
	taskCmd.AddCommand(newPassCmd())
	return taskCmd.Command
}
