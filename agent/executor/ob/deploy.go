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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
)

type DeployTask struct {
	RemoteExecutableTask
	dirs map[string]string
}

func newDeployTask() *DeployTask {
	newTask := &DeployTask{
		RemoteExecutableTask: *newRemoteExecutableTask(TASK_NAME_DEPLOY),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func CreateDeploySelfDag(dirs map[string]string) (*task.DagDetailDTO, error) {
	if err := checkItem(dirs); err != nil {
		return nil, err
	}
	subTask := newDeployTask()
	builder := task.NewTemplateBuilder(subTask.GetName())
	builder.AddTask(subTask, false).SetMaintenance(task.GlobalMaintenance())
	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext().SetAgentData(meta.OCS_AGENT, PARAM_DIRS, dirs))
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func (t *DeployTask) Execute() error {
	agent := t.GetExecuteAgent()
	if err := t.GetContext().GetAgentDataWithValue(&agent, PARAM_DIRS, &t.dirs); err != nil {
		return err
	}
	if !agent.Equal(meta.OCS_AGENT) {
		return t.remoteDeploy()
	}
	if err := t.checkItem(); err != nil {
		return err
	}
	if err := t.mkdDataDir(); err != nil {
		return err
	}
	if err := t.symLinkDataDir(); err != nil {
		return err
	}
	if err := t.mkUnConfigDir(); err != nil {
		return err
	}
	return nil
}

func (t *DeployTask) Rollback() error {
	agent := t.GetExecuteAgent()
	if !agent.Equal(meta.OCS_AGENT) {
		return t.remoteDestroy()
	}
	if err := t.GetContext().GetAgentDataWithValue(&agent, PARAM_DIRS, &t.dirs); err != nil {
		return errors.Wrap(err, "agent data not set")
	}
	if err := rmStoreDir(t); err != nil {
		return err
	}
	if err := clearUnConfigDir(t); err != nil {
		return err
	}
	return nil
}

func (t *DeployTask) remoteDeploy() error {
	t.initial(constant.URI_OB_RPC_PREFIX+constant.URI_DEPLOY, http.POST, param.DeployTaskParams{Dirs: t.dirs})
	return t.retmoteExecute()
}

func (t *DeployTask) remoteDestroy() error {
	t.initial(constant.URI_OB_RPC_PREFIX+constant.URI_DESTROY, http.POST, nil)
	t.rollbackTaskName = TASK_NAME_DESTROY
	return t.remoteRollback()
}

func checkItem(dirs map[string]string) error {
	for _, key := range allDirOrder {
		if _, ok := dirs[key]; !ok {
			return errors.Errorf("dir '%s' unset", key)
		}
	}
	return nil
}

func (t *DeployTask) checkItem() error {
	return checkItem(t.dirs)
}

func (t *DeployTask) mkdDataDir() error {
	for _, key := range storeDirOrder {
		dir := t.dirs[key]
		t.ExecuteLogf("mkdir %s: %s", key, dir)
		if err := mkdir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (t *DeployTask) symLinkDataDir() error {
	obDirs := map[string]string{
		constant.CONFIG_HOME_PATH: t.dirs[constant.CONFIG_HOME_PATH],
	}
	if err := fillDir(obDirs); err != nil {
		return err
	}
	for _, key := range storeDirOrder {
		dst, src := obDirs[key], t.dirs[key]
		t.ExecuteLogf("symlink %s: %s", src, dst)
		if err := symLink(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func (t *DeployTask) mkUnConfigDir() error {
	homePath := t.dirs[constant.CONFIG_HOME_PATH]
	for _, dir := range unconfigurableDirList {
		path := filepath.Join(homePath, dir)
		t.ExecuteLogf("mkdir %s", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func symLink(src, dst string) error {
	if src == dst {
		return nil
	}
	realSrc, err := filepath.EvalSymlinks(src)
	if err != nil {
		return errors.Wrapf(err, "eval symlink '%s' failed", src)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if realSrc == dst {
		return nil
	}
	if err := os.Symlink(src, dst); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func mkdir(dir string) error {
	parentDir := filepath.Dir(dir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
		// If dir exists, check dir is empty
		if isEmpty, err := checkDirIsEmpty(dir); err != nil {
			return errors.Wrapf(err, "check dir '%s' is empty failed", dir)
		} else if !isEmpty {
			return errors.Errorf("dir '%s' is not empty", dir)
		}
	}
	return nil
}

func checkDirIsEmpty(dirName string) (bool, error) {
	if dir, err := os.ReadDir(dirName); err != nil {
		return false, err
	} else {
		return len(dir) == 0, nil
	}
}
