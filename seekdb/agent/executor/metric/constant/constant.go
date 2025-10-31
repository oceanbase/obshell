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
	PARAM_SCOPE            = "scope"
	SCOPE_CLUSTER          = "OBCLUSTER"
	SCOPE_CLUSTER_OVERVIEW = "OBCLUSTER_OVERVIEW"
	SCOPE_OBPROXY          = "OBPROXY"
	SCOPE_SEEKDB           = "SEEKDB"
)

const (
	METRIC_RANGE_QUERY_URL  = "/api/v1/query_range"
	DEFAULT_TIMEOUT         = 30
	METRIC_CONFIG_FILE_ENUS = "agent/assets/metric/metrics-en_US.yaml"
	METRIC_CONFIG_FILE_ZHCN = "agent/assets/metric/metrics-zh_CN.yaml"
	METRIC_EXPR_CONFIG_FILE = "agent/assets/metric/metric_expr.yaml"
	KEY_INTERVAL            = "@INTERVAL"
	KEY_LABELS              = "@LABELS"
	KEY_GROUP_LABELS        = "@GBLABELS"
)
