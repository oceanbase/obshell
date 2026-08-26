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
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	agenterrors "github.com/oceanbase/obshell/seekdb/agent/errors"
	metricconstant "github.com/oceanbase/obshell/seekdb/agent/executor/metric/constant"
	model "github.com/oceanbase/obshell/seekdb/model/metric"
	"gopkg.in/yaml.v2"
)

func useMetricExprConfig(t *testing.T, configs map[string]MetricExprConfig) {
	t.Helper()
	original := metricExprConfig
	validateMetricExprConfigs(configs)
	metricExprConfig = configs
	t.Cleanup(func() {
		metricExprConfig = original
	})
}

func TestMetricExprConfigUnmarshal(t *testing.T) {
	content := []byte(`
scalar: sum(metric_total)
object:
  expr: sum(legacy_metric_total)
  minObVersion: 1.3.0.0
  maxObVersion: 1.4.0.0
`)
	configs := make(map[string]MetricExprConfig)
	if err := yaml.Unmarshal(content, &configs); err != nil {
		t.Fatalf("unmarshal metric expressions: %v", err)
	}
	if configs["scalar"].Expr != "sum(metric_total)" {
		t.Fatalf("unexpected scalar expression: %+v", configs["scalar"])
	}
	object := configs["object"]
	if object.Expr != "sum(legacy_metric_total)" || object.MinObVersion != "1.3.0.0" || object.MaxObVersion != "1.4.0.0" {
		t.Fatalf("unexpected object expression: %+v", object)
	}

	if err := yaml.Unmarshal([]byte("invalid:\n  maxObVersion: 1.4.0.0\n"), &configs); err == nil {
		t.Fatal("expected an object without expr to be rejected")
	}
}

func TestStrictObVersionParsing(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{version: "1.3", valid: true},
		{version: "1.4.0.0", valid: true},
		{version: "0.0.0.0", valid: true},
		{version: "", valid: false},
		{version: "v1.4.0.0", valid: false},
		{version: "1.04.0.0", valid: false},
		{version: "1.4.x", valid: false},
		{version: "1.4.0.0.1", valid: false},
	}
	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			if valid := IsValidProductVersion(test.version); valid != test.valid {
				t.Fatalf("IsValidProductVersion(%q) = %v, want %v", test.version, valid, test.valid)
			}
		})
	}
}

func TestMetricVersionAvailability(t *testing.T) {
	useMetricExprConfig(t, map[string]MetricExprConfig{
		"public":  {Expr: "public_metric"},
		"legacy":  {Expr: "legacy_metric", MaxObVersion: "1.4.0.0"},
		"new":     {Expr: "new_metric", MinObVersion: "1.4.0.0"},
		"window":  {Expr: "window_metric", MinObVersion: "1.3.0.0", MaxObVersion: "1.4.0.0"},
		"invalid": {Expr: "invalid_metric", MaxObVersion: "not-a-version"},
	})

	tests := []struct {
		name         string
		metric       string
		version      string
		availability metricAvailability
	}{
		{name: "public without version", metric: "public", availability: metricAvailable},
		{name: "legacy on 1.3", metric: "legacy", version: "1.3.9.9", availability: metricAvailable},
		{name: "max is exclusive", metric: "legacy", version: "1.4.0.0", availability: metricVersionUnsupported},
		{name: "min is inclusive", metric: "new", version: "1.4.0.0", availability: metricAvailable},
		{name: "new metric on 1.3", metric: "new", version: "1.3.9.9", availability: metricVersionUnsupported},
		{name: "four and three segments compare equally", metric: "legacy", version: "1.4.0", availability: metricVersionUnsupported},
		{name: "unknown version", metric: "legacy", availability: metricVersionUnknown},
		{name: "invalid runtime version", metric: "legacy", version: "seekdb-1.4", availability: metricVersionUnknown},
		{name: "invalid boundary", metric: "invalid", version: "1.3.0.0", availability: metricConfigInvalid},
		{name: "missing metric", metric: "missing", version: "1.4.0.0", availability: metricNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			availability, _ := getMetricAvailability(test.metric, test.version)
			if availability != test.availability {
				t.Fatalf("availability = %v, want %v", availability, test.availability)
			}
		})
	}
}

