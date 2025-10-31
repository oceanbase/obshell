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
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	alarmconstant "github.com/oceanbase/obshell/seekdb/agent/executor/alarm/constant"
	"github.com/oceanbase/obshell/seekdb/model/external"
)

func newClient(address string, auth *external.Auth) *resty.Client {
	client := resty.New().SetTimeout(time.Duration(alarmconstant.DefaultAlarmQueryTimeout * time.Second)).SetHostURL(address)
	if auth != nil && auth.Username != "" {
		client.SetBasicAuth(auth.Username, auth.Password)
	}
	return client
}

func newAlertmanagerClient(cfg *external.AlertmanagerConfig) *resty.Client {
	return newClient(cfg.Address, cfg.Auth)
}

func newPrometheusClient(cfg *external.PrometheusConfig) *resty.Client {
	return newClient(cfg.Address, cfg.Auth)
}

func GetAlertmanagerClientFromConfig() (*resty.Client, error) {
	cfg, err := GetAlertmanagerConfig(context.TODO())
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrConfigGetFailed, err, ALERTMANAGER_CONFIG_KEY, err.Error())
	}
	if cfg == nil {
		return nil, errors.Occur(errors.ErrConfigNotFound, ALERTMANAGER_CONFIG_KEY)
	}
	client := newAlertmanagerClient(cfg)
	return client, nil
}

func GetPrometheusClientFromConfig() (*resty.Client, error) {
	cfg, err := GetPrometheusConfig(context.TODO())
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrConfigGetFailed, err, PROMETHEUS_CONFIG_KEY, err.Error())
	}
	if cfg == nil {
		return nil, errors.Occur(errors.ErrConfigNotFound, ALERTMANAGER_CONFIG_KEY)
	}
	client := newPrometheusClient(cfg)
	return client, nil
}
