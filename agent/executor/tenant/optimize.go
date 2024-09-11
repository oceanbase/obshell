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

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/param"
)

type OptimizeTenantTask struct {
	task.Task
	tenantId               int
	template               string
	createTenantVariables  map[string]interface{}
	createTenantParameters map[string]interface{}
}

func newOptimizeTenantTask() *OptimizeTenantTask {
	newTask := &OptimizeTenantTask{
		Task: *task.NewSubTask(TASK_NAME_OPTIMIZE_TENANT),
	}
	newTask.SetCanRollback().SetCanRetry().SetCanCancel().SetCanContinue().SetCanPass()
	return newTask
}

func newOptimizeTenantNode(template string, createTenantParam *param.CreateTenantParam) *task.Node {
	context := task.NewTaskContext().
		SetParam(PARAM_OPTIMIZE_TENANT, template).
		SetParam(PARAM_CREATE_TENANT_VARIABLES, createTenantParam.Variables).
		SetParam(PARAM_CREATE_TENANT_PARAMETERS, createTenantParam.Parameters)
	return task.NewNodeWithContext(newOptimizeTenantTask(), false, context)
}

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

func (t *OptimizeTenantTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	if err := t.GetContext().GetParamWithValue(PARAM_OPTIMIZE_TENANT, &t.template); err != nil {
		return errors.Wrap(err, "Get template failed")
	}
	if err := t.GetContext().GetParamWithValue(PARAM_CREATE_TENANT_VARIABLES, &t.createTenantVariables); err != nil {
		return errors.Wrap(err, "Get create tenant variables failed")
	}
	if err := t.GetContext().GetParamWithValue(PARAM_CREATE_TENANT_PARAMETERS, &t.createTenantParameters); err != nil {
		return errors.Wrap(err, "Get create tenant parameters failed")
	}

	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}

	variables, err := parseTemplate(VARIABLES_TEMPLATE, path.ObshellDefaultVariablePath(), t.template)
	if err != nil {
		return errors.Wrap(err, "Parse variable template failed")
	}
	for key := range t.createTenantVariables {
		delete(variables, key)
	}
	transferNumber(variables)
	t.ExecuteLogf("optimize variables: %v\n", variables)

	parameters, err := parseTemplate(PARAMETERS_TEMPLATE, path.ObshellDefaultParameterPath(), t.template)
	if err != nil {
		return errors.Wrap(err, "Parse parameter template failed")
	}
	for key := range t.createTenantParameters {
		delete(parameters, key)
	}
	transferNumber(parameters)
	t.ExecuteLogf("optimize parameters: %v\n", parameters)

	if err = tenantService.SetTenantVariables(tenantName, variables); err != nil {
		return errors.Wrap(err, "Set tenant variables failed")
	}

	if err = tenantService.SetTenantParameters(tenantName, parameters); err != nil {
		return errors.Wrap(err, "Set tenant parameters failed")
	}
	return nil
}
