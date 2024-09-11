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
	"strconv"
	"strings"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/param"
)

func TenantRestore(p *param.RestoreParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if err := p.Check(); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err)
	}

	template := buildRestoreTemplate(p)
	ctx := buildRestoreTaskContext(p)
	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func newRestoreDagName(tenantName string) string {
	return fmt.Sprintf("%s_%s", DAG_RESTORE_BACKUP, tenantName)
}

func buildRestoreTemplate(p *param.RestoreParam) *task.Template {
	name := newRestoreDagName(p.TenantName)
	return task.NewTemplateBuilder(name).
		SetMaintenance(task.TenantMaintenance(p.TenantName)).
		AddTask(newPreRestoreCheckTask(), false).
		AddTask(newCreateResourceTask(), false).
		AddTask(newRestoreTask(), false).
		AddTask(newWaitRestoreFinshTask(), false).
		AddTask(newActiveTenantTask(), false).
		AddTask(newUpgradeTenantTask(), false).
		Build()
}

func buildRestoreTaskContext(p *param.RestoreParam) *task.TaskContext {
	taskTime := strconv.Itoa(int(time.Now().UnixMilli()))
	ctx := task.NewTaskContext().
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true).
		SetParam(PARAM_RESTORE, *p).
		SetParam(PARAM_TENANT_NAME, p.TenantName).
		SetParam(PARAM_POOL_NAME, fmt.Sprintf("%s_%s_pool", p.TenantName, taskTime)).
		SetParam(PARAM_TASK_TIME, taskTime).
		SetParam(PARAM_HA_HIGH_THREAD_SCORE, p.HaHighThreadScore)
	if p.KmsEncryptInfo != nil {
		ctx.SetParam(PARAM_KMS_ENCRYPT_INFO, *p.KmsEncryptInfo)
	}

	var scn string
	if p.SCN != nil && *p.SCN != 0 {
		scn = fmt.Sprint(*p.SCN)
		ctx.SetParam(PARAM_RESTORE_SCN, scn)
	}
	return ctx
}

type PreRestoreCheckTask struct {
	task.Task
	param *param.RestoreParam
	scn   string
}

func newPreRestoreCheckTask() *PreRestoreCheckTask {
	t := &PreRestoreCheckTask{
		Task: *task.NewSubTask(TASK_PRE_RESTORE_CHECK),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *PreRestoreCheckTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_RESTORE, &t.param); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_RESTORE)
	}
	if t.GetContext().GetParam(PARAM_RESTORE_SCN) != nil {
		if err = t.GetContext().GetParamWithValue(PARAM_RESTORE_SCN, &t.scn); err != nil {
			return errors.Wrapf(err, "get %s", PARAM_RESTORE_SCN)
		}
	}

	if !system.IsFileExist(path.OBAdmin()) || (t.scn == "" && t.param.Timestamp == nil) {
		t.ExecuteLog("Not need to check ob_admin")
		return
	}

	if t.param.Timestamp != nil {
		t.ExecuteLogf("Check restore time '%s'", t.param.Timestamp.Format("2006-01-02 15:04:05.00"))
		t.scn = fmt.Sprint((*(t.param.Timestamp)).UnixNano())
	} else {
		t.ExecuteLogf("Check restore time '%s'", t.scn)
	}

	if err = system.CheckRestoreTime(t.param.DataBackupUri, *t.param.ArchiveLogUri, t.scn); err != nil {
		return errors.Wrapf(err, "check restore time")
	}
	return nil
}

func (t *PreRestoreCheckTask) GetAdditionalData() map[string]any {
	return map[string]any{
		ADDL_KEY_RESTORE_JOB_ID: 0,
	}
}

type CreateResourceTask struct {
	task.Task
	taskTime string
	poolName string
	p        *param.RestoreParam
}

func newCreateResourceTask() *CreateResourceTask {
	t := &CreateResourceTask{
		Task: *task.NewSubTask(TASK_CREATE_RESOURCE),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *CreateResourceTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	pool, err := tenantService.GetResourcePoolsByName(t.poolName)
	if err != nil {
		return errors.Wrapf(err, "get resource pool")
	}
	if pool != nil {
		t.ExecuteLogf("Resource pool '%s' already exists", t.poolName)
		return nil
	}

	t.ExecuteLogf("Create resource pool '%s'", t.poolName)
	if err = tenantService.CreateResourcePool(t.poolName, t.p.UnitConfigName, *t.p.UnitNum, t.p.ZoneList); err != nil {
		return errors.Wrapf(err, "create tenant")
	}

	return nil
}

func (t *CreateResourceTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TASK_TIME, &t.taskTime); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TASK_TIME)
	}
	if err = t.GetContext().GetParamWithValue(PARAM_RESTORE, &t.p); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_RESTORE)
	}
	t.poolName = fmt.Sprintf("%s_%s_pool", t.p.TenantName, t.taskTime)
	return nil
}

