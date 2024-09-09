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
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func ObclusterStartBackup(p *param.BackupParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	allTenants, err := tenantService.GetAllUserTenants()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	if len(allTenants) == 0 {
		return nil, errors.Occur(errors.ErrKnown, "no user tenants")
	}
	for _, tenant := range allTenants {
		if err = checkAllDest(&tenant); err != nil {
			return nil, errors.Occur(errors.ErrKnown, err)
		}
	}

	ctx, err := buildStartClusterBackupCtx(p)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	template := buildStartClusterBackupTemplate(*p.Mode)

	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func checkAllDest(tenant *oceanbase.DbaOBTenants) error {
	dest, err := tenantService.GetArchiveDestByID(tenant.TenantID)
	if err != nil {
		return errors.Wrapf(err, "get archive dest of %s(%d)", tenant.TenantName, tenant.TenantID)
	}
	if dest == "" {
		return fmt.Errorf("archive dest is empty, tenant: %s(%d)", tenant.TenantName, tenant.TenantID)
	}

	dest, err = tenantService.GetDataBackupDestByID(tenant.TenantID)
	if err != nil {
		return errors.Wrapf(err, "get data dest of %s(%d)", tenant.TenantName, tenant.TenantID)
	}
	if dest == "" {
		return fmt.Errorf("data dest is empty, tenant: %s(%d)", tenant.TenantName, tenant.TenantID)
	}
	return nil
}

func buildStartTenantBackupTemplate(mode string, tenantName string) *task.Template {
	template := buildStartClusterBackupTemplate(mode)
	name := fmt.Sprintf("%s for %s", template.Name, tenantName)
	builder := task.NewTemplateBuilder(name).AddTemplate(template)
	return builder.Build()
}

func buildStartClusterBackupTemplate(mode string) *task.Template {
	name := DAG_OBCLUSTER_START_FULL_BACKUP
	if mode == constant.BACKUP_MODE_INCREMENTAL {
		name = DAG_OBCLUSTER_START_INCREMENT_BACKUP
	}

	builder := task.NewTemplateBuilder(name).
		AddTask(newCheckDestTask(), true).
		AddTask(newOpenArchiveLogTask(), false).
		AddTask(newStartBackupTask(), false).
		AddTask(newWaitBackupTask(), false)
	return builder.Build()
}

func buildStartTenantBackupCtx(p *param.BackupParam, tenantName string) (*task.TaskContext, error) {
	ctx, err := newStartBackupCtx(p)
	if err != nil {
		return nil, err
	}

	ctx.SetParam(PARAM_NEED_BACKUP_TENANT, tenantName)
	return ctx, nil
}

func buildStartClusterBackupCtx(p *param.BackupParam) (*task.TaskContext, error) {
	ctx, err := newStartBackupCtx(p)
	if err != nil {
		return nil, err
	}
	ctx.SetParam(PARAM_ALL_TENANTS, true)
	return ctx, nil
}

func newStartBackupCtx(p *param.BackupParam) (*task.TaskContext, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	if err = p.Check(); err != nil {
		return nil, err
	}

	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_BACKUP_MODE, p.Mode)

	if p.Encryption != nil && *p.Encryption != "" {
		ctx.SetParam(PARAM_BACKUP_ENCRYPTION, p.Encryption)
	}

	if p.PlusArchive != nil && *p.PlusArchive {
		ctx.SetParam(PARAM_BACKUP_PLUS_ARCHIVE, true)
	}
	return ctx, nil
}

type CheckDestTask struct {
	task.Task
	tenants []oceanbase.DbaOBTenants
}

