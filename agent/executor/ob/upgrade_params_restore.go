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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

type RestoreParametersTask struct {
	task.Task
	params []oceanbase.ObParameters
}

func newRestoreParametersTask() *RestoreParametersTask {
	newTask := &RestoreParametersTask{
		Task: *task.NewSubTask(TASK_RESTORE_PARAMETERS),
	}
	newTask.SetCanContinue().SetCanRetry()
	return newTask
}

func (t *RestoreParametersTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_OB_PARAMETERS, &t.params); err != nil {
		return err
	}
	return nil
}

func (t *RestoreParametersTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}
	t.ExecuteLog("start to restore parameters")
	for _, param := range t.params {
		t.ExecuteLogf("restore param: %v", param.Name)
	}
	if err = obclusterService.RestoreParamsForUpgrade(t.params); err != nil {
		return err
	}
	return nil
}

func ParamsRestore(param param.RestoreParams) error {
	log.Infof("restore params: %v", param.Params)
	if err := obclusterService.RestoreParamsForUpgrade(param.Params); err != nil {
		return err
	}
	return nil
}
