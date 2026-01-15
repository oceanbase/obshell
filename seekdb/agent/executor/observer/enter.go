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

package observer

import (
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/service/agent"
	"github.com/oceanbase/obshell/seekdb/agent/service/obcluster"
	taskservice "github.com/oceanbase/obshell/seekdb/agent/service/task"
	"github.com/oceanbase/obshell/seekdb/agent/service/tenant"
	"github.com/oceanbase/obshell/seekdb/agent/service/user"
)

const (
	DATA_SKIP_START_TASK = "skipStartTask"

	// task name
	TASK_NAME_START        = "Start seekdb"
	TASK_NAME_STOP         = "Stop seekdb"
	TASK_NAME_MINOR_FREEZE = "Minor freeze before stop server"

	// dag name
	DAG_EMERGENCY_START  = "Start local seekdb"
	DAG_EMERGENCY_STOP   = "Stop local seekdb"
	DAG_START_OBSERVER   = "Start seekdb"
	DAG_STOP_OBSERVER    = "Stop seekdb"
	DAG_RESTART_OBSERVER = "Restart seekdb"

	// rpc retry times
	MAX_RETRY_RPC_TIMES = 3
	RPC_RETRY_INTERVAL  = 1

	// stop ob retry times
	STOP_OB_MAX_RETRY_TIME     = 15
	STOP_OB_MAX_RETRY_INTERVAL = 5

	PARAM_OBSERVER_PID = "observerPid"
)

var (
	agentService     = agent.AgentService{}
	observerService  = obcluster.ObserverService{}
	obclusterService = obcluster.ObclusterService{}
	localTaskService = taskservice.NewLocalTaskService()
	tenantService    = tenant.TenantService{}
	userService      = user.UserService{}
)

func RegisterObStopTask() {
	task.RegisterTaskType(MinorFreezeTask{})
	task.RegisterTaskType(StopObserverTask{})
}

func RegisterObStartTask() {
	task.RegisterTaskType(StartObserverTask{})
}
