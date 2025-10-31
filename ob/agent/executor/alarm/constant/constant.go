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
	DefaultAlarmQueryTimeout = 20
)

const (
	AlertUrl              = "/api/v2/alerts"
	SingleSilencerUrl     = "/api/v2/silence"
	MultiSilencerUrl      = "/api/v2/silences"
	RuleUrl               = "/api/v1/rules"
	PrometheusReloadUrl   = "/-/reload"
	AlertmanagerReloadUrl = "/-/reload"
	StatusUrl             = "/api/v2/status"
)

const (
	LabelOBCluster = "ob_cluster_name"
	LabelOBZone    = "obzone"
	LabelOBServer  = "svr_ip"
	LabelOBTenant  = "tenant_name"
)

const (
	LabelRuleName     = "rule_name"
	LabelRuleType     = "rule_type"
	LabelSeverity     = "severity"
	LabelInstanceType = "instance_type"
)

const (
	AnnoSummary     = "summary"
	AnnoDescription = "description"
)

const (
	OBRuleGroupName = "ob-rule"
)

const (
	RegexOR = "|"
)
