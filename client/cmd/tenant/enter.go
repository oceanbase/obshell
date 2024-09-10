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

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/cmd/tenant/parameter"
	"github.com/oceanbase/obshell/client/cmd/tenant/replica"
	"github.com/oceanbase/obshell/client/cmd/tenant/variable"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

const (
	// obshell tenant create
	CMD_CREATE           = "create"
	FLAG_ZONE            = "zone"
	FLAG_ZONE_SH         = "z"
	FLAG_MODE            = "mode"
	FLAG_PRIMARY_ZONE    = "primary_zone"
	FLAG_PRIMARY_ZONE_SH = "p"
	FLAG_UNIT_NUM        = "unit_num"
	FLAG_UNIT_NUM_SH     = "n"
	FLAG_UNIT            = "unit"
	FLAG_UNIT_SH         = "u"
	FLAG_REPLICA_TYPE    = "replica_type"
	FLAG_CHARSET         = "charset"
	FLAG_COLLATE         = "collate"
	FLAG_INFO            = "info"
	FLAG_READ_ONLY       = "read_only"
	FLAG_PARAMETERS      = "parameters"
	FLAG_VARIABLES       = "variables"
	FLAG_SCENARIO        = "scenario"
	FLAG_WHITELIST       = "whitelist"
	FLAG_ROOT_PASSWORD   = "root_password"

	// obshell tenant restore
	CMD_RESTORE                  = "restore"
	FLAG_TENANT_NAME             = "tenant_name"
	FLAG_TENANT_NAME_SH          = "t"
	FLAG_ARCHIVE_LOG_URI         = "archive_log_uri"
	FLAG_ARCHIVE_LOG_URI_SH      = "a"
	FLAG_DATA_BACKUP_URI         = "data_backup_uri"
	FLAG_DATA_BACKUP_URI_SH      = "d"
	FLAG_UNIT_CONFIG_NAME        = "unit_config_name"
	FLAG_UNIT_CONFIG_NAME_SH     = "u"
	FLAG_TIMESTAMP               = "timestamp"
	FLAG_TIMESTAMP_SH            = "T"
	FLAG_SCN                     = "scn"
	FLAG_SCN_SH                  = "S"
	FLAG_HA_HIGH_THREAD_SCORE    = "ha_high_thread_score"
	FLAG_HA_HIGH_THREAD_SCORE_SH = "s"
	FLAG_ZONE_LIST               = "zone_list"
	FLAG_ZONE_LIST_SH            = "z"
	FLAG_LOCALITY                = "locality"
	FLAG_LOCALITY_SH             = "l"
	FLAG_CONCURRENCY             = "concurrency"
	FLAG_CONCURRENCY_SH          = "c"
	FLAG_DECRYPTION              = "decryption"
	FLAG_DECRYPTION_SH           = "D"

	FLAG_KMS_ENCRYPT_INFO    = "kms_encrypt_info"
	FLAG_KMS_ENCRYPT_INFO_SH = "k"

	// obshell tenant modify
	CMD_MODIFY        = "modify"
	FLAG_OLD_PASSWORD = "old_password"
	FLAG_NEW_PASSWORD = "new_password"
	FLAG_PASSWORD     = "password"

	// obshell tenant replica
	CMD_REPLICA = "replica"

	// obshell tenant drop
	CMD_DROP     = "drop"
	FLAG_RECYCLE = "recycle"

	// obshell tenant show
	CMD_SHOW = "show"

	// obshell tenant lock
	CMD_LOCK = "lock"

	// obshell tenant purge
	CMD_PURGE = "purge"

	// obshell tenant flashback
	CMD_FLASHBACK    = "flashback"
	FLAG_NEW_NAME    = "new_name"
	FLAG_NEW_NAME_SH = "n"

	// obshell tenant unlock
	CMD_UNLOCK = "unlock"

	// obshell tenant rename
	CMD_RENAME = "rename"
)

func NewTenantCmd() *cobra.Command {
	tenantCmd := command.NewCommand(&cobra.Command{
		Use: clientconst.CMD_TENANT,
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
	tenantCmd.AddCommand(newCreateCmd())
	tenantCmd.AddCommand(newModifyCmd())
	tenantCmd.AddCommand(newDropCmd())
	tenantCmd.AddCommand(newShowCmd())
	tenantCmd.AddCommand(newLockCmd())
	tenantCmd.AddCommand(newUnlockCmd())
	tenantCmd.AddCommand(replica.NewReplicaCmd())
	tenantCmd.AddCommand(variable.NewVariableCmd())
	tenantCmd.AddCommand(parameter.NewParameterCmd())
	tenantCmd.AddCommand(newRenameCmd())
	tenantCmd.AddCommand(newBackupCmd())
	tenantCmd.AddCommand(newRestoreCmd())

	return tenantCmd.Command
}
