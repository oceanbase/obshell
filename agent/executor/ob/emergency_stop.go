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
	"github.com/oceanbase/obshell/agent/errors"
)

func EmergencyStop() (*task.DagDetailDTO, *errors.OcsAgentError) {
	template := task.NewTemplateBuilder(DAG_EMERGENCY_STOP).
		SetMaintenance(false).
		AddTask(newStopObserverTask(), false)

	taskCtx := task.NewTaskContext()

	dag, err := localTaskService.CreateDagInstanceByTemplate(template.Build(), taskCtx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}
