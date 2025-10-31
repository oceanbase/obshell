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
	"os"
	"path/filepath"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

func DeleteObproxy() (*task.DagDetailDTO, error) {
	if !meta.IsObproxyAgent() {
		return nil, nil
	}

	templateBuilder := task.NewTemplateBuilder(DAG_DELETE_OBPROXY).
		SetMaintenance(task.ObproxyMaintenance()).
		SetType(task.DAG_OBPROXY).
		AddNode(newPrepareForObproxyAgentNode(true)).
		AddTask(newStopObproxyTask(), false).
		AddTask(newDeleteObproxyTask(), false).
		AddTask(newCleanObproxyDirTask(), false)

	context := task.NewTaskContext().SetParam(PARAM_OBPROXY_HOME_PATH, meta.OBPROXY_HOME_PATH)
	dag, err := localTaskService.CreateDagInstanceByTemplate(templateBuilder.Build(), context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

// DeleteObproxyTask will delete the obproxy home path
type DeleteObproxyTask struct {
	task.Task
}

func newDeleteObproxyTask() *DeleteObproxyTask {
	newTask := &DeleteObproxyTask{
		Task: *task.NewSubTask(TASK_DELETE_OBPROXY),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanPass()
	return newTask
}

func (t *DeleteObproxyTask) Execute() (err error) {
	if err := agentService.DeleteObproxy(); err != nil {
		return err
	}
	return nil
}

type CleanObproxyDirTask struct {
	task.Task
	obproxyHomePath string
}

func newCleanObproxyDirTask() *CleanObproxyDirTask {
	newTask := &CleanObproxyDirTask{
		Task: *task.NewSubTask(TASK_CLEAN_OBPROXY_DIR),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanPass()
	return newTask
}

func (t *CleanObproxyDirTask) Execute() (err error) {
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_HOME_PATH, &t.obproxyHomePath); err != nil {
		return err
	}
	deleteFiles := []string{constant.OBPROXY_DIR_ETC, constant.OBPROXY_DIR_LOG, constant.OBPROXY_DIR_RUN,
		constant.OBPROXY_DIR_BIN, constant.OBPROXY_DIR_LIB}
	for _, file := range deleteFiles {
		if err := os.RemoveAll(filepath.Join(t.obproxyHomePath, file)); err != nil {
			return err
		}
	}
	return nil
}
