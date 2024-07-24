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

package agent

import (
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
)

type AgentJoinSelfTask struct {
	task.Task
}

func CreateJoinSelfDag(zone string) (*task.Dag, error) {
	// Agent receive api to join self, then create a task to be master.
	builder := task.NewTemplateBuilder(DAG_JOIN_SELF)
	newTask := &AgentJoinSelfTask{
		Task: *task.NewSubTask(TASK_JOIN_SELF),
	}
	newTask.SetCanContinue()
	builder.AddTask(newTask, false)

	builder.SetMaintenance(task.GlobalMaintenance())
	template := builder.Build()

	ctx := task.NewTaskContext().SetParam(PARAM_ZONE, zone)
	return localTaskService.CreateDagInstanceByTemplate(template, ctx)
}

func (t *AgentJoinSelfTask) Execute() error {
	if t.IsContinue() && meta.OCS_AGENT.IsMasterAgent() {
		t.ExecuteLog("agent is master agent")
		return nil
	}
	if !meta.OCS_AGENT.IsSingleAgent() {
		return errors.New("agent is not single")
	}

	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("zone is not set")
	}
	if err := agentService.BeMasterAgent(zone); err != nil {
		return err
	}
	t.ExecuteLog("set agent identity to master")
	return nil
}
