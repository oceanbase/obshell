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
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	tenantservice "github.com/oceanbase/obshell/agent/service/tenant"

	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
)

type ModifyReplicaOption struct {
	replicaTypeChanged map[string]string
	unitNumChanged     int
	unitConfChanged    map[string]string
	needSplitedPools   []string
}

func checkModifyReplicaZoneParams(tenant *oceanbase.DbaObTenant, param []param.ModifyReplicaZoneParam, replicaNum int) error {
	nums := 0    // The number of zones need to modify unit num in param.
	unitNum := 0 // Target unit num will be changed to.
	existZones := make([]string, 0)
	for _, zone := range param {
		if utils.ContainsString(existZones, zone.Name) {
			return errors.Errorf("Zone %s is repeated.", zone.Name)
		}

		// Check replica type.
		if zone.ReplicaType != nil {
			if err := checkReplicaType(*zone.ReplicaType); err != nil {
				return err
			}
		}

		// Check unit num.
		if zone.UnitNum != nil {
			if *zone.UnitNum <= 0 {
				return errors.New("unit_num should be positive.")
			}
			servers, err := obclusterService.GetServerByZone(zone.Name)
			if err != nil {
				return err
			}
			if len(servers) < *zone.UnitNum {
				return errors.Errorf("The number of servers in zone %s is %d, less than the number of units %d.", zone.Name, len(servers), *zone.UnitNum)
			}
			if *zone.UnitNum != unitNum && unitNum != 0 {
				return errors.New("unit_num should be same in all zones.")
			}
			unitNum = *zone.UnitNum
			nums += 1
		}

		if zone.UnitConfigName != nil {
			// Check unit config if exsits.
			if exist, err := unitService.IsUnitConfigExist(*zone.UnitConfigName); err != nil {
				return err
			} else if !exist {
				return errors.Errorf("Resource unit config '%s' is not exist.", *zone.UnitConfigName)
			}
		}
	}

	if unitNum != 0 && nums != replicaNum {
		return errors.New("Could not modify unit num partially.")
	}

	currentUnitNum, err := tenantService.GetTenantUnitNum(tenant.Id)
	if err != nil {
		return err
	}
	if unitNum != 0 && unitNum != currentUnitNum {
		// Check if enable_rebalance is true.
		if enableRebalance, err := tenantService.GetTenantParameter(tenant.Id, "enable_rebalance"); err != nil {
			return err
		} else {
			if enableRebalance == nil {
				return errors.New("Get enable_rebalance failed.")
			} else if enableRebalance.Value != "True" {
				return errors.New("Could not modify unit num when enable_rebalance is false.")
			}
		}
	}

	return nil
}

func getUnitName(zoneName string, poolList []oceanbase.DbaOBResourcePool) (string, error) {
	for _, pool := range poolList {
		zones := buildZoneList(pool.ZoneList)
		if utils.ContainsString(zones, zoneName) {
			return unitService.GetUnitConfigNameById(pool.UnitConfigId)
		}
	}
	return "", nil
}

