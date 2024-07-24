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
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

func TakeOver() (err error) {
	if err = oceanbase.ParallelAutoMigrateObTables(); err != nil {
		return err
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
		return createTakeOverDag()
	} else {
		return agentService.CreateTakeOverAgent(meta.TAKE_OVER_FOLLOWER)
	}
}

func createTakeOverDag() error {
	isRunning, err := localTaskService.IsRunning()
	if err != nil {
		return err
	} else if !isRunning {
		log.Infof("The agent is already under maintenance.")
		return nil
	}
	log.Info("create take over dag")
	template := task.NewTemplateBuilder(DAG_TAKE_OVER).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newAgentSyncTask(), false).
		AddTemplate(newConvertClusterTemplate()).
		AddTask(newAgentSyncTask(), true).
		Build()

	ctx, err := newConvertClusterContext()
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
