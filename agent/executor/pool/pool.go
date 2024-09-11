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

package pool

import (
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func DropResourcePool(poolName string) *errors.OcsAgentError {
	pool, err := tenantService.GetResourcePoolByName(poolName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Get resource pool %s failed: %s", poolName, err.Error())
	} else if pool == nil {
		return nil
	} else if pool.TenantId != 0 {
		return errors.Occurf(errors.ErrBadRequest, "resource pool '%s' has already been granted to a tenant", poolName)
	}

	if err := tenantService.DropResourcePool(poolName, false); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Drop resource pool %s failed: %s", poolName, err.Error())
	}
	return nil
}

func GetAllResourcePools() ([]oceanbase.DbaObResourcePool, *errors.OcsAgentError) {
	pools, err := tenantService.GetAllResourcePool()
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Get all resource pool failed: %s", err.Error())
	}
	return pools, nil
}

type DropResourcePoolTask struct {
	task.Task
	resourcePools []string
}

// DropResourcePoolTask drop resource pools for `drop tenant`
func NewDropResourcePoolTask() *DropResourcePoolTask {
	newTask := &DropResourcePoolTask{
		Task: *task.NewSubTask(TASK_NAME_DROP_RESOURCE_POOL),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *DropResourcePoolTask) Execute() error {
	if err := t.GetContext().GetDataWithValue(PARAM_DROP_RESOURCE_POOL_LIST, &t.resourcePools); err != nil {
		return errors.New("Get resource pools failed.")
	}
	for _, pool := range t.resourcePools {
		t.ExecuteLogf("Drop resource pool %s", pool)
		if err := tenantService.DropResourcePool(pool, true); err != nil {
			return err
		}
	}
	return nil
}

func CreatePools(t task.Task, poolParam []param.CreateResourcePoolTaskParam) error {
	var createdResourcePool []param.CreateResourcePoolTaskParam
	for _, p := range poolParam {
		t.ExecuteLogf("Create resource pool: %v", p)
		if err := tenantService.CreateResourcePool(p.PoolName, p.UnitConfigName, p.UnitNum, []string{p.ZoneName}); err != nil {
			// drop all created resource pool
			if err := DropFreeResourcePools(t, createdResourcePool); err != nil {
				t.ExecuteWarnLog(errors.Wrapf(err, "Drop created resource pool failed"))
			}
			return errors.Wrapf(err, "Create resource pool %s failed", p.PoolName)
		}
		t.ExecuteLogf("Create resource pool %s success", p.PoolName)
		createdResourcePool = append(createdResourcePool, p)
	}
	return nil
}

func DropFreeResourcePools(t task.Task, param []param.CreateResourcePoolTaskParam) error {
	// drop all created resource pool
	for _, p := range param {
		t.ExecuteLogf("Drop resource pool: %v\n", p)
		if exist, err := tenantService.IsResourcePoolExistAndFreed(p.PoolName, p.UnitConfigName, p.UnitNum, p.ZoneName); err != nil {
			return err
		} else if exist {
			if err := tenantService.DropResourcePool(p.PoolName, false); err != nil {
				return err
			}
		} else {
			t.ExecuteWarnLogf("Resource pool %s is not exist or be attached, drop failed.", p.PoolName)
		}
	}
	return nil
}
