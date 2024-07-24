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
	"os"
	"path/filepath"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
)

type DestroyTask struct {
	RemoteExecutableTask
}

func newDestroyTask() *DestroyTask {
	newTask := &DestroyTask{
		RemoteExecutableTask: *newRemoteExecutableTask(TASK_NAME_DESTROY),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry()
	return newTask
}

func CreateDestroyDag() (*task.DagDetailDTO, error) {
	subTask := newDestroyTask()
	builder := task.NewTemplateBuilder(subTask.GetName())
	builder.AddTask(subTask, false).SetMaintenance(task.GlobalMaintenance())
	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext())
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func (t *DestroyTask) Execute() error {
	if err := clearUnConfigDir(t); err != nil {
		return err
	}
	if err := rmStoreDir(t); err != nil {
		return err
	}
	return nil
}

func rmStoreDir(t task.ExecutableTask) error {
	obDirs := map[string]string{
		constant.CONFIG_HOME_PATH: global.HomePath,
	}
	if err := fillDir(obDirs); err != nil {
		return err
	}

	needClearDirs := buildNeedClearDirs(obDirs)
	for _, path := range needClearDirs {
		t.ExecuteLogf("clear dir %s", path)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func buildNeedClearDirs(dirs map[string]string) []string {
	var needClearDirs []string
	for _, key := range storeDirOrder {
		path := dirs[key]
		if _, err := os.Stat(path); err != nil {
			continue
		}
		needClearDirs = append(needClearDirs, path)

		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		if realPath != path {
			needClearDirs = append(needClearDirs, realPath)
		}
	}
	return needClearDirs
}

func clearUnConfigDir(t task.ExecutableTask) error {
	for dirName, prefix := range clearDirMap {
		path := filepath.Join(global.HomePath, dirName)
		t.ExecuteLogf("clear dir %s", path)
		if err := clearDir(path, prefix); err != nil {
			return err
		}
	}
	return nil
}

func clearDir(path string, prefix string) error {
	if prefix == "" {
		return os.RemoveAll(path)
	}
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), prefix) {
			if err := os.Remove(path); err != nil {
				return errors.Wrapf(err, "remove file %s failed", path)
			}
		}
		return nil
	})
}
