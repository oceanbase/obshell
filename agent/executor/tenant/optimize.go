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

	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/lib/path"
)

func getAllSupportedScenarios() (scenarios []string) {
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
			scenarios = append(scenarios, val.(string))
		}
	}
	return
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
