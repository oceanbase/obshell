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
	"errors"
	"time"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

const WAIT_START_OBSERVER_TIMES = 360
const WAIT_START_OBSERVER_INTERVAL = 5 * time.Second

func CreateKillObserverDag() (*task.DagDetailDTO, error) {
	template := task.NewTemplateBuilder(DAG_KILL_OBSERVER).
		AddTask(newKillObserverTask(), false).
		SetMaintenance(task.GlobalMaintenance()).
		Build()
	context := task.NewTaskContext().
		SetParam(PARAM_DELETE_AGENTS, append([]meta.AgentInfo{}, meta.OCS_AGENT.GetAgentInfo())).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

type KillObserverTask struct {
	task.Task
}

func newKillObserverTask() *KillObserverTask {
	newTask := &KillObserverTask{
		Task: *task.NewSubTask(TASK_NAME_KILL_OBSERVER),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *KillObserverTask) Execute() error {
	return stopObserver(t)
}

func CreateStartObserverDag() (*task.DagDetailDTO, error) {
	if oceanbase.GetState() != oceanbase.STATE_PROCESS_NOT_RUNNING {
		return nil, nil
	}
	isRunning, err := localTaskService.IsRunning()
	if err != nil {
		return nil, err
	}
	if !isRunning {
		dag, err := localTaskService.GetLastMaintenanceDag()
		if err != nil {
			return nil, err
		}
		if dag != nil && dag.GetName() == DAG_START_OBSERVER_FOR_SCALE_IN_ROLLBACK {
			return task.NewDagDetailDTO(dag), nil
		} else if dag.GetName() == DAG_KILL_OBSERVER {
			if err := localTaskService.CancelDag(dag); err != nil {
				return nil, err
			}
		}
	}

	template := task.NewTemplateBuilder(DAG_START_OBSERVER_FOR_SCALE_IN_ROLLBACK).
		AddTask(newStartObserverForScaleInRollbackTask(), false).
		SetMaintenance(task.GlobalMaintenance()).
		Build()
	context := task.NewTaskContext().
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

type StartObserverForScaleInRollbackTask struct {
	task.Task
}

func newStartObserverForScaleInRollbackTask() *StartObserverForScaleInRollbackTask {
	newTask := &StartObserverForScaleInRollbackTask{
		Task: *task.NewSubTask(TASK_NAME_START_OBSERVER_FOR_SCALE_IN_ROLLBACK),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *StartObserverForScaleInRollbackTask) Execute() error {
	if err := startObserver(t, nil); err != nil {
		return err
	}

	// It will time out after 30 minutes.
	t.ExecuteInfoLog("Waiting for observer to start")
	for retryCount := 1; retryCount <= WAIT_START_OBSERVER_TIMES; retryCount++ {
		t.TimeoutCheck()
		_, err := oceanbase.GetInstance()
		if err == nil {
			return nil
		}
		if errors.Is(err, oceanbase.ERR_OBSERVER_NOT_EXIST) {
			return err
		}
		time.Sleep(WAIT_START_OBSERVER_INTERVAL)
	}
	return errors.New("launch observer failed")
}
