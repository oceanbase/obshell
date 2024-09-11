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

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func PostObclusterBackupConfig(p *param.ClusterBackupConfigParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if p.BackupBaseUri == nil || *p.BackupBaseUri == "" {
		return nil, errors.Occur(errors.ErrIllegalArgument, errors.New("backup_base_uri cannot be empty"))
	}
	return obclusterBackupConfig(p)
}

func PatchObclusterBackupConfig(p *param.ClusterBackupConfigParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if p.BackupBaseUri == nil || *p.BackupBaseUri == "" {
		if p.Binding != nil && *p.Binding != "" {
			return nil, errors.Occur(errors.ErrIllegalArgument, errors.New("If binding is set, backup_base_uri must be set"))
		}
		if p.PieceSwitchInterval != nil && *p.PieceSwitchInterval != "" {
			return nil, errors.Occur(errors.ErrIllegalArgument, errors.New("If piece_switch_interval is set, backup_base_uri must be set"))
		}
	}
	return obclusterBackupConfig(p)
}

func obclusterBackupConfig(clusterParam *param.ClusterBackupConfigParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	allTenants, err := tenantService.GetAllUserTenants()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	if len(allTenants) == 0 {
		return nil, errors.Occur(errors.ErrKnown, "no user tenants")
	}

	p := param.NewBackupConfigParamForCluster(clusterParam)

	backupConf, err := p.Check()
	if err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err)
	}
	template := buildSetBackupConfigTemplate(nil)
	taskCtx, err := buildSetBackupConfigTaskContext(backupConf, nil)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	dag, err := taskService.CreateDagInstanceByTemplate(template, taskCtx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildSetBackupConfigTemplate(tenantName *string) *task.Template {
	name := DAG_SET_BACKUP_CONFIG
	if tenantName != nil {
		name = fmt.Sprintf("%s for %s", name, *tenantName)
	}

	t := task.NewTemplateBuilder(name).
		AddTask(newCheckBackupConfigTask(), true).
		AddTask(newSetBackupConfigTask(), false)
	if tenantName != nil {
		log.Infof("Set maintenance for %s", *tenantName)
		t.SetMaintenance(task.TenantMaintenance(*tenantName))
	}
	return t.Build()
}

func buildSetBackupConfigTaskContext(p *param.BackupConf, tenantName *string) (*task.TaskContext, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}

	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_BACKUP_CONFIG, p)

	if tenantName != nil {
		ctx.SetParam(PARAM_NEED_BACKUP_TENANT, *tenantName)
	} else {
		ctx.SetParam(PARAM_ALL_TENANTS, true)
	}
	return ctx, nil
}

type CheckBackupConfigTask struct {
	task.Task
	conf *param.BackupConf
	backupConfigTaskContext
}

func newCheckBackupConfigTask() *CheckBackupConfigTask {
	t := &CheckBackupConfigTask{
		Task: *task.NewSubTask(TASK_CHECK_BACKUP_CONFIG),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *CheckBackupConfigTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_BACKUP_CONFIG, &t.conf); err != nil {
		return errors.Wrap(err, "get backup config")
	}
	t.tenants, err = getTenantFromCtx(t.GetContext())
	if err != nil {
		return errors.Wrap(err, "get tenant from context")
	}
	if t.clusterID, err = getClusterID(); err != nil {
		return errors.Wrap(err, "get cluster id")
	}
	if t.GetContext().GetParam(PARAM_ALL_TENANTS) != nil {
		t.isAllTenants = true
	}

	t.ExecuteLog("Check archive path for all tenants")
	if err = t.checkAllPath(t.conf.ArchiveDest); err != nil {
		return errors.Wrap(err, "check archive path")
	}

	t.ExecuteLog("Check data path for all tenants")
	if err = t.checkAllPath(t.conf.DataDest); err != nil {
		return errors.Wrap(err, "check data path")
	}
	return nil
}

func (t *CheckBackupConfigTask) checkAllPath(conf *param.DestConf) error {
	if conf == nil || conf.BaseURI == "" {
		t.ExecuteInfoLog("Path is empty, skip check")
		return nil
	}

	storage, err := system.GetStorageInterfaceByURI(conf.BaseURI)
	if err != nil {
		return errors.Wrap(err, "get archive storage interface")
	}
	t.ExecuteLogf("Storage type is '%s'", storage.GetResourceType())

	for _, tenant := range t.tenants {
		subpath := t.GetDestSubpath(tenant.TenantID, conf.JoinedDir)
		subStorage := storage.NewWithObjectKey(subpath)
		t.ExecuteLogf("Check '%s' wirte permission", tenant.TenantName)
		if err = subStorage.CheckWritePermission(); err != nil {
			return errors.Wrap(err, "check write permission")
		}
	}
	return nil
}

