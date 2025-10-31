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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	alarmconstant "github.com/oceanbase/obshell/seekdb/agent/executor/alarm/constant"
	"github.com/oceanbase/obshell/seekdb/agent/executor/external"
	"github.com/oceanbase/obshell/seekdb/model/alarm/silence"
	"github.com/oceanbase/obshell/seekdb/model/oceanbase"

	"github.com/go-openapi/strfmt"
	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	amsilence "github.com/prometheus/alertmanager/api/v2/restapi/operations/silence"
	log "github.com/sirupsen/logrus"
)

func DeleteSilencer(ctx context.Context, id string) error {
	client, err := external.GetAlertmanagerClientFromConfig()
	if err != nil {
		return errors.WrapRetain(errors.ErrAlarmClientFailed, err)
	}
	resp, err := client.R().SetContext(ctx).SetHeader("content-type", "application/json").Delete(fmt.Sprintf("%s/%s", alarmconstant.SingleSilencerUrl, id))
	if err != nil {
		return errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	} else if resp.StatusCode() != http.StatusOK {
		return errors.Occur(errors.ErrAlarmUnexpectedStatus, resp.StatusCode())
	}
	return nil
}

func GetSilencer(ctx context.Context, id string) (*silence.SilencerResponse, error) {
	gettableSilencer := ammodels.GettableSilence{}
	client, err := external.GetAlertmanagerClientFromConfig()
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmClientFailed, err)
	}
	resp, err := client.R().SetContext(ctx).SetHeader("content-type", "application/json").SetResult(&gettableSilencer).Get(fmt.Sprintf("%s/%s", alarmconstant.SingleSilencerUrl, id))
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Occur(errors.ErrAlarmUnexpectedStatus, resp.StatusCode())
	}
	return silence.NewSilencerResponse(&gettableSilencer), nil
}

func CreateOrUpdateSilencer(ctx context.Context, param *silence.SilencerParam) (*silence.SilencerResponse, error) {
	startTime := strfmt.DateTime(time.Now())
	endTime := strfmt.DateTime(time.Unix(param.EndsAt, 0))
	matchers := make(ammodels.Matchers, 0)
	matcherMap := make(map[string]*ammodels.Matcher)
	rules := strings.Join(param.Rules, alarmconstant.RegexOR)
	falseValue := false
	trueValue := true
	ruleName := alarmconstant.LabelRuleName
	ruleMatcher := &ammodels.Matcher{
		IsEqual: &trueValue,
		IsRegex: &trueValue,
		Name:    &ruleName,
		Value:   &rules,
	}
	matchers = append(matchers, ruleMatcher)
	matcherMap[ruleName] = ruleMatcher
	instanceType := oceanbase.TypeUnknown
	labelOBCluster := alarmconstant.LabelOBCluster
	labelInstance := alarmconstant.LabelOBCluster
	obcluster := ""
	instances := make([]string, 0, len(param.Instances))
	for _, instance := range param.Instances {
		if instanceType == oceanbase.TypeUnknown {
			instanceType = instance.Type
		}
		if instance.Type != instanceType {
			return nil, errors.Occur(errors.ErrAlarmSilencerInstanceTypeMismatch)
		}
		if instanceType != oceanbase.TypeOBCluster && obcluster != "" && obcluster != instance.OBCluster {
			return nil, errors.Occur(errors.ErrAlarmSilencerOBClusterMismatch)
		}
		obcluster = instance.OBCluster
		switch instance.Type {
		case oceanbase.TypeOBCluster:
			instances = append(instances, instance.OBCluster)
		case oceanbase.TypeOBServer:
			instances = append(instances, instance.OBServer)
			labelInstance = alarmconstant.LabelOBServer
		case oceanbase.TypeOBZone:
			instances = append(instances, instance.OBZone)
			labelInstance = alarmconstant.LabelOBZone
		case oceanbase.TypeOBTenant:
			instances = append(instances, instance.OBTenant)
			labelInstance = alarmconstant.LabelOBTenant
		default:
			return nil, errors.Occur(errors.ErrAlarmSilencerUnknownInstanceType, instance.Type)
		}
	}
	instanceValues := strings.Join(instances, alarmconstant.RegexOR)
	if instanceType != oceanbase.TypeOBCluster {
		clusterMatcher := &ammodels.Matcher{
			IsEqual: &trueValue,
			IsRegex: &falseValue,
			Name:    &labelOBCluster,
			Value:   &obcluster,
		}
		matchers = append(matchers, clusterMatcher)
		matcherMap[labelOBCluster] = clusterMatcher
	}

	instanceMatcher := &ammodels.Matcher{
		IsEqual: &trueValue,
		IsRegex: &trueValue,
		Name:    &labelInstance,
		Value:   &instanceValues,
	}

	matchers = append(matchers, instanceMatcher)
	matcherMap[labelInstance] = instanceMatcher
	for idx, m := range param.Matchers {
		matcher := &ammodels.Matcher{
			IsEqual: &trueValue,
			IsRegex: &param.Matchers[idx].IsRegex,
			Name:    &param.Matchers[idx].Name,
			Value:   &param.Matchers[idx].Value,
		}
		_, exists := matcherMap[m.Name]
		if !exists {
			log.Infof("matcher %s not exists, add it", m.Name)
			matchers = append(matchers, matcher)
			matcherMap[m.Name] = matcher
		} else {
			log.Infof("matcher %s exists, skip it", m.Name)
		}
	}

	silencer := ammodels.Silence{
		Comment:   &param.Comment,
		CreatedBy: &param.CreatedBy,
		StartsAt:  &startTime,
		EndsAt:    &endTime,
		Matchers:  matchers,
	}
	postableSilence := &ammodels.PostableSilence{
		ID:      param.Id,
		Silence: silencer,
	}
	okBody := amsilence.PostSilencesOKBody{}
	client, err := external.GetAlertmanagerClientFromConfig()
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmClientFailed, err)
	}
	resp, err := client.R().SetContext(ctx).SetHeader("content-type", "application/json").SetBody(postableSilence).SetResult(&okBody).Post(alarmconstant.MultiSilencerUrl)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Occur(errors.ErrAlarmUnexpectedStatus, resp.StatusCode())
	}
	state := string(silence.StateActive)
	gettableSilencer := ammodels.GettableSilence{
		Silence: silencer,
		ID:      &okBody.SilenceID,
		Status: &ammodels.SilenceStatus{
			State: &state,
		},
		UpdatedAt: &startTime,
	}
	silencerResponse := silence.NewSilencerResponse(&gettableSilencer)
	return silencerResponse, nil
}

