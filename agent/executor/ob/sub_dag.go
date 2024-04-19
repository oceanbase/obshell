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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

type SubDagInfo struct {
	GenericID     string `json:"genecric_id" binding:"required"`
	ExpectedStage int    `json:"expected_stage" binding:"required"`
}

type CreateSubDagResp struct {
	SubDagInfo
	param.ForcePassDagParam
}

type CreateSubDagTask struct {
	task.Task
	// main dag id, used to record the relationship between main dag and sub dag
	mainDagID string
	// whether force pass the specific dag
	forcePassDag param.ForcePassDagParam
	// sub dag expect main dag's next stage
	expectedStage int
	// sub dag whether need execute command
	needExecCmd bool
	// send create sub dag rpc request to uri
	uri string
	// params for remote create sub dag
	param *CreateSubDagParam
}

type CreateSubDagParam struct {
	Agent       meta.AgentInfo
	NeedExecCmd bool
	SubDagInfo
	param.ForcePassDagParam
}

func NewCreateSubDagTask(name string) *CreateSubDagTask {
	newTask := &CreateSubDagTask{
		Task: *task.NewSubTask(name),
	}
	newTask.SetCanCancel()
	return newTask
}

func (t *CreateSubDagTask) GetAdditionalData() map[string]any {
	var dagInfo SubDagInfo
	agent := t.GetExecuteAgent()
	if err := t.GetContext().GetAgentDataWithValue(&agent, DATA_SUB_DAG_INFO, &dagInfo); err != nil {
		return nil
	}

	sub_dags := map[string]any{
		agent.String(): dagInfo.GenericID,
	}
	return map[string]any{
		ADDL_KEY_SUB_DAGS: sub_dags,
	}
}

func (t *CreateSubDagTask) getParams(agent meta.AgentInfo) (err error) {
	ctx := t.GetContext()
	t.mainDagID, err = getDagGenericIDBySubTaskId(t.GetID())
	if err != nil {
		return errors.Wrap(err, "get dag generic id failed")
	}
	log.Infof("main dag id is %v", t.mainDagID)

	if err := ctx.GetParamWithValue(PARAM_FORCE_PASS_DAG, &t.forcePassDag); err != nil {
		return errors.Wrap(err, "get force pass dag failed")
	}
	if err := ctx.GetParamWithValue(PARAM_URI, &t.uri); err != nil {
		return errors.Wrap(err, "get uri failed")
	}
	if ctx.GetAgentData(&agent, DATA_SUB_DAG_NEED_EXEC_CMD) != nil {
		t.needExecCmd = true
	}
	if err := ctx.GetParamWithValue(PARAM_EXPECT_MAIN_NEXT_STAGE, &t.expectedStage); err != nil {
		return errors.Wrap(err, "get expected stage failed")
	}

	t.param = &CreateSubDagParam{
		Agent:             meta.OCS_AGENT.GetAgentInfo(),
		ForcePassDagParam: t.forcePassDag,
		NeedExecCmd:       t.needExecCmd,
		SubDagInfo: SubDagInfo{
			GenericID:     t.mainDagID,
			ExpectedStage: t.expectedStage,
		},
	}
	return nil
}

func (t *CreateSubDagTask) execute() error {
	agent := t.GetExecuteAgent()
	t.ExecuteLogf("Inform %s to create the task", agent.String())
	if err := t.getParams(agent); err != nil {
		return errors.Wrap(err, "get params failed")
	}

	if !agent.Equal(meta.OCS_AGENT) {
		return t.remoteCreateSubDag(agent)
	}

	var resp *CreateSubDagResp
	var err *errors.OcsAgentError
	switch t.uri {
	case constant.URI_OB_RPC_PREFIX + constant.URI_STOP:
		resp, err = CreateStopDag(*t.param)
	case constant.URI_OB_RPC_PREFIX + constant.URI_START:
		resp, err = CreateStartDag(*t.param)
	}
	if err != nil {
		t.ExecuteErrorLog(errors.New(err.Error()))
	} else {
		if resp != nil {
			t.ExecuteLogf("create task %s successfully", resp.GenericID)
			t.SetLocalData(DATA_SUB_DAG_INFO, SubDagInfo{
				GenericID:     resp.GenericID,
				ExpectedStage: resp.ExpectedStage,
			})
		}
	}
	return nil
}

const (
	// rpc retry times
	create_sub_dag_max_retry_times = 3
	create_sub_dag_retry_interval  = 3
)

// remoteCreateSubDag remote create sub dag, and will always return nil.
func (t *CreateSubDagTask) remoteCreateSubDag(agent meta.AgentInfo) (err error) {
	var resp CreateSubDagResp
	for i := 1; i <= create_sub_dag_max_retry_times; i++ {
		err = secure.SendPostRequest(&agent, t.uri, t.param, &resp)
		if err == nil {
			t.ExecuteLogf("create task %s successfully", resp.GenericID)
			t.GetContext().SetAgentData(&agent, DATA_SUB_DAG_INFO, SubDagInfo{
				GenericID:     resp.GenericID,
				ExpectedStage: resp.ExpectedStage,
			})
			return nil
		}
		t.ExecuteErrorLogf("%s failed to create the task due to '%v' and will try again[%d/%d].", agent.String(), err, i, create_sub_dag_max_retry_times)
		time.Sleep(time.Second * create_sub_dag_retry_interval)
	}
	return nil
}

