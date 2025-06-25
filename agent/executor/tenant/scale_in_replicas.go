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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/zone"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
)

func checkScaleInTenantReplicasParam(tenant *oceanbase.DbaObTenant, param *param.ScaleInTenantReplicasParam) error {
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenant.TenantID)
	if err != nil {
		return err
	}
	if len(replicaInfoMap) == 1 {
		return errors.Occur(errors.ErrObTenantReplicaOnlyOne, tenant.TenantName)
	}

	if err := checkScaleInLocalityValid(replicaInfoMap, param.Zones); err != nil {
		return err
	}

	primaryZone, err := tenantService.GetTenantPrimaryZone(tenant.TenantID)
	if err != nil {
		return err
	}

	// build new replica info map
	for zone := range replicaInfoMap {
		if utils.ContainsString(param.Zones, zone) {
			delete(replicaInfoMap, zone)
		}
	}

	if err := zone.CheckFirstPriorityPrimaryZoneChangedWhenAlterLocality(tenant, buildLocality(replicaInfoMap)); err != nil {
		return err
	}

	if err = zone.CheckPrimaryZoneAndLocality(primaryZone, replicaInfoMap); err != nil {
		return err
	}

	return nil
}

// this function won't change replicaInfoMap, be carefull
func checkScaleInLocalityValid(replicaInfoMap map[string]string, zoneList []string) error {
	var curPaxosNum, prePaxosNum int
	for zone, replicaType := range replicaInfoMap {
		if replicaType == constant.REPLICA_TYPE_FULL {
			prePaxosNum++
			if !utils.ContainsString(zoneList, zone) {
				curPaxosNum++
			}
		}
	}
	if curPaxosNum > 1 || curPaxosNum == 1 && prePaxosNum == 1 {
		return nil
	}
	return errors.Occur(errors.ErrObTenantLocalityPrincipalNotAllowed) // tenant with arb service should only support 4->2
}

func scaleInLocality(tenantId int, zone string) (map[string]string, error) {
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenantId)
	if err != nil {
		return nil, err
	}
	if _, ok := replicaInfoMap[zone]; !ok {
		return nil, nil
	}
	delete(replicaInfoMap, zone)
	if len(replicaInfoMap) == 0 { // double check
		return nil, errors.Occur(errors.ErrObTenantReplicaDeleteAll)
	}
	return replicaInfoMap, nil
}

func filterZones(tenantId int, param *param.ScaleInTenantReplicasParam) error {
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenantId)
	if err != nil {
		return err
	}
	filterZones := make([]string, 0)
	for _, zone := range param.Zones {
		if _, ok := replicaInfoMap[zone]; ok {
			filterZones = append(filterZones, zone)
		}
	}
	param.Zones = filterZones
	return nil
}