func ListSilencers(ctx context.Context, filter *silence.SilencerFilter) ([]silence.SilencerResponse, error) {
	gettableSilencers := make(ammodels.GettableSilences, 0)
	client, err := external.GetAlertmanagerClientFromConfig()
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmClientFailed, err)
	}
	req := client.R().SetContext(ctx).SetHeader("content-type", "application/json")
	resp, err := req.SetResult(&gettableSilencers).Get(alarmconstant.MultiSilencerUrl)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Occur(errors.ErrAlarmUnexpectedStatus, resp.StatusCode())
	}
	log.Infof("resp: %v", resp)
	log.Infof("silencers: %v", gettableSilencers)
	filteredSilencers := make([]silence.SilencerResponse, 0)
	for _, gettableSilencer := range gettableSilencers {
		silencer := silence.NewSilencerResponse(gettableSilencer)
		if filterSilencer(silencer, filter) {
			filteredSilencers = append(filteredSilencers, *silencer)
		}
	}
	return filteredSilencers, nil
}

func filterSilencer(silencer *silence.SilencerResponse, filter *silence.SilencerFilter) bool {
	matched := true
	if filter != nil {
		if filter.Keyword != "" {
			matched = matched && strings.Contains(silencer.Comment, filter.Keyword)
		}
		// require at least one instance matches
		// TODO: whether to consider a cluster in filter matches a tenant or observer if the cluster names are same

		if filter.Instance != nil {
			instanceMatched := false
			for _, instance := range silencer.Instances {
				if instance.Equals(filter.Instance) {
					instanceMatched = true
					break
				}
			}
			matched = matched && instanceMatched
		}
		if filter.InstanceType != "" {
			instanceTypeMatched := false
			for _, instance := range silencer.Instances {
				if instance.Type == filter.InstanceType {
					instanceTypeMatched = true
					break
				}
			}
			matched = matched && instanceTypeMatched
		}
	}
	return matched
}
