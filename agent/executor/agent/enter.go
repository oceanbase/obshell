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

package agent

import (
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/service/agent"
	"github.com/oceanbase/obshell/agent/service/obcluster"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
)

var (
	agentService     = agent.AgentService{}
	observerService  = obcluster.ObserverService{}
	clusterService   = obcluster.ObclusterService{}
	localTaskService = taskservice.NewLocalTaskService()
)

const (
	// task param
	PARAM_MASTER_AGENT           = "masterAgent"
	PARAM_MASTER_AGENT_PASSWORD  = "masterAgentPassword"
	PARAM_ZONE                   = "zone"
	PARAM_AGENT                  = "agent"
	PARAM_TAKE_OVER_MASTER_AGENT = "takeOverMasterAgent"

	// dag name
	DAG_FOLLOWER_REMOVE_SELF = "Follower remove self"
	DAG_AGENT_TO_SINGLE      = "Agent to single"
	DAG_JOIN_TO_MASTER       = "Join to master"
	DAG_JOIN_SELF            = "Join self"
	DAG_REMOVE_ALL_AGENTS    = "Remove all agents"

	// task name
	TASK_FOLLOWER_REMOVE_SELF = "Follower remove self"
	TASK_AGENT_TO_SINGLE      = "Agent to single"
	TASK_JOIN_TO_MASTER       = "Join to master"
	TASK_BE_FOLLOWER          = "Be follower"
	TASK_JOIN_SELF            = "Join self"
	TASK_REMOVE_MASTER        = "Remove master"
	TASK_REMOVE_FOLLOWER      = "Remove follower"
)

func RegisterAgentTask() {
	task.RegisterTaskType(AgentJoinSelfTask{})
	task.RegisterTaskType(AgentJoinMasterTask{})
	task.RegisterTaskType(AgentBeFollowerTask{})
	task.RegisterTaskType(AgentToSingleTask{})
	task.RegisterTaskType(RemoveFollowerAgentTask{})
	task.RegisterTaskType(AgentRemoveFollowerRPCTask{})
	task.RegisterTaskType(AgentRemoveMasterTask{})
	task.RegisterTaskType(SendFollowerRemoveSelfRPCTask{})
}
