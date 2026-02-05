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

package metric

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/oceanbase/obshell/ob/agent/bindata"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/external"
	metricconstant "github.com/oceanbase/obshell/ob/agent/executor/metric/constant"
	"github.com/oceanbase/obshell/ob/model/common"
	model "github.com/oceanbase/obshell/ob/model/metric"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var metricExprConfig map[string]string

func init() {
	metricExprConfig = make(map[string]string)
	metricExprConfigContent, err := bindata.Asset(metricconstant.METRIC_EXPR_CONFIG_FILE)
	if err != nil {
		log.WithError(err).Error("load metric expr config failed")
	}
	err = yaml.Unmarshal(metricExprConfigContent, &metricExprConfig)
	if err != nil {
		log.WithError(err).Error("parse metric expr config data failed")
	}
}

func ListMetricClasses(scope, language string) ([]model.MetricClass, error) {
	metricClasses := make([]model.MetricClass, 0)
	configFile := metricconstant.METRIC_CONFIG_FILE_ENUS
	switch language {
	case constant.LANGUAGE_EN_US:
		configFile = metricconstant.METRIC_CONFIG_FILE_ENUS
	case constant.LANGUAGE_ZH_CN:
		configFile = metricconstant.METRIC_CONFIG_FILE_ZHCN
	default:
		log.Infof("Not supported language %s, return default", language)
	}

	metricConfigContent, err := bindata.Asset(configFile)
	if err != nil {
		return metricClasses, errors.Occur(errors.ErrCommonUnexpected, err.Error())
	}
	metricConfigMap := make(map[string][]model.MetricClass)
	err = yaml.Unmarshal(metricConfigContent, &metricConfigMap)
	if err != nil {
		return metricClasses, errors.Occur(errors.ErrJsonUnmarshal, err.Error())
	}
	log.Debugf("metric configs: %v", metricConfigMap)
	metricClasses, found := metricConfigMap[scope]
	if !found {
		err = errors.Occur(errors.ErrMetricConfigNotFound, scope)
	}
	return metricClasses, err
}

func replaceQueryVariables(exprTemplate string, labels []common.KVPair, groupLabels []string, step int64) string {
	labelStrParts := make([]string, 0, len(labels))
	for _, label := range labels {
		labelStrParts = append(labelStrParts, fmt.Sprintf("%s=\"%s\"", label.Key, label.Value))
	}
	labelStr := strings.Join(labelStrParts, ",")
	groupLabelStr := strings.Join(groupLabels, ",")
	replacer := strings.NewReplacer(metricconstant.KEY_INTERVAL, fmt.Sprintf("%ss", strconv.FormatInt(step, 10)), metricconstant.KEY_LABELS, labelStr, metricconstant.KEY_GROUP_LABELS, groupLabelStr)
	return replacer.Replace(exprTemplate)
}

func extractMetricData(name string, resp *model.PrometheusQueryRangeResponse, groupLabels []string) []model.MetricData {
	metricDatas := make([]model.MetricData, 0)
	for _, result := range resp.Data.Result {
		values := make([]model.MetricValue, 0)

		labels := make([]common.KVPair, 0, len(result.Metric))
		for k, v := range result.Metric {
			labels = append(labels, common.KVPair{Key: k, Value: v})
		}
		// When group_labels are requested, skip series with no labels (Prometheus can return
		// such series when some underlying series lack the group dimension).
		if len(groupLabels) > 0 && len(labels) == 0 {
			continue
		}

		metric := model.Metric{
			Name:   name,
			Labels: labels,
		}
		lastValid := math.NaN()
		invalidTimestamps := make([]float64, 0)
		// one loop to handle invalid timestamps interpolation
		for _, value := range result.Values {
			t := value[0].(float64)
			v, err := strconv.ParseFloat(value[1].(string), 64)
			if err != nil {
				log.Warnf("Failed to parse value %v", err)
				invalidTimestamps = append(invalidTimestamps, t)
			} else if math.IsNaN(v) {
				log.Debugf("value at timestamp %f is NaN", t)
				invalidTimestamps = append(invalidTimestamps, t)
			} else {
				// if there are invalid timestamps, interpolate them
				if len(invalidTimestamps) > 0 {
					var interpolated float64
					if math.IsNaN(lastValid) {
						interpolated = v
					} else {
						interpolated = (lastValid + v) / 2
					}
					// interpolate invalid slots with last valid value
					for _, it := range invalidTimestamps {
						values = append(values, model.MetricValue{
							Timestamp: it,
							Value:     interpolated,
						})
					}
					invalidTimestamps = invalidTimestamps[:0]
				}
				values = append(values, model.MetricValue{
					Timestamp: t,
					Value:     v,
				})
				lastValid = v
			}
		}
		if math.IsNaN(lastValid) {
			lastValid = 0.0
		}
		for _, it := range invalidTimestamps {
			values = append(values, model.MetricValue{
				Timestamp: it,
				Value:     lastValid,
			})
		}
		metricDatas = append(metricDatas, model.MetricData{
			Metric: metric,
			Values: values,
		})
	}
	return metricDatas
}

func QueryMetricData(queryParam *model.MetricQuery) []model.MetricData {
	metricDatas := make([]model.MetricData, 0, len(queryParam.Metrics))
	client, err := external.GetPrometheusClientFromConfig()
	if err != nil {
		return metricDatas
	}
	wg := sync.WaitGroup{}
	metricDataCh := make(chan []model.MetricData, len(queryParam.Metrics))
	for _, m := range queryParam.Metrics {
		exprTemplate, found := metricExprConfig[m]
		if found {
			wg.Add(1)
			go func(m string, ch chan []model.MetricData) {
				defer wg.Done()
				expr := replaceQueryVariables(exprTemplate, queryParam.Labels, queryParam.GroupLabels, queryParam.QueryRange.Step)
				log.Infof("Query with expr: %s, range: %v", expr, queryParam.QueryRange)
				queryRangeResp := &model.PrometheusQueryRangeResponse{}
				resp, err := client.R().SetQueryParams(map[string]string{
					"start": strconv.FormatFloat(queryParam.QueryRange.StartTimestamp, 'f', 3, 64),
					"end":   strconv.FormatFloat(queryParam.QueryRange.EndTimestamp, 'f', 3, 64),
					"step":  strconv.FormatInt(queryParam.QueryRange.Step, 10),
					"query": expr,
				}).SetHeader("content-type", "application/json").
					SetResult(queryRangeResp).
					Get(metricconstant.METRIC_RANGE_QUERY_URL)
				if err != nil {
					log.WithError(err).Error("Query expression expr got error")
				} else if resp.StatusCode() == http.StatusOK {
					ch <- extractMetricData(m, queryRangeResp, queryParam.GroupLabels)
				} else {
					log.Errorf("Query metrics from prometheus got unexpected status: %d", resp.StatusCode())
				}
			}(m, metricDataCh)
		} else {
			log.Errorf("Metric expression for %s not found", m)
		}
	}
	wg.Wait()
	close(metricDataCh)
	for metricDataArray := range metricDataCh {
		metricDatas = append(metricDatas, metricDataArray...)
	}
	return metricDatas
}
