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

package observer

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
)

type StartObserverTask struct {
	task.Task
}

func newStartObServerTask() *StartObserverTask {
	newTask := &StartObserverTask{
		Task: *task.NewSubTask(TASK_NAME_START),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanRollback().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *StartObserverTask) Execute() error {
	exist, err := process.CheckObserverProcess()
	if err != nil {
		return errors.Wrap(err, "check seekdb process failed")
	}
	if exist {
		t.GetContext().SetData(DATA_SKIP_START_TASK, true)
		t.ExecuteLog("seekdb started.")
	} else {
		t.ExecuteLog("start seekdb")
		if err := startObserver(t); err != nil {
			return err
		}
	}

	if err := t.observerHealthCheck(); err != nil {
		return errors.Wrap(err, "seekdb health check failed")
	}
	t.ExecuteLog("start seekdb success")

	return nil
}

func (t *StartObserverTask) observerHealthCheck() error {
	dsConfig := config.NewObMysqlDataSourceConfig().
		SetTryTimes(1).
		SetDBName("").
		SetTimeout(10).
		SetPort(meta.MYSQL_PORT).
		SetPassword(meta.GetOceanbasePwd())

	const (
		maxRetries    = 60 // It should be small for seekdb
		retryInterval = 2 * time.Second
	)

	for retryCount := 1; retryCount <= maxRetries; retryCount++ {
		time.Sleep(retryInterval)
		if retryCount%10 == 0 {
			t.TimeoutCheck()
		} else {
			t.ExecuteLogf("seekdb health check, retry [%d/%d]", retryCount, maxRetries)
		}

		// Check if the seekdb process exists
		if exist, err := process.CheckObserverProcess(); err != nil {
			return errors.Occur(errors.ErrObServerProcessCheckFailed, err.Error())
		} else if !exist {
			return errors.Occur(errors.ErrObServerProcessNotExist)
		}

		// Attempt to connect to the OceanBase instance for testing
		if err := oceanbase.LoadOceanbaseInstanceForTest(dsConfig); err != nil {
			continue // Connection failed, retry
		}

		// All checks passed, exit the loop
		return nil
	}

	// If retries run out, return a timeout error
	return errors.Occur(errors.ErrTaskDagExecuteTimeout, "seekdb health check")
}

func (t *StartObserverTask) Rollback() error {
	if _, ok := t.GetContext().GetData(DATA_SKIP_START_TASK).(bool); ok {
		return nil
	}

	return stopObserver(t)
}

func startObserver(t task.ExecutableTask) error {
	t.ExecuteLog("check if first start")
	if err := requireCheck(t); err != nil {
		return err
	}
	t.ExecuteLogf("start cmd: %s --base-dir %s", path.ObserverBinPath(), global.HomePath)
	return execStartCmd(path.ObserverBinPath(), "--base-dir", global.HomePath)
}

// SafeStartObserver is a safe method to start the seekdb, ensuring that it has been successfully started at least once before using.
// This method allows an empty config and does not check whether the config contains the necessary startup configuration items.
func SafeStartObserver() error {
	if isFirst, err := isFirstStart(); err != nil {
		return err
	} else if isFirst {
		return errors.Occur(errors.ErrObServerHasNotBeenStarted)
	}

	cmd := fmt.Sprintf("%s --base-dir %s", path.ObserverBinPath(), global.HomePath)
	log.Info("safty start seekdb, cmd: ", cmd)
	return execStartCmd(path.ObserverBinPath(), "--base-dir", global.HomePath)
}

func requireCheck(t task.ExecutableTask) error {
	if isFirst, err := isFirstStart(); err != nil {
		return err
	} else if !isFirst {
		t.ExecuteLog("not first start, skip require check")
		return nil
	}
	return nil
}

func execStartCmd(bin string, args ...string) error {
	if err := os.Chdir(global.HomePath); err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	if stderr, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, string(stderr))
	}
	return nil
}

func CreateStartDag() (*task.DagDetailDTO, error) {
	template := task.NewTemplateBuilder(DAG_START_OBSERVER).
		SetMaintenance(task.GlobalMaintenance())
	// If not need start, then skip start seekdb.
	template.AddTask(newStartObServerTask(), false)

	dag, err := localTaskService.CreateDagInstanceByTemplate(template.Build(), task.NewTaskContext().SetParam(task.FAILURE_EXIT_MAINTENANCE, true))
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}
