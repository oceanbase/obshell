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
	"fmt"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
)

type StopZoneTask struct {
	task.Task
	zone string
}

// newStopZoneNode will create a node which can't be rollback.
func newStopZoneNode(zone string) *task.Node {
	ctx := task.NewTaskContext().SetParam(PARAM_ZONE, zone)
	name := fmt.Sprintf("Stop %s", zone)
	return task.NewNodeWithContext(&StopZoneTask{
		Task: *task.NewSubTask(name).
			SetCanContinue().
			SetCanRetry()},
		false, ctx)
}

// newStopZoneNodeForDelete will create a node which can be rollback.
func newStopZoneNodeForDelete(zone string) *task.Node {
	ctx := task.NewTaskContext().SetParam(PARAM_ZONE, zone)
	subTask := *task.NewSubTask(fmt.Sprintf(TASK_NAME_STOP_ZONE, zone)).
		SetCanContinue().
		SetCanRetry().
		SetCanRollback().
		SetCanPass().
		SetCanCancel()
	return task.NewNodeWithContext(
		&StopZoneTask{Task: subTask},
		false,
		ctx)
}

func (t *StopZoneTask) getParams() (err error) {
	return t.GetContext().GetParamWithValue(PARAM_ZONE, &t.zone)
}

func (t *StopZoneTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}
	t.ExecuteLog("stop zone " + t.zone)
	if err = obclusterService.StopZone(t.zone); err != nil {
		return
	}
	t.ExecuteLogf("wait for %s to be inactive", t.zone)
	for i := 0; i < constant.TICK_NUM_FOR_OB_STATUS_CHECK; i++ {
		zoneIsInactive, err := obclusterService.IsZoneInactive(t.zone)
		if err != nil {
			return err
		}
		if zoneIsInactive {
			return nil
		}
		time.Sleep(constant.TICK_INTERVAL_FOR_OB_STATUS_CHECK)
		t.TimeoutCheck()
	}
	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, "stop zone")
}

// Rollback will start the zone if it is inactive.
func (t *StopZoneTask) Rollback() error {
	if retry, err := clusterTaskService.IsRetryTask(t.Task.GetID()); err != nil {
		return err
	} else if retry {
		return nil
	}

	if err := t.getParams(); err != nil {
		return err
	}
	t.ExecuteLogf("rollback stop zone %s", t.zone)
	// Check if the zone is inactive, if not, start it
	zoneIsInactive, err := obclusterService.IsZoneInactive(t.zone)
	if err != nil {
		return errors.Wrap(err, "check zone status failed")
	}
	if zoneIsInactive {
		return obclusterService.StartZone(t.zone)
	} else {
		t.ExecuteLogf("no need to start '%s' because it is not exist or active", t.zone)
	}
	return nil
}
