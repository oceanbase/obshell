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
	URI_RPC_V1 = "/rpc/v1"
	URI_API_V1 = "/api/v1"

	URI_TASK_GROUP      = "/task"
	URI_AGENT_GROUP     = "/agent"
	URI_AGENTS_GROUP    = "/agents"
	URI_OB_GROUP        = "/ob"
	URI_OBCLUSTER_GROUP = "/obcluster"
	URI_OBSERVER_GROUP  = "/observer"
	URI_TENANT_GROUP    = "/tenant"

	URI_INFO     = "/info"
	URI_TIME     = "/time"
	URI_GIT_INFO = "/git-info"
	URI_STATUS   = "/status"
	URI_SECRET   = "secret"

	URI_JOIN   = "/join"
	URI_REMOVE = "/remove"

	URI_DAG      = "/dag"
	URI_NODE     = "/node"
	URI_SUB_TASK = "/sub_task"
	URI_LOG      = "/log"
	URI_MAINTAIN = "/maintain"
	URI_UNFINISH = "/unfinish"

	// OB api
	URI_CONFIG      = "/config"
	URI_START_CHECK = "/startcheck"
	URI_DEPLOY      = "/deploy"
	URI_START       = "/start"
	URI_STOP        = "/stop"
	URI_UPDATE      = "/update"
	URI_INIT        = "/init"
	URI_DESTROY     = "/destroy"
	URI_SCALE_OUT   = "/scale_out"
	URI_AGENTS      = "/agents"

	// Used for upgrade
	URI_UPGRADE = "/upgrade"
	URI_CHECK   = "/check"
	URI_PACKAGE = "/package"
	URI_PARAMS  = "/params"
	URI_BACKUP  = "/backup"
	URI_RESTORE = "/restore"

	// Used for backup
	URI_OVERVIEW = "/overview"
	URI_ARCHIVE  = "/log"

	// Used for tenant
	URI_TENANT = "/tenant"

	URI_TASK_API_PREFIX   = URI_API_V1 + URI_TASK_GROUP
	URI_AGENT_API_PREFIX  = URI_API_V1 + URI_AGENT_GROUP
	URI_AGENTS_API_PREFIX = URI_API_V1 + URI_AGENTS_GROUP
	URI_OB_API_PREFIX     = URI_API_V1 + URI_OB_GROUP

	URI_TASK_RPC_PREFIX     = URI_RPC_V1 + URI_TASK_GROUP
	URI_AGENT_RPC_PREFIX    = URI_RPC_V1 + URI_AGENT_GROUP
	URI_OBSERVER_RPC_PREFIX = URI_RPC_V1 + URI_OBSERVER_GROUP
	URI_OB_RPC_PREFIX       = URI_RPC_V1 + URI_OB_GROUP
)
