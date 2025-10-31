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

package restore

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/client/cmd/cluster"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
)

const (
	CMD_SHOW   = "show"
	CMD_CANCEL = "cancel"
)

func NewRestoreCmd() *cobra.Command {
	taskCmd := command.NewCommand(&cobra.Command{
		Use:   clientconst.CMD_RESTORE,
		Short: "Display and manage the OceanBase tenant restore.",
		PersistentPreRunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			defer stdio.StopLoading()
			global.InitGlobalVariable()
			return cluster.CheckAndStartDaemon()
		}),
	})
	taskCmd.AddCommand(newShowCmd())
	taskCmd.AddCommand(newCancelCmd())
	return taskCmd.Command
}
