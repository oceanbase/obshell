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
	"time"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/pool"
	"github.com/oceanbase/obshell/agent/executor/zone"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
)

func checkScaleOutTenantReplicasParam(tenant *oceanbase.DbaObTenant, param *param.ScaleOutTenantReplicasParam) error {
	if err := zone.CheckZoneParams(param.ZoneList); err != nil {
		return err
	}

	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenant.TenantID)
	if err != nil {
		return err
	}
	// Check whether there is already has a replica in the zone
	for _, zone := range param.ZoneList {
		if _, ok := replicaInfoMap[zone.Name]; ok {
			return errors.Occur(errors.ErrObTenantZoneAlreadyHasReplica, zone.Name)
		}
	}

	currentTenantUnitNum, err := tenantService.GetTenantUnitNum(tenant.TenantID)
	if err != nil {
		return err
	}
	if currentTenantUnitNum != param.ZoneList[0].UnitNum {
		return errors.Occur(errors.ErrObTenantUnitNumInconsistent)
	}

	for _, zone := range param.ZoneList {
		replicaInfoMap[zone.Name] = zone.ReplicaType
		// Check if tenant already have a pool located in the zone
		if exist, err := tenantService.CheckTenantHasPoolOnZone(tenant.TenantID, zone.Name); err != nil {
			return err
		} else if exist {
			return errors.Occur(errors.ErrObTenantHasPoolOnZone, zone.Name)
		}
	}

	if err := zone.CheckFirstPriorityPrimaryZoneChangedWhenAlterLocality(tenant, buildLocality(replicaInfoMap)); err != nil {
		return err
	}

	if err = zone.CheckPrimaryZoneAndLocality(tenant.PrimaryZone, replicaInfoMap); err != nil {
		return err
	}

	return nil
}

func ScaleOutTenantReplicas(tenantName string, param *param.ScaleOutTenantReplicasParam) (*task.DagDetailDTO, error) {
	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, err
	}

	zone.RenderZoneParams(param.ZoneList)
	if err := checkScaleOutTenantReplicasParam(tenant, param); err != nil {
		return nil, err
	}

	if err := CheckResourceEnough(param.ZoneList); err != nil {
		return nil, err
	}

	// Create 'Scale out tenant replicas' dag instance.
	template := buildScaleoutTenantReplicasDagTemplate(tenant, param)
	context := buildScaleoutTenantReplicasDagContext(tenant)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildScaleoutTenantReplicasDagTemplate(tenant *oceanbase.DbaObTenant, replicaParam *param.ScaleOutTenantReplicasParam) *task.Template {
	templateBuild := task.NewTemplateBuilder(DAG_SCALE_OUT_TENANT_REPLICA).
		SetMaintenance(task.TenantMaintenance(tenant.TenantName)).
		AddNode(newBatchCreateResourcePoolNode(replicaParam.ZoneList))
	for _, zone := range replicaParam.ZoneList {
		templateBuild.AddNode(newAlterLocalityNode(tenant.TenantID, SCALE_OUT_REPLICA, zone.Name, zone.ReplicaType))
	}
	return templateBuild.Build()
}

func buildScaleoutTenantReplicasDagContext(tenant *oceanbase.DbaObTenant) *task.TaskContext {
	return task.NewTaskContext().
		SetParam(PARAM_TENANT_ID, tenant.TenantID).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
}

type BatchCreateResourcePoolTask struct {
	task.Task
	createResourcePoolParam []param.CreateResourcePoolTaskParam
	tenantId                int
	timestamp               int64 // use for pool name
	zoneParam               []param.ZoneParam
}

func newBatchCreateResourcePoolTask() *BatchCreateResourcePoolTask {
	newTask := &BatchCreateResourcePoolTask{
		Task: *task.NewSubTask(TASK_NAME_CREATE_AND_ATTACH_RESOURCE_POOL),
	}
	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel()
	return newTask
}

