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
	"github.com/oceanbase/obshell/agent/executor/pool"
	"github.com/oceanbase/obshell/agent/executor/zone"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/param"
)

func TenantRestore(p *param.RestoreParam) (*task.DagDetailDTO, error) {
	if err := checkRestoreParam(p); err != nil {
		return nil, err
	}

	template := buildRestoreTemplate(p)
	ctx := buildRestoreTaskContext(p)
	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func checkRestoreParam(p *param.RestoreParam) error {
	if err := p.Check(); err != nil {
		return err
	}

	zone.RenderZoneParams(p.ZoneList)

	if err := zone.CheckZoneParams(p.ZoneList); err != nil {
		return err
	}

	if err := zone.CheckAtLeastOnePaxosReplica(p.ZoneList); err != nil {
		return err
	}

	zoneList := make([]string, 0)
	for _, zone := range p.ZoneList {
		zoneList = append(zoneList, zone.Name)
	}
	if err := zone.CheckPrimaryZone(*p.PrimaryZone, zoneList); err != nil {
		return err
	}

	locality := make(map[string]string, 0)
	for _, zone := range p.ZoneList {
		locality[zone.Name] = zone.ReplicaType
	}
	if err := zone.CheckPrimaryZoneAndLocality(*p.PrimaryZone, locality); err != nil {
		return err
	}
	return nil
}

func newRestoreDagName(tenantName string) string {
	return fmt.Sprintf("%s_%s", DAG_RESTORE_BACKUP, tenantName)
}

func buildRestoreTemplate(p *param.RestoreParam) *task.Template {
	name := newRestoreDagName(p.TenantName)
	return task.NewTemplateBuilder(name).
		SetMaintenance(task.TenantMaintenance(p.TenantName)).
		AddTask(newPreRestoreCheckTask(), false).
		AddTask(newStartRestoreTask(), false).
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
		SetParam(PARAM_TASK_TIME, taskTime).
		SetParam(PARAM_HA_HIGH_THREAD_SCORE, p.HaHighThreadScore)
	if p.KmsEncryptInfo != nil {
		ctx.SetParam(PARAM_KMS_ENCRYPT_INFO, *p.KmsEncryptInfo)
	}

	if p.SCN != nil && *p.SCN != 0 {
		ctx.SetParam(PARAM_RESTORE_SCN, *p.SCN)
	}
	return ctx
}

type PreRestoreCheckTask struct {
	task.Task
	param *param.RestoreParam
	scn   int64
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
		return err
	}
	if t.GetContext().GetParam(PARAM_RESTORE_SCN) != nil {
		if err = t.GetContext().GetParamWithValue(PARAM_RESTORE_SCN, &t.scn); err != nil {
			return err
		}
	}

	if !system.IsFileExist(path.OBAdmin()) || (t.scn == 0 && t.param.Timestamp == nil) {
		t.ExecuteLog("Not need to check ob_admin")
		return
	}

	if t.param.Timestamp != nil {
		t.ExecuteLogf("Check restore time '%s'", t.param.Timestamp.Format("2006-01-02 15:04:05.00"))
		t.scn = t.param.Timestamp.UnixNano()
	} else {
		t.ExecuteLogf("Check restore time '%d'", t.scn)
	}

	if err = system.CheckRestoreTime(t.param.DataBackupUri, *t.param.ArchiveLogUri, t.scn); err != nil {
		return errors.Wrap(err, "check restore time")
	}
	return nil
}

func (t *PreRestoreCheckTask) GetAdditionalData() map[string]any {
	return map[string]any{
		ADDL_KEY_RESTORE_JOB_ID: 0,
	}
}

type StartRestoreTask struct {
	task.Task
	tenantName              string
	restoreScn              int64
	param                   *param.RestoreParam
	jobID                   int64
	haHighThreadScore       int
	createResourcePoolParam []param.CreateResourcePoolTaskParam
	timeStamp               string
}

func newStartRestoreTask() *StartRestoreTask {
	t := &StartRestoreTask{
		Task: *task.NewSubTask(TASK_START_RESTORE),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *StartRestoreTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return err
	}

	if err = t.GetContext().GetParamWithValue(PARAM_RESTORE, &t.param); err != nil {
		return err
	}
	if t.GetContext().GetParam(PARAM_RESTORE_SCN) != nil {
		if err = t.GetContext().GetParamWithValue(PARAM_RESTORE_SCN, &t.restoreScn); err != nil {
			return err
		}
	}
	if err = t.GetContext().GetParamWithValue(PARAM_TASK_TIME, &t.timeStamp); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_HA_HIGH_THREAD_SCORE, &t.haHighThreadScore); err != nil {
		return err
	}

	return nil
}

