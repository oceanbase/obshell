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

package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/config"
	agentconst "github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
)

const (
	// CMD_JOIN represents the "join" command.
	CMD_JOIN = "join"
	// Flags for the "join" command.
	FLAG_MYSQL_PORT    = "mysql_port"
	FLAG_MYSQL_PORT_SH = "p"
	FLAG_RPC_PORT      = "rpc_port"
	FLAG_RPC_PORT_SH   = "P"
	FLAG_DATA_DIR      = "data_dir"
	FLAG_DATA_DIR_SH   = "d"
	FLAG_REDO_DIR      = "redo_dir"
	FLAG_REDO_DIR_SH   = "r"
	FLAG_OPT_STR       = "opt_str"
	FLAG_OPT_STR_SH    = "o"
	FLAG_LOG_LEVEL     = "log_level"
	FLAG_LOG_LEVEL_SH  = "l"

	// CMD_REMOVE represents the "remove" command used to remove an obshell agent.
	CMD_REMOVE = "remove"

	// CMD_INIT represents the "init" command used to initialize the cluster.
	CMD_INIT = "init"
	// Flags for the "init" command.
	FLAG_PASSWORD        = "rootpassword"
	FLAG_PASSWORD_ALIAS  = "rp"
	FLAG_CLUSTER_NAME    = "cluster_name"
	FLAG_CLUSTER_NAME_SH = "n"
	FLAG_CLUSTER_ID      = "cluster_id"
	FLAG_CLUSTER_ID_SH   = "i"
	FLAG_RS_LIST         = "rs_list"
	FLAG_RS_LIST_ALIAS   = "rs"

	// CMD_START represents the "start" command used to start observers.
	CMD_START = "start"
	// Flags for the "start" command.
	FLAG_SERVER    = "server"
	FLAG_SERVER_SH = "s"
	FLAG_ZONE      = "zone"
	FLAG_ZONE_SH   = "z"
	FLAG_ID        = "id"
	FLAG_ID_SH     = "i"
	FLAG_ALL       = "all"
	FLAG_ALL_SH    = "a"

	// Flags for SSH configuration.
	FLAG_SSH_USER           = "ssh_user"
	FLAG_SSH_PORT           = "ssh_port"
	FLAG_SSH_KEY_FILE       = "key_file"
	FLAG_SSH_KEY_PASSPHRASE = "key_passphrase"
	FLAG_USER_PASSWORD      = "user_password"

	// CMD_STOP represents the "stop" command used to stop observers.
	CMD_STOP = "stop"
	// Flags for the "stop" command.
	FLAG_FORCE        = "force"
	FLAG_FORCE_SH     = "f"
	FLAG_TERMINATE    = "terminate"
	FLAG_TERMINATE_SH = "t"
	FLAG_IMMEDIATE    = "immediate"
	FLAG_IMMEDIATE_SH = "I"

	// CMD_SCALE_OUT represents the "scale-out" command.
	CMD_SCALE_OUT = "scale-out"

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

	// CMD_SHOW represents the "show" command used to display information about the cluster status.
	CMD_SHOW = "show"
)

var (
	CLUSTER_CMD          = fmt.Sprintf("%s %s", agentconst.PROC_OBSHELL, clientconst.CMD_CLUSTER)
	CLUSTER_CMD_TEMPLATE = CLUSTER_CMD + " %s [flags]"
)

func NewClusterCmd() *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:   clientconst.CMD_CLUSTER,
		Short: "Deploy and manage the OceanBase cluster.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			global.InitGlobalVariable()
			switch cmd.Use {
			case CMD_START:
				AsyncCheckAndStartDaemon()
				fmt.Println("Starting the OceanBase cluster, please wait...")
			case CMD_STOP, CMD_SHOW, CMD_SCALE_OUT, CMD_UPGRADE:
				if err := CheckAndStartDaemon(true); err != nil {
					stdio.StopLoading()
					stdio.Error(err.Error())
					return nil
				}
			default:
				if err := CheckAndStartDaemon(); err != nil {
					stdio.StopLoading()
					stdio.Error(err.Error())
					return nil
				}
			}
			return nil
		},
	}
	clusterCmd.AddCommand(newJoinCmd())
	clusterCmd.AddCommand(newRemoveCmd())
	clusterCmd.AddCommand(newInitCmd())
	clusterCmd.AddCommand(newStartCmd())
	clusterCmd.AddCommand(newUpgradeCmd())
	clusterCmd.AddCommand(NewScaleOutCmd())
	clusterCmd.AddCommand(newShowCmd())
	clusterCmd.AddCommand(newStopCmd())
	return clusterCmd
}
