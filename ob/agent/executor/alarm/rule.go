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

package alarm

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	errors "github.com/oceanbase/obshell/ob/agent/errors"
	alarmconstant "github.com/oceanbase/obshell/ob/agent/executor/alarm/constant"
	"github.com/oceanbase/obshell/ob/agent/executor/external"
	"github.com/oceanbase/obshell/ob/model/alarm/rule"

	promv1 "github.com/prometheus/prometheus/web/api/v1"
	log "github.com/sirupsen/logrus"
)

func GetRule(ctx context.Context, name string) (*rule.RuleResponse, error) {
	rules, err := ListRules(ctx, nil)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	}
	for _, rule := range rules {
		if rule.Name == name {
			return &rule, nil
		}
	}
	return nil, errors.Occur(errors.ErrAlarmRuleNotFound, name)
}

func ListRules(ctx context.Context, filter *rule.RuleFilter) ([]rule.RuleResponse, error) {
	promRuleResponse := &rule.PromRuleResponse{}
	client, err := external.GetPrometheusClientFromConfig()
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmClientFailed, err)
	}
	resp, err := client.R().SetContext(ctx).SetQueryParam("type", "alert").SetHeader("content-type", "application/json").SetResult(promRuleResponse).Get(alarmconstant.RuleUrl)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Occur(errors.ErrAlarmUnexpectedStatus, resp.StatusCode())
	}
	log.Debugf("Response from prometheus: %v", resp)
	filteredRules := make([]rule.RuleResponse, 0)
	for _, ruleGroup := range promRuleResponse.Data.RuleGroups {
		for _, promRule := range ruleGroup.Rules {
			encodedPromRule, err := json.Marshal(promRule)
			if err != nil {
				log.Errorf("Got an error when encoding rule %v", promRule)
				continue
			}
			log.Debugf("Process prometheus rule: %s", string(encodedPromRule))
			alertingRule := &promv1.AlertingRule{}
			err = json.Unmarshal(encodedPromRule, alertingRule)
			if err != nil {
				log.Errorf("Got an error when decoding rule %v", promRule)
				continue
			}
			ruleResp := rule.NewRuleResponse(alertingRule)
			log.Debugf("Parsed prometheus rule: %v", ruleResp)
			if filterRule(ruleResp, filter) {
				filteredRules = append(filteredRules, *ruleResp)
			}
		}
	}
	return filteredRules, nil
}

func filterRule(rule *rule.RuleResponse, filter *rule.RuleFilter) bool {
	matched := true
	if filter != nil {
		if filter.Keyword != "" {
			matched = matched && strings.Contains(rule.Name, filter.Keyword)
		}
		if filter.InstanceType != "" {
			matched = matched && (rule.InstanceType == filter.InstanceType)
		}
		if filter.Severity != "" {
			matched = matched && (rule.Severity == filter.Severity)
		}
	}
	return matched
}
