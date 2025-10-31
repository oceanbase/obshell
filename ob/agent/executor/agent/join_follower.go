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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
)

type AgentJoinMasterTask struct {
	task.Task
	masterPassword string
}

type AgentBeFollowerTask struct {
	task.Task
}

func SendTokenToMaster(agentInfo meta.AgentInfo, masterPassword string) error {
	token, err := secure.NewToken(&agentInfo)
	if err != nil {
		return errors.Wrap(err, "get token failed")
	}
	param := param.AddTokenParam{
		AgentInfo: *meta.NewAgentInfoByInterface(meta.OCS_AGENT),
		Token:     token,
	}
	if err := secure.SendRequestWithPassword(&agentInfo, constant.URI_AGENT_RPC_PREFIX+constant.URI_TOKEN, http.POST, masterPassword, param, nil); err != nil {
		return errors.Wrap(err, "send post request failed")
	}
	return nil
}

func CreateJoinMasterDag(masterAgent meta.AgentInfo, zone string, masterPassword string) (*task.Dag, error) {
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

	// Encrypt master agent password.
	agentPassword, err := secure.Encrypt(masterPassword)
	if err != nil {
		return nil, err
	}
	template := builder.Build()
	ctx := task.NewTaskContext().SetParam(PARAM_ZONE, zone).SetParam(PARAM_MASTER_AGENT, masterAgent).
		SetParam(PARAM_MASTER_AGENT_PASSWORD, agentPassword)

	return localTaskService.CreateDagInstanceByTemplate(template, ctx)
}

func (t *AgentJoinMasterTask) Execute() error {
	var masterAgent meta.AgentInfo
	taskCtx := t.GetContext()
	if err := taskCtx.GetParamWithValue(PARAM_MASTER_AGENT, &masterAgent); err != nil {
		return err
	}
	if err := taskCtx.GetParamWithValue(PARAM_MASTER_AGENT_PASSWORD, &t.masterPassword); err != nil {
		return err
	}
	// Decrypt master agent password.
	masterPassword, err := secure.Decrypt(t.masterPassword)
	if err != nil {
		return errors.Occur(errors.ErrSecurityDecryptFailed, err.Error())
	}
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.Occur(errors.ErrTaskParamNotSet, PARAM_ZONE)
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
	if err := secure.SendRequestWithPassword(&masterAgent, constant.URI_AGENT_RPC_PREFIX, http.POST, masterPassword, param, &masterAgentInstance); err != nil {
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
		return errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.SINGLE)
	}

	var masterAgent meta.AgentInstance
	taskCtx := t.GetContext()
	if err := taskCtx.GetDataWithValue(PARAM_MASTER_AGENT, &masterAgent); err != nil {
		return err
	}
	zone, ok := taskCtx.GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.Occur(errors.ErrTaskParamNotSet, PARAM_ZONE)
	}
	if err := agentService.BeFollowerAgent(masterAgent, zone); err != nil {
		return err
	}
	t.ExecuteLog("set agent identity to follower")
	return nil
}

func AddFollowerAgent(param param.JoinMasterParam) error {
	targetToken, err := secure.Crypter.Decrypt(param.Token)
	if err != nil {
		return errors.Wrapf(err, "decrypt token of '%s' failed", param.JoinApiParam.AgentInfo.String())
	}

	agentInstance := meta.NewAgentInstanceByAgentInfo(&param.JoinApiParam.AgentInfo, param.JoinApiParam.ZoneName, meta.FOLLOWER, param.Version)
	if err = agentService.AddAgent(agentInstance, param.HomePath, param.Os, param.Architecture, param.PublicKey, targetToken); err != nil {
		return errors.Wrap(err, "insert agent failed")
	}
	return nil
}

func AddSingleToken(param param.AddTokenParam) error {
	targetToken, err := secure.Crypter.Decrypt(param.Token)
	if err != nil {
		return errors.Wrapf(err, "decrypt token of '%s' failed", param.AgentInfo.String())
	}

	if err = agentService.AddSingleToken(&param.AgentInfo, targetToken); err != nil {
		return errors.Wrap(err, "insert token failed")
	}
	return nil
}