func newBatchCreateResourcePoolNode(param []param.ZoneParam) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(PARAM_TIMESTAMP, time.Now().Unix()).SetParam(PARAM_ZONE_PARAM, param)
	return task.NewNodeWithContext(newBatchCreateResourcePoolTask(), false, ctx)
}

func mergePoolList(tenantPoolList []string, pool string) []string {
	if !utils.ContainsString(tenantPoolList, pool) {
		tenantPoolList = append(tenantPoolList, pool)
	}
	return tenantPoolList
}

func cullPoolList(tenantPoolList []string, pool string) []string {
	var result []string
	for _, p := range tenantPoolList {
		if p != pool {
			result = append(result, p)
		}
	}
	return result
}

func (t *BatchCreateResourcePoolTask) detachResourcePools() error {
	tenantPoolList, err := tenantService.GetTenantResourcePoolNames(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant resource pool list failed.")
	}
	// detach
	for _, pool := range t.createResourcePoolParam {
		targetPoolList := cullPoolList(tenantPoolList, pool.PoolName)
		t.ExecuteLogf("Modify tenant %d resource pool list from %v to %v", t.tenantId, tenantPoolList, targetPoolList)
		if err := tenantService.AlterResourcePoolList(t.tenantId, targetPoolList); err != nil {
			return err
		}
		tenantPoolList = targetPoolList
	}
	return nil
}

func (t *BatchCreateResourcePoolTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TIMESTAMP, &t.timestamp); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_ZONE_PARAM, &t.zoneParam); err != nil {
		return err
	}

	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed.")
	}
	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(tenantName, t.zoneParam, t.timestamp)

	tenantPoolList, err := tenantService.GetTenantResourcePoolNames(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant resource pool list failed.")
	}
	if err := pool.CreatePools(t.Task, t.createResourcePoolParam); err != nil {
		return err
	}
	for _, p := range t.createResourcePoolParam {
		targetPoolList := mergePoolList(tenantPoolList, p.PoolName)
		t.ExecuteLogf("Modify tenant %s resource pool list from %v to %v", tenantName, tenantPoolList, targetPoolList)
		if err := tenantService.AlterResourcePoolList(t.tenantId, targetPoolList); err != nil {
			// detach and drop
			if err := t.detachResourcePools(); err != nil {
				t.ExecuteWarnLog(errors.Wrap(err, "Detach resource pool failed."))
			} else if err := pool.DropFreeResourcePools(t.Task, t.createResourcePoolParam); err != nil {
				t.ExecuteWarnLog(errors.Wrap(err, "Drop created resource pool failed."))
			}
			return errors.Wrap(err, "Modify tenant resource pool failed.")
		}
		tenantPoolList = targetPoolList
	}
	return nil
}

func (t *BatchCreateResourcePoolTask) Rollback() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TIMESTAMP, &t.timestamp); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_ZONE_PARAM, &t.zoneParam); err != nil {
		return err
	}

	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed.")
	}
	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(tenantName, t.zoneParam, t.timestamp)

	// detach and drop
	if err := t.detachResourcePools(); err != nil {
		return errors.Wrap(err, "Detach resource pool failed.")
	}
	if err := pool.DropFreeResourcePools(t.Task, t.createResourcePoolParam); err != nil {
		return errors.Wrap(err, "Drop created resource pool failed.")
	}
	return nil
}

func scaleOutLocality(tenantId int, zone string, localityType string) (map[string]string, error) {
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenantId)
	if err != nil {
		return nil, err
	}
	if _, ok := replicaInfoMap[zone]; ok {
		if replicaInfoMap[zone] != localityType {
			return nil, errors.Occur(errors.ErrObTenantZoneAlreadyHasReplica, zone)
		} else {
			return nil, nil
		}
	}
	replicaInfoMap[zone] = localityType
	return replicaInfoMap, nil
}
