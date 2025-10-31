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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
)

type StartOneZoneTask struct {
	task.Task
	zone string
}

func newStartZoneNode(zone string) *task.Node {
	ctx := task.NewTaskContext().SetParam(PARAM_ZONE, zone)
	name := fmt.Sprintf("Start %s", zone)
	return task.NewNodeWithContext(&StartOneZoneTask{
		Task: *task.NewSubTask(name).
			SetCanRetry().
			SetCanContinue()},
		false, ctx)
}

func (t *StartOneZoneTask) getParams() (err error) {
	return t.GetContext().GetParamWithValue(PARAM_ZONE, &t.zone)
}

func (t *StartOneZoneTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}
	t.ExecuteLogf("start %s", t.zone)
	if err = obclusterService.StartZone(t.zone); err != nil {
		return
	}
	t.ExecuteLogf("wait for %s to be active", t.zone)
	for i := 0; i < constant.TICK_NUM_FOR_OB_STATUS_CHECK; i++ {
		zoneIsActive, _ := obclusterService.IsZoneActive(t.zone)
		if zoneIsActive {
			return nil
		}
		time.Sleep(constant.TICK_INTERVAL_FOR_OB_STATUS_CHECK)
		t.TimeoutCheck()
	}
	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, "start zone")
}
