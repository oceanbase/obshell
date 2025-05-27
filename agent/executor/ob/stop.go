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

package ob

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func HandleObStop(param param.ObStopParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occurf(errors.ErrKnown, "agent identity is '%v'", meta.OCS_AGENT.GetIdentity())
	}
	if err := CheckStopObParam(&param); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err)
	}

	template := buildStopTemplate(param.Force, param.Terminate)
	taskCtx, err := buildStopTaskContext(param)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, taskCtx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildStopTaskContext(param param.ObStopParam) (*task.TaskContext, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	needStopAgents, err := GenerateTargetAgentList(param.Scope)
	if err != nil {
		return nil, err
	}
	log.Infof("need stop agents are %v", needStopAgents)
	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_ALL_AGENTS, agents).
		SetParam(PARAM_SCOPE, param.Scope).
		SetParam(PARAM_FORCE_PASS_DAG, param.ForcePassDagParam).
		SetParam(PARAM_URI, constant.URI_OB_RPC_PREFIX+constant.URI_STOP).
		SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, SUB_STOP_DAG_EXPECT_MAIN_NEXT_STAGE)
	if param.Force || param.Terminate {
		for _, agent := range needStopAgents {
			ctx.SetAgentData(&agent, DATA_SUB_DAG_NEED_EXEC_CMD, true)
		}
	}
	return ctx, nil
}

func buildStopTemplate(force bool, terminate bool) *task.Template {
	task := task.NewTemplateBuilder(DAG_STOP_OB).
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCreateSubStopDagTask(), true).
		AddTask(newCheckSubStopDagReadyTask(), false)
	if terminate {
		task.AddTask(newMinorFreezeTask(), false)
	}
	task.AddTask(newRetrySubStopDagTask(), false).
		AddTask(newWaitSubStopDagFinishTask(), true)
	if !force && !terminate {
		task.AddTask(newExecStopSqlTask(), false)
	}
	task.AddTask(newPassSubStopDagTask(), false)
	return task.Build()
}

type CreateSubStopDagTask struct {
	CreateSubDagTask
}

type CheckSubStopDagReadyTask struct {
	CheckSubDagReadyTask
}

type RetrySubStopDagTask struct {
	RetrySubDagTask
}

type WaitSubStopDagFinishTask struct {
	WaitSubDagFinishTask
}

type ExecStopSqlTask struct {
	PassSubDagTask
	scope param.Scope
}

type PassSubStopDagTask struct {
	PassSubDagTask
}

func newCreateSubStopDagTask() *CreateSubStopDagTask {
	return &CreateSubStopDagTask{
		CreateSubDagTask: *NewCreateSubDagTask("Inform all agents to prepare to stop observer"),
	}
}

func newCheckSubStopDagReadyTask() *CheckSubStopDagReadyTask {
	return &CheckSubStopDagReadyTask{
		*NewCheckSubDagReadyTask(),
	}
}

func newRetrySubStopDagTask() *RetrySubStopDagTask {
	return &RetrySubStopDagTask{
		*NewRetrySubDagTask(),
	}
}

func newWaitSubStopDagFinishTask() *WaitSubStopDagFinishTask {
	return &WaitSubStopDagFinishTask{
		*NewWaitSubDagFinishTask(),
	}
}

func newExecStopSqlTask() *ExecStopSqlTask {
	newTask := &ExecStopSqlTask{
		PassSubDagTask: *NewPassSubDagTask("Execute stop sql"),
	}
	newTask.SetCanContinue().SetCanCancel()
	return newTask
}

func newPassSubStopDagTask() *PassSubStopDagTask {
	newTask := &PassSubStopDagTask{
		*NewPassSubDagTask("Inform all agents to end the task"),
	}
	newTask.SetCanCancel()
	return newTask
}

const (
	SUB_STOP_DAG_EXPECT_MAIN_NEXT_STAGE   = 3
	MAIN_STOP_DAG_EXPECTED_SUB_NEXT_STAGE = 2
)

func (t *CreateSubStopDagTask) Execute() error {
	return t.execute()
}

func (t *CheckSubStopDagReadyTask) Execute() (err error) {
	return t.execute()
}

func (t *RetrySubStopDagTask) Execute() (err error) {
	return t.execute()
}

func (t *WaitSubStopDagFinishTask) Execute() (err error) {
	return t.WaitSubDagFinishTask.Execute()
}

func (t *ExecStopSqlTask) Execute() (err error) {
	t.ExecuteLog("Stop observer")
	ctx := t.GetContext()
	if err := ctx.GetDataWithValue(DATA_ALL_AGENT_DAG_MAP, &t.allAgentDagMap); err != nil {
		return errors.Wrap(err, "get all agent dag map failed")
	}
	defer func() {
		if err != nil {
			t.pass()
		}
	}()

	if exist, err := process.CheckObserverProcess(); err != nil || !exist {
		return fmt.Errorf("check observer process exist: %v, %v,", exist, err)
	}

	if err = ctx.GetParamWithValue(PARAM_SCOPE, &t.scope); err != nil {
		return errors.Wrap(err, "get scope failed")
	}

	if err := getOceanbaseInstance(); err != nil {
		return err
	}

	switch t.scope.Type {
	case SCOPE_ZONE:
		return t.stopZone()
	case SCOPE_SERVER:
		return t.stopServer()
	}
	return errors.Errorf("invalid scope '%v'", t.scope)
}

func (t *ExecStopSqlTask) stopZone() (err error) {
	t.ExecuteLog("Stop Zone")
	for _, zone := range t.scope.Target {
		if err = obclusterService.StopZone(zone); err != nil {
			return err
		}

		active, err := obclusterService.IsZoneActive(zone)
		if !active {
			t.ExecuteLogf("%s stopped", zone)
			continue
		}
		if err != nil {
			t.ExecuteErrorLog(err)
		}
	}
	return nil
}

func (t *ExecStopSqlTask) stopServer() (err error) {
	agents, err := agentService.GetAllAgentsDOFromOB()
	if err != nil {
		return err
	}
	for _, server := range t.scope.Target {
		t.ExecuteLogf("Stop %s", server)
		agentInfo, err := meta.ConvertAddressToAgentInfo(server)
		if err != nil {
			return errors.Errorf("convert server '%s' to agent info failed: %v", server, err)
		}
		for _, agent := range agents {
			if agentInfo.Ip == agent.Ip && agentInfo.Port == agent.Port {
				serverInfo := meta.NewAgentInfo(agent.Ip, agent.RpcPort)
				sql := fmt.Sprintf("alter system stop server '%s'", serverInfo.String())
				log.Info(sql)
				if err = obclusterService.ExecuteSql(sql); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (t *PassSubStopDagTask) Execute() (err error) {
	return t.execute()
}

func CheckStopObParam(param *param.ObStopParam) error {
	if param.Scope.Type == SCOPE_GLOBAL && (!param.Force && !param.Terminate) {
		return errors.New("cannot stop all observer without 'force'")
	}
	if param.Force && param.Terminate {
		return errors.New("cannot stop observer with 'force' and 'terminate' at the same time")
	}
	if !param.Force && oceanbase.GetState() != oceanbase.STATE_CONNECTION_AVAILABLE {
		return errors.New("The current observer is not available, please stop with 'force'")
	}
	return nil
}