func (t *CreateResourceTask) Rollback() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	pool, err := tenantService.GetResourcePoolsByName(t.poolName)
	if err != nil {
		return errors.Wrapf(err, "get resource pool")
	}
	if pool == nil {
		t.ExecuteLogf("Resource pool '%s' not exists", t.poolName)
		return nil
	}

	t.ExecuteLogf("Delete resource pool '%s'", t.poolName)
	if err = tenantService.DeleteResourcePool(t.poolName); err != nil {
		return errors.Wrapf(err, "delete resource pool")
	}
	return nil
}

type RestoreTask struct {
	task.Task
	tenantName        string
	poolName          string
	restoreScn        string
	param             *param.RestoreParam
	jobID             int64
	haHighThreadScore int
}

func newRestoreTask() *RestoreTask {
	t := &RestoreTask{
		Task: *task.NewSubTask(TASK_RESTORE),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *RestoreTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TENANT_NAME)
	}

	if err = t.GetContext().GetParamWithValue(PARAM_RESTORE, &t.param); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_RESTORE)
	}
	if t.GetContext().GetParam(PARAM_RESTORE_SCN) != nil {
		if err = t.GetContext().GetParamWithValue(PARAM_RESTORE_SCN, &t.restoreScn); err != nil {
			return errors.Wrapf(err, "get %s", PARAM_RESTORE_SCN)
		}
	}
	if err = t.GetContext().GetParamWithValue(PARAM_POOL_NAME, &t.poolName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_POOL_NAME)
	}
	if err = t.GetContext().GetParamWithValue(PARAM_HA_HIGH_THREAD_SCORE, &t.haHighThreadScore); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_HA_HIGH_THREAD_SCORE)
	}

	return nil
}

func (t *RestoreTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	restoreJob, err := tenantService.GetRunningRestoreTask(t.tenantName)
	if err != nil {
		return errors.Wrapf(err, "get running restore task")
	}

	if restoreJob == nil {
		// If there is no running restore task and the tenant does not exist, restore the tenant. Otherwise, skip the restore
		tenant, err := tenantService.GetTenantByName(t.tenantName)
		if err != nil {
			return errors.Wrapf(err, "get tenant")
		}
		if tenant != nil {
			t.ExecuteLogf("Tenant '%s' already exists", t.tenantName)
			return nil
		} else {
			t.ExecuteLogf("Restore tenant '%s'", t.tenantName)
			if err = tenantService.Restore(t.param, t.poolName, t.restoreScn); err != nil {
				return errors.Wrapf(err, "restore tenant")
			}
		}

		restoreJob, err = tenantService.GetRunningRestoreTask(t.tenantName)
		if err != nil {
			return errors.Wrapf(err, "get running restore task")
		}
	}
	t.jobID = restoreJob.JobID
	t.GetContext().SetData(ADDL_KEY_RESTORE_JOB_ID, t.jobID)

	t.ExecuteLogf("Wait for create tenant '%s'", t.tenantName)
	var metaTenantNormal bool
	for i := 0; i < waitForCreateTenant; i++ {
		metaTenantNormal, err = tenantService.IsMetaTenantStatusNormal(t.tenantName)
		if err != nil {
			return errors.Wrapf(err, "get tenant")
		}
		if metaTenantNormal {
			break
		}
		time.Sleep(time.Second)
	}
	if !metaTenantNormal {
		return errors.New("create tenant timeout")
	}

	t.ExecuteLogf("Set ha_high_thread_score to %d", t.haHighThreadScore)
	if err = tenantService.SetHaHighThreadScore(t.tenantName, t.haHighThreadScore); err != nil {
		return err
	}
	return nil
}

func (t *RestoreTask) Rollback() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TENANT_NAME)
	}

	t.ExecuteLogf("Try cancel restore job")
	if err = tenantService.CancelRestore(t.tenantName); err != nil {
		t.ExecuteLog("Cancel restore job failed, try to get tenant")
		tenant, err := tenantService.GetTenantByName(t.tenantName)
		if err != nil {
			return errors.Wrapf(err, "get tenant")
		}

		if tenant != nil {
			t.ExecuteLog("Try to delete tenant")
			if err = tenantService.DeleteTenant(t.tenantName); err != nil {
				return errors.Wrapf(err, "delete tenant")
			}
		}
	}

	t.ExecuteLog("Cancel restore job success. Wait for restore task finish")
	for i := 0; i < waitForRestoreTaskFinish; i++ {
		job, err := tenantService.GetRunningRestoreTask(t.tenantName)
		if err != nil {
			return errors.Wrapf(err, "get running restore task")
		}
		if job == nil {
			t.ExecuteLog("Restore task has finished successfully")
			return nil
		}
		time.Sleep(time.Second)
	}

	return nil
}

type WaitRestoreFinshTask struct {
	task.Task
	tenantName string
	jobID      int64
}

