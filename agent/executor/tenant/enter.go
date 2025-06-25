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
	"strings"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/service/agent"
	"github.com/oceanbase/obshell/agent/service/obcluster"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
	"github.com/oceanbase/obshell/agent/service/tenant"
	"github.com/oceanbase/obshell/agent/service/unit"
)

var (
	tenantService      tenant.TenantService
	obclusterService   obcluster.ObclusterService
	clusterTaskService = taskservice.NewClusterTaskService()
	unitService        unit.UnitService
	agentService       agent.AgentService
)

const (
	// task param name
	PARAM_CREATE_TENANT                = "createTenant"
	PARAM_TENANT_VARIABLES             = "tenantVariables"
	PARAM_TENANT_NAME                  = "tenantName"
	PARAM_TENANT_TIME_ZONE             = "timeZone"
	PARAM_TENANT_ID                    = "tenantId"
	PARAM_ALTER_LOCALITY_TYPE          = "alterLocalityType"
	PARAM_TENANT_NEW_PASSWORD          = "newPassword"
	PARAM_TENANT_WHITELIST             = "whitelist"
	PARAM_TENANT_PARAMETER             = "tenantParameter"
	PARAM_DROP_RESOURCE_POOL_LIST      = "dropResourcePoolList"
	PARAM_SPLIT_RESOURCE_POOL_LIST     = "splitResourcePoolList"
	PARAM_TENANT_LOCALITY_ZONE         = "localityZone"
	PARAM_LOCALITY_TYPE                = "localityType"
	PARAM_TARGET_LOCALITY              = "targetLocality"
	PARAM_ZONE_PARAM                   = "zoneParam"
	PARAM_MODIFY_TENANT_REPLICAS_PARAM = "modifyTenantReplicasParam"
	PARAM_DELETE_TENANT_REPLICAS_PARAM = "deleteTenantReplicasParam"
	PARAM_ZONE_LIST                    = "zoneList"
	PARAM_TENANT_UNIT_NUM              = "tenantUnitNum"
	PARAM_TENANT_UNIT_NAME             = "tenantUnitName"
	PARAM_ZONE_NAME                    = "zoneName"
	PARAM_PRIMARY_ZONE                 = "primaryZone"
	PARAM_ZONE_WITH_UNIT               = "zoneWithUnit"
	PARAM_TIMESTAMP                    = "timestamp"

	// tenant task
	TASK_NAME_CREATE_AND_ATTACH_RESOURCE_POOL = "Create and attach resource pools"
	TASK_NAME_CREATE_TENANT                   = "Create tenant"
	TASK_NAME_OPTIMIZE_TENANT                 = "Optimize tenant"
	TASK_NAME_SET_TENANT_TIME_ZONE            = "Set tenant time zone"
	Task_NAME_MODIFY_WHITELIST                = "Modify tenant whitelist"
	TASK_NAME_MODIFY_PRIMARY_ZONE             = "Modify tenant primary zone"
	TASK_NAME_SET_ROOT_PWD                    = "Set root password"
	TASK_NAME_SET_TENANT_PARAM                = "Set tenant parameters"
	TASK_NAME_SET_TENANT_VARIABLE             = "Set tenant variables"
	TASK_NAME_DROP_RESOURCE_POOL              = "Drop resource pools"
	TASK_NAME_SET_TENANT_PARAMETER            = "Set tenant parameter"
	TASK_NAME_DROP_TENANT                     = "Drop tenant"
	TASK_NAME_RECYCLE_TENANT                  = "Recycle tenant"
	TASK_NAME_FLASHBACK_TENANT                = "Flashback tenant"
	TASK_NAME_SCALE_OUT_TENANT_LOCALITY       = "Scale out tenant locality"
	TASK_NAME_SPLIT_RESOURCE_POOL             = "Split resource pool"
	TASK_NAME_ALTER_RESOURCE_POOL_UNIT_CONF   = "Alter resource pool unit config"
	TASK_NAME_ALTER_RESOURCE_POOL_UNIT_NUM    = "Alter resource pool unit num"
	TASK_NAME_ATTACH_TENANT_RESOURCE_POOL     = "Attach tenant resource pool"
	TASK_NAME_ALTER_TENANT_LOCALITY           = "Alter tenant locality"
	TASK_NAME_ALTER_TENANT_PRIMARY_ZONE       = "Alter tenant primary zone"

	// tenant dag
	DAG_CREATE_TENANT              = "Create tenant %s"
	DAG_SET_ROOTPASSWORD           = "Set root password"
	DAG_DROP_TENANT                = "Drop tenant"
	DAG_SCALE_OUT_TENANT_REPLICA   = "Scale out tenant replicas"
	DAG_SCALE_IN_TENANT_REPLICA    = "Scale in tenant replicas"
	DAG_MODIFY_TENANT_REPLICA      = "Modify tenant replicas"
	DAG_MODIFY_TENANT_PRIMARY_ZONE = "Modify tenant primary zone"

	TENANT_NAME_PATTERN = `^[a-zA-Z0-9-_~#+]+$`

	EXPRESS_OLTP = "express_oltp"
	COMPLEX_OLTP = "complex_oltp"
	OLAP         = "olap"
	OLTP         = "htap"
	KV           = "kv"

	NORMAL_TENANT = "NORMAL"

	VARIABLES_TEMPLATE  = "variables"
	PARAMETERS_TEMPLATE = "parameters"
)

func checkTenantExist(name string) (*oceanbase.DbaObTenant, error) {
	tenant, err := tenantService.GetTenantByName(name)
	if err != nil {
		return tenant, errors.Wrapf(err, "Get tenant '%s' failed.", name)
	}
	if tenant == nil {
		return tenant, errors.Occur(errors.ErrObTenantNotExist, name)
	}
	return tenant, nil
}

func checkTenantExistAndStatus(name string) (*oceanbase.DbaObTenant, error) {
	tenant, err := checkTenantExist(name)
	if err != nil {
		return tenant, err
	}
	if tenant.Status != NORMAL_TENANT {
		return tenant, errors.Occur(errors.ErrObTenantStatusNotNormal, name, tenant.Status)
	}
	return tenant, nil
}

func transfer(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	return str
}

func RegisterTenantTask() {
	task.RegisterTaskType(CreateTenantTask{})
	task.RegisterTaskType(ModifyPrimaryZoneTask{})
	task.RegisterTaskType(SetRootPwdTask{})
	task.RegisterTaskType(SetTenantTimeZoneTask{})
	task.RegisterTaskType(SetTenantParamterTask{})
	task.RegisterTaskType(SetTenantVariableTask{})
	task.RegisterTaskType(DropTenantTask{})
	task.RegisterTaskType(RecycleTenantTask{})
	task.RegisterTaskType(BatchCreateResourcePoolTask{})
	task.RegisterTaskType(AlterLocalityTask{})
	task.RegisterTaskType(SplitResourcePoolTask{})
	task.RegisterTaskType(BatchDropResourcePoolTask{})
	task.RegisterTaskType(AlterResourcePoolUnitNumTask{})
	task.RegisterTaskType(AlterResourcePoolUnitConfTask{})
	task.RegisterTaskType(ModifyTenantWhitelistTask{})
}
