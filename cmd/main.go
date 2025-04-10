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

package main

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"

	agentcmd "github.com/oceanbase/obshell/agent/cmd"
	"github.com/oceanbase/obshell/agent/cmd/admin"
	"github.com/oceanbase/obshell/agent/cmd/daemon"
	"github.com/oceanbase/obshell/agent/cmd/server"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/client/cmd/agent"
	"github.com/oceanbase/obshell/client/cmd/backup"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/cmd/pool"
	"github.com/oceanbase/obshell/client/cmd/recyclebin"
	"github.com/oceanbase/obshell/client/cmd/restore"
	"github.com/oceanbase/obshell/client/cmd/task"
	"github.com/oceanbase/obshell/client/cmd/tenant"
	"github.com/oceanbase/obshell/client/cmd/unit"
	"github.com/oceanbase/obshell/client/command"
)

func main() {
	runtime.GOMAXPROCS(1)
	cmds := newCmd()

	cmds.AddCommand(admin.NewAdminCmd())
	cmds.AddCommand(daemon.NewDaemonCmd())
	cmds.AddCommand(server.NewServerCmd())
	cmds.AddCommand(agentcmd.NewVersionCmd())
	cmds.AddCommand(agentcmd.NewInfoIpCmd())

	cmds.AddCommand(cluster.NewClusterCmd())
	cmds.AddCommand(agent.NewAgentCmd())
	cmds.AddCommand(task.NewTaskCmd())
	cmds.AddCommand(tenant.NewTenantCmd())
	cmds.AddCommand(unit.NewUnitCommand())
	cmds.AddCommand(pool.NewPoolCommand())
	cmds.AddCommand(recyclebin.NewRecyclebinCmd())
	cmds.AddCommand(backup.NewBackupCmd())
	cmds.AddCommand(restore.NewRestoreCmd())

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

	agentcmd.PreHandler()

	if err := cmds.Execute(); err != nil {
		os.Exit(-1)
	}
}

func newCmd() *cobra.Command {
	cmd := command.NewCommand(&cobra.Command{
		Use:   constant.PROC_OBSHELL,
		Short: "obshell is a CLI for OceanBase database management.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
			DisableNoDescFlag: true,
		},
	})
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	return cmd.Command
}
