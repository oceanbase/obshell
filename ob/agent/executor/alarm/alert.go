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
	"net/http"
	"strings"

	errors "github.com/oceanbase/obshell/ob/agent/errors"
	alarmconstant "github.com/oceanbase/obshell/ob/agent/executor/alarm/constant"
	"github.com/oceanbase/obshell/ob/agent/executor/external"
	"github.com/oceanbase/obshell/ob/model/alarm/alert"

	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	log "github.com/sirupsen/logrus"
)

func ListAlerts(ctx context.Context, filter *alert.AlertFilter) ([]alert.Alert, error) {
	gettableAlerts := make(ammodels.GettableAlerts, 0)

	client, err := external.GetAlertmanagerClientFromConfig()
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmClientFailed, err)
	}

	resp, err := client.R().SetContext(ctx).SetQueryParams(map[string]string{
		"active":      "true",
		"silenced":    "true",
		"inhibited":   "true",
		"unprocessed": "true",
		"receiver":    "",
	}).SetHeader("content-type", "application/json").SetResult(&gettableAlerts).Get(alarmconstant.AlertUrl)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrAlarmQueryFailed, err)
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Occur(errors.ErrAlarmUnexpectedStatus, resp.StatusCode())
	}
	filteredAlerts := make([]alert.Alert, 0)
	for _, gettableAlert := range gettableAlerts {
		alert, err := alert.NewAlert(gettableAlert)
		if err != nil {
			log.WithError(err).Error("Parse alert got error, just skip")
			continue
		}
		if filterAlert(alert, filter) {
			filteredAlerts = append(filteredAlerts, *alert)
		}
	}
	return filteredAlerts, nil
}

func filterAlert(alert *alert.Alert, filter *alert.AlertFilter) bool {
	matched := true
	if filter.Severity != "" {
		matched = matched && (filter.Severity == alert.Severity)
	}
	if filter.StartTime != 0 {
		matched = matched && (filter.StartTime <= alert.StartsAt)
	}
	if filter.EndTime != 0 {
		matched = matched && (filter.EndTime >= alert.StartsAt)
	}
	if filter.Keyword != "" {
		matched = matched && (strings.Contains(alert.Description, filter.Keyword) || strings.Contains(alert.Summary, filter.Keyword))
	}
	if filter.Instance != nil {
		matched = matched && filter.Instance.Equals(alert.Instance)
	}
	if filter.InstanceType != "" {
		matched = matched && (filter.InstanceType == alert.Instance.Type)
	}
	return matched
}