func buildCreateResourcePoolTaskParam(tenantName string, zoneParam []param.ZoneParam, timestamp string) []param.CreateResourcePoolTaskParam {
	createResourcePoolParams := make([]param.CreateResourcePoolTaskParam, 0)
	for _, zone := range zoneParam {
		createResourcePoolParams = append(createResourcePoolParams, param.CreateResourcePoolTaskParam{
			PoolName:       strings.Join([]string{tenantName, zone.Name, timestamp}, "_"),
			ZoneName:       zone.Name,
			UnitConfigName: zone.PoolParam.UnitConfigName,
			UnitNum:        zone.PoolParam.UnitNum,
		})
	}
	return createResourcePoolParams
}

func (t *StartRestoreTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	restoreJob, err := tenantService.GetRunningRestoreTask(t.tenantName)
	if err != nil {
		return errors.Wrap(err, "get running restore task")
	}

	if restoreJob == nil {
		// If there is no running restore task and the tenant does not exist, restore the tenant. Otherwise, skip the restore
		tenant, err := tenantService.GetTenantByName(t.tenantName)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		if tenant != nil {
			t.ExecuteLogf("Tenant '%s' already exists", t.tenantName)
			return nil
		} else {
			if err = t.restoreTenant(); err != nil {
				return err
			}
		}

		restoreJob, err = tenantService.GetRunningRestoreTask(t.tenantName)
		if err != nil {
			return errors.Wrap(err, "get running restore task")
		}
	}
	t.jobID = restoreJob.JobID
	t.GetContext().SetData(ADDL_KEY_RESTORE_JOB_ID, t.jobID)

	t.ExecuteLogf("Wait for create tenant '%s'", t.tenantName)
	var metaTenantNormal bool
	for i := 0; i < waitForCreateTenant; i++ {
		metaTenantNormal, err = tenantService.IsMetaTenantStatusNormal(t.tenantName)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		if metaTenantNormal {
			break
		}
		time.Sleep(time.Second)
		t.TimeoutCheck()
	}
	if !metaTenantNormal {
		return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, fmt.Sprintf("create tenant '%s'", t.tenantName))
	}

	t.ExecuteLogf("Set ha_high_thread_score to %d", t.haHighThreadScore)
	if err = tenantService.SetHaHighThreadScore(t.tenantName, t.haHighThreadScore); err != nil {
		return err
	}
	return nil
}

func (t *StartRestoreTask) restoreTenant() (err error) {
	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(t.tenantName, t.param.ZoneList, t.timeStamp)
	if err := pool.CreatePools(t.Task, t.createResourcePoolParam); err != nil {
		return err
	}

	var poolList []string
	for _, poolParam := range t.createResourcePoolParam {
		poolList = append(poolList, poolParam.PoolName)
	}
	resourcePoolList := strings.Join(poolList, ",")

	var localityList []string
	for _, zone := range t.param.ZoneList {
		if zone.ReplicaType == "" {
			localityList = append(localityList, strings.Join([]string{constant.REPLICA_TYPE_FULL, zone.Name}, "@"))
		} else {
			localityList = append(localityList, strings.Join([]string{zone.ReplicaType, zone.Name}, "@"))
		}
	}
	locality := strings.Join(localityList, ",")

	t.ExecuteLogf("Restore tenant '%s'", t.tenantName)
	if err = tenantService.Restore(t.param, locality, resourcePoolList, t.restoreScn); err != nil {
		// drop all created resource pool
		if err := pool.DropFreeResourcePools(t.Task, t.createResourcePoolParam); err != nil {
			t.ExecuteWarnLog(errors.Wrap(err, "Drop created resource pool failed"))
		}
		return errors.Wrap(err, "restore tenant")
	}
	return nil
}

func (t *StartRestoreTask) Rollback() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	t.ExecuteLogf("Try cancel restore job")
	if err = tenantService.CancelRestore(t.tenantName); err != nil {
		t.ExecuteLog("Cancel restore job failed, try to get tenant")
		tenant, err := tenantService.GetTenantByName(t.tenantName)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}

		if tenant != nil {
			t.ExecuteLog("Try to delete tenant")
			if err = tenantService.DeleteTenant(t.tenantName); err != nil {
				return errors.Wrap(err, "delete tenant")
			}
		}
	}

	t.ExecuteLog("Cancel restore job success. Wait for restore task finish")
	for i := 0; i < waitForRestoreTaskFinish; i++ {
		job, err := tenantService.GetRunningRestoreTask(t.tenantName)
		if err != nil {
			return errors.Wrap(err, "get running restore task")
		}
		if job == nil {
			t.ExecuteLog("Restore task has finished successfully")
			break
		}
		time.Sleep(time.Second)
		t.TimeoutCheck()
	}

	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(t.tenantName, t.param.ZoneList, t.timeStamp)
	if err := pool.DropFreeResourcePools(t.Task, t.createResourcePoolParam); err != nil {
		return err
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
		return err
	}

	t.ExecuteLog("Wait for restore task finish")
	for i := 0; i < waitForRestoreTaskFinish; i++ {
		restoreTask, err := tenantService.GetRunningRestoreTask(t.tenantName)
		if err != nil {
			return errors.Wrap(err, "get running restore task")
		}
		if restoreTask == nil {
			t.ExecuteLog("Restore task has finished successfully")
			return nil
		}
		time.Sleep(time.Second)
		t.TimeoutCheck()
	}
	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, fmt.Sprintf("restore tenant '%s'", t.tenantName))
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
		return err
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
		t.TimeoutCheck()
	}
	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, fmt.Sprintf("active tenant '%s'", t.tenantName))
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
		return err
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
		t.TimeoutCheck()
	}

	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, fmt.Sprintf("upgrade tenant '%s'", t.tenantName))
}

