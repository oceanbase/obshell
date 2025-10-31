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
	URI_API_V1 = "/api/v1"

	URI_TASK_GROUP      = "/task"
	URI_AGENT_GROUP     = "/agent"
	URI_OB_GROUP        = "/ob"
	URI_OBSERVER_GROUP  = "/observer"
	URI_USER_GROUP      = "/user"
	URI_USERS_GROUP     = "/users"
	URI_DATABASE_GROUP  = "/database"
	URI_DATABASES_GROUP = "/databases"

	URI_METRIC_GROUP   = "/metrics"
	URI_SYSTEM_GROUP   = "/system"
	URI_EXTERNAL_GROUP = "/external"
	URI_PROMETHEUS     = "/prometheus"
	URI_ALERTMANAGER   = "/alertmanager"

	URI_INFO      = "/info"
	URI_STATE     = "/state"
	URI_TIME      = "/time"
	URI_GIT_INFO  = "/git-info"
	URI_HOST_INFO = "/host-info"
	URI_STATUS    = "/status"
	URI_SECRET    = "secret"

	URI_TOKEN = "/token"

	URI_SYNC_BIN = "/sync-bin"

	URI_DAG      = "/dag"
	URI_DAGS     = "/dags"
	URI_NODE     = "/node"
	URI_SUB_TASK = "/sub_task"
	URI_LOG      = "/log"
	URI_LOGS     = "/logs"
	URI_MAINTAIN = "/maintain"
	URI_UNFINISH = "/unfinish"

	// OB api
	URI_START      = "/start"
	URI_STOP       = "/stop"
	URI_RESTART    = "/restart"
	URI_AGENTS     = "/agents"
	URI_CHARSETS   = "/charsets"
	URI_STATISTICS = "/statistics"

	// Used for upgrade
	URI_UPGRADE = "/upgrade"
	URI_CHECK   = "/check"
	URI_PACKAGE = "/package"

	// Used for tenant
	URI_PASSWORD          = "/password"
	URI_WHITELIST         = "/whitelist"
	URI_PARAMETERS        = "/parameters"
	URI_VARIABLES         = "/variables"
	URI_VARIABLE          = "/variable"
	URI_PARAMETER         = "/parameter"
	URI_USER              = "/user"
	URI_USERS             = "/users"
	URI_COMPACT           = "/compact"
	URI_COMPACTION        = "/compaction"
	URI_COMPACTION_ERROR  = "/compaction-error"
	URI_DATABASES         = "/databases"
	URI_DB_PRIVILEGE      = "/db-privilege"
	URI_DB_PRIVILEGES     = "/db-privileges"
	URI_GLOBAL_PRIVILEGES = "/global-privileges"
	URI_GLOBAL_PRIVILEGE  = "/global-privilege"
	URI_STATS             = "/stats"
	URI_LOCK              = "/lock"

	URI_PARAM_NAME          = "name"
	URI_PATH_PARAM_NAME     = "/:" + URI_PARAM_NAME
	URI_PARAM_VAR           = "variable"
	URI_PATH_PARAM_VAR      = "/:" + URI_PARAM_VAR
	URI_PARAM_PARA          = "parameter"
	URI_PATH_PARAM_PARA     = "/:" + URI_PARAM_PARA
	URI_PARAM_USER          = "user"
	URI_PATH_PARAM_USER     = "/:" + URI_PARAM_USER
	URI_PARAM_DATABASE      = "database"
	URI_PATH_PARAM_DATABASE = "/:" + URI_PARAM_DATABASE

	URI_TASK_API_PREFIX     = URI_API_V1 + URI_TASK_GROUP
	URI_AGENT_API_PREFIX    = URI_API_V1 + URI_AGENT_GROUP
	URI_OBSERVER_API_PREFIX = URI_API_V1 + URI_OBSERVER_GROUP

	// Used for alarm
	URI_ALARM_GROUP = "/alarm"
	URI_ALERT       = "/alert"
	URI_ALERTS      = "/alerts"
	URI_SILENCER    = "/silencer"
	URI_SILENCERS   = "/silencers"
	URI_RULE        = "/rule"
	URI_RULES       = "/rules"

	URI_PARAM_ID      = "id"
	URI_PATH_PARAM_ID = "/:" + URI_PARAM_ID
)
