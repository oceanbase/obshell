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
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
)

type ConvertFollowerToClusterAgentTask struct {
	task.Task
}

type ConvertMasterToClusterAgentTask struct {
	task.Task
}

func newConvertFollowerToClusterAgentTask() *ConvertFollowerToClusterAgentTask {
	newTask := &ConvertFollowerToClusterAgentTask{
		Task: *task.NewSubTask(TASK_CONVERT_FOLLOWER_TO_CLUSTER),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func newConvertMasterToClusterAgentTask() *ConvertMasterToClusterAgentTask {
	newTask := &ConvertMasterToClusterAgentTask{
		Task: *task.NewSubTask(TASK_CONVERT_MASTER_TO_CLUSTER),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func newConvertClusterTemplate() *task.Template {
	return task.NewTemplateBuilder("").
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newConvertFollowerToClusterAgentTask(), false).
		AddTask(newAgentSyncTask(), true).
		AddTask(newConvertMasterToClusterAgentTask(), false).
		Build()
}

func newConvertClusterContext() (*task.TaskContext, error) {
	password, err := secure.Encrypt(meta.OCEANBASE_PWD)
	if err != nil {
		return nil, err
	}
	agents, err := agentService.GetAllActiveServerAgentsFromOB()
	if err != nil {
		return nil, err
	} else if len(agents) == 0 {
		return nil, errors.Occur(errors.ErrAgentNoActiveServer)
	}
	return task.NewTaskContext().SetParam(task.EXECUTE_AGENTS, agents).SetData(PARAM_ROOT_PWD, password), nil
}

func (t *ConvertFollowerToClusterAgentTask) Execute() (err error) {
	if _, err = oceanbase.GetOcsInstance(); err != nil {
		t.ExecuteLog("connect to oceanbase")
		if err := oceanbase.LoadOceanbaseInstance(config.NewObDataSourceConfig().SetPassword(meta.OCEANBASE_PWD)); err != nil {
			return err
		}
	}
	var agents []meta.AgentInstance
	if meta.OCS_AGENT.IsMasterAgent() {
		agents, err = agentService.GetFollowerAgentsFromOB()
		if err != nil {
			return err
		}
	} else {
		agents, err = agentService.GetTakeOverFollowerAgentsFromOB()
		if err != nil {
			return err
		}
	}
	for _, agent := range agents {
		t.ExecuteLogf("convert agent %s to cluster agent", agent.String())
		if err := agentService.ConvertToClusterAgent(&agent); err != nil {
			return err
		}
	}
	return nil
}

func (t *ConvertMasterToClusterAgentTask) Execute() error {
	// update self's identity in all_agent.
	t.ExecuteLog("convert self to cluster agent")
	if err := agentService.ConvertToClusterAgent(meta.OCS_AGENT); err != nil {
		return err
	}

	t.ExecuteLog("sync agent binary")
	if err := syncAgentBinary(); err != nil {
		return err
	}

	t.ExecuteLog("synchronize agent from oceanbase")
	if err := agentService.SyncAgentData(); err != nil {
		return err
	}
	return nil
}
