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
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/pool"
	"github.com/oceanbase/obshell/param"
)

func DropTenant(param *param.DropTenantParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if param.Name == constant.SYS_TENANT {
		return nil, errors.Occur(errors.ErrIllegalArgument, "Can't drop sys tenant.")
	}

	tenant, err := tenantService.GetTenantByName(param.Name)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Get tenant '%s' failed.", param.Name)
	}
	if tenant == nil {
		return nil, nil
	}
	if tenant.Status != NORMAL_TENANT {
		return nil, errors.Occurf(errors.ErrKnown, "Tenant '%s' status is '%s'.", param.Name, tenant.Status)
	}

	// Create 'Drop Tenant' dag instance.
	template := buildDropTenantDagTemplate(param)
	context := task.NewTaskContext().
		SetParam(PARAM_TENANT_ID, tenant.Id).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "create '%s' dag instance failed: %s", DAG_DROP_TENANT, err.Error())
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildDropTenantDagTemplate(param *param.DropTenantParam) *task.Template {
	templateBuild := task.NewTemplateBuilder(DAG_DROP_TENANT).SetMaintenance(task.TenantMaintenance(param.Name))
	if param.NeedRecycle != nil && *param.NeedRecycle {
		return templateBuild.AddTask(newRecycleTenantTask(), false).Build()
	}
	templateBuild.AddTask(newDropTenantTask(), false)
	templateBuild.AddTask(pool.NewDropResourcePoolTask(), false)
	return templateBuild.Build()
}

type RecycleTenantTask struct {
	task.Task
	tenantId int
}

func newRecycleTenantTask() *RecycleTenantTask {
	newTask := &RecycleTenantTask{
		Task: *task.NewSubTask(TASK_NAME_RECYCLE_TENANT),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *RecycleTenantTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return err
	}
	t.ExecuteLogf("Recycle tenant %s", tenantName)
	return tenantService.RecycleTenant(tenantName)
}

type DropTenantTask struct {
	task.Task
	id int
}

func newDropTenantTask() *DropTenantTask {
	newTask := &DropTenantTask{
		Task: *task.NewSubTask(TASK_NAME_DROP_TENANT),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *DropTenantTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.id); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	tenantName, err := tenantService.GetTenantName(t.id)
	if err != nil {
		return errors.New("Get tenant name failed.")
	}
	t.ExecuteLogf("Drop tenant %s", tenantName)
	// Get all resource pool
	resourcePools, err := tenantService.GetTenantResourcePoolNames(t.id)
	if err != nil {
		return errors.New("Get resource pool failed.")
	}
	t.ExecuteLogf("Resource pool list: %v", resourcePools)
	t.GetContext().SetData(PARAM_DROP_RESOURCE_POOL_LIST, resourcePools)
	return tenantService.DropTenant(tenantName)
}
