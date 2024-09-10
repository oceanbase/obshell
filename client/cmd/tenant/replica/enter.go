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

package replica

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/client/command"
)

const (
	CMD_REPLICA = "replica"

	// obshell tenant create
	CMD_ADD           = "add"
	FLAG_ZONE         = "zone"
	FLAG_ZONE_SH      = "z"
	FLAG_UNIT_NUM     = "unit_num"
	FLAG_UNIT         = "unit"
	FLAG_UNIT_SH      = "u"
	FLAG_REPLICA_TYPE = "replica_type"

	// obshell tenant modify
	CMD_MODIFY        = "modify"
	FLAG_OLD_PASSWORD = "old_password"

	CMD_DELETE = "delete"
)

func NewReplicaCmd() *cobra.Command {
	tenantCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_REPLICA,
		Short: "Manage the replicas of tenant.",
	})
	tenantCmd.AddCommand(newAddCmd())
	tenantCmd.AddCommand(newDeleteCmd())
	tenantCmd.AddCommand(newModifyCmd())
	return tenantCmd.Command
}