type CheckSubDagReadyTask struct {
	PassSubDagTask
	agents []meta.AgentInfo
}

func NewCheckSubDagReadyTask() *CheckSubDagReadyTask {
	newTask := &CheckSubDagReadyTask{
		PassSubDagTask: *NewPassSubDagTask("Make sure all agents are ready"),
	}
	newTask.SetCanCancel()
	return newTask
}

func (t *CheckSubDagReadyTask) execute() (err error) {
	t.allAgentDagMap = make(map[string]SubDagInfo)

	defer func() {
		if err != nil {
			t.pass()
		}
	}()

	if err = t.checkSubDagCreated(); err != nil {
		return errors.Wrap(err, "check sub dag created failed")
	}
	if err = t.checkCanAdvanceSubDag(); err != nil {
		return errors.Wrap(err, "check can advance sub dag failed")
	}
	return nil
}

func (t *CheckSubDagReadyTask) checkSubDagCreated() (err error) {
	ctx := t.GetContext()
	if err = ctx.GetParamWithValue(PARAM_ALL_AGENTS, &t.agents); err != nil {
		return errors.Wrap(err, "get all agents failed")
	}

	for _, agent := range t.agents {
		var dagInfo SubDagInfo
		if err = ctx.GetAgentDataWithValue(&agent, DATA_SUB_DAG_INFO, &dagInfo); err == nil {
			t.allAgentDagMap[agent.String()] = dagInfo
		}
	}

	ctx.SetData(DATA_ALL_AGENT_DAG_MAP, t.allAgentDagMap)

	if len(t.agents) != len(t.allAgentDagMap) {
		return errors.New("Not all tasks created. main dag failed")
	}

	t.ExecuteLog("All agents have created the tasks successfully")
	return nil
}

const (
	maxQueryDagDetailTimes    = 900
	queryDagDetailInterval    = 1 * time.Second
	maxQuerySubDagDetailTimes = 300
)

func (t *CheckSubDagReadyTask) checkCanAdvanceSubDag() (err error) {
	t.ExecuteLog("Check if all agents can be advanced")
	for agent, dag := range t.allAgentDagMap {

		canBeAdvanced := false
		for i := 0; i < maxQueryDagDetailTimes; i++ {
			dagDetailDTO, err := sendGetDagDetailRequest(dag.GenericID)
			if err != nil {
				t.ExecuteErrorLog(err)
			} else if dagDetailDTO.IsFailed() {
				t.ExecuteLogf("%s is ready", agent)
				canBeAdvanced = true
				break
			}
			t.ExecuteInfoLogf("%s is not ready", agent)
			time.Sleep(queryDagDetailInterval)
		}

		if !canBeAdvanced {
			return fmt.Errorf("wait for %s to be ready timeout", agent)
		}
	}
	return nil
}

type RetrySubDagTask struct {
	PassSubDagTask
}

func NewRetrySubDagTask() *RetrySubDagTask {
	newTask := &RetrySubDagTask{
		*NewPassSubDagTask("Advance agents to execute the task"),
	}
	newTask.SetCanCancel()
	return newTask
}

func (t *RetrySubDagTask) execute() (err error) {
	ctx := t.GetContext()
	if err = ctx.GetDataWithValue(DATA_ALL_AGENT_DAG_MAP, &t.allAgentDagMap); err != nil {
		return errors.Wrap(err, "get all agent dag map failed")
	}

	defer func() {
		if err != nil {
			t.pass()
		}
	}()

	success := true
	for agent, dag := range t.allAgentDagMap {
		t.ExecuteLogf("advance %s to execute the task", agent)
		if err := sendDagOperatorRequest(task.RETRY, dag.GenericID); err != nil {
			t.ExecuteErrorLog(err)
			success = false
		}
	}
	if !success {
		return errors.New("failed to advance all agents")
	}
	return nil
}

type WaitSubDagFinishTask struct {
	PassSubDagTask
}

func NewWaitSubDagFinishTask() *WaitSubDagFinishTask {
	newTask := &WaitSubDagFinishTask{
		PassSubDagTask: *NewPassSubDagTask("Wait for all agents to execute tasks successfully"),
	}
	newTask.SetCanContinue().SetCanCancel()
	return newTask
}

func (t *WaitSubDagFinishTask) Execute() (err error) {
	t.ExecuteLog("Wait for task to succeed")

	ctx := t.GetContext()
	if err = ctx.GetDataWithValue(DATA_ALL_AGENT_DAG_MAP, &t.allAgentDagMap); err != nil {
		t.ExecuteErrorLog(err)
		return
	}

	return t.syncResult()
}

