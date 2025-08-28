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

package tenant

import (
	"io"
	"os"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/lib/path"
)

var descriptionZhMap = map[string]string{
	"COMPLEX_OLTP": "适用于银行、保险系统等工作负载。它们通常具有复杂的联接、复杂的相关子查询、用 PL 编写的批处理作业，以及长事务和大事务。有时对短时间运行的查询使用并行执行",
	"OLAP":         "用于实时数据仓库分析场景",
	"HTAP":         "适用于混合 OLAP 和 OLTP 工作负载。通常用于从活动运营数据、欺诈检测和个性化建议中获取即时见解",
	"KV":           "用于键值工作负载和类似 hbase 的宽列工作负载，这些工作负载通常具有非常高的吞吐量并且对延迟敏感",
	"EXPRESS_OLTP": "适用于贸易、支付核心系统、互联网高吞吐量应用程序等工作负载。没有外键等限制，没有存储过程，没有长交易，没有大交易，没有复杂的连接，没有复杂的子查询",
}

var descriptionEnMap = map[string]string{
	"COMPLEX_OLTP": "for workloads like bank, insurance system. they often have complex join, complex correlated subquery, batch jobs written in PL, have both long and large transactions. Sometimes use parallel execution for short running queries",
	"OLAP":         "for real-time data warehouse analytics scenarios.",
	"HTAP":         "for mixed OLAP and OLTP workload. Typically utilized for obtaining instant insights from active operational data, fraud detection, and personalized recommendations",
	"KV":           "for key-value workloads and hbase-like wide-column workloads, which commonly experience very high throughput and are sensitive to latency",
	"EXPRESS_OLTP": "for workloads like trade, payment core system, internet high throughput application, etc. no restrictions like foreign key, no stored procedure, no long transaction, no large transaction, no complex join, no complex subquery",
}

type ParameterTemplate struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
}

func GetAllSupportedScenarios(language string) (res []ParameterTemplate) {
	res = make([]ParameterTemplate, 0)
	supportedScenarios := make(map[string]string)
	if _, err := os.Stat(path.ObshellDefaultVariablePath()); err != nil {
		return
	}
	// Read default template from file
	file, err := os.Open(path.ObshellDefaultVariablePath())
	if err != nil {
		return
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return
	}
	var data []map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return
	}

	var descriptionMap map[string]string
	switch language {
	case constant.LANGUAGE_EN_US:
		descriptionMap = descriptionEnMap
	case constant.LANGUAGE_ZH_CN:
		descriptionMap = descriptionZhMap
	default:
		descriptionMap = descriptionEnMap
	}
	for _, item := range data {
		var template map[string]interface{}
		data, err := json.Marshal(item)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(data, &template); err != nil {
			return nil
		}
		if val, ok := template["scenario"]; ok {
			if valStr, ok := val.(string); ok {
				if _, ok := descriptionMap[strings.ToUpper(valStr)]; ok {
					supportedScenarios[strings.ToUpper(valStr)] = descriptionMap[strings.ToUpper(valStr)]
				}
			}
		}
	}
	for scenario, description := range supportedScenarios {
		res = append(res, ParameterTemplate{
			Scenario:    scenario,
			Description: description,
		})
	}
	return res
}

func parseTemplate(templateType, filepath, scenario string) (map[string]interface{}, error) {
	if _, err := os.Stat(filepath); err != nil {
		return nil, err
	}
	// Read default template from file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var data []map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})

	for _, item := range data {
		var template map[string]interface{}
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &template); err != nil {
			return nil, err
		}

		val, ok := template["scenario"]
		if !(ok && val.(string) == scenario) {
			continue
		}

		for key, value := range template {
			if key == templateType {
				var paramters map[string]interface{}
				data, _ := json.Marshal(value)
				if err := json.Unmarshal(data, &paramters); err != nil {
					return nil, err
				}
				for key, value := range paramters {
					if key == "tenant" {
						var tenant []map[string]interface{}
						data, _ := json.Marshal(value)
						if err := json.Unmarshal(data, &tenant); err != nil {
							return nil, err
						}
						for _, data := range tenant {
							var k string
							var v interface{}
							for key, value := range data {
								if key == "name" {
									k = value.(string)
								} else if key == "value" {
									v = value
								}
							}
							res[k] = v
						}
					}
				}
			}
		}
		return res, nil
	}
	return res, nil
}
