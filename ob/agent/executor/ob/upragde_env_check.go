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
	"bytes"
	"os/exec"
	"strings"

	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
)

type CheckEnvTask struct {
	task.Task
}

func newCheckEnvTask() *CheckEnvTask {
	newTask := &CheckEnvTask{
		Task: *task.NewSubTask(TASK_CHECK_PYTHON_ENV),
	}
	newTask.
		SetCanContinue().
		SetCanRollback().
		SetCanRetry().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *CheckEnvTask) Execute() (err error) {
	if isRealExecuteAgent, _, err := isRealExecuteAgent(t); err != nil {
		return err
	} else if !isRealExecuteAgent {
		return nil
	}

	if err = t.checkEnv(); err != nil {
		return
	}
	t.ExecuteLog("check env success")
	return nil
}

func (t *CheckEnvTask) checkEnv() (err error) {
	t.ExecuteLog("Checking if python is installed.")
	cmd := exec.Command("python", "-c", "import sys; print(sys.version_info.major)")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return errors.Occur(errors.ErrEnvironmentWithoutPython)
	}
	output := strings.TrimSpace(out.String())
	t.ExecuteLogf("Python major version %s", output)

	for _, module := range modules {
		t.ExecuteLogf("Checking if python module '%s' is installed.", module)
		cmd = exec.Command("python", "-c", "import "+module)
		if err = cmd.Run(); err != nil {
			return errors.Occur(errors.ErrEnvironmentWithoutPythonModule, module)
		}
	}
	return nil
}