func TestFilterMetricClassesPrunesEmptyGroupsAndClasses(t *testing.T) {
	useMetricExprConfig(t, map[string]MetricExprConfig{
		"public": {Expr: "public_metric"},
		"legacy": {Expr: "legacy_metric", MaxObVersion: "1.4.0.0"},
	})
	classes := []model.MetricClass{
		{
			Name: "mixed class",
			MetricGroups: []model.MetricGroup{
				{Name: "legacy group", Metrics: []model.MetricMeta{{Key: "legacy"}}},
				{Name: "mixed group", Metrics: []model.MetricMeta{{Key: "legacy"}, {Key: "public"}}},
			},
		},
		{Name: "legacy class", MetricGroups: []model.MetricGroup{{Name: "legacy group", Metrics: []model.MetricMeta{{Key: "legacy"}}}}},
	}

	filtered := filterMetricClasses(classes, "1.4.0.0")
	if len(filtered) != 1 || len(filtered[0].MetricGroups) != 1 || len(filtered[0].MetricGroups[0].Metrics) != 1 {
		t.Fatalf("unexpected filtered catalog: %+v", filtered)
	}
	if filtered[0].MetricGroups[0].Metrics[0].Key != "public" {
		t.Fatalf("unexpected remaining metric: %+v", filtered[0].MetricGroups[0].Metrics)
	}
}

func TestMetricAssetsAreFilteredForSeekdb14(t *testing.T) {
	if len(metricExprConfig) != 172 {
		t.Fatalf("expression count = %d, want 172", len(metricExprConfig))
	}
	boundedCount := 0
	for _, config := range metricExprConfig {
		if config.MaxObVersion == "1.4.0.0" {
			boundedCount++
		}
	}
	if boundedCount != 114 {
		t.Fatalf("expressions restricted below 1.4 = %d, want 114", boundedCount)
	}
	memoryUsageConfig, exists := metricExprConfig["seekdb_ob_memory_percent"]
	if !exists || memoryUsageConfig.MaxObVersion != "1.4.0.0" {
		t.Fatalf("seekdb memory usage metric must be restricted below 1.4: %+v", memoryUsageConfig)
	}
	for _, name := range []string{
		"seekdb_location_cache_hit",
		"seekdb_location_cache_hit_ratio",
		"seekdb_location_cache_miss",
		"seekdb_location_cache_req_total",
		"seekdb_location_cache_size",
		"seekdb_location_cache_size_mb",
	} {
		config, exists := metricExprConfig[name]
		if !exists || config.MaxObVersion != "1.4.0.0" {
			t.Fatalf("location cache expression %s must only be available below 1.4: %+v", name, config)
		}
	}

	for _, language := range []string{constant.LANGUAGE_EN_US, constant.LANGUAGE_ZH_CN} {
		t.Run(language, func(t *testing.T) {
			classes13, err := ListMetricClasses(metricconstant.SCOPE_SEEKDB, language, "1.3.0.0")
			if err != nil {
				t.Fatalf("list 1.3 metric classes: %v", err)
			}
			if countMetricMetas(classes13) != 143 {
				t.Fatalf("1.3 metric count = %d, want 143", countMetricMetas(classes13))
			}
			if !containsMetric(classes13, "seekdb_ob_memory_percent") {
				t.Fatal("seekdb memory usage metric is missing from the 1.3 catalog")
			}
			legacyCatalogMetrics := map[string]struct{}{
				"seekdb_location_cache_size_mb":   {},
				"seekdb_location_cache_hit_ratio": {},
				"seekdb_location_cache_req_total": {},
			}
			foundLegacyMetrics := make(map[string]struct{}, len(legacyCatalogMetrics))
			for _, class := range classes13 {
				for _, group := range class.MetricGroups {
					for _, metric := range group.Metrics {
						if _, legacy := legacyCatalogMetrics[metric.Key]; legacy {
							foundLegacyMetrics[metric.Key] = struct{}{}
						}
					}
				}
			}
			if len(foundLegacyMetrics) != len(legacyCatalogMetrics) {
				t.Fatalf("1.3 location cache catalog is incomplete: got %v, want %v", foundLegacyMetrics, legacyCatalogMetrics)
			}

			classes14, err := ListMetricClasses(metricconstant.SCOPE_SEEKDB, language, "1.4.0.0")
			if err != nil {
				t.Fatalf("list 1.4 metric classes: %v", err)
			}
			if countMetricMetas(classes14) != 44 {
				t.Fatalf("1.4 metric count = %d, want 44", countMetricMetas(classes14))
			}
			if containsMetric(classes14, "seekdb_ob_memory_percent") {
				t.Fatal("seekdb memory usage metric is still listed for 1.4")
			}
			for key := range legacyCatalogMetrics {
				if containsMetric(classes14, key) {
					t.Fatalf("legacy location cache metric %s is still listed for 1.4", key)
				}
			}
			classesUnknown, err := ListMetricClasses(metricconstant.SCOPE_SEEKDB, language, "")
			if err != nil {
				t.Fatalf("list metric classes without product version: %v", err)
			}
			if countMetricMetas(classesUnknown) != 44 {
				t.Fatalf("unknown-version metric count = %d, want 44", countMetricMetas(classesUnknown))
			}
			if containsMetric(classesUnknown, "seekdb_ob_memory_percent") {
				t.Fatal("seekdb memory usage metric is listed without a known product version")
			}
			for key := range legacyCatalogMetrics {
				if containsMetric(classesUnknown, key) {
					t.Fatalf("legacy location cache metric %s is listed without a known product version", key)
				}
			}
			for _, class := range classes14 {
				if len(class.MetricGroups) == 0 {
					t.Fatalf("empty metric class was not pruned: %s", class.Name)
				}
				for _, group := range class.MetricGroups {
					if len(group.Metrics) == 0 {
						t.Fatalf("empty metric group was not pruned: %s/%s", class.Name, group.Name)
					}
				}
			}
		})
	}
}

