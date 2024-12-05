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
package tenant

import (
	"strings"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
)

const (
	SCALE_IN_REPLICA = iota
	SCALE_OUT_REPLICA
	MODIFY_REPLICA_TYPE
)

type AlterLocalityTask struct {
	task.Task
	option       int // SCALE_IN_REPLICA, SCALE_OUT_REPLICA or MODIFY_REPLICA_TYPE
	tenantId     int
	zone         string
	localityType string
}

func newAlterLocalityNode(tenantId int, op int, zoneName string, locality ...string) *task.Node {
	ctx := task.NewTaskContext().
		SetParam(PARAM_TENANT_ID, tenantId).
		SetParam(PARAM_ZONE_NAME, zoneName).
		SetParam(PARAM_ALTER_LOCALITY_TYPE, op)
	node := task.NewNodeWithContext(newAlterLocalityTask(), false, ctx)
	if len(locality) > 0 {
		node.GetContext().SetParam(PARAM_LOCALITY_TYPE, locality[0])
	}
	return node
}

func newAlterLocalityTask() *AlterLocalityTask {
	newTask := &AlterLocalityTask{
		Task: *task.NewSubTask(TASK_NAME_ALTER_TENANT_LOCALITY),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanRollback().SetCanPass()
	return newTask
}

func (t *AlterLocalityTask) Execute() error {
	var err error
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.tenantId); err != nil {
		return errors.Wrap(err, "Get tenant id failed")
	}
	if err = t.GetContext().GetParamWithValue(PARAM_ZONE_NAME, &t.zone); err != nil {
		return errors.Wrap(err, "Get zone failed")
	}
	if err = t.GetContext().GetParamWithValue(PARAM_ALTER_LOCALITY_TYPE, &t.option); err != nil {
		return errors.Wrap(err, "Get alter locality type failed")
	}
	if t.option == SCALE_OUT_REPLICA || t.option == MODIFY_REPLICA_TYPE {
		if err = t.GetContext().GetParamWithValue(PARAM_LOCALITY_TYPE, &t.localityType); err != nil {
			return errors.Wrap(err, "Get locality type failed")
		}
	}

	targetReplicaInfpMap := make(map[string]string)
	switch t.option {
	case SCALE_IN_REPLICA:
		if targetReplicaInfpMap, err = scaleInLocality(t.tenantId, t.zone); err != nil {
			return errors.Wrap(err, "Scale in locality failed")
		}
	case SCALE_OUT_REPLICA:
		targetReplicaInfpMap, err = scaleOutLocality(t.tenantId, t.zone, t.localityType)
		if err != nil {
			return errors.Wrap(err, "Scale out locality failed")
		}
	case MODIFY_REPLICA_TYPE:
		if targetReplicaInfpMap, err = modifyLocality(t.tenantId, t.zone, t.localityType); err != nil {
			return errors.Wrap(err, "Modify replica type failed")
		}
	}
	if targetReplicaInfpMap == nil {
		t.ExecuteLogf("No need to alter tenant locality")
		return nil
	}

	targetLocality := buildLocality(targetReplicaInfpMap)
	if jobBo, err := tenantService.GetInProgressTenantJobBo(constant.ALTER_TENANT_LOCALITY, t.tenantId); err != nil {
		return errors.Wrap(err, "Get in progress tenant job failed")
	} else if jobBo != nil {
		if jobBo.TargetIs(targetReplicaInfpMap) {
			if err := waitTenantJobSucceed(t.Task, jobBo.JobId); err != nil {
				return errors.Wrap(err, "Wait for alter tenant locality succeed failed")
			}
		} else {
			t.ExecuteErrorLogf("There is already a job for altering tenant locality to %s", targetLocality)
			return errors.Errorf("There is already a job for altering tenant locality to %s", targetLocality)
		}
	} else {
		t.ExecuteLogf("Alter tenant locality to %s", targetLocality)
		if err := tenantService.AlterTenantLocality(t.tenantId, targetLocality); err != nil {
			return errors.Wrap(err, "Alter tenant locality failed")
		}
		// Wait for task execute successfully
		if err := waitAlterTenantLocalitySucceed(t.Task, t.tenantId, targetLocality); err != nil {
			return errors.Wrap(err, "Wait for alter tenant locality succeed failed")
		}
	}
	return nil
}

func waitAlterTenantLocalitySucceed(t task.Task, tenantId int, locality string) error {
	jobId, err := tenantService.GetTargetTenantJob(constant.ALTER_TENANT_LOCALITY, tenantId, transfer("%\""+locality+"\""))
	if err != nil {
		return errors.Wrap(err, "Get tenant job failed")
	} else if jobId == 0 {
		return errors.Occur(errors.ErrUnexpected, "There is no job for altering tenant locality to %s", locality)
	}
	return waitTenantJobSucceed(t, jobId)
}

func waitTenantJobSucceed(t task.Task, jobId int) error {
	// wait for success
	retryTimes := constant.CHECK_JOB_RETRY_TIMES
	for retryTimes > 0 {
		t.TimeoutCheck()
		jobStatus, err := tenantService.GetTenantJobStatus(jobId)
		if err != nil {
			return errors.Wrap(err, "Get tenant job status failed.")
		}
		if jobStatus == "SUCCESS" {
			return nil
		} else if jobStatus != "INPROGRESS" {
			return errors.Errorf("Job %d failed, job status is %s", jobId, jobStatus)
		} else {
			retryTimes--
			time.Sleep(constant.CHECK_JOB_INTERVAL)
		}
	}
	return errors.Errorf("Wait job %d timeout", jobId)
}

func buildLocality(replicaInfoMap map[string]string) string {
	var locality []string
	for zone, replicaType := range replicaInfoMap {
		locality = append(locality, replicaType+"@"+zone)
	}
	return strings.Join(locality, ",")
}
