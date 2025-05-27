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
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
)

type StopObserverTask struct {
	RemoteExecutableTask
}

func newStopObserverTask() *StopObserverTask {
	newTask := &StopObserverTask{
		RemoteExecutableTask: *newRemoteExecutableTask(TASK_NAME_STOP),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func CreateStopSelfDag() (*task.DagDetailDTO, error) {
	subTask := newStopObserverTask()
	builder := task.NewTemplateBuilder(subTask.GetName())
	builder.AddTask(subTask, false).SetMaintenance(task.GlobalMaintenance())
	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext())
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func (t *StopObserverTask) Execute() error {
	return stopObserver(t)
}

func stopObserver(t task.ExecutableTask) error {
	t.ExecuteLog("Get observer Pid")
	pid, err := process.GetObserverPid()
	if err != nil {
		return err
	}
	if pid == "" {
		t.ExecuteLog("Observer is not running")
		return nil
	}
	for i := 0; i < STOP_OB_MAX_RETRY_TIME; i++ {
		t.ExecuteLogf("Kill observer process %s", pid)
		res := exec.Command("kill", "-9", pid)
		if err := res.Run(); err != nil {
			log.Warn("Kill observer process failed")
		}

		time.Sleep(time.Second * STOP_OB_MAX_RETRY_INTERVAL)
		t.TimeoutCheck()

		t.ExecuteLog("Check observer process")
		exist, err := process.CheckObserverProcess()
		if err != nil {
			log.Warnf("Check observer process failed: %v", err)
		} else if !exist {
			t.ExecuteLog("Successfully killed the observer process")
			return nil
		}
	}
	return errors.New("kill observer process timeout")
}

func CreateStopDag(params CreateSubDagParam) (*CreateSubDagResp, *errors.OcsAgentError) {
	lastMainDagId, err := checkMaintenanceAndPassDag(params.ForcePassDagParam.ID)
	if err != nil {
		return &CreateSubDagResp{ForcePassDagParam: param.ForcePassDagParam{ID: []string{lastMainDagId}}}, errors.Occur(errors.ErrKnown, err)
	}

	stage, template := buildSubStopTemplate(params)
	ctx := buildSubStopTaskCtx(params)
	resp := &CreateSubDagResp{
		SubDagInfo: SubDagInfo{
			ExpectedStage: stage,
		},
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return resp, errors.Occur(errors.ErrUnexpected, err)
	}
	resp.GenericID = task.NewDagDetailDTO(dag).GenericID
	return resp, nil
}

func buildSubStopTaskCtx(param CreateSubDagParam) *task.TaskContext {
	ctx := task.NewTaskContext().
		SetParam(PARAM_MAIN_DAG_ID, param.GenericID).
		SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, param.ExpectedStage).
		SetParam(PARAM_MAIN_AGENT, param.Agent)
	return ctx
}

func buildSubStopTemplate(param CreateSubDagParam) (stage int, t *task.Template) {
	stage = MAIN_STOP_DAG_EXPECTED_SUB_NEXT_STAGE
	template := task.NewTemplateBuilder(DAG_STOP_OBSERVER).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newCheckDagStageTask(), false)
	if param.NeedExecCmd {
		template.AddTask(newStopObserverTask(), false)
		stage += 1
	}
	template.AddTask(newWaitPassOperatorTask(), false)
	return stage, template.Build()
}

func checkMaintenanceAndPassDag(ids []string) (lastMainDagId string, err error) {
	log.Info("check if in maintenance")
	isRunning, err := localTaskService.IsRunning()
	if err != nil {
		return
	}

	if !isRunning {
		var dag *task.Dag
		dag, lastMainDagId, err = findLastMaintenanceDag(ids)
		if err != nil {
			return
		}
		log.Infof("handle sub dag %d created by main dag %s", dag.GetID(), lastMainDagId)
		if err = cancelAndPassSubDag(dag); err != nil {
			return "", err
		}
	}
	return lastMainDagId, nil
}

func findLastMaintenanceDag(ids []string) (dag *task.Dag, lastMainDagId string, err error) {
	dag, err = localTaskService.GetLastMaintenanceDag()
	if err != nil {
		return
	}
	if err = dag.GetContext().GetParamWithValue(PARAM_MAIN_DAG_ID, &lastMainDagId); err != nil {
		log.WithError(err).Error("get last maintenance dag id failed")
		return nil, "", errors.New("agent is under maintenance")
	}
	for _, id := range ids {
		if id == lastMainDagId {
			return
		}
	}
	return nil, lastMainDagId, fmt.Errorf("cluster is under maintenance, dag id: %s. This dag need to be forcibly terminated", lastMainDagId)
}

func cancelAndPassSubDag(dag *task.Dag) (err error) {
	if dag.IsRunning() {
		if err = cancelSubDag(dag); err != nil {
			return err
		}
	}

	return passSubDag(dag)
}

func passSubDag(dag *task.Dag) error {
	log.Infof("force pass dag %d", dag.GetID())
	if err := localTaskService.PassDag(dag); err != nil {
		log.WithError(err).Errorf("force pass %d failed", dag.GetID())
	} else {
		log.Infof("wait %d to be passed", dag.GetID())
		for i := 0; i < maxQuerySubDagDetailTimes; i++ {
			time.Sleep(time.Second)
			dag, err := localTaskService.GetDagInstance(dag.GetID())
			if err != nil {
				log.WithError(err).Errorf("get %d failed", dag.GetID())
				continue
			}
			if dag.IsSuccess() {
				return nil
			}
		}
	}
	return fmt.Errorf("pass %d timeout after %d seconds", dag.GetID(), maxQuerySubDagDetailTimes)
}

func cancelSubDag(dag *task.Dag) error {
	log.Infof("force cancel %d", dag.GetID())
	if err := localTaskService.CancelDag(dag); err != nil {
		log.WithError(err).Errorf("force cancel %d failed", dag.GetID())
	} else {
		// Wait sub dag to be cancelled.
		log.Infof("wait %d to be cancelled", dag.GetID())
		for i := 0; i < maxQuerySubDagDetailTimes; i++ {
			time.Sleep(time.Second)
			dag, err := localTaskService.GetDagInstance(dag.GetID())
			if err != nil {
				log.WithError(err).Error("get last maintenance dag failed")
				continue
			}
			if dag.IsFail() {
				return nil
			}
		}
	}
	return fmt.Errorf("cancel %d timeout after %d seconds", dag.GetID(), maxQuerySubDagDetailTimes)
}

func GenerateTargetAgentList(scope param.Scope) ([]meta.AgentInfo, error) {
	targetAgents := make([]meta.AgentInfo, 0)
	agentList, err := agentService.GetAllAgentInstances()
	if err != nil {
		log.WithError(err).Error("get agent list failed")
		return nil, err
	}
	switch scope.Type {
	case SCOPE_GLOBAL:
		for _, a := range agentList {
			targetAgents = append(targetAgents, a.AgentInfo)
		}
	case SCOPE_ZONE:
		for _, a := range agentList {
			for _, zone := range scope.Target {
				if a.Zone == zone {
					targetAgents = append(targetAgents, a.AgentInfo)
					break
				}
			}
		}
	case SCOPE_SERVER:
		for _, server := range scope.Target {
			info, err := meta.ConvertAddressToAgentInfo(server)
			if err != nil {
				log.WithError(err).Errorf("parse server '%s' failed", server)
				return nil, err
			}

			targetAgents = append(targetAgents, *info)
		}
	}
	return targetAgents, nil
}
