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
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func GetTenantParameters(tenantName string, filter string) ([]oceanbase.GvObParameter, error) {
	if filter == "" {
		filter = "%"
	}
	parameters, err := tenantService.GetTenantParameters(tenantName, filter)
	if err != nil {
		return nil, err
	}

	return parameters, nil
}

func GetTenantParameter(tenantName string, parameterName string) (*oceanbase.GvObParameter, error) {
	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, err
	}
	parameter, err := tenantService.GetTenantParameter(tenant.TenantID, parameterName)
	if err != nil {
		return nil, err
	}
	if parameter == nil {
		return nil, errors.Occur(errors.ErrObTenantParameterNotExist, parameterName)
	}
	return parameter, nil
}

func SetTenantParameters(tenantName string, parameters map[string]interface{}) error {
	if err := checkParameters(parameters); err != nil {
		return err
	}

	transferNumber(parameters)
	if err := tenantService.SetTenantParameters(tenantName, parameters); err != nil {
		return errors.Wrap(err, "set tenant parameters failed")
	}
	return nil
}

type SetTenantParamterTask struct {
	task.Task
	parameters map[string]interface{}
	tenantId   int
}

func newSetTenantParameterNode(parameters map[string]interface{}) *task.Node {
	subtask := newSetTenantParameterTask()
	ctx := task.NewTaskContext().SetParam(PARAM_TENANT_PARAMETER, parameters)
	return task.NewNodeWithContext(subtask, false, ctx)
}

func newSetTenantParameterTask() *SetTenantParamterTask {
	newTask := &SetTenantParamterTask{
		Task: *task.NewSubTask(TASK_NAME_SET_TENANT_PARAMETER),
	}
	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *SetTenantParamterTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_PARAMETER, &t.parameters); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return err
	}

	transferNumber(t.parameters)
	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	t.ExecuteLogf("Set tenant parameter '%v' for tenant '%s'", t.parameters, tenantName)
	return tenantService.SetTenantParameters(tenantName, t.parameters)
}