func ScaleInTenantReplicas(tenantName string, param *param.ScaleInTenantReplicasParam) (*task.DagDetailDTO, error) {
	tenant, err := checkTenantExistAndStatus(tenantName)
	if err != nil {
		return nil, err
	}

	if err := filterZones(tenant.TenantID, param); err != nil {
		return nil, err
	}
	if len(param.Zones) == 0 {
		return nil, nil
	}

	if err := checkScaleInTenantReplicasParam(tenant, param); err != nil {
		return nil, err
	}

	// Create 'Scale in tenant replicas' dag instance.
	template, err := buildScaleInTenantReplicasDagTemplate(tenant, *param)
	if err != nil {
		return nil, err
	}
	context := buildScaleInTenantReplicasDagContext(tenant, *param)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildScaleInTenantReplicasDagTemplate(tenant *oceanbase.DbaObTenant, param param.ScaleInTenantReplicasParam) (*task.Template, error) {
	templateBuilder := task.NewTemplateBuilder(DAG_SCALE_IN_TENANT_REPLICA).SetMaintenance(task.TenantMaintenance(tenant.TenantName))

	for _, zone := range param.Zones {
		templateBuilder.AddNode(newAlterLocalityNode(tenant.TenantID, SCALE_IN_REPLICA, zone))
	}

	poolInfo, err := tenantService.GetTenantResourcePool(tenant.TenantID)
	if err != nil {
		return nil, errors.Wrap(err, "Get tenant resource pool info failed")
	}
	needSplitPools := make([]string, 0)
	for _, pool := range poolInfo {
		zones := buildZoneList(pool.ZoneList)
		for _, zone := range param.Zones {
			if len(zones) > 1 {
				if utils.ContainsString(zones, zone) {
					needSplitPools = append(needSplitPools, pool.Name)
					break
				}
			}
		}
	}
	if len(needSplitPools) != 0 {
		// Add split resource pool node to ensure the resource pool is split before dropping
		templateBuilder.AddNode(newSplitResourcePoolNode(needSplitPools))
	}

	templateBuilder.AddTask(newBatchDropResourcePoolTask(), false)
	return templateBuilder.Build(), nil
}

func buildScaleInTenantReplicasDagContext(tenant *oceanbase.DbaObTenant, param param.ScaleInTenantReplicasParam) *task.TaskContext {
	context := task.NewTaskContext().
		SetParam(PARAM_TENANT_ID, tenant.TenantID).
		SetParam(PARAM_ZONE_LIST, param.Zones).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	return context
}

type SplitResourcePoolTask struct {
	task.Task
	poolList  []string
	tenantId  int
	timestamp string
}

func newSplitResourcePoolNode(pools []string) *task.Node {
	ctx := task.NewTaskContext().
		SetParam(PARAM_SPLIT_RESOURCE_POOL_LIST, pools).
		SetParam(PARAM_TIMESTAMP, fmt.Sprint(time.Now().Unix()))
	return task.NewNodeWithContext(newSplitResourcePoolTask(), false, ctx)
}

func newSplitResourcePoolTask() *SplitResourcePoolTask {
	newTask := &SplitResourcePoolTask{
		Task: *task.NewSubTask(TASK_NAME_SPLIT_RESOURCE_POOL),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *SplitResourcePoolTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_SPLIT_RESOURCE_POOL_LIST, &t.poolList); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TIMESTAMP, &t.timestamp); err != nil {
		return err
	}

	poolInfos, err := tenantService.GetTenantResourcePool(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant resource pool list failed.")
	}

	for _, pool := range poolInfos {
		// get resource pool name
		if utils.ContainsString(t.poolList, pool.Name) {
			zones := buildZoneList(pool.ZoneList)
			// splite
			t.splitResourcePool(t.tenantId, pool.Name, zones, t.timestamp)
		}
	}
	return nil
}

func (t *SplitResourcePoolTask) splitResourcePool(tenantId int, poolName string, zoneList []string, timestamp string) error {
	targetPoolList := make([]string, 0)
	if len(zoneList) > 1 {
		// gen pool name
		for _, z := range zoneList {
			targetPoolList = append(targetPoolList, strings.Join([]string{strconv.Itoa(tenantId), z, timestamp}, "_"))
		}
		// splite
		t.ExecuteLogf("Split resource pool %s on %v", poolName, zoneList)
		if err := tenantService.SplitResourcePool(poolName, targetPoolList, zoneList); err != nil {
			return errors.Wrap(err, "Split resource pool failed.")
		}
	}
	return nil
}

type BatchDropResourcePoolTask struct {
	task.Task
	zoneList []string
	tenantId int
}

func newBatchDropResourcePoolTask() *BatchDropResourcePoolTask {
	newTask := &BatchDropResourcePoolTask{
		Task: *task.NewSubTask(TASK_NAME_DROP_RESOURCE_POOL),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func buildZoneList(zoneList string) []string {
	return strings.Split(zoneList, ";")
}

func (t *BatchDropResourcePoolTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_ZONE_LIST, &t.zoneList); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return err
	}

	poolInfos, err := tenantService.GetTenantResourcePool(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant resource pool list failed.")
	}

	tenantPoolList, err := tenantService.GetTenantResourcePoolNames(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant resource pool list failed.")
	}

	for _, zone := range t.zoneList {
		for _, pool := range poolInfos {
			// pool.ZoneList must only contain one zone.
			// beacuse we have splited the pool before.
			if zone == pool.ZoneList {
				t.ExecuteLogf("Detach and drop resource pool %s", pool.ZoneList)
				targetPoolList := cullPoolList(tenantPoolList, pool.Name)
				// detach resource pool
				if err := tenantService.AlterResourcePoolList(t.tenantId, targetPoolList); err != nil {
					return err
				}
				// drop resource pool
				if err := tenantService.DropResourcePool(pool.Name, true); err != nil {
					return errors.Wrap(err, "Drop resource pool failed.")
				}
				tenantPoolList = targetPoolList
			}
		}
	}
	return nil
}
