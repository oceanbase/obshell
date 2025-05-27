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

package obproxy

import (
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
)

func StopObproxy() (*task.DagDetailDTO, *errors.OcsAgentError) {
	if !meta.IsObproxyAgent() {
		return nil, errors.Occur(errors.ErrBadRequest, "This is not an obproxy agent")
	}

	template := task.NewTemplateBuilder(DAG_STOP_OBPROXY).
		SetMaintenance(task.ObproxyMaintenance()).
		SetType(task.DAG_OBPROXY).
		AddNode(newPrepareForObproxyAgentNode(true)).
		AddTask(newStopObproxyTask(), false).Build()

	ctx := task.NewTaskContext().SetParam(PARAM_OBPROXY_HOME_PATH, meta.OBPROXY_HOME_PATH)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

// StopObproxyTask will stop obproyxd and obproxy.
type StopObproxyTask struct {
	task.Task
}

func newStopObproxyTask() *StopObproxyTask {
	newTask := &StopObproxyTask{
		Task: *task.NewSubTask(TASK_STOP_OBPROXY),
	}
	newTask.SetCanRetry().SetCanContinue()
	return newTask
}

func (t *StopObproxyTask) Execute() error {
	// if err := t.stopObproxyd(); err != nil {
	// 	return err
	// }
	if err := t.stopObproxy(); err != nil {
		return err
	}
	return nil
}

func (t *StopObproxyTask) stopObproxy() error {
	pid, err := process.GetObproxyPid()
	if err != nil {
		return err
	}
	t.ExecuteLogf("Get obproxy pid: %s", pid)
	if pid == "" {
		t.ExecuteLog("Obproxy is not running")
		return nil
	}
	for i := 0; i < STOP_PROCESS_MAX_RETRY_TIME; i++ {
		t.ExecuteLogf("Kill obproxy process %s", pid)
		res := exec.Command("kill", "-9", pid)
		if err := res.Run(); err != nil {
			log.Warn("Kill obproxy process failed")
		}

		time.Sleep(time.Second * time.Duration(STOP_PROCESS_RETRY_INTERVAL))
		t.TimeoutCheck()

		t.ExecuteLog("Check obproxy process")
		exist, err := process.CheckObproxyProcess()
		if err != nil {
			log.Warnf("Check obproxy process failed: %v", err)
		} else if !exist {
			t.ExecuteLog("Successfully killed the obproxy process")
			return nil
		}
	}
	return errors.New("kill obproxy process timeout")
}

func (t *StopObproxyTask) stopObproxyd() error {
	pid, err := process.GetObproxydPid()
	if err != nil {
		return err
	}
	t.ExecuteLogf("Get obproxyd pid: %s", pid)
	if pid == "" {
		t.ExecuteLog("Obproxyd is not running")
		return nil
	}
	for i := 0; i < STOP_PROCESS_MAX_RETRY_TIME; i++ {
		t.ExecuteLogf("Kill obproxyd process %s", pid)
		res := exec.Command("kill", "-9", pid)
		if err := res.Run(); err != nil {
			log.Warn("Kill obproxyd process failed")
		}

		time.Sleep(time.Second * time.Duration(STOP_PROCESS_RETRY_INTERVAL))
		t.TimeoutCheck()

		t.ExecuteLog("Check obproxyd process")
		exist, err := process.CheckObproxydProcess()
		if err != nil {
			log.Warnf("Check obproxyd process failed: %v", err)
		} else if !exist {
			t.ExecuteLog("Successfully killed the obproxyd process")
			return nil
		}
	}
	return errors.New("kill obproxyd process timeout")
}
