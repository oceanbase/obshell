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

package external

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/oceanbase/obshell/ob/agent/errors"
	configservice "github.com/oceanbase/obshell/ob/agent/service/config"
	"github.com/oceanbase/obshell/ob/model/external"
	log "github.com/sirupsen/logrus"
)

const (
	PROMETHEUS                            = "prometheus"
	ALERTMANAGER                          = "alertmanager"
	PROMETHEUS_CONFIG_KEY                 = "prometheus_config"
	ALERTMANAGER_CONFIG_KEY               = "alertmanager_config"
	PROMETHEUS_READINESS_CHECK_URL        = "/-/ready"
	ALERTMANAGER_READINESS_CHECK_URL      = "/-/ready"
	PROMETHEUS_READINESS_CHECK_RESPONSE   = "Prometheus Server is Ready.\n"
	ALERTMANAGER_READINESS_CHECK_RESPONSE = "OK"
)

func SavePrometheusConfig(ctx context.Context, cfg *external.PrometheusConfig) error {
	client := newPrometheusClient(cfg)
	resp, err := client.R().SetContext(ctx).Get(PROMETHEUS_READINESS_CHECK_URL)
	if err != nil {
		return errors.WrapRetain(errors.ErrExternalComponentNotReady, err, PROMETHEUS)
	} else if resp.StatusCode() != http.StatusOK || string(resp.Body()) != PROMETHEUS_READINESS_CHECK_RESPONSE {
		log.Warnf("check prometheus readiness failed response %s", string(resp.Body()))
		return errors.Occur(errors.ErrExternalComponentNotReady, PROMETHEUS)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return errors.Occur(errors.ErrJsonMarshal, err.Error())
	}
	return configservice.SaveOcsConfig(PROMETHEUS_CONFIG_KEY, string(data), "Prometheus configuration")
}

func GetPrometheusConfig(ctx context.Context) (*external.PrometheusConfig, error) {
	ocsConfig, err := configservice.GetOcsConfig(PROMETHEUS_CONFIG_KEY)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrConfigGetFailed, err, PROMETHEUS_CONFIG_KEY, err.Error())
	}
	if ocsConfig == nil {
		return nil, errors.Occurf(errors.ErrConfigNotFound, PROMETHEUS_CONFIG_KEY)
	}
	var cfg external.PrometheusConfig
	err = json.Unmarshal([]byte(ocsConfig.Value), &cfg)
	if err != nil {
		return nil, errors.Occur(errors.ErrJsonUnmarshal, err.Error())
	}
	client := newPrometheusClient(&cfg)
	resp, err := client.R().SetContext(ctx).Get(PROMETHEUS_READINESS_CHECK_URL)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrExternalComponentNotReady, err, PROMETHEUS)
	} else if resp.StatusCode() != http.StatusOK || string(resp.Body()) != PROMETHEUS_READINESS_CHECK_RESPONSE {
		log.Warnf("check prometheus readiness failed response %s", string(resp.Body()))
		return nil, errors.Occur(errors.ErrExternalComponentNotReady, PROMETHEUS)
	}
	return &cfg, nil
}

func SaveAlertmanagerConfig(ctx context.Context, cfg *external.AlertmanagerConfig) error {
	client := newAlertmanagerClient(cfg)
	resp, err := client.R().SetContext(ctx).Get(ALERTMANAGER_READINESS_CHECK_URL)
	if err != nil {
		return errors.WrapRetain(errors.ErrExternalComponentNotReady, err, ALERTMANAGER)
	} else if resp.StatusCode() != http.StatusOK || string(resp.Body()) != ALERTMANAGER_READINESS_CHECK_RESPONSE {
		log.Warnf("check alertmanager readiness failed response %s", string(resp.Body()))
		return errors.Occur(errors.ErrExternalComponentNotReady, ALERTMANAGER)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return errors.Occur(errors.ErrJsonMarshal, err.Error())
	}
	return configservice.SaveOcsConfig(ALERTMANAGER_CONFIG_KEY, string(data), "Alertmanager configuration")
}

func GetAlertmanagerConfig(ctx context.Context) (*external.AlertmanagerConfig, error) {
	ocsConfig, err := configservice.GetOcsConfig(ALERTMANAGER_CONFIG_KEY)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrConfigGetFailed, err, ALERTMANAGER_CONFIG_KEY, err.Error())
	}
	if ocsConfig == nil {
		return nil, errors.Occur(errors.ErrConfigNotFound, ALERTMANAGER_CONFIG_KEY)
	}
	var cfg external.AlertmanagerConfig
	err = json.Unmarshal([]byte(ocsConfig.Value), &cfg)
	if err != nil {
		return nil, errors.Occur(errors.ErrJsonUnmarshal, err.Error())
	}
	client := newAlertmanagerClient(&cfg)
	resp, err := client.R().SetContext(ctx).Get(ALERTMANAGER_READINESS_CHECK_URL)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrExternalComponentNotReady, err, ALERTMANAGER)
	} else if resp.StatusCode() != http.StatusOK || string(resp.Body()) != ALERTMANAGER_READINESS_CHECK_RESPONSE {
		log.Warnf("check alertmanager readiness failed response %s", string(resp.Body()))
		return nil, errors.Occur(errors.ErrExternalComponentNotReady, ALERTMANAGER)
	}
	return &cfg, nil
}
