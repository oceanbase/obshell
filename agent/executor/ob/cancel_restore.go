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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
)

func CancelRestoreTaskForTenant(tenantName string) (*task.DagDetailDTO, *errors.OcsAgentError) {
	// Get the running restore task registered in ob by tenant name.
	job, err := tenantService.GetRunningRestoreTask(tenantName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err, "get running restore task")
	}
	// Get the restore dag id by tenant name from obshell.
	id, err := tenantService.GetTenantLevelDagIDByTenantName(tenantName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err, "get restore dag id")
	}

	if job == nil {
		log.Infof("There is no running restore task registered in ob for '%s'", tenantName)
		// If there is no running restore task, return directly.
		if id == nil {
			log.Infof("There is no running restore dag for tenant '%s'", tenantName)
			tenant, err := tenantService.GetTenantsByName(tenantName)
			if err != nil {
				return nil, errors.Occur(errors.ErrUnexpected, err, "get tenant by name")
			}
			if tenant != nil {
				return nil, errors.Occurf(errors.ErrBadRequest, "tenant '%s' not in restore", tenantName)
			}
			// TODO: Return a new error type
			return nil, nil

		} else {
			return cmpDagAncCancelRestoreForJobNil(*id, tenantName)
		}

	} else {
		log.Infof("There is a running restore task registered in ob for '%s', job id is %d", tenantName, job.JobID)
		if id == nil {
			// The restore task is caused by sql, not by obshell.
			// Create a cancel restore task, including cancel the restore and drop the rp.
			log.Infof("There is no running restore dag for tenant '%s'", tenantName)
			return createCancelRestoreDag(tenantName)

		} else {
			// Compare the dag id from obshell and the expected, if they are the same, cancel and rollback the dag.
			return cmpDagAndCancelRestore(*id, job.JobID, tenantName)
		}

	}

}

func cmpDagAncCancelRestoreForJobNil(dagID int64, tenantName string) (*task.DagDetailDTO, *errors.OcsAgentError) {
	log.Infof("The current dag of tenant '%s' is %d", tenantName, dagID)
	dag, err := taskService.GetDagInstance(dagID)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	expectName := newRestoreDagName(tenantName)
	log.Infof("dag name is %s, expect name is %s", dag.GetName(), expectName)
	if dag.GetName() != expectName {
		return nil, errors.Occur(errors.ErrKnown, "there is no restore dag, %s(%d) is not the restore dag", dag.GetName(), dagID)
	}

	// There is a restore task, so we need to compare the additional data.
	dagDetail := task.NewDagDetailDTO(dag)
	uri := fmt.Sprintf("%s%s/%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, dagDetail.GenericID)
	if err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, dagDetail); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	log.Infof("dag additional data is %v", *dagDetail.AdditionalData)
	if dagDetail.AdditionalData == nil {
		return nil, errors.Occurf(errors.ErrBadRequest, "additional data is nil, %s(%d) is not the restore backup task", dag.GetName(), dagID)
	}

	data := *(dagDetail.AdditionalData)
	jobID, ok := data[ADDL_KEY_RESTORE_JOB_ID].(float64)
	// If the dag additional data has no restore job id, which means the dag is not the restore backup task, return directly.
	if !ok {
		return nil, errors.Occurf(errors.ErrBadRequest, "additional data has no restore job id, dag is not the restore backup task")
	}

	if jobID == 0 {
		return cancelAndRollbackRestoreDag(dag)
	} else {
		return nil, errors.Occur(errors.ErrBadRequest, "restore task was succeed, can not cancel")
	}

}

func cmpDagAndCancelRestore(dagID, expectedJobID int64, tenantName string) (*task.DagDetailDTO, *errors.OcsAgentError) {
	// Get the dag detail by dag id from obshell.
	dag, err := taskService.GetDagInstance(dagID)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	if dag.IsSuccess() {
		//  If the dag is successful, which means the previous dag has finished,
		// so we need to create a new cancel restore dag.
		return createCancelRestoreDag(tenantName)

	} else {
		dagDetail := task.NewDagDetailDTO(dag)
		uri := fmt.Sprintf("%s%s/%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, dagDetail.GenericID)
		if err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, dagDetail); err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err)
		}

		// If the dag does not have additional data, which means the dag is not the restore backup task, return directly.
		if dagDetail.AdditionalData == nil {
			return nil, errors.Occurf(errors.ErrBadRequest, "additional data is nil, %s(%d) is not the restore backup task", dag.GetName(), dagID)
		} else {
			log.Infof("dag additional data is %v", *dagDetail.AdditionalData)
			data := *(dagDetail.AdditionalData)

			jobID, ok := data[ADDL_KEY_RESTORE_JOB_ID].(float64)
			if !ok {
				return nil, errors.Occurf(errors.ErrBadRequest, "additional data has no restore job id, %s(%d) is not the restore backup task", dag.GetName(), dagID)
			}

			if jobID == 0 || jobID == float64(expectedJobID) {
				return cancelAndRollbackRestoreDag(dag)
			} else {
				return nil, errors.Occurf(errors.ErrBadRequest, "%s(%d) is not the restore backup task", dag.GetName(), dagID)
			}

		}
	}

}

