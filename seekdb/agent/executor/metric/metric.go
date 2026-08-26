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
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/oceanbase/obshell/seekdb/agent/bindata"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/executor/external"
	metricconstant "github.com/oceanbase/obshell/seekdb/agent/executor/metric/constant"
	"github.com/oceanbase/obshell/seekdb/model/common"
	model "github.com/oceanbase/obshell/seekdb/model/metric"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var seekdbVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.(0|[1-9][0-9]*)){1,3}$`)

type MetricExprConfig struct {
	Expr         string `yaml:"expr"`
	MinObVersion string `yaml:"minObVersion"`
	MaxObVersion string `yaml:"maxObVersion"`
	configError  string
}

func (c *MetricExprConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var expr string
	if err := unmarshal(&expr); err == nil {
		c.Expr = expr
		return nil
	}

	type rawMetricExprConfig MetricExprConfig
	var raw rawMetricExprConfig
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*c = MetricExprConfig(raw)
	if c.Expr == "" {
		return fmt.Errorf("metric expression is empty")
	}
	return nil
}

var metricExprConfig map[string]MetricExprConfig
var getPrometheusClient = external.GetPrometheusClientFromConfig

func init() {
	metricExprConfig = make(map[string]MetricExprConfig)
	metricExprConfigContent, err := bindata.Asset(metricconstant.METRIC_EXPR_CONFIG_FILE)
	if err != nil {
		log.WithError(err).Error("load metric expr config failed")
		return
	}
	err = yaml.Unmarshal(metricExprConfigContent, &metricExprConfig)
	if err != nil {
		log.WithError(err).Error("parse metric expr config data failed")
		return
	}
	validateMetricExprConfigs(metricExprConfig)
}

func validateMetricExprConfigs(configs map[string]MetricExprConfig) {
	for name, config := range configs {
		config.configError = ""
		if err := config.validate(); err != nil {
			config.configError = err.Error()
			log.WithError(err).Errorf("disable metric %s because its version boundary is invalid", name)
		}
		configs[name] = config
	}
}

type obVersion struct {
	segments [4]uint64
}

func parseObVersion(raw string) (*obVersion, error) {
	if !seekdbVersionPattern.MatchString(raw) {
		return nil, fmt.Errorf("invalid seekdb product version %q", raw)
	}
	parsed := &obVersion{}
	for i, segment := range strings.Split(raw, ".") {
		value, err := strconv.ParseUint(segment, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid seekdb product version %q: %w", raw, err)
		}
		parsed.segments[i] = value
	}
	return parsed, nil
}

func (v *obVersion) lessThan(other *obVersion) bool {
	for i := range v.segments {
		if v.segments[i] < other.segments[i] {
			return true
		}
		if v.segments[i] > other.segments[i] {
			return false
		}
	}
	return false
}

func IsValidProductVersion(raw string) bool {
	_, err := parseObVersion(raw)
	return err == nil
}

func (c MetricExprConfig) validate() error {
	if c.Expr == "" {
		return fmt.Errorf("metric expression is empty")
	}
	var minVersion, maxVersion *obVersion
	var err error
	if c.MinObVersion != "" {
		minVersion, err = parseObVersion(c.MinObVersion)
		if err != nil {
			return fmt.Errorf("invalid minObVersion: %w", err)
		}
	}
	if c.MaxObVersion != "" {
		maxVersion, err = parseObVersion(c.MaxObVersion)
		if err != nil {
			return fmt.Errorf("invalid maxObVersion: %w", err)
		}
	}
	if minVersion != nil && maxVersion != nil && !minVersion.lessThan(maxVersion) {
		return fmt.Errorf("minObVersion %s must be less than maxObVersion %s", c.MinObVersion, c.MaxObVersion)
	}
	return nil
}

type metricAvailability int

const (
	metricAvailable metricAvailability = iota
	metricNotFound
	metricConfigInvalid
	metricVersionUnknown
	metricVersionUnsupported
)

func getMetricAvailability(name, obVersion string) (metricAvailability, string) {
	config, found := metricExprConfig[name]
	if !found {
		return metricNotFound, fmt.Sprintf("metric expression for %s not found", name)
	}
	if config.configError != "" {
		return metricConfigInvalid, fmt.Sprintf("metric %s is disabled: %s", name, config.configError)
	}
	if config.MinObVersion == "" && config.MaxObVersion == "" {
		return metricAvailable, ""
	}
	currentVersion, err := parseObVersion(obVersion)
	if err != nil {
		return metricVersionUnknown, fmt.Sprintf("cannot determine seekdb product version for version-restricted metric %s: %v", name, err)
	}
	if config.MinObVersion != "" {
		minVersion, _ := parseObVersion(config.MinObVersion)
		if currentVersion.lessThan(minVersion) {
			return metricVersionUnsupported, fmt.Sprintf("metric %s is not supported by seekdb %s", name, obVersion)
		}
	}
	if config.MaxObVersion != "" {
		maxVersion, _ := parseObVersion(config.MaxObVersion)
		if !currentVersion.lessThan(maxVersion) {
			return metricVersionUnsupported, fmt.Sprintf("metric %s is not supported by seekdb %s", name, obVersion)
		}
	}
	return metricAvailable, ""
}

func ListMetricClasses(scope, language, obVersion string) ([]model.MetricClass, error) {
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
		return metricClasses, err
	}
	return filterMetricClasses(metricClasses, obVersion), nil
}

func filterMetricClasses(metricClasses []model.MetricClass, obVersion string) []model.MetricClass {
	filteredClasses := make([]model.MetricClass, 0, len(metricClasses))
	for _, class := range metricClasses {
		filteredGroups := make([]model.MetricGroup, 0, len(class.MetricGroups))
		for _, group := range class.MetricGroups {
			filteredMetrics := make([]model.MetricMeta, 0, len(group.Metrics))
			for _, metricMeta := range group.Metrics {
				availability, _ := getMetricAvailability(metricMeta.Key, obVersion)
				if availability == metricAvailable {
					filteredMetrics = append(filteredMetrics, metricMeta)
				}
			}
			if len(filteredMetrics) > 0 {
				group.Metrics = filteredMetrics
				filteredGroups = append(filteredGroups, group)
			}
		}
		if len(filteredGroups) > 0 {
			class.MetricGroups = filteredGroups
			filteredClasses = append(filteredClasses, class)
		}
	}
	return filteredClasses
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

func extractMetricData(name string, resp *model.PrometheusQueryRangeResponse) []model.MetricData {
	metricDatas := make([]model.MetricData, 0)
	for _, result := range resp.Data.Result {
		values := make([]model.MetricValue, 0)

		labels := make([]common.KVPair, 0, len(result.Metric))
		for k, v := range result.Metric {
			labels = append(labels, common.KVPair{Key: k, Value: v})
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
			} else if math.IsNaN(v) || math.IsInf(v, 0) {
				log.Debugf("value at timestamp %f is invalid (%v)", t, v)
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

func validateMetricQuery(metrics []string, obVersion string) error {
	for _, name := range metrics {
		availability, reason := getMetricAvailability(name, obVersion)
		switch availability {
		case metricAvailable:
			continue
		case metricNotFound, metricVersionUnsupported:
			return errors.Occur(errors.ErrCommonBadRequest, reason)
		case metricConfigInvalid, metricVersionUnknown:
			return errors.Occur(errors.ErrCommonUnexpected, reason)
		}
	}
	return nil
}

func QueryMetricData(queryParam *model.MetricQuery, obVersion string) ([]model.MetricData, error) {
	metricDatas := make([]model.MetricData, 0, len(queryParam.Metrics))
	if err := validateMetricQuery(queryParam.Metrics, obVersion); err != nil {
		return metricDatas, err
	}
	client, err := getPrometheusClient()
	if err != nil {
		return metricDatas, errors.Occur(errors.ErrCommonUnexpected, err.Error())
	}
	wg := sync.WaitGroup{}
	metricDataCh := make(chan []model.MetricData, len(queryParam.Metrics))
	for _, m := range queryParam.Metrics {
		exprConfig, found := metricExprConfig[m]
		if found {
			wg.Add(1)
			go func(m string, ch chan []model.MetricData) {
				defer wg.Done()
				expr := replaceQueryVariables(exprConfig.Expr, queryParam.Labels, queryParam.GroupLabels, queryParam.QueryRange.Step)
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
					ch <- extractMetricData(m, queryRangeResp)
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
	return metricDatas, nil
}
