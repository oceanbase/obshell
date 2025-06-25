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
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

func EmergencyStart() (*task.DagDetailDTO, error) {
	template := task.NewTemplateBuilder(DAG_EMERGENCY_START).
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCheckObserverForStartTask(), false).
		AddTask(newStartObServerTask(), false).
		AddTask(newGetConnForEStartTask(), false)

	taskCtx := task.NewTaskContext().SetParam(PARAM_START_OWN_OBSVR, true)

	dag, err := localTaskService.CreateDagInstanceByTemplate(template.Build(), taskCtx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

type GetConnForEStartTask struct {
	task.Task
}

func newGetConnForEStartTask() *GetConnForEStartTask {
	newTask := &GetConnForEStartTask{
		Task: *task.NewSubTask(TASK_NAME_GET_CONN_FOR_EMERGENCY_START),
	}
	newTask.SetCanContinue().SetCanPass()
	return newTask
}

func (t *GetConnForEStartTask) Execute() error {
	t.ExecuteLog("try to get db connection")
	var err error
	var db *gorm.DB
	for i := 0; i < constant.MAX_GET_INSTANCE_RETRIES; i++ {
		if db, err = oceanbase.GetRestrictedInstance(); db != nil {
			return nil
		} else {
			log.Error("get db connection failed", err)
		}
		time.Sleep(constant.GET_INSTANCE_RETRY_INTERVAL * time.Second)
		t.TimeoutCheck()
	}
	return errors.Wrap(err, "get connection failed")
}