func buildModifyReplicaOptions(tenant *oceanbase.DbaObTenant, param []param.ModifyReplicaZoneParam) (*ModifyReplicaOption, error) {
	options := &ModifyReplicaOption{
		replicaTypeChanged: make(map[string]string),
		unitConfChanged:    make(map[string]string),
	}

	poolList, err := tenantService.GetTenantResourcePool(tenant.Id)
	if err != nil {
		return nil, err
	}
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenant.Id)
	if err != nil {
		return nil, err
	}
	for _, zone := range param {
		if zone.ReplicaType != nil {
			if replicaInfoMap[zone.Name] != *zone.ReplicaType { // replicaInfoMap must contain zone.Name, has been checked
				options.replicaTypeChanged[zone.Name] = *zone.ReplicaType
			}
		}
		if zone.UnitNum != nil {
			currentUnitNum, err := tenantService.GetTenantUnitNum(tenant.Id)
			if err != nil {
				return nil, err
			}
			// Check if unit num changed.
			if currentUnitNum != *zone.UnitNum {
				options.unitNumChanged = *zone.UnitNum
			}
		}
		if zone.UnitConfigName != nil {
			if unitName, err := getUnitName(zone.Name, poolList); err != nil {
				return nil, err
			} else if unitName != *zone.UnitConfigName {
				options.unitConfChanged[zone.Name] = *zone.UnitConfigName
			}
		}
	}

	// build option for splite resource pool
	for _, pool := range poolList {
		zones := buildZoneList(pool.ZoneList)
		var targetUnitConfig string
		var num int
		var needSplit bool
		for _, zone := range zones {
			if _, ok := options.unitConfChanged[zone]; ok {
				num++
				if targetUnitConfig == "" {
					targetUnitConfig = options.unitConfChanged[zone]
				}
				if targetUnitConfig != options.unitConfChanged[zone] {
					needSplit = true
					break
				}
			}
		}
		if num != 0 && num != len(zones) {
			needSplit = true
		}
		if needSplit {
			options.needSplitedPools = append(options.needSplitedPools, pool.Name)
		}
	}

	return options, nil
}

func renderModifyReplicasParam(param *param.ModifyReplicasParam) {
	for i := range param.ZoneList {
		if param.ZoneList[i].ReplicaType != nil && *param.ZoneList[i].ReplicaType == "" {
			*param.ZoneList[i].ReplicaType = strings.ToUpper(*param.ZoneList[i].ReplicaType)
		}
	}
}

func checkModifyTenantReplicasParam(tenant *oceanbase.DbaObTenant, modifyReplicasParam *param.ModifyReplicasParam) error {
	if modifyReplicasParam.ZoneList == nil || len(modifyReplicasParam.ZoneList) == 0 {
		return errors.New("zone_list is empty.")
	}

	renderModifyReplicasParam(modifyReplicasParam)

	// Check whether there is already has a replica in the zone
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenant.Id)
	if err != nil {
		return err
	}
	for _, zone := range modifyReplicasParam.ZoneList {
		if _, ok := replicaInfoMap[zone.Name]; !ok {
			return errors.Errorf("zone '%s' does not have a replica", zone.Name)
		}
	}

	if err := checkModifyReplicaZoneParams(tenant, modifyReplicasParam.ZoneList, len(replicaInfoMap)); err != nil {
		return err
	}

	if err := checkModifyLocalityValid(replicaInfoMap, modifyReplicasParam.ZoneList); err != nil {
		return err
	}

	primaryZone, err := tenantService.GetTenantPrimaryZone(tenant.Id)
	if err != nil {
		return err
	}

	// build new replica info map
	for _, zone := range modifyReplicasParam.ZoneList {
		if zone.ReplicaType != nil {
			replicaInfoMap[zone.Name] = *zone.ReplicaType
		}
	}
	if err = checkPrimaryZoneAndLocality(primaryZone, replicaInfoMap); err != nil {
		return err
	}

	return nil
}

// this function will change replicaInfoMap, be carefull
func checkModifyLocalityValid(replicaInfoMap map[string]string, zoneList []param.ModifyReplicaZoneParam) error {
	var curPaxosNum, prePaxosNum int
	for _, replicaType := range replicaInfoMap {
		if replicaType == constant.FULL_REPLICA {
			prePaxosNum++
		}
	}
	curPaxosNum = prePaxosNum
	for _, zone := range zoneList {
		if zone.ReplicaType != nil {
			if *zone.ReplicaType == constant.FULL_REPLICA && replicaInfoMap[zone.Name] != constant.FULL_REPLICA { // replicaInfoMap must contain zone.Name, has been checked
				curPaxosNum++
			} else if *zone.ReplicaType == constant.READONLY_REPLICA && replicaInfoMap[zone.Name] == constant.FULL_REPLICA {
				curPaxosNum--
			}
		}
	}
	if curPaxosNum < 1 || curPaxosNum == 1 && prePaxosNum > 1 {
		return errors.New("violate locality principal not allowed")
	}
	return nil
}

