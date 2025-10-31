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
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
)

const DEFAULT_MINOR_FREEZE_TIMEOUT = 120

type MinorFreezeTask struct {
	task.Task
}

func newMinorFreezeTask() *MinorFreezeTask {
	newTask := &MinorFreezeTask{
		Task: *task.NewSubTask(TASK_NAME_MINOR_FREEZE),
	}
	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

func (t *MinorFreezeTask) Execute() error {
	checkpointScn, err := obclusterService.GetServerCheckpointScn()
	if err != nil {
		return errors.Wrap(err, "get server checkpoint_scn failed")
	}
	t.ExecuteLogf("checkpoint_scn before minor freeze: %v", checkpointScn)

	if err := obclusterService.MinorFreeze(); err != nil {
		return errors.Wrap(err, "minor freeze failed")
	}
	t.ExecuteLogf("minor freeze server")

	for count := 0; count < DEFAULT_MINOR_FREEZE_TIMEOUT; count++ {
		t.TimeoutCheck()
		time.Sleep(10 * time.Second)
		if ok, err := t.isMinorFreezeOver(checkpointScn); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
	return errors.Occur(errors.ErrObClusterMinorFreezeTimeout)
}

func (t *MinorFreezeTask) isMinorFreezeOver(oldCheckpointScn uint64) (bool, error) {
	if checkpointScn, err := obclusterService.GetServerCheckpointScn(); err != nil {
		return false, errors.Wrap(err, "check minor freeze failed")
	} else if checkpointScn == 0 {
		// checkpoint_scn is 0, means there is no ls in this server
		return false, nil
	} else if checkpointScn > oldCheckpointScn {
		t.ExecuteLogf("smallest checkpoint_scn %+v bigger than expired timestamp %+v, check pass ", checkpointScn, oldCheckpointScn)
		return true, nil
	} else {
		t.ExecuteLogf("smallest checkpoint_scn: %+v smaller than expired timestamp %+v, waiting...", checkpointScn, oldCheckpointScn)
		return false, nil
	}
}
