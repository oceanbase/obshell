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
	"github.com/oceanbase/obshell/ob/agent/config"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
)

type AgentSyncTask struct {
	RemoteExecutableTask
	password       string
	cipherPassword string
}

func newAgentSyncTask() *AgentSyncTask {
	subTask := *newRemoteExecutableTask(TASK_NAME_AGENT_SYNC)
	newTask := &AgentSyncTask{
		RemoteExecutableTask: subTask,
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry()
	return newTask
}

func CreateAgentSyncDag(rootPWD string) (*task.DagDetailDTO, error) {
	if _, err := secure.Decrypt(rootPWD); err != nil {
		return nil, err
	}
	subTask := newAgentSyncTask()
	template := task.NewTemplateBuilder(subTask.GetName()).AddTask(subTask, false).SetMaintenance(task.GlobalMaintenance()).Build()
	ctx := task.NewTaskContext().SetData(PARAM_ROOT_PWD, rootPWD)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func (t *AgentSyncTask) Execute() error {
	agent := t.GetExecuteAgent()
	if err := t.setRootPWD(); err != nil {
		return err
	}

	if !meta.OCS_AGENT.Equal(&agent) {
		return t.retmoteSync()
	}

	t.ExecuteLog("try to connect")
	if err := oceanbase.LoadOceanbaseInstance(config.NewObMysqlDataSourceConfig().SetPassword(t.password)); err != nil {
		t.ExecuteLog("connect failed. try to get connection from pool")
		if _, err := oceanbase.GetOcsInstance(); err != nil {
			return err
		}
	} else {
		t.ExecuteLog("connect succeed, dump password")
		if err := secure.UpdateObPassword(t.cipherPassword); err != nil {
			return err
		}
	}
	t.ExecuteLog("synchronize agent from oceanbase")
	return t.syncAgentData()
}

func (t *AgentSyncTask) setRootPWD() (err error) {
	cipherPassword := t.GetContext().GetData(PARAM_ROOT_PWD)
	if cipherPassword != nil {
		t.cipherPassword = cipherPassword.(string)
		t.password, err = secure.Decrypt(t.cipherPassword)
	}
	return
}

func (t *AgentSyncTask) retmoteSync() error {
	agent := t.GetExecuteAgent()
	t.ExecuteLog("encrypt password for agent")
	cipherPassword, err := secure.EncryptForAgent(t.password, &agent)
	if err != nil {
		return err
	}
	params := param.SyncAgentParams{Password: cipherPassword}
	t.initial(constant.URI_AGENT_RPC_PREFIX+constant.URI_UPDATE, http.POST, params)
	return t.retmoteExecute()
}

func (t *AgentSyncTask) syncAgentData() error {
	t.ExecuteLog("sync agent data")
	if err := agentService.SyncAgentData(); err != nil {
		return err
	}
	return obclusterService.MigrateObSysParameter()
}
