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

package recyclebin

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/cmd/recyclebin/tenant"
	"github.com/oceanbase/obshell/client/command"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

const (
	// obshell recyclebin
	CMD_RECYCLEBIN = "recyclebin"
)

func NewRecyclebinCmd() *cobra.Command {
	recyclebinCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_RECYCLEBIN,
		Short: "Manage the recyclebin.",
		Args:  cobra.NoArgs,
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
	recyclebinCmd.AddCommand(tenant.NewTenantCmd())
	return recyclebinCmd.Command
}