func (t *WaitSubDagFinishTask) syncResult() (err error) {
	agent := t.GetExecuteAgent()
	isLocal := agent.Equal(meta.OCS_AGENT)
	dag := t.allAgentDagMap[agent.String()]
	var dagDetailDTO *task.DagDetailDTO
	for i := 0; i < maxQueryDagDetailTimes; i++ {
		log.Info("query dag detail")
		if isLocal {
			dagID, _, _ := task.ConvertGenericID(dag.GenericID)
			dagDetailDTO, err = localTaskService.GetDagDetail(dagID)
		} else {
			dagDetailDTO, err = sendGetDagDetailRequest(dag.GenericID)
		}
		if err != nil {
			t.ExecuteWarnLog(err)
		} else {
			if t.isSubDagFinished(dagDetailDTO, dag, agent) {
				break
			}
			if dagDetailDTO.IsRunning() && i == maxQueryDagDetailTimes-1 {
				t.ExecuteWarnLogf("since %s is timeout, try to cancel it", dag.GenericID)
				if err := sendDagOperatorRequest(task.CANCEL, dag.GenericID); err != nil {
					t.ExecuteWarnLog(err)
				} else {
					t.ExecuteLogf("wait %s to be finished", dag.GenericID)
					i = maxQueryDagDetailTimes / 2
				}
			}
		}
		time.Sleep(queryDagDetailInterval)
	}

	return nil
}

func (t *WaitSubDagFinishTask) isSubDagFinished(dagDetailDTO *task.DagDetailDTO, dag SubDagInfo, agent meta.AgentInfo) bool {
	if dagDetailDTO.IsFailed() {
		getResult(t, dagDetailDTO)
		if dagDetailDTO.Stage == dag.ExpectedStage {
			t.ExecuteLogf("%s succeed", dagDetailDTO.GenericID)
			t.GetContext().SetAgentData(&agent, DATA_SUB_DAG_SUCCEED, true)
		}
		return true
	}
	return false
}

func getResult(t task.ExecutableTask, dagDetailDTO *task.DagDetailDTO) {
	for _, node := range dagDetailDTO.Nodes {
		for _, task := range node.SubTasks {
			for _, log := range task.TaskLogs {
				t.ExecuteLogf("%s: %s", task.Name, log)
			}
		}
	}
}

type PassSubDagTask struct {
	task.Task
	allAgentDagMap map[string]SubDagInfo
	agents         []meta.AgentInfo
}

func NewPassSubDagTask(name string) *PassSubDagTask {
	return &PassSubDagTask{
		Task: *task.NewSubTask(name),
	}
}

func (t *PassSubDagTask) pass() bool {
	succeed := true
	for agent, dag := range t.allAgentDagMap {
		if err := sendDagOperatorRequest(task.PASS, dag.GenericID); err != nil {
			t.ExecuteErrorLogf("%s failed to end the task : %v", agent, err)
			succeed = false
		}
		t.ExecuteLogf("%s end the task", agent)
	}
	return succeed
}

func (t *PassSubDagTask) cancel() {
	var dagDetailDTO *task.DagDetailDTO
	var err error
	for agent, dag := range t.allAgentDagMap {
		if agent == meta.OCS_AGENT.String() {
			dagID, _, _ := task.ConvertGenericID(dag.GenericID)
			dagDetailDTO, err = localTaskService.GetDagDetail(dagID)
		} else {
			dagDetailDTO, err = sendGetDagDetailRequest(dag.GenericID)
		}
		if err != nil {
			t.ExecuteWarnLogf("%s failed to get dag %s detail : %v", agent, dag.GenericID, err)
			continue
		}

		if dagDetailDTO.IsRunning() {
			if err := sendDagOperatorRequest(task.CANCEL, dag.GenericID); err != nil {
				t.ExecuteWarnLogf("%s failed to cancel %s : %v", agent, dag.GenericID, err)
			}
		}
	}
}

func (t *PassSubDagTask) checkSubDagsucceed() (err error) {
	ctx := t.GetContext()
	if err = ctx.GetParamWithValue(PARAM_ALL_AGENTS, &t.agents); err != nil {
		return errors.Wrap(err, "get all agents failed")
	}

	succeed := true
	for _, agent := range t.agents {
		if _, ok := t.GetContext().GetAgentData(&agent, DATA_SUB_DAG_SUCCEED).(bool); !ok {
			t.ExecuteErrorLogf("agent %s not succeed", agent.String())
			succeed = false
		}
	}
	if succeed {
		t.ExecuteInfoLog("Check the final result: SUCCEED!")
		return nil
	}
	return errors.New("not all agents succeed. main dag failed")
}

func (t *PassSubDagTask) execute() (err error) {
	ctx := t.GetContext()
	if err := ctx.GetDataWithValue(DATA_ALL_AGENT_DAG_MAP, &t.allAgentDagMap); err != nil {
		return errors.Wrap(err, "get all agent dag map failed")
	}

	defer func() {
		t.cancel()
		if !t.pass() && err == nil {
			err = errors.New("failed to end all tasks")
		}
	}()

	if err := t.checkSubDagsucceed(); err != nil {
		return errors.Wrap(err, "check sub dag succeed failed")
	}

	t.ExecuteLog("End all tasks")
	return nil
}
