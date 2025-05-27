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
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/secure"
)

// CreateDefaultUserTask creates default user, currently limited to 'proxyro'.
type CreateDefaultUserTask struct {
	task.Task
	encryptProxyroPassword string
}

func newCreateDefaultUserNode(proxyroPassword string) (*task.Node, error) {
	subtask := newCreateDefaultUserTask()
	encryptedProxyroPassword, err := secure.Encrypt(proxyroPassword)
	if err != nil {
		return nil, err
	}
	ctx := task.NewTaskContext().SetParam(PARAM_PROXYRO_PASSWORD, encryptedProxyroPassword)
	return task.NewNodeWithContext(subtask, false, ctx), nil
}

func newCreateDefaultUserTask() *CreateDefaultUserTask {
	newTask := &CreateDefaultUserTask{
		Task: *task.NewSubTask(TASK_NAME_CREATE_USER),
	}
	newTask.SetCanRetry().SetCanContinue().SetCanCancel()
	return newTask
}

func (t *CreateDefaultUserTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_PROXYRO_PASSWORD, &t.encryptProxyroPassword); err != nil {
		return err
	}

	// decrypt password
	proxyroPassword, err := secure.Decrypt(t.encryptProxyroPassword)
	if err != nil {
		return err
	}

	if err := obclusterService.CreateProxyroUser(proxyroPassword); err != nil {
		return err
	}
	return nil
}
