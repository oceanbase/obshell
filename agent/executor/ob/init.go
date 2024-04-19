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

import "github.com/oceanbase/obshell/agent/engine/task"

func CreateInitDag() (*task.DagDetailDTO, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}

	template := task.NewTemplateBuilder(DAG_INIT_CLUSTER).
		SetMaintenance(true).
		AddTask(newIntegrateObConfigTask(), false).
		AddTask(newDeployTask(), true).
		AddTask(newStartObServerTask(), true).
		AddTask(newClusterBoostrapTask(), false).
		AddTask(newMigrateTableTask(), false).
		AddTask(newModifyPwdTask(), false).
		AddTask(newMigrateDataTask(), false).
		AddTemplate(newConvertClusterTemplate()).
		AddTask(newAgentSyncTask(), true).
		Build()

	ctx := task.NewTaskContext().SetParam(task.EXECUTE_AGENTS, agents)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}
