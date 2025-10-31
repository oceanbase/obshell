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

package param

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

type ClusterBackupConfigParam struct {
	BackupBaseUri *string `json:"backup_base_uri"`
	LogArchiveDestConf
	BaseBackupConfigParam
}

func NewBackupConfigParamForCluster(p *ClusterBackupConfigParam) *BackupConfigParam {
	return &BackupConfigParam{
		BackupBaseUri:         p.BackupBaseUri,
		LogArchiveDestConf:    p.LogArchiveDestConf,
		BaseBackupConfigParam: p.BaseBackupConfigParam,
	}
}

func NewBackupConfigParamForTenant(p *TenantBackupConfigParam) *BackupConfigParam {
	return &BackupConfigParam{
		DataBaseUri:           p.DataBaseUri,
		ArchiveBaseUri:        p.ArchiveBaseUri,
		LogArchiveDestConf:    p.LogArchiveDestConf,
		BaseBackupConfigParam: p.BaseBackupConfigParam,
	}
}

type TenantBackupConfigParam struct {
	DataBaseUri    *string `json:"data_base_uri"`
	ArchiveBaseUri *string `json:"archive_base_uri"`
	LogArchiveDestConf
	BaseBackupConfigParam
}

type BaseBackupConfigParam struct {
	LogArchiveConcurrency *int                `json:"log_archive_concurrency"`
	ArchiveLagTarget      *string             `json:"archive_lag_target"`
	HaLowThreadScore      *int                `json:"ha_low_thread_score"`
	DeletePolicy          *BackupDeletePolicy `json:"delete_policy"`
}

type BackupConfigParam struct {
	BackupBaseUri  *string
	ArchiveBaseUri *string
	DataBaseUri    *string

	LogArchiveDestConf
	BaseBackupConfigParam
}

type LogArchiveDestConf struct {
	Location            *string
	Binding             *string `json:"binding"`
	PieceSwitchInterval *string `json:"piece_switch_interval"`
}

type BackupConf struct {
	BackupConfigParam

	ArchiveDest *DestConf
	DataDest    *DestConf
}

type DestConf struct {
	BaseURI     string
	StorageType string
	JoinedDir   string
}

type BackupDeletePolicy struct {
	Policy         string `json:"policy"`
	RecoveryWindow string `json:"recovery_window"`
}

func (p *BackupConfigParam) Format() {
	if p.Binding != nil {
		if *p.Binding != "" {
			*p.Binding = strings.ToUpper(*p.Binding)
		} else {
			*p.Binding = constant.BINDING_MODE_OPTIONAL
		}
	}
	if p.ArchiveLagTarget != nil {
		*p.ArchiveLagTarget = strings.ToLower(*p.ArchiveLagTarget)
	}
	// The policy of DeletePolicy is the name of the policy, which is only supported by the 'default' policy.
	if p.DeletePolicy != nil && p.DeletePolicy.Policy != "" {
		p.DeletePolicy.Policy = strings.ToLower(p.DeletePolicy.Policy)
	}

}

func (p *BackupConfigParam) newBackupConf() *BackupConf {
	conf := &BackupConf{
		BackupConfigParam: *p,
	}

	if p.BackupBaseUri != nil && (p.ArchiveBaseUri == nil || *p.ArchiveBaseUri == "") {
		conf.ArchiveDest = &DestConf{
			BaseURI:   *p.BackupBaseUri,
			JoinedDir: constant.BACKUP_DIR_CLOG,
		}
	} else if p.ArchiveBaseUri != nil && *p.ArchiveBaseUri != "" {
		conf.ArchiveDest = &DestConf{
			BaseURI:   *p.ArchiveBaseUri,
			JoinedDir: constant.BACKUP_DIR_CLOG,
		}
	}

	if p.BackupBaseUri != nil && (p.DataBaseUri == nil || *p.DataBaseUri == "") {
		conf.DataDest = &DestConf{
			BaseURI:   *p.BackupBaseUri,
			JoinedDir: constant.BACKUP_DIR_DATA,
		}
	} else if p.DataBaseUri != nil && *p.DataBaseUri != "" {
		conf.DataDest = &DestConf{
			BaseURI:   *p.DataBaseUri,
			JoinedDir: constant.BACKUP_DIR_DATA,
		}
	}
	return conf
}

