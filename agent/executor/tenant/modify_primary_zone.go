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
)

func checkModifyPrimaryZoneParam(tenant *oceanbase.DbaObTenant, param *param.ModifyTenantPrimaryZoneParam) *errors.OcsAgentError {
	var zoneList = make([]string, 0)
	replicaInfoMap, err := tenantService.GetTenantReplicaInfoMap(tenant.TenantID)
	if err != nil {
		return errors.Occur(errors.ErrUnexpected, err.Error())
	}
	for zone := range replicaInfoMap {
		zoneList = append(zoneList, zone)
	}

	if err := checkPrimaryZone(*param.PrimaryZone, zoneList); err != nil {
		return errors.Occur(errors.ErrIllegalArgument, err.Error())
	}

	if err := checkPrimaryZoneAndLocality(*param.PrimaryZone, replicaInfoMap); err != nil {
		return errors.Occur(errors.ErrIllegalArgument, err.Error())
	}
	return nil
}

func renderModifyTenantPrimaryZone(param *param.ModifyTenantPrimaryZoneParam) {
	if param.PrimaryZone != nil && *param.PrimaryZone == "" {
		*param.PrimaryZone = constant.PRIMARY_ZONE_RANDOM
	}
	if strings.ToUpper(*param.PrimaryZone) == constant.PRIMARY_ZONE_RANDOM {
		*param.PrimaryZone = constant.PRIMARY_ZONE_RANDOM
	}
}

func ModifyTenantPrimaryZone(tenantName string, param *param.ModifyTenantPrimaryZoneParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	tenant, ocsErr := checkTenantExistAndStatus(tenantName)
	if ocsErr != nil {
		return nil, ocsErr
	}

	renderModifyTenantPrimaryZone(param)

	if err := checkModifyPrimaryZoneParam(tenant, param); err != nil {
		return nil, err
	}

	if err := tenantService.AlterTenantPrimaryZone(tenant.TenantName, *param.PrimaryZone); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}

	template := task.NewTemplateBuilder(DAG_MODIFY_TENANT_PRIMARY_ZONE).
		SetMaintenance(task.TenantMaintenance(tenantName)).
		AddTask(newModifyPrimaryZoneTask(), false).Build()
	context := task.NewTaskContext().
		SetParam(PARAM_TENANT_ID, tenant.TenantID).
		SetParam(PARAM_PRIMARY_ZONE, param.PrimaryZone).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "create '%s' dag instance failed", DAG_MODIFY_TENANT_PRIMARY_ZONE)
	}
	return task.NewDagDetailDTO(dag), nil
}

type ModifyPrimaryZoneTask struct {
	task.Task
	tenantId    int
	primaryZone string
}

func newModifyPrimaryZoneTask() *ModifyPrimaryZoneTask {
	newTask := &ModifyPrimaryZoneTask{
		Task: *task.NewSubTask(TASK_NAME_MODIFY_PRIMARY_ZONE),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback().SetCanPass()
	return newTask
}

func waitAlterPrimaryZoneSucceed(tenantId int, targetPrimaryZone string) error {
	tenantName, err := tenantService.GetTenantName(tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed.")
	}
	jobId, err := tenantService.GetTargetTenantJob(constant.ALTER_TENANT_PRIMARY_ZONE, tenantId, fmt.Sprintf(tenantservice.SQL_ALTER_TENANT_PRIMARY_ZONE, tenantName, targetPrimaryZone))
	if err != nil {
		return errors.Wrap(err, "Get tenant job failed.")
	}

	if jobId == 0 {
		return errors.Occurf(errors.ErrUnexpected, "There is no job for altering tenant %s primary zone to %s", tenantName, targetPrimaryZone)
	}
	return waitTenantJobSucceed(jobId)
}

func (t *ModifyPrimaryZoneTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	if err := t.GetContext().GetParamWithValue(PARAM_PRIMARY_ZONE, &t.primaryZone); err != nil {
		return errors.Wrap(err, "Get primary zone failed")
	}
	tenantName, err := tenantService.GetTenantName(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}

	// Check if the primary zone is already the target primary zone
	currPrimaryZone, err := tenantService.GetTenantPrimaryZone(t.tenantId)
	if err != nil {
		return errors.Wrap(err, "Get tenant primary zone failed")
	}
	parsedCurrPrimaryZoneList := tenantservice.ParsePrimaryZone(currPrimaryZone)
	parsedTargetPrimaryZoneList := tenantservice.ParsePrimaryZone(t.primaryZone)
	if len(parsedCurrPrimaryZoneList) >= len(parsedTargetPrimaryZoneList) && len(parsedCurrPrimaryZoneList)-len(parsedTargetPrimaryZoneList) <= 1 {
		for i, zonesStr := range parsedTargetPrimaryZoneList {
			if parsedCurrPrimaryZoneList[i] != zonesStr {
				break
			}
			if i == len(parsedTargetPrimaryZoneList)-1 {
				t.ExecuteLogf("Tenant %s primary zone is already %s", tenantName, t.primaryZone)
				return nil
			}
		}

	}

	if jobBo, err := tenantService.GetInProgressTenantJobBo(constant.ALTER_TENANT_PRIMARY_ZONE, t.tenantId); err != nil {
		return errors.Wrap(err, "Get in progress tenant job failed")
	} else if jobBo != nil {
		if jobBo.TargetIs(parsedTargetPrimaryZoneList) {
			if err := waitTenantJobSucceed(jobBo.JobId); err != nil {
				return errors.Wrap(err, "Wait for alter tenant primary zone succeed failed")
			}
		} else {
			t.ExecuteLogf("There already exists a inprogress job alter primary zone to %s", t.primaryZone)
			return errors.Errorf("There already exists a inprogress job alter primary zone to %s", t.primaryZone)
		}
	} else {
		t.ExecuteLogf("Alter tenant %s primary zone to %s", tenantName, t.primaryZone)
		if err := tenantService.AlterTenantPrimaryZone(tenantName, t.primaryZone); err != nil {
			return errors.Occur(errors.ErrUnexpected, err.Error())
		}
		t.ExecuteLogf("Wait for tenant %s primary zone to be altered to %s", tenantName, t.primaryZone)
		if err := waitAlterPrimaryZoneSucceed(t.tenantId, t.primaryZone); err != nil {
			return err
		}
	}

	return nil
}