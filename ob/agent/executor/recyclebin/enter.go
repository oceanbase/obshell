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
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	taskservice "github.com/oceanbase/obshell/ob/agent/service/task"
	"github.com/oceanbase/obshell/ob/agent/service/tenant"
)

func RegisterRecyclebinTask() {
	task.RegisterTaskType(WaitForPurgeFinishedTask{})
}

var (
	tenantService      tenant.TenantService
	clusterTaskService = taskservice.NewClusterTaskService()
)

const (
	// task name
	TASK_NAME_WAIT_PURGE_TENANT_FINISHED = "Wait tenant purge finished"

	// dag name
	DAG_WAIT_PURGE_TENANT_FINISHED = "Wait tenant purge finished"

	// task param name
	PARAM_OBECJT_NAME             = "object_name"
	PARAM_ORIGINA_TENANTL_NAME    = "original_tenant_name"
	PARAM_DROP_RESOURCE_POOL_LIST = "dropResourcePoolList"
)