func cancelAndRollbackRestoreDag(dag *task.Dag) (*task.DagDetailDTO, *errors.OcsAgentError) {
	var err error
	// If the dag is running, cancel the dag.
	if dag.IsRunning() {
		if err = taskService.CancelDag(dag); err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err)
		}
		// Wait for the dag to be failed.
		for i := 0; i < 60; i++ {
			dag, err = taskService.GetDagInstance(dag.GetID())
			if err != nil {
				return nil, errors.Occur(errors.ErrUnexpected, err)
			}
			if dag.IsFail() {
				break
			}
			time.Sleep(time.Second)
		}
	}

	// If the dag is failed, set the dag to rollback.
	if err = taskService.SetDagRollback(dag); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	dagDetail, err := taskService.GetDagDetail(dag.GetID())
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return dagDetail, nil
}

func createCancelRestoreDag(tenantName string) (*task.DagDetailDTO, *errors.OcsAgentError) {
	tenant, err := tenantService.GetTenantsByName(tenantName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err, "get tenant by name")
	}
	if tenant == nil {
		return nil, errors.Occurf(errors.ErrBadRequest, "tenant '%s' not found", tenantName)
	}
	pools, err := tenantService.GetResourcePoolsNameByTenantID(tenant.TenantID)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err, "get resource pools")
	}
	log.Infof("Get resource pools '%v' by tenant '%s'", pools, tenantName)

	// Create a cancel restore task, including cancel the restore and drop the rp.
	t := task.NewTemplateBuilder(fmt.Sprintf("%s %s", DAG_CANCEL_RESTORE, tenantName)).
		SetMaintenance(task.TenantMaintenance(tenantName)).
		AddTask(newCancelRestoreTask(), false).
		AddTask(newDropResourcePoolTask(), false).
		Build()

	ctx := task.NewTaskContext().
		SetParam(PARAM_TENANT_NAME, tenantName).
		SetParam(PARAM_POOLS_NAME, pools)

	dag, err := taskService.CreateDagInstanceByTemplate(t, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

type CancelRestoreTask struct {
	task.Task
	tenantName string
}

func newCancelRestoreTask() *CancelRestoreTask {
	t := &CancelRestoreTask{
		Task: *task.NewSubTask(TASK_CANCEL_RESTORE),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *CancelRestoreTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TENANT_NAME)
	}

	job, err := tenantService.GetRunningRestoreTask(t.tenantName)
	if err != nil {
		return errors.Wrapf(err, "get running restore task")
	}
	if job == nil {
		t.ExecuteLog("There is no running restore task")
		return nil
	}

	t.ExecuteLogf("Cancel restore job of tenant '%s'", t.tenantName)
	if err = tenantService.CancelRestore(t.tenantName); err != nil {
		if strings.Contains(err.Error(), "not in restore") {
			t.ExecuteLog("tenant is not in restore")
			return nil
		}
		return errors.Wrapf(err, "cancel restore %s", t.tenantName)
	}

	t.GetContext().SetParam(PARAM_NEED_DELETE_RP, true)
	t.ExecuteLog("Need to drop resource pool")
	return nil
}

type DropResourcePoolTask struct {
	task.Task
	poolsName []string
}

func newDropResourcePoolTask() *DropResourcePoolTask {
	t := &DropResourcePoolTask{
		Task: *task.NewSubTask(TASK_DROP_RESOURCE_POOL),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *DropResourcePoolTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_POOLS_NAME, &t.poolsName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_POOLS_NAME)
	}

	if t.GetContext().GetParam(PARAM_NEED_DELETE_RP) != nil {
		for _, pool := range t.poolsName {
			t.ExecuteLogf("Delete resource pool '%s'", pool)
			if err = tenantService.DeleteResourcePool(pool); err != nil {
				return errors.Wrapf(err, "delete resource pool %s", pool)
			}
		}
	}

	return nil
}