func GetRestoreOverview(tenantName string) (res *param.RestoreOverview, e error) {
	runningTask, err := tenantService.GetRunningRestoreTask(tenantName)
	if err != nil {
		return nil, err
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
			return nil, err
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
		return nil, errors.Occur(errors.ErrCommonNotFound, "restore task")
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

func GetAllRestoreTasks(p *param.QueryRestoreTasksParam) (res *bo.PaginatedRestoreTaskResponse, e error) {
	res = &bo.PaginatedRestoreTaskResponse{
		Contents: make([]bo.RestoreTask, 0),
		Page:     bo.CustomPage{},
	}
	runningTasks, err := tenantService.ListRunningRestoreTasks(p)
	if err != nil {
		return nil, err
	}

	for _, task := range runningTasks {
		restoreOverview := bo.RestoreTask{
			RestoreInfo:       task.RestoreInfo,
			RecoverScn:        task.RecoverScn,
			RecoverScnDisplay: task.RecoverScnDisplay,
			RecoverProgress:   task.RecoverProgress,
			RestoreProgress:   task.RestoreProgress,
		}
		res.Contents = append(res.Contents, restoreOverview)
	}

	history, err := tenantService.GetRestoreHistory(p)
	if err != nil {
		return nil, err
	}
	for _, task := range history {
		restoreOverview := bo.RestoreTask{
			RestoreInfo:          task.RestoreInfo,
			BackupClusterVersion: task.BackupClusterVersion,
			LsCount:              task.LsCount,
			FinishLsCount:        task.FinishLsCount,
			Comment:              task.Comment,
			FinishTimestamp:      task.FinishTimestamp,
		}
		res.Contents = append(res.Contents, restoreOverview)
	}

	for i := range res.Contents {
		dataStorageInterface, _ := system.GetStorageInterfaceByURI(res.Contents[i].RestoreInfo.BackupSetList)
		if dataStorageInterface != nil {
			res.Contents[i].BackupSetList = dataStorageInterface.GenerateURIWithoutSecret()
		} else {
			res.Contents[i].BackupSetList = ""
		}
		clogStorageInterface, _ := system.GetStorageInterfaceByURI(res.Contents[i].RestoreInfo.BackupPieceList)
		if clogStorageInterface != nil {
			res.Contents[i].BackupPieceList = clogStorageInterface.GenerateURIWithoutSecret()
		} else {
			res.Contents[i].BackupPieceList = ""
		}
	}

	res.Page = bo.CustomPage{
		Number:        p.Page,
		Size:          p.Size,
		TotalPages:    calculateTotalPages(uint64(len(res.Contents)), p.Size),
		TotalElements: uint64(len(res.Contents)),
	}
	return res, nil
}

func GetRestoreWindows(param *param.RestoreWindowsParam) (res *system.RestoreWindows, e error) {
	if !system.IsFileExist(path.OBAdmin()) {
		return nil, errors.Occur(errors.ErrEnvironmentWithoutObAdmin)
	}

	if param.ArchiveLogUri == nil || *param.ArchiveLogUri == "" {
		*param.ArchiveLogUri = param.DataBackupUri
	}
	res, err := system.GetRestoreWindows(param.DataBackupUri, *param.ArchiveLogUri)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetRestoreSourceTenantInfo(param *param.RestoreStorageParam) (res *system.RestoreTenantInfo, e error) {
	if !system.IsFileExist(path.OBAdmin()) {
		return nil, errors.Occur(errors.ErrEnvironmentWithoutObAdmin)
	}

	if param.ArchiveLogUri == nil || *param.ArchiveLogUri == "" {
		*param.ArchiveLogUri = param.DataBackupUri
	}

	res, err := system.GetRestoreSourceTenantInfo(param.DataBackupUri, *param.ArchiveLogUri)
	if err != nil {
		return nil, err
	}
	return res, nil
}
