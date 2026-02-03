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
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
)

func HandleObStart(param param.StartObParam) (*task.DagDetailDTO, error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}
	template := buildStartObclusterTemplate(param.Scope.Type)
	taskCtx, err := buildStartObclusterTaskContext(param)
	if err != nil {
		return nil, err
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, taskCtx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildStartObclusterTaskContext(param param.StartObParam) (*task.TaskContext, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	needStartAgents, err := GenerateTargetAgentList(param.Scope)
	if err != nil {
		return nil, err
	}
	log.Infof("need start agents %v", needStartAgents)
	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_ALL_AGENTS, agents).
		SetParam(PARAM_SCOPE, param.Scope).
		SetParam(PARAM_FORCE_PASS_DAG, param.ForcePassDagParam).
		SetParam(PARAM_URI, constant.URI_OB_RPC_PREFIX+constant.URI_START).
		SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, SUB_START_DAG_EXPECT_MAIN_NEXT_STAGE).
		SetParam(PARAM_MAIN_DAG_NAME, DAG_START_OB)
	for _, agent := range needStartAgents {
		ctx.SetAgentData(&agent, DATA_SUB_DAG_NEED_EXEC_CMD, true)
	}
	return ctx, nil
}

func buildStartObclusterTemplate(t string) *task.Template {
	task := task.NewTemplateBuilder(DAG_START_OB).
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCreateSubStartDagTask(), true).
		AddTask(newCheckSubStartDagReadyTask(), false).
		AddTask(newRetrySubStartDagTask(), false).
		AddTask(newWaitSubStartDagFinishTask(), true)
	if t == SCOPE_ZONE || t == SCOPE_GLOBAL {
		task.AddTask(newStartZoneTask(), false)
	}
	task.AddTask(newPassSubStartDagTask(), false)
	return task.Build()
}

type CreateSubStartDagTask struct {
	CreateSubDagTask
}

type CheckSubStartDagReadyTask struct {
	CheckSubDagReadyTask
}

type RetrySubStartDagTask struct {
	RetrySubDagTask
}

type WaitSubStartDagFinishTask struct {
	WaitSubDagFinishTask
}

type StartZoneTask struct {
	PassSubDagTask
}

type PassSubStartDagTask struct {
	PassSubDagTask
}

func newCreateSubStartDagTask() *CreateSubStartDagTask {
	return &CreateSubStartDagTask{
		*NewCreateSubDagTask("Inform all agents to start observer"),
	}
}

func newCheckSubStartDagReadyTask() *CheckSubStartDagReadyTask {
	return &CheckSubStartDagReadyTask{
		*NewCheckSubDagReadyTask(),
	}
}

func newRetrySubStartDagTask() *RetrySubStartDagTask {
	return &RetrySubStartDagTask{
		*NewRetrySubDagTask(),
	}
}

func newWaitSubStartDagFinishTask() *WaitSubStartDagFinishTask {
	return &WaitSubStartDagFinishTask{
		*NewWaitSubDagFinishTask(),
	}
}

func newStartZoneTask() *StartZoneTask {
	newTask := &StartZoneTask{
		*NewPassSubDagTask("Start Zone"),
	}
	newTask.
		SetCanContinue().
		SetCanCancel().
		SetCanRetry()
	return newTask
}

func newPassSubStartDagTask() *PassSubStartDagTask {
	newTask := &PassSubStartDagTask{
		*NewPassSubDagTask("Inform all agents to end the task"),
	}
	newTask.SetCanCancel()
	return newTask
}

const (
	// sub start dag will check the main dag's next stage is 3
	SUB_START_DAG_EXPECT_MAIN_NEXT_STAGE = 3
	// main start dag will check the sub dag's next stage is 2
	MAIN_START_DAG_EXPECTED_SUB_NEXT_STAGE = 2
)

func (t *CreateSubStartDagTask) Execute() error {
	return t.execute()
}

func (t *CheckSubStartDagReadyTask) Execute() (err error) {
	return t.execute()
}

func (t *RetrySubStartDagTask) Execute() (err error) {
	return t.execute()
}

func (t *WaitSubStartDagFinishTask) Execute() (err error) {
	return t.WaitSubDagFinishTask.Execute()
}

func (t *StartZoneTask) Execute() (err error) {
	t.ExecuteLog("start zone")
	ctx := t.GetContext()
	scope := param.Scope{}
	if err := ctx.GetDataWithValue(DATA_ALL_AGENT_DAG_MAP, &t.allAgentDagMap); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			t.pass()
		}
	}()

	if err = ctx.GetParamWithValue(PARAM_SCOPE, &scope); err != nil {
		return err
	}

	if err := getOceanbaseInstance(); err != nil {
		return err
	}

	var zoneList []string
	success := true
	if scope.Type == SCOPE_ZONE {
		zoneList = scope.Target
	} else {
		zoneList, err = obclusterService.GetObZonesName()
		if err != nil {
			return errors.Wrap(err, "get ob zones name failed")
		}
	}
	for _, zone := range zoneList {
		if err = obclusterService.StartZone(zone); err != nil {
			if isZoneStatusNotMatchError(err) {
				active, err := obclusterService.IsZoneActive(zone)
				if active {
					t.ExecuteLogf("%s started", zone)
					continue
				}
				if err != nil {
					t.ExecuteErrorLog(err)
				}
			}
			t.ExecuteErrorLog(err)
			success = false
		}
	}
	if !success {
		return errors.Occur(errors.ErrCommonUnexpected, "start zone failed")
	}
	return nil
}

func isZoneStatusNotMatchError(err error) bool {
	return strings.Contains(err.Error(), "zone status not match")
}

func (t *PassSubStartDagTask) Execute() error {
	return t.execute()
}

func sendDagOperatorRequest(operator int, id string) error {
	dagOperator := task.DagOperator{Operator: task.OPERATOR_MAP[operator]}
	return secure.SendPostRequest(meta.OCS_AGENT, constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+id, dagOperator, nil)
}

func sendGetDagDetailRequest(id string) (*task.DagDetailDTO, error) {
	var dagDetailDTO *task.DagDetailDTO
	if err := secure.SendGetRequest(meta.OCS_AGENT, constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+id, nil, &dagDetailDTO); err != nil {
		return nil, err
	}
	return dagDetailDTO, nil
}

func getDagGenericIDBySubTaskId(id int64) (string, error) {
	dag, err := localTaskService.GetDagBySubTaskId(id)
	return task.ConvertLocalIDToGenericID(dag.GetID(), dag.GetDagType()), err
}
