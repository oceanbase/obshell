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
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func (s *TenantService) GetArchiveLogByID(tenantID int) (res *oceanbase.CdbOBArchivelogStatus, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_ARCHIVELOG).Where("tenant_id = ?", tenantID).Scan(&res).Error
	return
}

func (s *TenantService) GetArchiveDestByID(tenantID int) (value string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_ARCHIVE_DEST).Where("tenant_id = ? and NAME = 'path'", tenantID).Select("value").Scan(&value).Error
	return
}

func (s *TenantService) ListArchiveLogTasks(tenantID int) (summarys []oceanbase.CdbOBArchivelogSummary, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}

	err = oceanbaseDb.Table(CDB_OB_ARCHIVELOG_SUMMARY).
		Select("*, truncate((time_to_usec(now()) - checkpoint_scn / 1000) / 1000000, 4) AS DELAY").
		Where("tenant_id = ? and status in ('BEGINNING', 'DOING', 'STOPPING', 'PREPARE', 'SUSPEND')", tenantID).Scan(&summarys).Error
	return
}

func (s *TenantService) GetLatestArchiveLog(tenantID int) (summary *oceanbase.CdbOBArchivelogSummary, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_ARCHIVELOG_SUMMARY).
		Select("*, truncate((time_to_usec(now()) - checkpoint_scn / 1000) / 1000000, 4) AS DELAY").
		Where("tenant_id = ?", tenantID).
		Order("ROUND_ID desc").
		Limit(1).
		Scan(&summary).
		Error
	return
}
func (s *TenantService) IsArchiveLogClosed(tenantName string) (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var mode string
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where("TENANT_NAME = ?", tenantName).Select("LOG_MODE").Scan(&mode).Error
	return mode == constant.LOG_MODE_NOARCHIVELOG, err
}

func (s *TenantService) CloseArchiveLog(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM NOARCHIVELOG TENANT = %s;", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) OpenArchiveLog(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM ARCHIVELOG TENANT = %s;", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) GetArchiveLogStatus(tenantID int) (status string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_ARCHIVELOG).Where("tenant_id = ?", tenantID).Select("status").Scan(&status).Error
	return
}

func (s *TenantService) SetLogArchiveDest(tenantName string, dest param.LogArchiveDestConf) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	if dest.Location == nil {
		return errors.Occur(errors.ErrCommonUnexpected, "location is required")
	}

	subsql := fmt.Sprintf("LOCATION=%s", *dest.Location)
	if dest.Binding != nil {
		subsql = fmt.Sprintf("%s BINDING=%s", subsql, *dest.Binding)
	}
	if dest.PieceSwitchInterval != nil {
		subsql = fmt.Sprintf("%s PIECE_SWITCH_INTERVAL=%s", subsql, *dest.PieceSwitchInterval)
	}

	sql := fmt.Sprintf("ALTER SYSTEM SET LOG_ARCHIVE_DEST = '%s' TENANT = %s;", subsql, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) EnableArchiveLogDest(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET LOG_ARCHIVE_DEST_STATE='ENABLE' TENANT = %s;", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) DeferArchiveLogDest(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET LOG_ARCHIVE_DEST_STATE='DEFER' TENANT = %s;", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) SetLogArchiveConcurrency(tenantName string, concurrency int) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET LOG_ARCHIVE_CONCURRENCY = %d TENANT = %s;", concurrency, tenantName)
	log.Info(sql)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) SetArchiveLagTarget(tenantName string, target string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET ARCHIVE_LAG_TARGET = '%s' TENANT = %s;", target, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) SetEncryption(pwd string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	pwd = strings.ReplaceAll(pwd, "\"", "\\\"")
	sql := fmt.Sprintf("SET ENCRYPTION ON IDENTIFIED BY \"%s\" ONLY", pwd)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) SetDataBackupDest(tenantName, dest string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET DATA_BACKUP_DEST = '%s' TENANT = %s;", dest, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) GetDataBackupDestByID(tenantID int) (value string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_BACKUP_PARAMETER).Where("NAME = 'data_backup_dest' and TENANT_ID = ?", tenantID).Select("value").Scan(&value).Error
	return
}

func (s *TenantService) IsBackupFinished(tenantID int) (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = oceanbaseDb.Table(CDB_OB_BACKUP_JOBS).Where("TENANT_ID = ?", tenantID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *TenantService) StopBackup(tenantName string) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM CANCEL BACKUP TENANT = %s;", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) GetDeletePolicy(tenantID int) (*oceanbase.CdbObBackupDeletePolicy, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var policy *oceanbase.CdbObBackupDeletePolicy
	err = oceanbaseDb.Table(CDB_OB_BACKUP_DELETE_POLICY).Where("TENANT_ID = ?", tenantID).Scan(&policy).Error
	return policy, err
}

