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
	"fmt"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
)

// AgentToSingleTask will let follower agent register itself to master agent
type AgentToSingleTask struct {
	task.Task
}

// RemoveFollowerAgentTask will let master agent remove follower agent
type RemoveFollowerAgentTask struct {
	task.Task
}

type SendFollowerRemoveSelfRPCTask struct {
	task.Task
}

func CreaetFollowerRemoveSelfDag() (*task.Dag, error) {
	builder := task.NewTemplateBuilder(DAG_FOLLOWER_REMOVE_SELF)
	rpcTask := &SendFollowerRemoveSelfRPCTask{
		Task: *task.NewSubTask(TASK_FOLLOWER_REMOVE_SELF),
	}
	rpcTask.SetCanContinue()
	builder.AddTask(rpcTask, false)

	newTask := &AgentToSingleTask{
		Task: *task.NewSubTask(TASK_AGENT_TO_SINGLE),
	}
	newTask.SetCanContinue()
	builder.AddTask(newTask, false)

	builder.SetMaintenance(true)
	return localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext())

}

func CreateToSingleDag() (*task.Dag, error) {
	// Master agent send rpc to follower agent to remove itself, then follower agent create a task to clear itself
	builder := task.NewTemplateBuilder(DAG_AGENT_TO_SINGLE)
	newTask := &AgentToSingleTask{
		Task: *task.NewSubTask(TASK_AGENT_TO_SINGLE),
	}
	newTask.SetCanContinue()
	builder.AddTask(newTask, false)

	builder.SetMaintenance(true)
	return localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext())
}

func CreateRemoveFollowerAgentDag(agent meta.AgentInfo, fromAPI bool) (*task.Dag, error) {
	// Follower agent send rpc to master agent to remove itself or master agent receive api to remove follower agent.
	// Then, master agent create a task to remove follower agent.
	// Master will clear observer and zone config if there is no other follower agent in the zone.
	name := fmt.Sprintf("Remove follower agent %s:%d", agent.Ip, agent.Port)
	builder := task.NewTemplateBuilder(name)
	if fromAPI {
		builder.AddNode(newAgentRemoveFollowerRPCNode([]meta.AgentInfo{agent}))
	}
	newTask := &RemoveFollowerAgentTask{
		Task: *task.NewSubTask(name),
	}
	newTask.SetCanContinue()
	builder.AddTask(newTask, false)

	builder.SetMaintenance(false)
	template := builder.Build()
	ctx := task.NewTaskContext().SetParam(PARAM_AGENT, agent)
	return localTaskService.CreateDagInstanceByTemplate(template, ctx)
}

func (t *AgentToSingleTask) Execute() error {
	if meta.OCS_AGENT.IsSingleAgent() {
		t.ExecuteLog("Agent is aready single agent")
		return nil
	}
	t.ExecuteLog("clearing ob config")
	if err := observerService.ClearObConfig(); err != nil {
		return errors.Wrap(err, "clear ob config failed")
	}
	t.ExecuteLog("change agent to single agent")
	if err := agentService.BeSingleAgent(); err != nil {
		return errors.Wrap(err, "be single agent failed")
	}
	t.ExecuteLog("set agent to single agent success")
	return nil
}

func (t *RemoveFollowerAgentTask) Execute() (err error) {
	var agent meta.AgentInfo
	ctx := t.GetContext()
	if err = ctx.GetParamWithValue(PARAM_AGENT, &agent); err != nil {
		return errors.Wrap(err, "get param failed")
	}

	t.ExecuteLogf("finding agent %s:%d info", agent.Ip, agent.Port)
	agentInstance, err := GetFollowerAgent(&agent)
	if err != nil {
		return errors.Wrap(err, "get follower agent failed")
	}
	if agentInstance == nil {
		t.ExecuteLogf("agent %s:%d is not exists", agent.Ip, agent.Port)
		return nil
	}

	t.ExecuteLogf("get zone '%s' agents", agentInstance.Zone)
	zoneAgents, err := agentService.GetAgentInstanceByZone(agentInstance.Zone)
	if err != nil {
		return errors.Wrap(err, "find agent instance by zone failed")
	}
	if len(zoneAgents) < 2 {
		t.ExecuteLogf("zone '%s' has no other follower agent, clearing zone and observer config", agentInstance.Zone)
		if err = observerService.ClearObserverAndZoneConfig(agentInstance); err != nil {
			return errors.Wrap(err, "clear observer and zone config failed")
		}
	} else {
		t.ExecuteLogf("zone '%s' has other follower agent, deleting observer config", agentInstance.Zone)
		if err = observerService.DeleteObServerConfig(agentInstance); err != nil {
			return errors.Wrap(err, "delete ob server config failed")
		}
	}

	t.ExecuteLogf("deleting agent %s:%d", agent.Ip, agent.Port)
	if err = agentService.DeleteAgent(&agent); err != nil {
		return errors.Wrap(err, "delete agent failed")
	}
	t.ExecuteLogf("remove follower agent %s:%d success", agent.Ip, agent.Port)
	return nil
}

func (t *SendFollowerRemoveSelfRPCTask) Execute() error {
	if meta.OCS_AGENT.IsSingleAgent() {
		t.ExecuteLog("Agent is aready single agent")
		return nil
	}

	var dagDTO task.DagDetailDTO
	masterAgent := agentService.GetMasterAgentInfo()
	agent := t.GetExecuteAgent()
	t.ExecuteInfoLog("sending remove this agent request to master agent")
	for count := 0; count < 30; count++ {
		// Send rpc to master agent.
		resp, err := secure.SendDeleteRequestAndReturnResponse(masterAgent, constant.URI_AGENT_RPC_PREFIX, agent, &dagDTO)
		if resp != nil && resp.IsError() {
			return errors.Errorf("send remove agent request to %s failed: %v", masterAgent.String(), resp.Error())
		}
		if err != nil {
			t.ExecuteWarnLogf("send remove this agent request failed, err: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		t.ExecuteInfoLogf("send remove this agent request succeed, remote dag id: %s", dagDTO.GenericDTO)
		break
	}
	return nil
}

func GetFollowerAgent(agent meta.AgentInfoInterface) (agentInstance *meta.AgentInstance, err error) {
	agentInstance, err = agentService.FindAgentInstance(agent)
	if err != nil {
		err = errors.Wrap(err, "get agent instance failed")
	} else if agentInstance != nil && !agentInstance.IsFollowerAgent() {
		err = errors.Errorf("agent %s:%d is not follower", agent.GetIp(), agent.GetPort())
	}
	return
}
