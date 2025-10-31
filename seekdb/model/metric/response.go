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

import "github.com/oceanbase/obshell/seekdb/model/common"

type MetricClass struct {
	Name         string        `json:"name" yaml:"name" binding:"required"`
	Description  string        `json:"description" yaml:"description" binding:"required"`
	MetricGroups []MetricGroup `json:"metric_groups" yaml:"metricGroups" binding:"required"`
}

type MetricGroup struct {
	Name        string       `json:"name" yaml:"name" binding:"required"`
	Description string       `json:"description" yaml:"description" binding:"required"`
	Metrics     []MetricMeta `json:"metrics" yaml:"metrics" binding:"required"`
}

type MetricMeta struct {
	Name        string `json:"name" yaml:"name" binding:"required"`
	Unit        string `json:"unit" yaml:"unit" binding:"required"`
	Description string `json:"description" yaml:"description" binding:"required"`
	Key         string `json:"key" yaml:"key" binding:"required"`
}

type Metric struct {
	Name   string          `json:"name" yaml:"name"`
	Labels []common.KVPair `json:"labels" yaml:"labels"`
}

type MetricValue struct {
	Value     float64 `json:"value" yaml:"value" binding:"required"`
	Timestamp float64 `json:"timestamp" yaml:"timestamp" binding:"required"`
}

type MetricData struct {
	Metric Metric        `json:"metric" yaml:"metric" binding:"required"`
	Values []MetricValue `json:"values" yaml:"values" binding:"required"`
}
