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

package constant

import (
	"fmt"

	"github.com/oceanbase/obshell/seekdb/model/oceanbase"
)

const (
	ENV_OBSHELL_TELEMETRY_ENABLED = "OBSHELL_TELEMETRY_ENABLED"
)

const (
	OCSAGENT_META_PATH = ".meta"
)

// Defaulte congfig for ocsagent
const (
	DEFAULT_AGENT_PORT = 2886
)

const (
	AGENT_V4241 = "4.2.4.1"
)

var (
	VERSION             = ""
	RELEASE             = ""
	DIST                = ""
	VERSION_RELEASE     = fmt.Sprintf("%s-%s", VERSION, RELEASE)
	SUPPORT_MIN_VERSION = "1.0.0.0"
)

// key of TABLE_OCS_INFO
const (
	OCS_INFO_IP           = "ip"
	OCS_INFO_PORT         = "port"
	OCS_INFO_IDENTITY     = "identity"
	OCS_INFO_VERSION      = "version"
	OCS_INFO_STATUS       = "status"
	OCS_INFO_OS           = "os"
	OCS_INFO_ARCHITECTURE = "architecture"
)

const (
	AGENT_START_TIMEOUT = 600
)

// command flag
const (
	FLAG_PORT      = "port"
	FLAG_PORT_SH   = "P"
	FLAG_PID       = "pid"
	FLAG_START_OB  = "ob"
	FLAG_TAKE_OVER = "takeover"
	FLAG_ROOT_PWD  = "password"
	FLAG_BASE_DIR  = "base-dir"

	FLAG_NEED_BE_CLUSTER = "cluster"
)

// proc name
const (
	PROC_OBSHELL        = "obshell"
	PROC_OBSHELL_SERVER = "server"
	PROC_OBSHELL_DAEMON = "daemon"
	PROC_OBSHELL_ADMIN  = "admin"
	PROC_OBSHELL_CLIENT = "client"

	PROC_OBSERVER = "observer"
)

// upload pkg names
const (
	PKG_OBSHELL              = "obshell"
	PKG_OCEANBASE_CE         = "oceanbase-ce"
	PKG_OCEANBASE_CE_LIBS    = "oceanbase-ce-libs"
	PKG_OCEANBASE            = "oceanbase"
	PKG_OCEANBASE_STANDALONE = "oceanbase-standalone"
)

var SUPPORT_PKG_NAMES_MAP = map[oceanbase.OBType][]string{
	oceanbase.OBTypeCommunity: {
		PKG_OBSHELL,
		PKG_OCEANBASE_CE,
		PKG_OCEANBASE_CE_LIBS,
	},
	oceanbase.OBTypeBusiness: {
		PKG_OBSHELL,
		PKG_OCEANBASE,
	},
	oceanbase.OBTypeStandalone: {
		PKG_OBSHELL,
		PKG_OCEANBASE_STANDALONE,
	},
}

var REQUIRE_UPGRADE_PKG_NAMES_MAP = map[oceanbase.OBType][]string{
	oceanbase.OBTypeCommunity: {
		PKG_OCEANBASE_CE,
		PKG_OCEANBASE_CE_LIBS,
	},
	oceanbase.OBTypeBusiness: {
		PKG_OCEANBASE,
	},
	oceanbase.OBTypeStandalone: {
		PKG_OCEANBASE_STANDALONE,
	},
}

const (
	DIR_RUN         = "run"
	DIR_BIN         = "bin"
	DIR_CA          = "ca"
	DIR_LOG_OBSHELL = "log_obshell"
)

// exit code
const (
	// killed by command
	EXIT_CODE_CMD_KILL = iota + 16
	// exit code for agent
	EXIT_CODE_ERROR_GET_IP_FAILED
	EXIT_CODE_ERROR_INVAILD_AGENT
	EXIT_CODE_ERROR_IP_NOT_MATCH
	EXIT_CODE_ERROR_AGENT_START_FAILED
	EXIT_CODE_ERROR_SERVER_LISTEN
	EXIT_CODE_ERROR_TAKE_OVER_FAILED
	EXIT_CODE_NOTIFY_SIGNAL
	EXIT_CODE_ERROR_NOT_CLUSTER_AGENT
	// exit code for ob
	EXIT_CODE_ERROR_OB_START_FAILED
	EXIT_CODE_ERROR_OB_CONN_TIMEOUT
	// exit code for permission denied
	EXIT_CODE_ERROR_PERMISSION_DENIED
	EXIT_CODE_ERROR_OB_PWD_ERROR
	// exit code for daemon
	EXIT_CODE_ERROR_DAEMON_START_FAILED
	// exit code for admin
	EXIT_CODE_ERROR_ADMIN_START_FAILED
	// exit code for backup binary
	EXIT_CODE_ERROR_BACKUP_BINARY_FAILED
	EXIT_CODE_ERROR_EXEC_BINARY_FAILED

	// exit code for base dir invalid
	EXIT_CODE_ERROR_BASE_DIR_INVALID
)

var AGENT_NEED_EXIT_CODE_LIST = []int{}
