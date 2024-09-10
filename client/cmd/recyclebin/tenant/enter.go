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

package tenant

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/client/command"
)

const (
	// obshell recyclebin tenant
	CMD_TENANT = "tenant"

	// obshell recyclebin tenant purge
	CMD_PURGE = "purge"

	// obshell recyclebin tenant flashback
	CMD_FLASHBACK    = "flashback"
	FLAG_NEW_NAME_SH = "n"
	FLAG_NEW_NAME    = "new_name"

	// obshell recyclebin tenatn show
	CMD_SHOW = "show"
)

func NewTenantCmd() *cobra.Command {
	tenantCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_TENANT,
		Short: "Manage tenant in recyclebin.",
	})
	tenantCmd.AddCommand(newPurgeCmd())
	tenantCmd.AddCommand(newFlashbackCmd())
	tenantCmd.AddCommand(newShowCmd())
	return tenantCmd.Command
}
