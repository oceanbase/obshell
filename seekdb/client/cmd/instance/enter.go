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

package instance

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/seekdb/agent/log"
	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
)

const (
	// CMD_START represents the "start" command used to start observers.
	CMD_START = "start"
	// Flags for the "start" command.
	FLAG_FORCE_PASS = "force-pass"

	// CMD_RESTART represents the "restart" command used to restart observers.
	CMD_RESTART = "restart"

	// CMD_STOP represents the "stop" command used to stop observers.
	CMD_STOP = "stop"
	// Flags for the "stop" command.
	FLAG_TERMINATE    = "terminate"
	FLAG_TERMINATE_SH = "t"

	// CMD_UPGRADE represents the "upgrade" command for upgrading the cluster.
	CMD_UPGRADE = "upgrade"
	// Flags for the "upgrade" command.
	FLAG_PKG_DIR        = "pkg_directory"
	FLAG_PKG_DIR_SH     = "d"
	FLAG_VERSION        = "target_version"
	FLAG_VERSION_SH     = "V"
	FLAG_MODE           = "mode"
	FLAG_MODE_SH        = "m"
	FLAG_UPGRADE_DIR    = "tmp_directory"
	FLAG_UPGRADE_DIR_SH = "t"

	CMD_PARAMETER = "parameter"

	CMD_VARIABLE = "variable"

	// CMD_SHOW represents the "show" command used to display information about the cluster status.
	CMD_SHOW = "show"
)

const (
	DAG_START_OBSERVER   = "Start observer"   // should be the same as the DAG_START_OBSERVER in seekdb/agent/executor/observer/enter.go
	DAG_STOP_OBSERVER    = "Stop observer"    // should be the same as the DAG_STOP_OBSERVER in seekdb/agent/executor/observer/enter.go
	DAG_RESTART_OBSERVER = "Restart observer" // should be the same as the DAG_RESTART_OBSERVER in seekdb/agent/executor/observer/enter.go
)

func NewSeekdbCmd() *cobra.Command {
	seekdbCmd := command.NewCommand(&cobra.Command{
		Use:   clientconst.CMD_SEEKDB,
		Short: "Deploy and manage the Seekdb instance.",
		PersistentPreRunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			global.InitGlobalVariable()
			return nil
		}),
	})

	seekdbCmd.AddCommand(newRestartCmd())
	seekdbCmd.AddCommand(newStartCmd())
	seekdbCmd.AddCommand(newShowCmd())
	seekdbCmd.AddCommand(newStopCmd())
	return seekdbCmd.Command
}
