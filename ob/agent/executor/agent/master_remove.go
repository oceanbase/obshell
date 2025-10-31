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
	"time"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

// Master Agent sends an rpc to Follower to transition to a Single
type AgentRemoveFollowerRPCTask struct {
	task.Task
}

// Master remove self.
type AgentRemoveMasterTask struct {
	task.Task
}

func newAgentRemoveFollowerRPCNode(agents []meta.AgentInfo) *task.Node {
	removeFollowerTask := &AgentRemoveFollowerRPCTask{
		Task: *task.NewSubTask(TASK_REMOVE_FOLLOWER),
	}
	removeFollowerTask.SetCanContinue().SetCanPass()
	return task.NewNodeWithContext(removeFollowerTask, true, task.NewTaskContext().SetParam(task.EXECUTE_AGENTS, agents))
}

func CreateRemoveAllAgentsDag() (dag *task.Dag, err error) {
	// Master receive api to remove self, then create a task to remove all agents
	var agents []meta.AgentInfo
	agentInstances, err := agentService.GetFollowerAgents()
	if err != nil {
		err = errors.Wrap(err, "get all agent instance failed")
		return
	}
	for _, agent := range agentInstances {
		agents = append(agents, agent.AgentInfo)
	}

	builder := task.NewTemplateBuilder(DAG_REMOVE_ALL_AGENTS)
	if len(agents) > 0 {
		builder.AddNode(newAgentRemoveFollowerRPCNode(agents))
	}

	removeMasterTask := &AgentRemoveMasterTask{
		Task: *task.NewSubTask(TASK_REMOVE_MASTER),
	}
	builder.AddTask(removeMasterTask, false)

	builder.SetMaintenance(task.GlobalMaintenance())
	template := builder.Build()
	ctx := task.NewTaskContext()
	return localTaskService.CreateDagInstanceByTemplate(template, ctx)
}

func (t *AgentRemoveFollowerRPCTask) Execute() error {
	var dagDTO task.DagDetailDTO
	agent := t.GetExecuteAgent()
	t.ExecuteLogf("send remove agent request to follower, agent: %v", agent)
	for count := 0; count < 30; count++ {
		// Send rpc to follower agent.
		resp, err := secure.SendDeleteRequestAndReturnResponse(&agent, constant.URI_AGENT_RPC_PREFIX, agent, &dagDTO)

		if resp != nil && resp.IsError() {
			return errors.Occur(errors.ErrAgentRPCRequestError, http.DELETE, constant.URI_AGENT_RPC_PREFIX, agent.String(), resp.Error())
		}
		if err != nil {
			t.ExecuteWarnLogf("send remove agent request failed, err: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		t.ExecuteInfoLogf("send remove agent request succeed, remote dag id: %s", dagDTO.GenericDTO)
		break
	}
	return nil
}

func (t *AgentRemoveMasterTask) Execute() error {
	t.ExecuteLog("clearing ob config")
	if err := observerService.ClearObConfig(); err != nil {
		return errors.Wrap(err, "clear config failed")
	}
	t.ExecuteLog("clearing global config")
	if err := observerService.ClearGlobalConfig(); err != nil {
		return errors.Wrap(err, "clear config failed")
	}

	t.ExecuteLog("change agent to single agent")
	if err := agentService.BeSingleAgent(); err != nil {
		return errors.Wrap(err, "be single agent failed")
	}
	t.ExecuteLog("set agent to single agent success")
	return nil
}
