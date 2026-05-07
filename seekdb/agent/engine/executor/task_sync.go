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

package executor

import (
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
)

func finishRemoteTaskByService(remoteTaskId int64, subTask task.ExecutableTask) (err error) {
	// Finish task in remote, try to get remote task from ob
	remoteSubTask, err := clusterTaskService.GetSubTaskByTaskID(remoteTaskId)
	if err != nil {
		return err
	}
	// Only finish remote task when execute times is equal
	if remoteSubTask.GetExecuteTimes() == subTask.GetExecuteTimes() {
		if remoteSubTask.IsFinished() {
			log.Debugf("remote task %d is finished, execute times %d", remoteTaskId, remoteSubTask.GetExecuteTimes())
		} else {
			// Try to finish remote task in ob
			remoteSubTask.SetContext(subTask.GetContext())
			if err = clusterTaskService.FinishSubTask(remoteSubTask, subTask.GetState()); err != nil {
				return err
			}
		}
		return nil
	} else {
		log.Warnf("remote task %d execute times %d != local task %d execute times %d", remoteTaskId, remoteSubTask.GetExecuteTimes(), subTask.GetID(), subTask.GetExecuteTimes())
		return nil
	}

}
