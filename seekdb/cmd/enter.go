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

package cmd

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"

	agentcmd "github.com/oceanbase/obshell/seekdb/agent/cmd"
	"github.com/oceanbase/obshell/seekdb/agent/cmd/admin"
	"github.com/oceanbase/obshell/seekdb/agent/cmd/daemon"
	"github.com/oceanbase/obshell/seekdb/agent/cmd/server"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/client/cmd/agent"
	"github.com/oceanbase/obshell/seekdb/client/cmd/instance"
	"github.com/oceanbase/obshell/seekdb/client/cmd/task"
	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
)

func SeekdbMain() {
	runtime.GOMAXPROCS(1)
	cmds := newCmd()

	cmds.AddCommand(admin.NewAdminCmd())
	cmds.AddCommand(daemon.NewDaemonCmd())
	cmds.AddCommand(server.NewServerCmd())
	cmds.AddCommand(agentcmd.NewVersionCmd())
	cmds.AddCommand(agentcmd.NewInfoIpCmd())

	cmds.AddCommand(instance.NewSeekdbCmd())
	cmds.AddCommand(agent.NewAgentCmd())
	cmds.AddCommand(task.NewTaskCmd())

	var showDetailedVersion bool
	cmds.Flags().BoolVarP(&showDetailedVersion, agentcmd.CMD_VERSION, agentcmd.CMD_V, false, "Display version for obshell and exit")
	cmds.Flags().MarkHidden(agentcmd.CMD_VERSION)

	cmds.Run = func(cmd *cobra.Command, args []string) {
		if showDetailedVersion && cmd.Parent() == nil {
			agentcmd.HandleVersionFlag()
		} else {
			cmd.Help()
		}
	}

	var port int
	cmds.PersistentFlags().IntVar(&port, clientconst.FLAG_OBSHELL_PORT, 2886, "Specify the port of the obshell, or specify it by environment variable OBSHELL_PORT_FOR_SEEKDB")
	var isForSeekdb bool
	cmds.PersistentFlags().BoolVar(&isForSeekdb, "seekdb", true, "Specify if the command is for seekdb")
	var useIPv6 bool
	cmds.PersistentFlags().BoolVarP(&useIPv6, "use-ipv6", "6", false, "Specify if the command should use IPv6, only used for seekdb")

	agentcmd.PreHandler()

	if err := cmds.Execute(); err != nil {
		os.Exit(-1)
	}
}

func newCmd() *cobra.Command {
	cmd := command.NewCommand(&cobra.Command{
		Use:   constant.PROC_OBSHELL,
		Short: "obshell is a CLI for seekdb management.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
			DisableNoDescFlag: true,
		},
	})
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	return cmd.Command
}