func modifyLocality(tenantId int, zone string, replicaType string) (map[string]string, error) {
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenantId)
	if err != nil {
		return nil, err
	}
	if _, ok := replicaInfoMap[zone]; !ok {
		return nil, errors.Errorf("zone '%s' does not have a replica", zone)
	}
	replicaInfoMap[zone] = replicaType
	return replicaInfoMap, nil
}

func ModifyTenantReplica(tenantName string, param *param.ModifyReplicasParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	tenant, ocsErr := checkTenantExistAndStatus(tenantName)
	if ocsErr != nil {
		return nil, ocsErr
	}

	if err := checkModifyTenantReplicasParam(tenant, param); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err.Error())
	}

	options, err := buildModifyReplicaOptions(tenant, param.ZoneList)
	if err != nil {
		return nil, errors.Occur(errors.ErrBadRequest, "Build modify replica options failed: %s.", err.Error())
	}

	template, err := buildModifyReplicaTemplate(tenant, options)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Build modify replica template failed: %s.", err.Error())
	}
	if template.IsEmpty() {
		return nil, nil
	}
	context := task.NewTaskContext().
		SetParam(PARAM_TENANT_ID, tenant.Id).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "create '%s' dag instance failed", DAG_MODIFY_TENANT_REPLICA)
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildModifyReplicaTemplate(tenant *oceanbase.DbaObTenant, options *ModifyReplicaOption) (*task.Template, error) {
	templateBuilder := task.NewTemplateBuilder(DAG_MODIFY_TENANT_REPLICA).SetMaintenance(task.TenantMaintenance(tenant.Name))
	if options.replicaTypeChanged != nil && len(options.replicaTypeChanged) != 0 {
		// Modify 'FULL' replica to 'READONLY' replica firstly.
		for zone, replicaType := range options.replicaTypeChanged {
			if replicaType == constant.FULL_REPLICA {
				templateBuilder.AddNode(newAlterLocalityNode(tenant.Id, MODIFY_REPLICA_TYPE, zone, replicaType))
			}
		}
		// Modify 'READONLY' replica to 'FULL' replica secondly.
		for zone, replicaType := range options.replicaTypeChanged {
			if replicaType == constant.READONLY_REPLICA {
				templateBuilder.AddNode(newAlterLocalityNode(tenant.Id, MODIFY_REPLICA_TYPE, zone, replicaType))
			}
		}

	}

	if len(options.needSplitedPools) != 0 {
		templateBuilder.AddNode(newSplitResourcePoolNode(options.needSplitedPools))
	}

	if len(options.unitConfChanged) != 0 {
		templateBuilder.AddNode(newAlterResourcePoolUnitConfNode(options.unitConfChanged))
	}

	if options.unitNumChanged != 0 {
		ctx := task.NewTaskContext().SetParam(PARAM_TENANT_UNIT_NUM, options.unitNumChanged)
		templateBuilder.AddNode(task.NewNodeWithContext(newAlterResourcePoolUnitNumTask(), false, ctx))
	}
	return templateBuilder.Build(), nil
}

type AlterResourcePoolUnitConfTask struct {
	task.Task
	tenantId         int
	zoneWithUnitConf map[string]string
}

func newAlterResourcePoolUnitConfNode(zoneWithUnitConf map[string]string) *task.Node {
	ctx := task.NewTaskContext().SetParam(PARAM_ZONE_WITH_UNIT, zoneWithUnitConf)
	return task.NewNodeWithContext(newAlterResourcePoolUnitConfTask(), false, ctx)
}