func newCheckDestTask() *CheckDestTask {
	t := &CheckDestTask{
		Task: *task.NewSubTask(TASK_CHECK_DEST),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *CheckDestTask) Execute() error {
	if err := t.getParams(); err != nil {
		return errors.Wrap(err, "get params")
	}

	for _, tenant := range t.tenants {
		archiveLogClosed, err := tenantService.IsArchiveLogClosed(tenant.TenantName)
		if err != nil {
			return errors.Wrap(err, "check archive log closed")
		}
		if archiveLogClosed {
			if err = t.checkArchiveDest(&tenant); err != nil {
				return errors.Wrap(err, "check archive dest")
			}
		}

		if err = t.checkDataBackupDest(&tenant); err != nil {
			return errors.Wrap(err, "check data backup dest")
		}

	}
	return nil
}

func (t *CheckDestTask) getParams() (err error) {
	t.tenants, err = getTenantFromCtx(t.GetContext())
	if err != nil {
		return errors.Wrap(err, "get tenant from context")
	}
	return nil
}

func (t *CheckDestTask) checkArchiveDest(tenant *oceanbase.DbaOBTenants) (err error) {
	t.ExecuteLogf("Check archive log dest of %s(%d)", tenant.TenantName, tenant.TenantID)
	dest, err := tenantService.GetArchiveDestByID(tenant.TenantID)
	if err != nil {
		return errors.Wrap(err, "get archive dest")
	}

	if dest == "" {
		return errors.New("archive dest is empty")
	}

	storage, err := system.GetStorageInterfaceByURI(dest)
	if err != nil {
		return errors.Wrap(err, "get storage interface")
	}
	t.ExecuteLogf("Archive dest is %s", storage.GenerateURIWithoutSecret())

	if err = storage.CheckWritePermission(); err != nil {
		return errors.Wrap(err, "check storage")
	}
	return nil
}

func (t *CheckDestTask) checkDataBackupDest(tenant *oceanbase.DbaOBTenants) (err error) {
	t.ExecuteLogf("Check data backup dest of %s(%d)", tenant.TenantName, tenant.TenantID)
	dest, err := tenantService.GetDataBackupDestByID(tenant.TenantID)
	if err != nil {
		return errors.Wrap(err, "get data dest")
	}

	if dest == "" {
		return errors.New("data dest is empty")
	}

	storage, err := system.GetStorageInterfaceByURI(dest)
	if err != nil {
		return errors.Wrap(err, "get storage interface")
	}

	if err = storage.CheckWritePermission(); err != nil {
		return errors.Wrap(err, "check storage")
	}
	t.ExecuteLogf("Data backup dest of %s(%d) is '%s'", tenant.TenantName, tenant.TenantID, storage.GenerateURIWithoutSecret())
	return nil
}

type OpenArchiveLogTask struct {
	task.Task
	tenants []oceanbase.DbaOBTenants
}

func newOpenArchiveLogTask() *OpenArchiveLogTask {
	t := &OpenArchiveLogTask{
		Task: *task.NewSubTask(TASK_OPEN_ARCHIVE_LOG),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *OpenArchiveLogTask) Execute() error {
	if err := t.getParams(); err != nil {
		return errors.Wrap(err, "get params")
	}

	for _, tenant := range t.tenants {
		archiveLogClosed, err := tenantService.IsArchiveLogClosed(tenant.TenantName)
		if err != nil {
			return errors.Wrap(err, "check archive log opened")
		}
		if archiveLogClosed {
			t.ExecuteLogf("Open archive log of %s(%d)", tenant.TenantName, tenant.TenantID)
			if err = tenantService.OpenArchiveLog(tenant.TenantName); err != nil {
				return err
			}
		}
	}

	for _, tenant := range t.tenants {
		archiveLogClosed, err := tenantService.IsArchiveLogClosed(tenant.TenantName)
		if err != nil {
			return errors.Wrap(err, "check archive log opened")
		}
		if archiveLogClosed {
			if err = t.waitArchiveLogOpened(&tenant); err != nil {
				return err
			}
		}

		status, err := tenantService.GetArchiveLogStatus(tenant.TenantID)
		if err != nil {
			return errors.Wrap(err, "get archive log status")
		}
		t.ExecuteLogf("Archive log status of %s(%d) is '%s'", tenant.TenantName, tenant.TenantID, status)

		switch status {
		case constant.ARCHIVELOG_STATUS_DOING:
			continue
		case constant.ARCHIVELOG_STATUS_NULL,
			constant.ARCHIVELOG_STATUS_PREPARE,
			constant.ARCHIVELOG_STATUS_BEGINNING:
			if err = t.waitArchiveLogDoing(&tenant); err != nil {
				return err
			}
		case constant.ARCHIVELOG_STATUS_STOP,
			constant.ARCHIVELOG_STATUS_STOPPING:
			if err = tenantService.OpenArchiveLog(tenant.TenantName); err != nil {
				return err
			}
			if err = t.waitArchiveLogOpened(&tenant); err != nil {
				return err
			}
			if err = t.waitArchiveLogDoing(&tenant); err != nil {
				return err
			}
		case constant.ARCHIVELOG_STATUS_SUSPEND,
			constant.ARCHIVELOG_STATUS_SUSPENDING:
			if err = tenantService.EnableArchiveLogDest(tenant.TenantName); err != nil {
				return err
			}
			if err = t.waitArchiveLogDoing(&tenant); err != nil {
				return err
			}
		case constant.ARCHIVELOG_STATUS_INTERRUPTED:
			return fmt.Errorf("tenant %s(%d) archive log interrupted,", tenant.TenantName, tenant.TenantID)
		}
	}
	return nil
}

func (t *OpenArchiveLogTask) getParams() (err error) {
	t.tenants, err = getTenantFromCtx(t.GetContext())
	if err != nil {
		return errors.Wrap(err, "get tenant from context")
	}
	return nil
}

func (t *OpenArchiveLogTask) waitArchiveLogOpened(tenant *oceanbase.DbaOBTenants) error {
	t.ExecuteLogf("Wait %s(%d) archive log opened", tenant.TenantName, tenant.TenantID)
	for i := 0; i < waitForArchiveLogOpened; i++ {
		archiveLogClosed, err := tenantService.IsArchiveLogClosed(tenant.TenantName)
		if err != nil {
			return err
		}
		if !archiveLogClosed {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("wait archive log opened timeout")
}

const (
	waitForArchiveLogDoing = 3600
	waitForArchiveLogStop  = 3600
)

func (t *OpenArchiveLogTask) waitArchiveLogDoing(tenant *oceanbase.DbaOBTenants) error {
	t.ExecuteLogf("Wait for %s(%d) archive log to be 'DOING'", tenant.TenantName, tenant.TenantID)
	var status string
	var err error
	for i := 0; i < waitForArchiveLogDoing; i++ {
		status, err = tenantService.GetArchiveLogStatus(tenant.TenantID)
		if err != nil {
			return err
		}
		if status == constant.ARCHIVELOG_STATUS_DOING {
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("wait archive log doing timeout, tenant: %s(%d), status: %s", tenant.TenantName, tenant.TenantID, status)
}

func (t *OpenArchiveLogTask) RollBack() (err error) {
	return nil
}

type StartBackupTask struct {
	task.Task
	mode        string
	encryption  string
	plusArchive bool
	tenants     []oceanbase.DbaOBTenants
}

func newStartBackupTask() *StartBackupTask {
	t := &StartBackupTask{
		Task: *task.NewSubTask(TASK_START_BACKUP),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *StartBackupTask) Execute() error {
	if err := t.getParams(); err != nil {
		return errors.Wrap(err, "get params")
	}

	if t.mode == constant.BACKUP_MODE_FULL {
		for _, tenant := range t.tenants {
			t.ExecuteLogf("Start full backup of %s(%d)", tenant.TenantName, tenant.TenantID)
			if err := tenantService.StartFullBackup(tenant.TenantName, t.encryption, t.plusArchive); err != nil {
				return errors.Wrap(err, "start full backup")
			}
		}
	} else {
		for _, tenant := range t.tenants {
			t.ExecuteLogf("Start incremental backup of %s(%d)", tenant.TenantName, tenant.TenantID)
			if err := tenantService.StartIncrementalBackup(tenant.TenantName, t.encryption, t.plusArchive); err != nil {
				return errors.Wrap(err, "start incremental backup")
			}
		}
	}
	return nil
}

func (t *StartBackupTask) getParams() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_BACKUP_MODE, &t.mode); err != nil {
		return errors.Wrap(err, "get backup mode")
	}
	if t.GetContext().GetParam(PARAM_BACKUP_ENCRYPTION) != nil {
		if err = t.GetContext().GetParamWithValue(PARAM_BACKUP_ENCRYPTION, &t.encryption); err != nil {
			return errors.Wrap(err, "get backup encryption")
		}
	}
	if t.GetContext().GetParam(PARAM_BACKUP_PLUS_ARCHIVE) != nil {
		t.plusArchive = true
	}
	if t.tenants, err = getTenantFromCtx(t.GetContext()); err != nil {
		return errors.Wrap(err, "get tenant from context")
	}
	return nil
}

func (t *StartBackupTask) RollBack() (err error) {
	if t.tenants, err = getTenantFromCtx(t.GetContext()); err != nil {
		return errors.Wrap(err, "get tenant from context")
	}
	for _, tenant := range t.tenants {
		if err = stopBackupAndWait(t, &tenant); err != nil {
			return err
		}
	}
	return nil
}

type WaitBackupTaskFinish struct {
	task.Task
	tenants []oceanbase.DbaOBTenants
}

func newWaitBackupTask() *WaitBackupTaskFinish {
	t := &WaitBackupTaskFinish{
		Task: *task.NewSubTask(TASK_WAIT_BACKUP),
	}
	t.SetCanRetry().SetCanRollback().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *WaitBackupTaskFinish) Execute() (err error) {
	t.tenants, err = getTenantFromCtx(t.GetContext())
	if err != nil {
		return errors.Wrap(err, "get tenant from context")
	}

	for _, tenant := range t.tenants {
		if err := waitBackupFinish(t, &tenant); err != nil {
			return err
		}
	}
	return nil
}