func (p *BackupConfigParam) Check() (backupConf *BackupConf, err error) {
	log.Info("check backup config param")
	p.Format()

	if p.LogArchiveConcurrency != nil && (*p.LogArchiveConcurrency < constant.LOG_ARCHIVE_CONCURRENCY_LOW || *p.LogArchiveConcurrency > constant.LOG_ARCHIVE_CONCURRENCY_HIGH) {
		return nil, errors.Occur(errors.ErrObBackupLogArchiveConcurrencyInvalid, constant.LOG_ARCHIVE_CONCURRENCY_LOW, constant.LOG_ARCHIVE_CONCURRENCY_HIGH)
	}

	if p.HaLowThreadScore != nil && (*p.HaLowThreadScore < constant.HA_LOW_THREAD_SCORE_LOW || *p.HaLowThreadScore > constant.HA_LOW_THREAD_SCORE_HIGH) {
		return nil, errors.Occur(errors.ErrObBackupHaLowThreadScoreInvalid, constant.HA_LOW_THREAD_SCORE_LOW, constant.HA_LOW_THREAD_SCORE_HIGH)
	}

	if p.Binding != nil &&
		(*p.Binding != "" && *p.Binding != constant.BINDING_MODE_OPTIONAL &&
			*p.Binding != constant.BINDING_MODE_MANDATORY) {
		return nil, errors.Occur(errors.ErrObBackupBindingInvalid, *p.Binding, constant.BINDING_MODE_OPTIONAL, constant.BINDING_MODE_MANDATORY)
	}

	if err = p.checkPieceSwitchInterval(); err != nil {
		return nil, err
	}

	if err = p.checkDeletePolicy(); err != nil {
		return nil, err
	}

	backupConf = p.newBackupConf()
	if backupConf.ArchiveDest != nil && backupConf.ArchiveDest.BaseURI != "" {
		backupConf.ArchiveDest.StorageType, err = system.GetResourceType(backupConf.ArchiveDest.BaseURI)
		if err != nil {
			return nil, err
		}
		log.Infof("archive storage type is %s", backupConf.ArchiveDest.StorageType)
	}
	if backupConf.DataDest != nil && backupConf.DataDest.BaseURI != "" {
		backupConf.DataDest.StorageType, err = system.GetResourceType(backupConf.DataDest.BaseURI)
		if err != nil {
			return nil, err
		}
		log.Infof("data storage type is %s", backupConf.DataDest.StorageType)
	}

	if backupConf.ArchiveDest != nil {
		if err = p.checkArchiveLagTarget(backupConf.ArchiveDest.StorageType); err != nil {
			return nil, err
		}
	}

	return backupConf, nil
}

func (p *BackupConfigParam) checkDeletePolicy() error {
	if p.DeletePolicy != nil {
		if p.DeletePolicy.Policy == "" {
			p.DeletePolicy.Policy = constant.DELETE_POLICY_DEFAULT
		} else if p.DeletePolicy.Policy != constant.DELETE_POLICY_DEFAULT {
			return errors.Occur(errors.ErrObBackupDeletePolicyInvalid, p.DeletePolicy.Policy, constant.DELETE_POLICY_DEFAULT)
		}
	}
	return nil
}

// checkPieceSwitchInterval checks the format and value of PieceSwitchInterval,
// which must be a valid time duration between 1d and 7d.
func (p *BackupConfigParam) checkPieceSwitchInterval() error {
	if p.PieceSwitchInterval == nil || *p.PieceSwitchInterval == "" {
		return nil
	}

	duration, err := system.ParseTime(*p.PieceSwitchInterval)
	if err != nil {
		return err
	}

	if duration < constant.PIECE_SWITCH_INTERVAL_LOW {
		return errors.Occur(errors.ErrObBackupPieceSwitchIntervalInvalid, constant.PIECE_SWITCH_INTERVAL_LOW, constant.PIECE_SWITCH_INTERVAL_HIGH)
	} else if duration > constant.PIECE_SWITCH_INTERVAL_HIGH {
		return errors.Occur(errors.ErrObBackupPieceSwitchIntervalInvalid, constant.PIECE_SWITCH_INTERVAL_LOW, constant.PIECE_SWITCH_INTERVAL_HIGH)
	}

	return nil
}

