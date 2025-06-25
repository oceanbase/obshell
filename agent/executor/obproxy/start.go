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
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
)

func StartObproxy() (*task.DagDetailDTO, error) {
	if !meta.IsObproxyAgent() {
		errors.Occur(errors.ErrOBProxyNotBeManaged)
	}

	template := task.NewTemplateBuilder(DAG_START_OBPROXY).
		SetType(task.DAG_OBPROXY).
		AddNode(newPrepareForObproxyAgentNode(true)).
		AddNode(newStartObproxyWithoutOptionsNode()).Build()
	context := task.NewTaskContext().SetParam(PARAM_OBPROXY_HOME_PATH, meta.OBPROXY_HOME_PATH)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}