type backupConfigTaskContext struct {
	isAllTenants bool
	clusterID    int
	tenants      []oceanbase.DbaObTenant
}

func (t *backupConfigTaskContext) GetDestSubpath(tenantID int, joinDir string) (subpath string) {
	if t.isAllTenants {
		subpath = fmt.Sprintf("%d/%d/%s", t.clusterID, tenantID, joinDir)
	}
	return
}

type SetBackupConfigTask struct {
	task.Task
	conf *param.BackupConf
	backupConfigTaskContext
}

func newSetBackupConfigTask() *SetBackupConfigTask {
	t := &SetBackupConfigTask{
		Task: *task.NewSubTask(TASK_SET_BACKUP_CONFIG),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SetBackupConfigTask) Execute() error {
	if err := t.getParams(); err != nil {
		return err
	}

	if err := t.setLogArchiveConcurrency(); err != nil {
		return err
	}

	if err := t.setLogArchiveDest(); err != nil {
		return errors.Wrap(err, "set log archive")
	}

	if err := t.setArchiveLagTarget(); err != nil {
		return err
	}

	if err := t.setDataBackupDest(); err != nil {
		return err
	}

	if err := t.setHaLowThreadScore(); err != nil {
		return err
	}

	if err := t.setDeletePolicy(); err != nil {
		return errors.Wrap(err, "set delete policy")
	}

	return nil
}

func (t *SetBackupConfigTask) getParams() (err error) {
	t.tenants, err = getTenantFromCtx(t.GetContext())
	if err != nil {
		return errors.Wrap(err, "get tenant from context")
	}
	if err = t.GetContext().GetParamWithValue(PARAM_BACKUP_CONFIG, &t.conf); err != nil {
		return errors.Wrap(err, "get backup config")
	}

	if t.clusterID, err = getClusterID(); err != nil {
		return errors.Wrap(err, "get cluster id")
	}
	if t.GetContext().GetParam(PARAM_ALL_TENANTS) != nil {
		t.isAllTenants = true
	}
	return nil
}

func (t *SetBackupConfigTask) setLogArchiveConcurrency() (err error) {
	if t.conf.LogArchiveConcurrency == nil {
		return
	}

	if err = t.closeAllArchiveLog(); err != nil {
		return errors.Wrap(err, "check archive log closed")
	}

	if err = t.waitForAllArchiveLogStop(); err != nil {
		return errors.Wrap(err, "wait archive log stop")
	}

	for _, tenant := range t.tenants {
		if err = tenantService.SetLogArchiveConcurrency(tenant.TenantName, *t.conf.LogArchiveConcurrency); err != nil {
			return errors.Wrap(err, "set log archive concurrency")
		}
		t.ExecuteLogf("Set log archive concurrency to '%d' for %s(%d)", *t.conf.LogArchiveConcurrency, tenant.TenantName, tenant.TenantID)
	}
	return nil
}

func (t *SetBackupConfigTask) waitForAllArchiveLogStop() (err error) {
	for _, tenant := range t.tenants {
		if err = t.waitForArchiveLogClosed(&tenant); err != nil {
			return errors.Wrap(err, "wait archive log closed")
		}
		if err = t.waitForArchiveLogStop(&tenant); err != nil {
			return errors.Wrap(err, "wait archive log stopped")
		}
		status, err := tenantService.GetArchiveLogStatus(tenant.TenantID)
		if err != nil {
			return errors.Wrap(err, "get archive log status")
		}
		log.Infof("current archive log status is %s", status)
		t.ExecuteLogf("Archive log has been closed for %s(%d)", tenant.TenantName, tenant.TenantID)
	}
	return nil
}

func (t *SetBackupConfigTask) closeAllArchiveLog() (err error) {
	for _, tenant := range t.tenants {
		archiveLogClosed, err := tenantService.IsArchiveLogClosed(tenant.TenantName)
		if err != nil {
			return errors.Wrap(err, "check archive log closed")
		}

		if !archiveLogClosed {
			t.ExecuteLogf("Close archive log for %s(%d)", tenant.TenantName, tenant.TenantID)
			if err = tenantService.CloseArchiveLog(tenant.TenantName); err != nil {
				return errors.Wrap(err, "close archive log")
			}
		}
		t.ExecuteLogf("Archive log has been closed for %s(%d)", tenant.TenantName, tenant.TenantID)
	}
	return nil
}

func (t *SetBackupConfigTask) waitForArchiveLogStop(tenant *oceanbase.DbaObTenant) error {
	t.ExecuteLogf("Wait for %s(%d) archive log to be '%s'", tenant.TenantName, tenant.TenantID, constant.ARCHIVELOG_STATUS_STOP)
	for i := 0; i < waitForArchiveLogStop; i++ {
		status, err := tenantService.GetArchiveLogStatus(tenant.TenantID)
		if err != nil {
			return err
		}
		if status == "" || status == constant.ARCHIVELOG_STATUS_STOP {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("wait archive log to be '%s' timeout", constant.ARCHIVELOG_STATUS_STOP)
}

func (t *SetBackupConfigTask) setLogArchiveDest() (err error) {
	if t.conf.ArchiveDest == nil || t.conf.ArchiveDest.BaseURI == "" {
		return nil
	}

	if err = t.closeAllArchiveLog(); err != nil {
		return errors.Wrap(err, "check archive log closed")
	}

	if err = t.waitForAllArchiveLogStop(); err != nil {
		return errors.Wrap(err, "wait archive log stop")
	}

	if t.conf.Binding != nil {
		t.ExecuteLogf("Set log archive dest binding to %s", *t.conf.Binding)
	}
	if t.conf.PieceSwitchInterval != nil {
		t.ExecuteLogf("Set log archive dest piece switch interval to %s", *t.conf.PieceSwitchInterval)
	}
	for _, tenant := range t.tenants {
		t.ExecuteLogf("Set log archive dest for %s(%d)", tenant.TenantName, tenant.TenantID)
		path, _ := t.getURI(t.conf.ArchiveDest, tenant.TenantID)
		dest := param.LogArchiveDestConf{
			Location:            &path,
			Binding:             t.conf.Binding,
			PieceSwitchInterval: t.conf.PieceSwitchInterval,
		}
		if err = tenantService.SetLogArchiveDest(tenant.TenantName, dest); err != nil {
			return err
		}
	}
	return nil
}

func (t *SetBackupConfigTask) getURI(conf *param.DestConf, tenantID int) (string, error) {
	storage, err := system.GetStorageInterfaceByURI(conf.BaseURI)
	if err != nil {
		return "", err
	}
	subpath := t.GetDestSubpath(tenantID, conf.JoinedDir)
	subStorage := storage.NewWithObjectKey(subpath)
	t.ExecuteLogf("URI is '%s'", subStorage.GenerateURIWithoutSecret())
	return subStorage.GenerateURI(), nil
}

func (t *SetBackupConfigTask) setArchiveLagTarget() (err error) {
	if t.conf.ArchiveLagTarget == nil {
		return nil
	}

	t.ExecuteLogf("Set archive lag target to %s", *t.conf.ArchiveLagTarget)
	for _, tenant := range t.tenants {
		if err = tenantService.SetArchiveLagTarget(tenant.TenantName, *t.conf.ArchiveLagTarget); err != nil {
			return errors.Wrapf(err, "set archive lag target for %s(%d)", tenant.TenantName, tenant.TenantID)
		}
	}
	return nil
}

const (
	waitForArchiveLogClosed = 600 // seconds
	waitForArchiveLogOpened = 600 // seconds
)

func (t *SetBackupConfigTask) waitForArchiveLogClosed(tenant *oceanbase.DbaObTenant) error {
	t.ExecuteLogf("Wait for %s(%d) to close archive log", tenant.TenantName, tenant.TenantID)
	for i := 0; i < waitForArchiveLogClosed; i++ {
		archiveLogClosed, err := tenantService.IsArchiveLogClosed(tenant.TenantName)
		if err != nil {
			return err
		}
		if archiveLogClosed {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("wait archive log closed timeout")
}

func (t *SetBackupConfigTask) setDataBackupDest() error {
	if t.conf.DataDest == nil || t.conf.DataDest.BaseURI == "" {
		return nil
	}

	for _, tenant := range t.tenants {
		t.ExecuteLogf("Set data backup dest for %s(%d)", tenant.TenantName, tenant.TenantID)
		if err := stopBackupAndWait(t, &tenant); err != nil {
			return errors.Wrap(err, "check and wait backup stopped")
		}

		path, _ := t.getURI(t.conf.DataDest, tenant.TenantID)
		if err := tenantService.SetDataBackupDest(tenant.TenantName, path); err != nil {
			return errors.Wrap(err, "set data path")
		}
	}
	return nil
}

func (t *SetBackupConfigTask) setDeletePolicy() (err error) {
	if t.conf.DeletePolicy == nil {
		return nil
	}

	for _, tenant := range t.tenants {
		currPolicy, err := tenantService.GetDeletePolicy(tenant.TenantID)
		if err != nil {
			return errors.Wrapf(err, "get delete policy count for %s(%d)", tenant.TenantName, tenant.TenantID)
		}

		log.Infof("current delete policy is %v", currPolicy)
		if currPolicy != nil {
			t.ExecuteLogf("Drop delete policy %v for %s(%d)", *currPolicy, tenant.TenantName, tenant.TenantID)
			if err = tenantService.DropDeletePolicy(tenant.TenantName, currPolicy.PolicyName); err != nil {
				return err
			}
		}

		t.ExecuteLogf("Set delete policy to %v for %s(%d)", *t.conf.DeletePolicy, tenant.TenantName, tenant.TenantID)
		if err = tenantService.SetDeletePolicy(tenant.TenantName, *t.conf.DeletePolicy); err != nil {
			return err
		}
	}
	return nil

}

func (t *SetBackupConfigTask) setHaLowThreadScore() (err error) {
	if t.conf.HaLowThreadScore == nil {
		return nil
	}

	for _, tenant := range t.tenants {
		t.ExecuteLogf("Set ha low thread score for %s(%d) to %d", tenant.TenantName, tenant.TenantID, *t.conf.HaLowThreadScore)
		if err = tenantService.SetHaLowThreadScore(tenant.TenantName, *t.conf.HaLowThreadScore); err != nil {
			return errors.Wrapf(err, "set ha low thread score for %s(%d)", tenant.TenantName, tenant.TenantID)
		}
	}
	return nil
}

func stopBackupAndWait(t task.ExecutableTask, tenant *oceanbase.DbaObTenant) error {
	backupStopped, err := tenantService.IsBackupFinished(tenant.TenantID)
	if err != nil {
		return errors.Wrap(err, "check backup stopped")
	}

	if !backupStopped {
		if err := tenantService.StopBackup(tenant.TenantName); err != nil {
			return errors.Wrap(err, "stop backup")
		}
		if err := waitBackupFinish(t, tenant); err != nil {
			return errors.Wrap(err, "wait backup stopped")
		}
	}
	return nil
}

const (
	waitForBackupStopped = 3600 // seconds
)

func waitBackupFinish(t task.ExecutableTask, tenant *oceanbase.DbaObTenant) error {
	t.ExecuteLogf("Wait %s(%d) backup finish", tenant.TenantName, tenant.TenantID)
	for i := 0; i < waitForBackupStopped; i++ {
		backupFinished, err := tenantService.IsBackupFinished(tenant.TenantID)
		if err != nil {
			return err
		}
		if backupFinished {
			t.ExecuteLogf("%s(%d) backup finished", tenant.TenantName, tenant.TenantID)
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("wait backup finish timeout")
}

func getTenantFromCtx(ctx *task.TaskContext) (tenants []oceanbase.DbaObTenant, err error) {
	if ctx.GetParam(PARAM_ALL_TENANTS) != nil {
		tenants, err = tenantService.GetAllUserTenants()
		if err != nil {
			return nil, errors.Wrap(err, "get all tenants")
		}
	} else {
		var tenantName string
		if err := ctx.GetParamWithValue(PARAM_NEED_BACKUP_TENANT, &tenantName); err != nil {
			return nil, errors.Wrap(err, "get need backup tenant")
		}
		tenant, err := tenantService.GetTenantByName(tenantName)
		if err != nil {
			return nil, errors.Wrap(err, "get tenant by name")
		}
		tenants = append(tenants, *tenant)
	}
	return tenants, nil
}
