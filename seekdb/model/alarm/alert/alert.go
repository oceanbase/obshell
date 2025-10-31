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

package alert

import (
	"errors"
	"time"

	alarmconstant "github.com/oceanbase/obshell/seekdb/agent/executor/alarm/constant"
	"github.com/oceanbase/obshell/seekdb/model/alarm"
	"github.com/oceanbase/obshell/seekdb/model/common"
	"github.com/oceanbase/obshell/seekdb/model/oceanbase"

	ammodels "github.com/prometheus/alertmanager/api/v2/models"
)

type Status struct {
	InhibitedBy []string `json:"inhibited_by" binding:"required"`
	SilencedBy  []string `json:"silenced_by" binding:"required"`
	State       State    `json:"state" binding:"required"`
}

type Alert struct {
	Fingerprint string                `json:"fingerprint" binding:"required"`
	Rule        string                `json:"rule" binding:"required"`
	Severity    alarm.Severity        `json:"severity" binding:"required"`
	Instance    *oceanbase.OBInstance `json:"instance" binding:"required"`
	StartsAt    int64                 `json:"starts_at" binding:"required"`
	UpdatedAt   int64                 `json:"updated_at" binding:"required"`
	EndsAt      int64                 `json:"ends_at" binding:"required"`
	Status      *Status               `json:"status" binding:"required"`
	Labels      []common.KVPair       `json:"labels,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
}

func NewAlert(alert *ammodels.GettableAlert) (*Alert, error) {
	rule, ok := alert.Labels[alarmconstant.LabelRuleName]
	if !ok {
		return nil, errors.New("Convert alert failed, no rule")
	}
	severity, ok := alert.Labels[alarmconstant.LabelSeverity]
	if !ok {
		return nil, errors.New("Convert alert failed, no severity")
	}
	labels := make([]common.KVPair, 0, len(alert.Labels))
	for k, v := range alert.Labels {
		labels = append(labels, common.KVPair{
			Key:   k,
			Value: v,
		})
	}
	instance := &oceanbase.OBInstance{}
	obcluster, exists := alert.Labels[alarmconstant.LabelOBCluster]
	if exists {
		instance.OBCluster = obcluster
		instance.Type = oceanbase.TypeOBCluster
	}

	summary, ok := alert.Annotations[alarmconstant.AnnoSummary]
	if !ok {
		return nil, errors.New("No summary info")
	}
	description, ok := alert.Annotations[alarmconstant.AnnoDescription]
	if !ok {
		return nil, errors.New("No description info")
	}
	return &Alert{
		Fingerprint: *alert.Fingerprint,
		Rule:        rule,
		Severity:    alarm.Severity(severity),
		Instance:    instance,
		StartsAt:    time.Time(*alert.StartsAt).Unix(),
		UpdatedAt:   time.Time(*alert.UpdatedAt).Unix(),
		EndsAt:      time.Time(*alert.EndsAt).Unix(),
		Status: &Status{
			InhibitedBy: alert.Status.InhibitedBy,
			SilencedBy:  alert.Status.SilencedBy,
			State:       State(*alert.Status.State),
		},
		Labels:      labels,
		Summary:     summary,
		Description: description,
	}, nil
}
