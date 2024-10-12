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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

func TakeOver() (err error) {
	if err = oceanbase.AutoMigrateObTables(true); err != nil {
		return err
	}

	targetVersion, err := needUpdateAgentBinary()
	if err != nil {
		return
	}

	lk, err := obclusterService.GetClusterStatusLock()
	if err != nil {
		return errors.Wrap(err, "get db lock failed")
	}

	if err = lk.Lock(); err != nil {
		return errors.New("lock cluster status failed")
	}
	defer func() {
		if _err := lk.Unlock(); _err != nil {
			log.WithError(_err).Error("unlock cluster status failed")
		} else {
			log.Info("unlock cluster status succeed")
		}
	}()
	log.Info("lock cluster status succeed")

	takeOverMaster, err := agentService.GetTakeOverMasterAgent()
	if err != nil {
		return errors.Wrap(err, "get take over master filed")
	}
	if takeOverMaster != nil {
		// Self become TAKE_OVER_FOLLOWER.
		return agentService.CreateTakeOverAgent(meta.FOLLOWER)
	}
	canBeTakeOverMaster, err := agentService.CheckCanBeTakeOverMaster()
	if err != nil {
		return errors.Wrap(err, "check can be take over master failed")
	}
	// If the last one, self become TAKE_OVER_MASTER.
	if canBeTakeOverMaster {
		if err = agentService.CreateTakeOverAgent(meta.TAKE_OVER_MASTER); err != nil {
			return errors.Wrap(err, "create take over agent failed")
		}
		return createTakeOverDag(targetVersion)
	} else {
		return agentService.CreateTakeOverAgent(meta.TAKE_OVER_FOLLOWER)
	}
}

func createTakeOverDag(targetVersion string) error {
	isRunning, err := localTaskService.IsRunning()
	if err != nil {
		return err
	} else if !isRunning {
		log.Infof("The agent is already under maintenance.")
		return nil
	}
	log.Info("create take over dag")
	builder := task.NewTemplateBuilder(DAG_TAKE_OVER).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newAgentSyncTask(), false)

	if targetVersion != "" {
		builder.AddTask(newTakeOverAgentUpdateBinaryTask(), true)
	}

	template := builder.
		AddTemplate(newConvertClusterTemplate()).
		AddTask(newAgentSyncTask(), true).
		Build()

	ctx, err := newTakeOverContext(targetVersion)
	if err != nil {
		return err
	}

	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return err
	}
	log.Infof("create takeover dag '%s' success", task.NewDagDetailDTO(dag).GenericID)
	return err
}

func newTakeOverContext(targetVersion string) (*task.TaskContext, error) {
	ctx, err := newConvertClusterContext()
	if err != nil {
		return nil, errors.Wrap(err, "new convert cluster context failed")
	}
	if targetVersion != "" {
		ctx.SetParam(PARAM_TARGET_AGENT_VERSION, targetVersion)
	}

	takeoverFollowers, err := agentService.GetTakeOverFollowerAgentsFromOB()
	if err != nil {
		return nil, errors.Wrap(err, "get take over follower agents failed")
	}
	var agents []meta.AgentInfo
	for _, follower := range takeoverFollowers {
		agents = append(agents, follower.AgentInfo)
	}
	agents = append(agents, meta.OCS_AGENT.GetAgentInfo())

	ctx.SetParam(task.EXECUTE_AGENTS, agents)
	return ctx, nil
}

type TakeOverAgentUpdateBinaryTask struct {
	task.Task

	RemoteExecutableTask
	UpgradeToClusterAgentVersionTask
}

func newTakeOverAgentUpdateBinaryTask() *TakeOverAgentUpdateBinaryTask {
	newTask := &TakeOverAgentUpdateBinaryTask{
		Task: *task.NewSubTask(DAG_TAKE_OVER_UPDATE_AGENT_VERSION),
	}
	newTask.
		SetCanRetry().
		SetCanRollback().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *TakeOverAgentUpdateBinaryTask) Execute() (err error) {
	agent := t.GetExecuteAgent()
	if !meta.OCS_AGENT.Equal(&agent) {
		return t.retmoteSync()
	}

	t.UpgradeToClusterAgentVersionTask.Task = t.Task
	return t.UpgradeToClusterAgentVersionTask.Execute()
}

func (t *TakeOverAgentUpdateBinaryTask) retmoteSync() error {
	t.RemoteExecutableTask.Task = t.Task
	t.initial(constant.URI_AGENT_RPC_PREFIX+constant.URI_SYNC_BIN, http.POST, nil)
	return t.retmoteExecute()
}
