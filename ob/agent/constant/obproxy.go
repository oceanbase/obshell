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

const (
	OBPROXY_INFO_SQL_PORT             = "sql_port"
	OBPROXY_INFO_OBPROXY_SYS_PASSWORD = "obproxy_sys_password"
	OBPROXY_INFO_HOME_PATH            = "home_path"
	OBPROXY_INFO_PROXYRO_PASSWORD     = "proxyro_password"
	OBPROXY_INFO_VERSION              = "version"

	OBPROXY_CONFIG_PROMETHUES_LISTEN_PORT       = "prometheus_listen_port"
	OBPROXY_CONFIG_RS_LIST                      = "rootservice_list"
	OBPROXY_CONFIG_CONFIG_SERVER_URL            = "obproxy_config_server_url"
	OBPROXY_CONFIG_LISTEN_PORT                  = "listen_port"
	OBPROXY_CONFIG_CLUSTER_NAME                 = "cluster_name"
	OBPROXY_CONFIG_RPC_LISTEN_PORT              = "rpc_listen_port"
	OBPROXY_CONFIG_OBPROXY_SYS_PASSWORD         = "obproxy_sys_password"
	OBPROXY_CONFIG_ROOT_SERVICE_CLUSTER_NAME    = "rootservice_cluster_name"
	OBPROXY_CONFIG_PROXYRO_PASSWORD             = "observer_sys_password"
	OBPROXY_CONFIG_PROXY_LOCAL_CMD              = "proxy_local_cmd"
	OBPROXY_CONFIG_HOT_UPGRADE_ROLLBACK_TIMEOUT = "hot_upgrade_rollback_timeout"
	OBPROXY_CONFIG_HOT_UPGRADE_EXIT_TIMEOUT     = "hot_upgrade_exit_timeout"

	OBPROXY_MIN_VERSION_SUPPORT = "4.0.0"

	OBPROXY_INFO_STATUS = "status"

	OBPROXY_DIR_ETC = "etc"
	OBPROXY_DIR_BIN = "bin"
	OBPROXY_DIR_LIB = "lib"
	OBPROXY_DIR_LOG = "log"
	OBPROXY_DIR_RUN = "run"
	BIN_OBPROXY     = "obproxy"
	BIN_OBPROXYD    = "obproxyd"

	RESTART_FOR_PROXY_LOCAL_CMD = "2"

	OBPROXY_DEFAULT_SQL_PORT      = 2883
	OBPROXY_DEFAULT_EXPORTER_PORT = 2884
	OBPROXY_DEFAULT_RPC_PORT      = 2885

	DEFAULT_HOT_RESTART_TIME_OUT = 1800 // 30 minutes
)
