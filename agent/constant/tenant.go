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

const (
	MYSQL_MODE  = "MYSQL"
	ORACAL_MODE = "ORACLE"

	TENANT_STATUS_NORMAL = "NORMAL"

	TENANT_ROLE_PRIMARY = "PRIMARY"

	TENANT_TYPE_USER = "USER"
	TENANT_TYPE_META = "META"

	TENANT_SYS    = "sys"
	TENANT_SYS_ID = 1

	REPLICA_TYPE_FULL     = "FULL"
	REPLICA_TYPE_READONLY = "READONLY"

	PRIMARY_ZONE_RANDOM = "RANDOM"

	// not support tenant name since ob4.2.1
	TENANT_ALL      = "all"
	TENANT_ALL_USER = "all_user"
	TENANT_ALL_META = "all_meta"

	CHECK_JOB_RETRY_TIMES         = 360
	CHECK_JOB_INTERVAL            = 10 * time.Second
	CHECK_TENANT_EXIST_INTERVAL   = 5 * time.Second
	RESOURCE_UNIT_CONFIG_CPU_MINE = 1
	SLOW_SQL_THRESHOLD            = 100000

	VARIABLE_TIME_ZONE            = "time_zone"
	VARIABLE_OB_TCP_INVITED_NODES = "ob_tcp_invited_nodes"
	VARIABLE_READ_ONLY            = "read_only"

	PARAMETER_ENABLE_REBALANCE = "enable_rebalance"

	ALTER_RESOURCE_TENANT_UNIT_NUM = "ALTER_RESOURCE_TENANT_UNIT_NUM"
	ALTER_TENANT_LOCALITY          = "ALTER_TENANT_LOCALITY"
	ALTER_TENANT_PRIMARY_ZONE      = "ALTER_TENANT_PRIMARY_ZONE"
)
