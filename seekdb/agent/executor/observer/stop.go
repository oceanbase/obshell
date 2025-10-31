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

package observer

import (
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/param"
)

type StopObserverTask struct {
	task.Task
}

func newStopObserverTask() *StopObserverTask {
	newTask := &StopObserverTask{
		Task: *task.NewSubTask(TASK_NAME_STOP),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func CreateStopDag(p param.ObStopParam) (*task.DagDetailDTO, error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}

	obState := oceanbase.GetState()

	if p.Terminate {
		if obState != oceanbase.STATE_CONNECTION_AVAILABLE {
			return nil, errors.Occur(errors.ErrAgentOceanbaseUesless) // when stop with terminate, the ob state must be available
		}
	}
	if exist, err := process.CheckObserverProcess(); err != nil {
		log.Warnf("Check observer process failed: %v", err)
	} else if !exist {
		return nil, nil // when observer process not exist, return nil directly
	}

	builder := task.NewTemplateBuilder(DAG_STOP_OBSERVER)
	if p.Terminate {
		builder.AddTask(newMinorFreezeTask(), false)
	}

	builder.AddTask(newStopObserverTask(), false).SetMaintenance(task.GlobalMaintenance())
	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext().SetParam(task.FAILURE_EXIT_MAINTENANCE, true))
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

		// get observer pid to avoid observer has been restarted by other

		pidNow, err := process.GetObserverPid()
		if err != nil {
			log.Warnf("Get observer pid failed: %v", err)
		} else if pidNow != "" && pidNow != pid {
			t.ExecuteLogf("Observer has been restarted by other, new pid: %s", pidNow)
			return nil
		}

		t.ExecuteLog("Check observer process")
		exist, err := process.CheckObserverProcess()
		if err != nil {
			log.Warnf("Check observer process failed: %v", err)
		} else if !exist {
			t.ExecuteLog("Successfully killed the observer process")
			return nil
		}
	}
	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, "kill observer process")
}
