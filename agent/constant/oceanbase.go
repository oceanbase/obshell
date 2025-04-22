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

import "time"

// Compatible with OBShell which version prior to 4.2.3
var OB_CONFIG_COMPATIBLE_MAP = map[string]string{
	CONFIG_RPC_PORT:   "rpcPort",
	CONFIG_MYSQL_PORT: "mysqlPort",
}

var OB_INNER_USERS = []string{"PUBLIC", "LBACSYS", "OBAAUDITOR", "__oceanbase_inner_standby_user", "ocp_monitor"}
var OB_NO_PRIVILEGE_DBS = []string{"information_schema"}
var OB_RESTRICTED_DBS = []string{"oceanbase", "mysql", "SYS", "LBACSYS", "OBAAUDITOR"}

const (
	OB_MYSQL_PRIVILEGE_ALTER          = "ALTER"
	OB_MYSQL_PRIVILEGE_CREATE         = "CREATE"
	OB_MYSQL_PRIVILEGE_DELETE         = "DELETE"
	OB_MYSQL_PRIVILEGE_DROP           = "DROP"
	OB_MYSQL_PRIVILEGE_INSERT         = "INSERT"
	OB_MYSQL_PRIVILEGE_SELECT         = "SELECT"
	OB_MYSQL_PRIVILEGE_UPDATE         = "UPDATE"
	OB_MYSQL_PRIVILEGE_INDEX          = "INDEX"
	OB_MYSQL_PRIVILEGE_CREATE_VIEW    = "CREATE_VIEW"
	OB_MYSQL_PRIVILEGE_SHOW_VIEW      = "SHOW_VIEW"
	OB_MYSQL_PRIVILEGE_CREATE_USER    = "CREATE_USER"
	OB_MYSQL_PRIVILEGE_PROCESS        = "PROCESS"
	OB_MYSQL_PRIVILEGE_SUPER          = "SUPER"
	OB_MYSQL_PRIVILEGE_SHOW_DATABASES = "SHOW_DATABASES"
	OB_MYSQL_PRIVILEGE_GRANT_OPTION   = "GRANT_OPTION"
)

const (
	OB_CONNECTION_TYPE_DIRECT = "DIRECT"
	OB_CONNECTION_TYPE_PROXY  = "PROXY"
)

var OB_MYSQL_PRIVILEGES = []string{OB_MYSQL_PRIVILEGE_ALTER, OB_MYSQL_PRIVILEGE_CREATE, OB_MYSQL_PRIVILEGE_DELETE, OB_MYSQL_PRIVILEGE_DROP, OB_MYSQL_PRIVILEGE_INSERT, OB_MYSQL_PRIVILEGE_INDEX, OB_MYSQL_PRIVILEGE_SELECT, OB_MYSQL_PRIVILEGE_UPDATE, OB_MYSQL_PRIVILEGE_CREATE_VIEW, OB_MYSQL_PRIVILEGE_SHOW_VIEW, OB_MYSQL_PRIVILEGE_CREATE_USER, OB_MYSQL_PRIVILEGE_PROCESS, OB_MYSQL_PRIVILEGE_SUPER, OB_MYSQL_PRIVILEGE_SHOW_DATABASES, OB_MYSQL_PRIVILEGE_GRANT_OPTION}

const (
	DB_OCEANBASE = "oceanbase"
	DB_OCS       = "ocs"

	DEFAULT_HOST = "%"

	SYS_USER_PROXYRO = "proxyro"

	CONFIG_RPC_PORT   = "rpc_port"
	CONFIG_MYSQL_PORT = "mysql_port"

	DEFAULT_MYSQL_PORT = 2881
	DEFAULT_RPC_PORT   = 2882

	CONFIG_HOME_PATH      = "homePath"
	CONFIG_ROOT_PWD       = "rootPwd"
	CONFIG_AGENT_PASSWORD = "agentRootPwd"
	CONFIG_DATA_DIR       = "data_dir"
	CONFIG_REDO_DIR       = "redo_dir"
	CONFIG_CLOG_DIR       = "clog_dir"
	CONFIG_SLOG_DIR       = "slog_dir"

	CONFIG_LOCAL_IP  = "local_ip"
	CONFIG_DEV_NAME  = "devname"
	CONFIG_LOG_LEVEL = "syslog_level"
	CONFIG_ZONE      = "zone"

	CONFIG_ROOT_PASSWORD = "rootPwd"
	CONFIG_RS_LIST       = "rootservice_list"
	CONFIG_CLUSTER_ID    = "cluster_id"
	CONFIG_CLUSTER_NAME  = "cluster"

	CONFIG_FILE_LOCAL_IP = "local_ip"

	OB_PARAM_CLUSTER_ID   = "cluster_id"
	OB_PARAM_CLUSTER_NAME = "cluster"

	// configurable dir
	OB_DIR_STORE = "store"
	OB_DIR_CLOG  = "clog"
	OB_DIR_SLOG  = "slog"

	// unconfigurable dir
	OB_DIR_ETC     = "etc"
	OB_DIR_LOG     = "log"
	OB_DIR_SSTABLE = "store/sstable"

	OB_CONFIG_FILE = "observer.config.bin"
	OB_ADMIN       = "ob_admin"
	OB_BLOCK_FILE  = "block_file"

	OB_IMPORT_TIME_ZONE_INFO_SCRIPT = "import_time_zone_info.py"
	OB_IMPORT_SRS_DATA_SCRIPT       = "import_srs_data.py"
	OB_IMPORT_TIME_ZONE_INFO_FILE   = "timezone_V1.log"
	OB_IMPORT_SRS_DATA_FILE         = "default_srs_data_mysql.sql"

	OB_MODULE_TIMEZONE = "timezone"
	OB_MODULE_GIS      = "gis"
	OB_MODULE_REDIS    = "redis"

	// env
	OB_ROOT_PASSWORD = "OB_ROOT_PASSWORD"
)

const (
	CHUNK_SIZE uint64 = 1024*1024*16 - 1
)

const (
	TICK_INTERVAL_FOR_OB_STATUS_CHECK = 5 * time.Second
	TICK_NUM_FOR_OB_STATUS_CHECK      = 60

	OBSERVER_STATUS_DELETING = "DELETING"
)
