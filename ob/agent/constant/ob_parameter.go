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
	PARAMETER_ENABLE_REBALANCE               = "enable_rebalance"
	PARAMETER_GLOBAL_INDEX_AUTO_SPLIT_POLICY = "global_index_auto_split_policy"
	PARAMETER_MIN_FULL_RESOURCE_POOL_MEMORY  = "__min_full_resource_pool_memory"

	VARIABLE_TIME_ZONE              = "time_zone"
	VARIABLE_OB_TCP_INVITED_NODES   = "ob_tcp_invited_nodes"
	VARIABLE_READ_ONLY              = "read_only"
	VARIABLE_LOWER_CASE_TABLE_NAMES = "lower_case_table_names"
)

var (
	// READONLY variables
	CREATE_TENANT_STATEMENT_VARIABLES = []string{"lower_case_table_names"}
	// Those variables could not set by sys tenant.
	VARIAbLES_NEED_TO_CONNEC_WHEN_SET = []string{
		"collation_server",
		"collation_database",
		"collation_connection",
		"character_set_server",
		"character_set_database",
		"character_set_connection",
		"plsql_warnings",
	}
)
