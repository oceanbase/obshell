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
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/executor/script"
	"github.com/oceanbase/obshell/agent/lib/binary"
	"github.com/oceanbase/obshell/param"
)

func CreateInitDag(param param.ObInitParam) (*task.DagDetailDTO, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}

	builder := task.NewTemplateBuilder(DAG_INIT_CLUSTER).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newIntegrateObConfigTask(), false).
		AddTask(newDeployTask(), true).
		AddTask(newStartObServerTask(), true).
		AddTask(newClusterBoostrapTask(), false).
		AddTask(newMigrateTableTask(), false).
		AddTask(newModifyPwdTask(), false)
	if param.CreateProxyroUser {
		createUserNode, err := newCreateDefaultUserNode(param.ProxyroPassword)
		if err != nil {
			return nil, err
		}
		builder.AddNode(createUserNode)
	}
	builder.AddTask(newMigrateDataTask(), false).
		AddTemplate(newConvertClusterTemplate())
	if param.ImportScript {
		builder.AddNode(script.NewImportScriptForTenantNode(false))
	}
	obVersion, isCommunityEdition, _ := binary.GetMyOBVersion() // ignore the error
	if obVersion > OB_VERSION_4_3_5_2 && isCommunityEdition {
		builder.AddTask(newPostInitClusterTask(), false)
	}

	template := builder.AddTask(newAgentSyncTask(), true).Build()

	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_HEALTH_CHECK, true).
		SetParam(PARAM_TENANT_NAME, constant.TENANT_SYS)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

type PostInitClusterTask struct {
	task.Task
}

func newPostInitClusterTask() *PostInitClusterTask {
	newTask := &PostInitClusterTask{
		Task: *task.NewSubTask(TASK_NAME_POST_INIT_CLUSTER),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *PostInitClusterTask) Execute() error {
	t.ExecuteLog("post init cluster")
	// Set parameter 'global_index_auto_split_policy' to 'ALL'
	// Ignore the error, because it won't affect the cluster initialization
	obclusterService.SetParameter(
		param.SetParameterParam{
			Name:  constant.PARAMETER_GLOBAL_INDEX_AUTO_SPLIT_POLICY,
			Value: "ALL",
		},
	)
	return nil
}
