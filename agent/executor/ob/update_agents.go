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
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

type UpdateAllAgentsTask struct {
	task.Task
}

func newUpdateAllAgentsTask() *UpdateAllAgentsTask {
	newTask := &UpdateAllAgentsTask{
		Task: *task.NewSubTask(TASK_NAME_UPDATE_AGENT),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *UpdateAllAgentsTask) Execute() error {
	if _, err := oceanbase.GetOcsInstance(); err != nil {
		t.ExecuteLog("connect to oceanbase")
		if err := oceanbase.LoadOceanbaseInstance(config.NewObDataSourceConfig().SetPassword(meta.OCEANBASE_PWD)); err != nil {
			return err
		}
	}
	t.ExecuteLog("get all agents")
	agents, err := agentService.GetAllAgents()
	if err != nil {
		return err
	}

	for _, agent := range agents {
		t.ExecuteLogf("convert agent %s to cluster agent", agent.String())
		if err := agentService.ConvertToClusterAgent(&agent); err != nil {
			return err
		}
	}
	return nil
}
