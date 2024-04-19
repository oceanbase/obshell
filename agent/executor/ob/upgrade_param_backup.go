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
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

type BackupParametersTask struct {
	task.Task
}

func newBackupParametersTask() *BackupParametersTask {
	newTask := &BackupParametersTask{
		Task: *task.NewSubTask(TASK_BACKUP_PARAMETERS),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanCancel()
	return newTask
}

func (t *BackupParametersTask) Execute() (err error) {
	t.ExecuteLog("Starting backup of parameters.")
	if err = t.BackupParameters(); err != nil {
		return
	}
	return nil
}

func (t *BackupParametersTask) BackupParameters() (err error) {
	t.ExecuteLogf("backup %v", needBackupParamName)
	paramsForUpgrade, err := obclusterService.GetObParametersForUpgrade(needBackupParamName)
	if err != nil {
		return err
	}
	t.GetContext().SetParam(PARAM_OB_PARAMETERS, paramsForUpgrade)
	return nil
}

func (t *BackupParametersTask) Rollback() (err error) {
	restoreParamTask := &RestoreParametersTask{
		Task: t.Task,
	}
	restoreParamTask.SetContext(t.GetContext())
	t.ExecuteLog("restore parameters")
	return restoreParamTask.Execute()
}

var needBackupParamName = []string{
	"server_permanent_offline_time",
	"enable_rebalance",
	"enable_rereplication",
}

func ParamsBackup() (params []oceanbase.ObParameters, error *errors.OcsAgentError) {
	log.Infof("backup params: %v", needBackupParamName)
	paramsForUpgrade, err := obclusterService.GetObParametersForUpgrade(needBackupParamName)
	if err != nil {
		log.WithError(err).Error("get ob parameters failed")
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return paramsForUpgrade, nil
}
