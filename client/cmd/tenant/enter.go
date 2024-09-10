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
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

const (
	CMD_RESTORE = "restore"
)

const (
	FLAG_TENANT_NAME             = "tenant_name"
	FLAG_TENANT_NAME_SH          = "t"
	FLAG_ARCHIVE_LOG_URI         = "archive_log_uri"
	FLAG_ARCHIVE_LOG_URI_SH      = "a"
	FLAG_DATA_BACKUP_URI         = "data_backup_uri"
	FLAG_DATA_BACKUP_URI_SH      = "d"
	FLAG_UNIT_CONFIG_NAME        = "unit_config_name"
	FLAG_UNIT_CONFIG_NAME_SH     = "u"
	FLAG_UNIT_NUM                = "unit_num"
	FLAG_UNIT_NUM_SH             = "n"
	FLAG_TIMESTAMP               = "timestamp"
	FLAG_TIMESTAMP_SH            = "T"
	FLAG_SCN                     = "scn"
	FLAG_SCN_SH                  = "S"
	FLAG_HA_HIGH_THREAD_SCORE    = "ha_high_thread_score"
	FLAG_HA_HIGH_THREAD_SCORE_SH = "s"
	FLAG_ZONE_LIST               = "zone_list"
	FLAG_ZONE_LIST_SH            = "z"
	FLAG_PRIMARY_ZONE            = "primary_zone"
	FLAG_PRIMARY_ZONE_SH         = "p"
	FLAG_LOCALITY                = "locality"
	FLAG_LOCALITY_SH             = "l"
	FLAG_CONCURRENCY             = "concurrency"
	FLAG_CONCURRENCY_SH          = "c"
	FLAG_DECRYPTION              = "decryption"
	FLAG_DECRYPTION_SH           = "D"

	FLAG_KMS_ENCRYPT_INFO    = "kms_encrypt_info"
	FLAG_KMS_ENCRYPT_INFO_SH = "k"
)

func NewTenantCmd() *cobra.Command {

	cmd := command.NewCommand(&cobra.Command{
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
	cmd.AddCommand(newBackupCmd())
	cmd.AddCommand(newRestoreCmd())

	return cmd.Command
}
