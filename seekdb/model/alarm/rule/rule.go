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

package rule

import (
	"time"

	alarmconstant "github.com/oceanbase/obshell/seekdb/agent/executor/alarm/constant"
	bizcommon "github.com/oceanbase/obshell/seekdb/agent/executor/common"
	"github.com/oceanbase/obshell/seekdb/model/alarm"
	"github.com/oceanbase/obshell/seekdb/model/common"
	"github.com/oceanbase/obshell/seekdb/model/oceanbase"

	prommodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	promv1 "github.com/prometheus/prometheus/web/api/v1"
)

type PromRuleResponse struct {
	Status string                `json:"status" binding:"required"`
	Data   *promv1.RuleDiscovery `json:"data" binding:"required"`
}

type Rule struct {
	Name         string                   `json:"name" binding:"required"`
	InstanceType oceanbase.OBInstanceType `json:"instance_type" binding:"required"`
	Type         RuleType                 `json:"type" default:"customized"`
	Query        string                   `json:"query" binding:"required"`
	Duration     int                      `json:"duration" binding:"required"`
	Labels       []common.KVPair          `json:"labels" binding:"required"`
	Severity     alarm.Severity           `json:"severity" binding:"required"`
	Summary      string                   `json:"summary" binding:"required"`
	Description  string                   `json:"description" binding:"required"`
}

type RuleResponse struct {
	State          RuleState  `json:"state" binding:"required"`
	KeepFiringFor  int        `json:"keep_firing_for" binding:"required"`
	Health         RuleHealth `json:"health" binding:"required"`
	LastEvaluation int64      `json:"last_evaluation" binding:"required"`
	EvaluationTime float64    `json:"evaluation_time" binding:"required"`
	LastError      string     `json:"last_error,omitempty"`
	Rule
}

type RuleIdentity struct {
	Name string `json:"name" binding:"required"`
}

type ConfigRuleGroups struct {
	Groups []ConfigRuleGroup `json:"groups"`
}

type ConfigRuleGroup struct {
	Name  string         `json:"name"`
	Rules []rulefmt.Rule `json:"rules"`
}

func (r *Rule) ToPromRule() *rulefmt.Rule {
	annotations := make(map[string]string)
	annotations[alarmconstant.AnnoSummary] = r.Summary
	annotations[alarmconstant.AnnoDescription] = r.Description
	labels := r.Labels
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelRuleType,
		Value: string(r.Type),
	})
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelSeverity,
		Value: string(r.Severity),
	})
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelRuleName,
		Value: r.Name,
	})
	labels = append(labels, common.KVPair{
		Key:   alarmconstant.LabelInstanceType,
		Value: string(r.InstanceType),
	})
	promRule := &rulefmt.Rule{
		Alert:       r.Name,
		Expr:        r.Query,
		For:         prommodel.Duration(r.Duration * int(time.Second)),
		Labels:      bizcommon.KVsToMap(labels),
		Annotations: annotations,
	}
	return promRule
}

func NewRuleResponse(promRule *promv1.AlertingRule) *RuleResponse {
	var instanceType oceanbase.OBInstanceType
	var ruleType RuleType
	severity := alarm.SeverityInfo
	summary := ""
	description := ""
	labels := make([]common.KVPair, 0, len(promRule.Labels))
	for _, label := range promRule.Labels {
		labels = append(labels, common.KVPair{
			Key:   label.Name,
			Value: label.Value,
		})
		if label.Name == alarmconstant.LabelSeverity {
			severity = alarm.Severity(label.Value)
		}
		if label.Name == alarmconstant.LabelInstanceType {
			instanceType = oceanbase.OBInstanceType(label.Value)
		}
		if label.Name == alarmconstant.LabelRuleType {
			ruleType = RuleType(label.Value)
		}
	}
	for _, annotation := range promRule.Annotations {
		if annotation.Name == alarmconstant.AnnoSummary {
			summary = annotation.Value
		}
		if annotation.Name == alarmconstant.AnnoDescription {
			description = annotation.Value
		}
	}
	rule := &Rule{
		Name:         promRule.Name,
		InstanceType: instanceType,
		Type:         ruleType,
		Query:        promRule.Query,
		Duration:     int(promRule.Duration),
		Labels:       labels,
		Severity:     severity,
		Summary:      summary,
		Description:  description,
	}
	return &RuleResponse{
		State:          RuleState(promRule.State),
		KeepFiringFor:  int(promRule.KeepFiringFor),
		Health:         RuleHealth(promRule.Health),
		LastEvaluation: promRule.LastEvaluation.Unix(),
		EvaluationTime: promRule.EvaluationTime,
		LastError:      promRule.LastError,
		Rule:           *rule,
	}
}