func (s *TenantService) SetDeletePolicy(tenantName string, policy param.BackupDeletePolicy) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM ADD DELETE BACKUP POLICY '%s' RECOVERY_WINDOW '%s' TENANT %s", policy.Policy, policy.RecoveryWindow, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) DropDeletePolicy(tenantName, policyName string) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM DROP DELETE BACKUP POLICY '%s' TENANT %s", policyName, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) SetHaLowThreadScore(tenantName string, score int) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET HA_LOW_THREAD_SCORE = %d TENANT = %s;", score, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) SetHaHighThreadScore(tenantName string, score int) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM SET HA_HIGH_THREAD_SCORE = %d TENANT = %s;", score, tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) StartFullBackup(tenantName, encryption string, plusArchive bool) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM BACKUP TENANT = %s", tenantName)
	if encryption != "" {
		encryption = strings.ReplaceAll(encryption, "\"", "\\\"")
		sql = fmt.Sprintf("SET ENCRYPTION ON IDENTIFIED BY \"%s\" ONLY; %s", encryption, sql)
	}
	if plusArchive {
		sql = fmt.Sprintf("%s PLUS ARCHIVELOG", sql)
	}
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) StartIncrementalBackup(tenantName, encryption string, plusArchive bool) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM BACKUP INCREMENTAL TENANT = %s", tenantName)
	if encryption != "" {
		encryption = strings.ReplaceAll(encryption, "\"", "\\\"")
		sql = fmt.Sprintf("SET ENCRYPTION ON IDENTIFIED BY \"%s\" ONLY; %s", encryption, sql)
	}
	if plusArchive {
		sql = fmt.Sprintf("%s PLUS ARCHIVELOG", sql)
	}
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) GetRunningBackupTask(tenantID int) (task *oceanbase.CdbObBackupTask, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_BACKUP_TASKS).Where("TENANT_ID = ?", tenantID).Scan(&task).Error
	return
}

func (s *TenantService) GetLastBackupTask(tenantID int) (task *oceanbase.CdbObBackupTask, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(CDB_OB_BACKUP_TASK_HISTORY).Where("TENANT_ID = ?", tenantID).Order("START_TIMESTAMP desc").Limit(1).Scan(&task).Error
	return
}

func (s *TenantService) buildBackupJobQuery(oceanbaseDb *gorm.DB, tenantID int, p *param.QueryBackupTasksParam) (historyQuery, runningQuery *gorm.DB) {
	historyQuery = oceanbaseDb.Table(CDB_OB_BACKUP_JOB_HISTORY).Where("TENANT_ID = ?", tenantID)
	runningQuery = oceanbaseDb.Table(CDB_OB_BACKUP_JOBS).Where("TENANT_ID = ?", tenantID)

	if p.StartTime != nil {
		historyQuery = historyQuery.Where("START_TIMESTAMP >= ?", p.StartTime)
		runningQuery = runningQuery.Where("START_TIMESTAMP >= ?", p.StartTime)
	}
	if p.EndTime != nil {
		historyQuery = historyQuery.Where("START_TIMESTAMP <= ?", p.EndTime)
		runningQuery = runningQuery.Where("START_TIMESTAMP <= ?", p.EndTime)
	}
	if len(p.ParsedStatus) > 0 {
		historyQuery = historyQuery.Where("STATUS IN ?", p.ParsedStatus)
		runningQuery = runningQuery.Where("STATUS IN ?", p.ParsedStatus)
	}

	return historyQuery, runningQuery
}

func (s *TenantService) GetTenantBackupJobs(tenantID int, p *param.QueryBackupTasksParam) (tasks []oceanbase.CdbObBackupJob, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}

	historyQuery, runningQuery := s.buildBackupJobQuery(oceanbaseDb, tenantID, p)

	var historyJobs, runningJobs []oceanbase.CdbObBackupJob

	err = historyQuery.Find(&historyJobs).Error
	if err != nil {
		return nil, err
	}

	err = runningQuery.Find(&runningJobs).Error
	if err != nil {
		return nil, err
	}

	tasks = append(historyJobs, runningJobs...)

	sort.Slice(tasks, func(i, j int) bool {
		if p.SortBy == "start_timestamp" {
			if tasks[i].StartTimestamp == nil {
				return false
			}
			if tasks[j].StartTimestamp == nil {
				return true
			}
			if p.SortOrder == "asc" {
				return tasks[i].StartTimestamp.Before(*tasks[j].StartTimestamp)
			} else {
				return tasks[i].StartTimestamp.After(*tasks[j].StartTimestamp)
			}
		} else if p.SortBy == "end_timestamp" {
			if tasks[i].EndTimestamp == nil {
				return false
			}
			if tasks[j].EndTimestamp == nil {
				return true
			}
			if p.SortOrder == "asc" {
				return tasks[i].EndTimestamp.Before(*tasks[j].EndTimestamp)
			} else {
				return tasks[i].EndTimestamp.After(*tasks[j].EndTimestamp)
			}
		}
		return false
	})

	start := int((p.Page - 1) * p.Size)
	end := start + int(p.Size)
	if start >= len(tasks) {
		tasks = []oceanbase.CdbObBackupJob{}
	} else if end > len(tasks) {
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	return tasks, nil
}

func (s *TenantService) GetTenantBackupJobCount(tenantID int, p *param.QueryBackupTasksParam) (count uint64, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}

	historyQuery, runningQuery := s.buildBackupJobQuery(oceanbaseDb, tenantID, p)

	var historyCount, runningCount int64

	err = historyQuery.Count(&historyCount).Error
	if err != nil {
		return 0, err
	}

	err = runningQuery.Count(&runningCount).Error
	if err != nil {
		return 0, err
	}

	return uint64(historyCount + runningCount), nil
}
