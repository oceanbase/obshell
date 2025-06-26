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

package unit

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
)

const (
	// obshell unit create
	CMD_CREATE          = "create"
	FLAG_MEMORY_SIZE    = "memory_size"
	FLAG_MEMORY_SIZE_SH = "m"
	FLAG_MAX_CPU        = "max_cpu"
	FLAG_MAX_CPU_SH     = "c"
	FLAG_MIN_CPU        = "min_cpu"
	FLAG_LOG_DISK_SIZE  = "log_disk_size"
	FLAG_MAX_IOPS       = "max_iops"
	FLAG_MIN_IOPS       = "min_iops"

	// obshell unit drop
	CMD_DROP = "drop"

	// obshell unit show
	CMD_SHOW = "show"
)

func NewUnitCommand() *cobra.Command {
	unitCommand := command.NewCommand(&cobra.Command{
		Use:   clientconst.CMD_UNIT,
		Short: "Manage the unit config.",
		PersistentPreRunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			global.InitGlobalVariable()
			return cluster.CheckAndStartDaemon()
		}),
	})
	unitCommand.AddCommand(newCreateCmd())
	unitCommand.AddCommand(newDropCmd())
	unitCommand.AddCommand(newShowCmd())
	return unitCommand.Command
}