func countMetricMetas(classes []model.MetricClass) int {
	count := 0
	for _, class := range classes {
		for _, group := range class.MetricGroups {
			count += len(group.Metrics)
		}
	}
	return count
}

func containsMetric(classes []model.MetricClass, key string) bool {
	for _, class := range classes {
		for _, group := range class.MetricGroups {
			for _, metric := range group.Metrics {
				if metric.Key == key {
					return true
				}
			}
		}
	}
	return false
}

func TestUnsupportedMetricDoesNotCreatePrometheusClient(t *testing.T) {
	useMetricExprConfig(t, map[string]MetricExprConfig{
		"public": {Expr: "public_metric"},
		"legacy": {Expr: "legacy_metric", MaxObVersion: "1.4.0.0"},
	})
	original := getPrometheusClient
	clientRequested := false
	getPrometheusClient = func() (*resty.Client, error) {
		clientRequested = true
		return nil, fmt.Errorf("test client sentinel")
	}
	t.Cleanup(func() {
		getPrometheusClient = original
	})

	_, err := QueryMetricData(&model.MetricQuery{Metrics: []string{"legacy"}}, "1.4.0.0")
	assertAgentErrorKind(t, err, http.StatusBadRequest)
	if clientRequested {
		t.Fatal("Prometheus client was requested for an unsupported metric")
	}

	_, err = QueryMetricData(&model.MetricQuery{Metrics: []string{"legacy"}}, "")
	assertAgentErrorKind(t, err, http.StatusInternalServerError)
	if clientRequested {
		t.Fatal("Prometheus client was requested while seekdb version was unknown")
	}

	_, err = QueryMetricData(&model.MetricQuery{Metrics: []string{"public"}}, "")
	if err == nil || !clientRequested {
		t.Fatalf("unversioned metric should reach Prometheus client setup, called=%v err=%v", clientRequested, err)
	}
}

func assertAgentErrorKind(t *testing.T, err error, kind int) {
	t.Helper()
	agentErr, ok := err.(agenterrors.OcsAgentErrorInterface)
	if !ok {
		t.Fatalf("expected OcsAgentErrorInterface, got %T: %v", err, err)
	}
	if agentErr.ErrorCode().Kind != kind {
		t.Fatalf("error kind = %d, want %d: %v", agentErr.ErrorCode().Kind, kind, err)
	}
}
