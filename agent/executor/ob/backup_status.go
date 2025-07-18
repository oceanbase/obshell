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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func PatchObclusterBackup(p *param.BackupStatusParam) error {
	if err := p.Check(); err != nil {
		return err
	}

	tenants, err := tenantService.GetAllUserTenants()
	if err != nil {
		return err
	}

	for _, tenant := range tenants {
		if err = stopBackup(&tenant); err != nil {
			return err
		}
	}
	return nil
}

func stopBackup(tenant *oceanbase.DbaObTenant) error {
	log.Infof("stop backup for %s", tenant.TenantName)

	backupStopped, err := tenantService.IsBackupFinished(tenant.TenantID)
	if err != nil {
		return errors.Wrap(err, "check backup stopped")
	}
	if backupStopped {
		log.Infof("%s backup is stopped", tenant.TenantName)
		return nil
	}

	return tenantService.StopBackup(tenant.TenantName)
}

func PatchTenantBackup(tenantName string, p *param.BackupStatusParam) error {
	if err := p.Check(); err != nil {
		return err
	}

	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return err
	}

	if err = stopBackup(tenant); err != nil {
		return err
	}
	return nil
}

func PatchObclusterArchiveLog(p *param.ArchiveLogStatusParam) error {
	if err := p.Check(); err != nil {
		return err
	}

	tenants, err := tenantService.GetAllUserTenants()
	if err != nil {
		return err
	}

	for _, tenant := range tenants {
		if err = patchArchiveLogStatus(&tenant, p); err != nil {
			return err
		}
	}
	return nil
}

func patchArchiveLogStatus(tenant *oceanbase.DbaObTenant, p *param.ArchiveLogStatusParam) (err error) {
	switch *p.Status {
	case constant.ARCHIVELOG_STATUS_DOING:
		return startArchiveLog(tenant)
	case constant.ARCHIVELOG_STATUS_STOP:
		return stopArchiveLog(tenant)
	default:
		return errors.Occur(errors.ErrObBackupArchiveLogStatusInvalid, *p.Status, constant.ARCHIVELOG_STATUS_STOP, constant.ARCHIVELOG_STATUS_DOING)
	}
}

func startArchiveLog(tenant *oceanbase.DbaObTenant) error {
	log.Infof("start archive log for %s", tenant.TenantName)
	status, err := tenantService.GetArchiveLogStatus(tenant.TenantID)
	if err != nil {
		return errors.Wrap(err, "get archive log status")
	}
	log.Infof("tenant %s archive log status is %s", tenant.TenantName, status)

	switch status {
	case constant.ARCHIVELOG_STATUS_DOING,
		constant.ARCHIVELOG_STATUS_PREPARE,
		constant.ARCHIVELOG_STATUS_BEGINNING:
		return nil
	case constant.ARCHIVELOG_STATUS_STOP,
		constant.ARCHIVELOG_STATUS_STOPPING:
		if err = tenantService.EnableArchiveLogDest(tenant.TenantName); err != nil {
			return err
		}
		return tenantService.OpenArchiveLog(tenant.TenantName)
	case constant.ARCHIVELOG_STATUS_SUSPEND,
		constant.ARCHIVELOG_STATUS_SUSPENDING:
		return tenantService.EnableArchiveLogDest(tenant.TenantName)
	case constant.ARCHIVELOG_STATUS_INTERRUPTED:
		return errors.Occurf(errors.ErrCommonUnexpected, "tenant %s archive log is interrupted", tenant.TenantName)
	}
	return nil
}

func stopArchiveLog(tenant *oceanbase.DbaObTenant) error {
	log.Infof("stop archive log for %s", tenant.TenantName)
	status, err := tenantService.GetArchiveLogStatus(tenant.TenantID)
	if err != nil {
		return errors.Wrap(err, "get archive log status")
	}
	log.Infof("tenant %s archive log status is %s", tenant.TenantName, status)

	switch status {
	case constant.ARCHIVELOG_STATUS_STOP,
		constant.ARCHIVELOG_STATUS_STOPPING:
		return nil
	case constant.ARCHIVELOG_STATUS_DOING,
		constant.ARCHIVELOG_STATUS_PREPARE,
		constant.ARCHIVELOG_STATUS_BEGINNING,
		constant.ARCHIVELOG_STATUS_SUSPEND,
		constant.ARCHIVELOG_STATUS_SUSPENDING:
		return tenantService.CloseArchiveLog(tenant.TenantName)
	case constant.ARCHIVELOG_STATUS_INTERRUPTED:
		return errors.Occurf(errors.ErrCommonUnexpected, "tenant %s archive log is interrupted", tenant.TenantName)
	}
	return nil
}

func PatchTenantArchiveLog(tenantName string, p *param.ArchiveLogStatusParam) error {
	if err := p.Check(); err != nil {
		return err
	}

	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return err
	}

	if err = patchArchiveLogStatus(tenant, p); err != nil {
		return err
	}
	return nil
}

func GetObclusterBackupOverview() (*param.BackupOverview, error) {
	tenants, err := tenantService.GetAllUserTenants()
	if err != nil {
		return nil, err
	}

	overview := &param.BackupOverview{
		Statuses: make([]oceanbase.CdbObBackupTask, 0),
	}

	for _, tenant := range tenants {
		task, err := getTenantBackupTask(tenant.TenantID)
		if err != nil {
			return nil, err
		}

		if task != nil {
			overview.Statuses = append(overview.Statuses, *task)
		}
	}

	if len(overview.Statuses) == 0 {
		return nil, errors.Occur(errors.ErrCommonNotFound, "backup task")
	}

	return overview, nil
}

func GetTenantBackupOverview(name string) (*param.TenantBackupOverview, error) {
	tenant, err := tenantService.GetTenantByName(name)
	if err != nil {
		return nil, err
	}

	task, err := getTenantBackupTask(tenant.TenantID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, errors.Occur(errors.ErrTaskNotFoundWithReason, "no backup task")
	}

	overview := &param.TenantBackupOverview{
		Status: *task,
	}
	return overview, nil
}

func getTenantBackupTask(tenantID int) (*oceanbase.CdbObBackupTask, error) {
	task, err := tenantService.GetRunningBackupTask(tenantID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		task, err = tenantService.GetLastBackupTask(tenantID)
		if err != nil {
			return nil, err
		}
	}

	return task, nil
}
