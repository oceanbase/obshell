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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

type AgentJoinMasterTask struct {
	task.Task
}

type AgentBeFollowerTask struct {
	task.Task
}

func CreateJoinMasterDag(masterAgent meta.AgentInfo, zone string) (*task.Dag, error) {
	// Agent receive api to join master, then create a task to be follower.
	builder := task.NewTemplateBuilder(DAG_JOIN_TO_MASTER)

	joinTask := &AgentJoinMasterTask{
		Task: *task.NewSubTask(TASK_JOIN_TO_MASTER),
	}
	joinTask.SetCanContinue()
	builder.AddTask(joinTask, false)

	beFollowerAgent := &AgentBeFollowerTask{
		Task: *task.NewSubTask(TASK_BE_FOLLOWER),
	}
	beFollowerAgent.SetCanContinue()
	builder.AddTask(beFollowerAgent, false)

	builder.SetMaintenance(task.GlobalMaintenance())
	template := builder.Build()
	ctx := task.NewTaskContext().SetParam(PARAM_ZONE, zone).SetParam(PARAM_MASTER_AGENT, masterAgent)
	return localTaskService.CreateDagInstanceByTemplate(template, ctx)
}

func (t *AgentJoinMasterTask) Execute() error {
	var masterAgent meta.AgentInfo
	taskCtx := t.GetContext()
	if err := taskCtx.GetParamWithValue(PARAM_MASTER_AGENT, &masterAgent); err != nil {
		return errors.Wrapf(err, "Get Param %s failed", PARAM_MASTER_AGENT)
	}
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("zone is not set")
	}

	t.ExecuteLog("creating token")
	token, err := secure.NewToken(&masterAgent)
	if err != nil {
		return errors.Wrap(err, "get token failed")
	}
	t.ExecuteLog("token created")

	param := param.JoinMasterParam{
		JoinApiParam: param.JoinApiParam{
			AgentInfo: *meta.NewAgentInfoByInterface(meta.OCS_AGENT),
			ZoneName:  zone,
		},
		HomePath:     global.HomePath,
		Version:      meta.OCS_AGENT.GetVersion(),
		Os:           global.Os,
		Architecture: global.Architecture,
		PublicKey:    secure.Public(),
		Token:        token,
	}
	t.ExecuteLog("send join rpc to master")
	var masterAgentInstance meta.AgentInstance
	if err := secure.SendPostRequest(&masterAgent, constant.URI_AGENT_RPC_PREFIX, param, &masterAgentInstance); err != nil {
		return errors.Wrap(err, "send post request failed")
	}
	t.ExecuteLog(fmt.Sprintf("join to master success, master agent info: %v", masterAgentInstance))
	taskCtx.SetData(PARAM_MASTER_AGENT, masterAgentInstance)
	return nil
}

func (t *AgentBeFollowerTask) Execute() error {
	if t.IsContinue() && meta.OCS_AGENT.IsFollowerAgent() {
		t.ExecuteLog("agent is follower agent")
		return nil
	}
	if !meta.OCS_AGENT.IsSingleAgent() {
		return errors.New("agent is not single")
	}

	var masterAgent meta.AgentInstance
	taskCtx := t.GetContext()
	if err := taskCtx.GetDataWithValue(PARAM_MASTER_AGENT, &masterAgent); err != nil {
		return errors.Wrap(err, "masterAgent is not found")
	}
	zone, ok := taskCtx.GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("zone is not set")
	}
	if err := agentService.BeFollowerAgent(masterAgent, zone); err != nil {
		return err
	}
	t.ExecuteLog("set agent identity to follower")
	return nil
}

func AddFollowerAgent(param param.JoinMasterParam) *errors.OcsAgentError {
	targetToken, err := secure.Crypter.Decrypt(param.Token)
	if err != nil {
		return errors.Occurf(errors.ErrKnown, "decrypt token of '%s:%d' failed: %v", param.JoinApiParam.AgentInfo.GetIp(), param.JoinApiParam.AgentInfo.GetPort(), err)
	}

	agentInstance := meta.NewAgentInstanceByAgentInfo(&param.JoinApiParam.AgentInfo, param.JoinApiParam.ZoneName, meta.FOLLOWER, param.Version)
	if err = agentService.AddAgent(agentInstance, param.HomePath, param.Os, param.Architecture, param.PublicKey, targetToken); err != nil {
		return errors.Occurf(errors.ErrKnown, "insert agent failed: %v", err)
	}
	return nil
}

func UpdateFollowerAgent(agentInstance meta.Agent, param param.JoinMasterParam) *errors.OcsAgentError {
	// Agent already exists.
	if agentInstance.GetIdentity() != meta.FOLLOWER || agentInstance.GetVersion() != param.Version || agentInstance.GetZone() != param.JoinApiParam.ZoneName {
		return errors.Occur(errors.ErrBadRequest, "agent already exists")
	}

	targetToken, err := secure.Crypter.Decrypt(param.Token)
	if err != nil {
		return errors.Occurf(errors.ErrKnown, "decrypt token of '%s:%d' failed: %v", param.JoinApiParam.AgentInfo.GetIp(), param.JoinApiParam.AgentInfo.GetPort(), err)
	}

	if err = agentService.UpdateAgent(agentInstance, param.HomePath, param.Os, param.Architecture, param.PublicKey, targetToken); err != nil {
		return errors.Occurf(errors.ErrKnown, "update agent failed: %v", err)
	}
	return nil
}