func newAlterResourcePoolUnitConfTask() *AlterResourcePoolUnitConfTask {
	newTask := &AlterResourcePoolUnitConfTask{
		Task: *task.NewSubTask(TASK_NAME_ALTER_RESOURCE_POOL_UNIT_CONF),
	}
	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *AlterResourcePoolUnitConfTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	if err := t.GetContext().GetParamWithValue(PARAM_ZONE_WITH_UNIT, &t.zoneWithUnitConf); err != nil {
		return errors.Wrap(err, "Get tenant zone name failed")
	}
	poolInfo, err := tenantService.GetTenantResourcePool(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant resource pool info failed.")
	}
	for _, pool := range poolInfo {
		zones := buildZoneList(pool.ZoneList)
		for zone, unitName := range t.zoneWithUnitConf {
			if utils.ContainsString(zones, zone) {
				if err := tenantService.AlterResourcePoolUnitConfig(pool.Name, unitName); err != nil {
					return errors.Wrap(err, "Alter resource pool unit configuration failed.")
				} else {
					t.ExecuteInfoLogf("Alter resource pool '%s' unit config to '%s' succeed.", pool.Name, unitName)
					break
				}
			}
		}
	}
	return nil
}

type AlterResourcePoolUnitNumTask struct {
	task.Task
	tenantId int
	unitNum  int
}

func newAlterResourcePoolUnitNumTask() *AlterResourcePoolUnitNumTask {
	newTask := &AlterResourcePoolUnitNumTask{
		Task: *task.NewSubTask(TASK_NAME_ALTER_RESOURCE_POOL_UNIT_NUM),
	}
	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func waitAlterTenantUnitNumSucceed(tenantId int, targetUnitNum int) error {
	tenantName, err := tenantService.GetTenantName(tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed.")
	}
	jobId, err := tenantService.GetTargetTenantJob(constant.ALTER_RESOURCE_TENANT_UNIT_NUM, tenantId, fmt.Sprintf(tenantservice.SQL_ALTER_TENANT_UNIT_NUM, tenantName, targetUnitNum))
	if err != nil {
		return errors.Wrap(err, "Get tenant job failed.")
	}

	if jobId == 0 {
		return errors.Occur(errors.ErrUnexpected, "There is no job for altering tenant %s unit num to %d", tenantName, targetUnitNum)
	} else {
		return waitTenantJobSucceed(jobId)
	}
}

func (t *AlterResourcePoolUnitNumTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_UNIT_NUM, &t.unitNum); err != nil {
		return errors.Wrap(err, "Get tenant unit num failed")
	}

	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed.")
	}

	currentUnitName, err := tenantService.GetTenantUnitNum(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant current unit num failed")
	} else if currentUnitName == t.unitNum {
		t.ExecuteLogf("Tenant '%s' unit num is already %d, skip.", tenantName, t.unitNum)
		return nil
	}

	if jobBo, err := tenantService.GetInProgressTenantJobBo(constant.ALTER_RESOURCE_TENANT_UNIT_NUM, t.tenantId); err != nil {
		return errors.Wrap(err, "Get in progress tenant job failed")
	} else if jobBo != nil {
		if jobBo.TargetIs(t.unitNum) {
			if err := waitTenantJobSucceed(jobBo.JobId); err != nil {
				return errors.Wrap(err, "Wait for alter tenant unit num succeed failed")
			}
		} else {
			t.ExecuteErrorLogf("There already exists a inprogress job alter unit num to %d", t.unitNum)
			return errors.Errorf("There already exists a inprogress job alter unit num to %d", t.unitNum)
		}
	} else {
		t.ExecuteLogf("Alter tenant '%s' unit num to %d", tenantName, t.unitNum)
		if err := tenantService.AlterTenantUnitNum(tenantName, t.unitNum); err != nil {
			return errors.Wrap(err, "Alter tenant unit num failed.")
		}
		// Wait for task execute successfully
		if err := waitAlterTenantUnitNumSucceed(t.tenantId, t.unitNum); err != nil {
			return errors.Wrap(err, "Wait for alter tenant unit num succeed failed.")
		}
	}
	return nil
}