func (p *BackupConfigParam) checkArchiveLagTarget(t string) error {
	if p.ArchiveLagTarget == nil || *p.ArchiveLagTarget == "" {
		return nil
	}

	duration, err := system.ParseTime(*p.ArchiveLagTarget)
	if err != nil {
		return err
	}

	if duration > constant.ARCHIVE_LAG_TARGET_HIGH {
		return errors.Occur(errors.ErrObBackupArchiveLagTargetInvalid, constant.ARCHIVE_LAG_TARGET_HIGH)
	}

	if t == constant.PROTOCOL_S3 && duration < constant.ARCHIVE_LAG_TARGET_LOW_FOR_S3 {
		return errors.Occur(errors.ErrObBackupArchiveLagTargetForS3Invalid, constant.ARCHIVE_LAG_TARGET_LOW_FOR_S3)
	}
	return nil
}

type BackupParam struct {
	Mode        *string `json:"mode"`
	Encryption  *string `json:"encryption"`
	PlusArchive *bool   `json:"plus_archive"`
}

func (p *BackupParam) Format() {
	if p.Mode == nil || *p.Mode == "" {
		full := constant.BACKUP_MODE_FULL
		p.Mode = &full
	}
	*p.Mode = strings.ToLower(*p.Mode)
}

func (p *BackupParam) Check() error {
	p.Format()

	switch *p.Mode {
	case constant.BACKUP_MODE_FULL,
		constant.BACKUP_MODE_INCREMENTAL:
		return nil
	default:
		return errors.Occur(errors.ErrObBackupModeInvalid, *p.Mode, constant.BACKUP_MODE_FULL, constant.BACKUP_MODE_INCREMENTAL)
	}
}

type BackupStatusParam struct {
	Status *string `json:"status"`
}

func (p *BackupStatusParam) Format() {
	if p.Status == nil || *p.Status == "" {
		status := constant.BACKUP_CANCELED
		p.Status = &status
	}
	*p.Status = strings.ToLower(*p.Status)
}

func (p *BackupStatusParam) Check() error {
	p.Format()

	switch *p.Status {
	case constant.BACKUP_CANCELED:
		return nil
	default:
		return errors.Occur(errors.ErrObBackupStatusInvalid, *p.Status, constant.BACKUP_CANCELED)
	}
}

type ArchiveLogStatusParam struct {
	Status *string `json:"status"`
}

func (p *ArchiveLogStatusParam) Format() {
	if p.Status == nil || *p.Status == "" {
		status := constant.ARCHIVELOG_STATUS_STOP
		p.Status = &status
	}
	*p.Status = strings.ToUpper(*p.Status)
}

func (p *ArchiveLogStatusParam) Check() error {
	p.Format()

	switch *p.Status {
	case constant.ARCHIVELOG_STATUS_STOP,
		constant.ARCHIVELOG_STATUS_DOING,
		constant.ARCHIVELOG_STATUS_SUSPEND:
		return nil
	default:
		return errors.Occur(errors.ErrObBackupArchiveLogStatusInvalid, *p.Status, constant.ARCHIVELOG_STATUS_STOP, constant.ARCHIVELOG_STATUS_DOING)
	}
}

type BackupOverview struct {
	Statuses []oceanbase.CdbObBackupTask `json:"statuses"`
}

type TenantBackupOverview struct {
	Status oceanbase.CdbObBackupTask `json:"status"`
}

type QueryBackupTasksParam struct {
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
	CustomPageQuery
	Status       string   `form:"status"`
	Sort         string   `form:"sort,default=start_timestamp,desc"`
	ParsedStatus []string `form:"-"`
	SortBy       string   `form:"-"`
	SortOrder    string   `form:"-"`
}

type CustomPageQuery struct {
	Page uint64 `form:"page,default=1"`
	Size uint64 `form:"size,default=2147483647"`
}

func (p *CustomPageQuery) Format() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size < 1 {
		p.Size = 2147483647
	}
}

func (p *QueryBackupTasksParam) Format() {
	if p.Status != "" {
		p.ParsedStatus = strings.Split(strings.ToUpper(p.Status), ",")
	}
	p.CustomPageQuery.Format()
	if p.Sort != "" {
		parts := strings.Split(p.Sort, ",")
		if len(parts) == 2 {
			p.SortBy = parts[0]
			p.SortOrder = parts[1]
		} else {
			p.SortBy = parts[0]
		}
	}
	if p.SortBy != "start_timestamp" && p.SortBy != "end_timestamp" {
		p.SortBy = "start_timestamp"
	}
	if p.SortOrder != "asc" && p.SortOrder != "desc" {
		p.SortOrder = "desc"
	}
}
