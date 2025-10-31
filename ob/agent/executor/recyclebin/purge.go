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

package recyclebin

import (
	"time"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/pool"
)

func PurgeRecyclebinTenant(name string) (*task.DagDetailDTO, error) {
	objectName, err := tenantService.GetRecycledTenantObjectName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "Check tenant '%s' exist in recyclebin failed", name)
	} else if objectName == "" {
		return nil, nil
	}

	// Get all resource pool.
	originalTenantId, err := tenantService.GetTenantId(objectName)
	if err != nil {
		return nil, errors.Wrapf(err, "Get tenant id of '%s' failed", name)
	}
	resourcePools, err := tenantService.GetTenantResourcePoolNames(originalTenantId)
	if err != nil {
		return nil, errors.Wrapf(err, "Get resource pools of tenant '%s' failed", name)
	}

	if err := tenantService.PurgeTenant(objectName); err != nil {
		return nil, errors.Wrapf(err, "Purge tenant '%s' failed", name)
	}

	template := task.NewTemplateBuilder(DAG_WAIT_PURGE_TENANT_FINISHED).
		SetMaintenance(task.TenantMaintenance(objectName)).
		AddTask(newWaitForPurgeFinishedTask(), false).
		AddTask(pool.NewDropResourcePoolTask(), false).Build()
	context := task.NewTaskContext().
		SetParam(PARAM_OBECJT_NAME, objectName).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true).
		SetData(PARAM_DROP_RESOURCE_POOL_LIST, resourcePools)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

type WaitForPurgeFinishedTask struct {
	task.Task
}

func newWaitForPurgeFinishedTask() *WaitForPurgeFinishedTask {
	newTask := &WaitForPurgeFinishedTask{
		Task: *task.NewSubTask(TASK_NAME_WAIT_PURGE_TENANT_FINISHED),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanRollback()
	return newTask
}

func (t *WaitForPurgeFinishedTask) Execute() error {
	var name string
	if err := t.GetContext().GetParamWithValue(PARAM_OBECJT_NAME, &name); err != nil {
		return err
	}
	t.ExecuteLogf("Wait for tenant %s purge finished", name)
	for {
		t.TimeoutCheck()
		if exist, err := tenantService.IsTenantExist(name); err != nil {
			return errors.Wrapf(err, "Check tenant '%s' exist failed", name)
		} else if !exist {
			break
		}
		time.Sleep(constant.CHECK_TENANT_EXIST_INTERVAL)
	}
	return nil
}