func newWaitRestoreFinshTask() *WaitRestoreFinshTask {
	t := &WaitRestoreFinshTask{
		Task: *task.NewSubTask(TASK_WAIT_RESTORE_FINISH),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

const (
	waitForCreateTenant      = 600  // seconds
	waitForRestoreTaskFinish = 3600 // seconds
)

func (t *WaitRestoreFinshTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TENANT_NAME)
	}

	t.ExecuteLog("Wait for restore task finish")
	for i := 0; i < waitForRestoreTaskFinish; i++ {
		restoreTask, err := tenantService.GetRunningRestoreTask(t.tenantName)
		if err != nil {
			return errors.Wrapf(err, "get running restore task")
		}
		if restoreTask == nil {
			t.ExecuteLog("Restore task has finished successfully")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("wait restore finish timeout")
}

func (t *WaitRestoreFinshTask) GetAdditionalData() map[string]any {
	if err := t.GetContext().GetParamWithValue(ADDL_KEY_RESTORE_JOB_ID, &t.jobID); err != nil {
		return nil
	}
	return map[string]any{
		ADDL_KEY_RESTORE_JOB_ID: t.jobID,
	}
}

type ActiveTenantTask struct {
	task.Task
	tenantName string
}

func newActiveTenantTask() *ActiveTenantTask {
	t := &ActiveTenantTask{
		Task: *task.NewSubTask(TASK_ACTIVE_TENANT),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

const (
	waitForTenantRolePrimary = 600 // seconds
)

func (t *ActiveTenantTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TENANT_NAME)
	}

	t.ExecuteLogf("Active tenant '%s'", t.tenantName)
	if err = tenantService.ActiveTenant(t.tenantName); err != nil {
		return err
	}

	t.ExecuteLogf("Wait for tenant to be primary role")
	for i := 0; i < waitForTenantRolePrimary; i++ {
		role, err := tenantService.GetTenantRole(t.tenantName)
		if err != nil {
			return err
		}
		if strings.ToUpper(role) == constant.TENANT_ROLE_PRIMARY {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("wait for tenant role primary timeout")
}

type UpgradeTenantTask struct {
	task.Task
	tenantName string
}

func newUpgradeTenantTask() *UpgradeTenantTask {
	t := &UpgradeTenantTask{
		Task: *task.NewSubTask(TASK_UPGRADE_TENANT),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

const (
	waitForUpgradeTenant = 600 // seconds
)

func (t *UpgradeTenantTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrapf(err, "get %s", PARAM_TENANT_NAME)
	}

	t.ExecuteLogf("Upgrade tenant %s", t.tenantName)
	if err = tenantService.Upgrade(t.tenantName); err != nil {
		return err
	}

	t.ExecuteLogf("Wait for upgrade tenant %s", t.tenantName)
	for i := 0; i < waitForUpgradeTenant; i++ {
		count, err := tenantService.GetUpgradeJobHistoryCount(t.tenantName)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		time.Sleep(time.Second)
	}

	return errors.New("wait for upgrade tenant timeout")
}

func GetRestoreOverview(tenantName string) (res *param.RestoreOverview, e *errors.OcsAgentError) {
	runningTask, err := tenantService.GetRunningRestoreTask(tenantName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	if runningTask != nil {
		res = &param.RestoreOverview{
			RestoreInfo:       runningTask.RestoreInfo,
			RecoverScn:        runningTask.RecoverScn,
			RecoverScnDisplay: runningTask.RecoverScnDisplay,
			RecoverProgress:   runningTask.RecoverProgress,
			RestoreProgress:   runningTask.RestoreProgress,
		}

	} else {

		lastTask, err := tenantService.GetLastRestoreTask(tenantName)
		if err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err)
		}

		if lastTask != nil {
			res = &param.RestoreOverview{
				RestoreInfo:          lastTask.RestoreInfo,
				BackupClusterVersion: lastTask.BackupClusterVersion,
				LsCount:              lastTask.LsCount,
				FinishLsCount:        lastTask.FinishLsCount,
				Comment:              lastTask.Comment,
				FinishTimestamp:      lastTask.FinishTimestamp,
			}
		}
	}

	if res == nil {
		return nil, errors.Occur(errors.ErrTaskNotFound, "restore task not found")
	}

	dataStorageInterface, _ := system.GetStorageInterfaceByURI(res.RestoreInfo.BackupSetList)
	if dataStorageInterface != nil {
		res.BackupSetList = dataStorageInterface.GenerateURIWithoutSecret()
	} else {
		res.BackupSetList = ""
	}

	clogStorageInterface, _ := system.GetStorageInterfaceByURI(res.RestoreInfo.BackupPieceList)
	if clogStorageInterface != nil {
		res.BackupPieceList = clogStorageInterface.GenerateURIWithoutSecret()
	} else {
		res.BackupPieceList = ""
	}

	return res, nil
}
