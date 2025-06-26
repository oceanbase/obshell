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

package pool

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
)

const (
	// obshell rp drop
	CMD_DROP = "drop"
	// obshell rp show
	CMD_SHOW = "show"
)

func NewPoolCommand() *cobra.Command {
	poolCommand := command.NewCommand(&cobra.Command{
		Use:   clientconst.CMD_POOL,
		Short: "Manage the resource poll.",
		PersistentPreRunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			global.InitGlobalVariable()
			return cluster.CheckAndStartDaemon()
		}),
	})
	poolCommand.AddCommand(newDropCmd())
	poolCommand.AddCommand(newShowCmd())
	return poolCommand.Command
}
