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
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

type ClusterBackupConfigParam struct {
	BackupBaseUri *string `json:"backup_base_uri"`
	LogArchiveDestConf
	baseBackupConfigParam
}

func NewBackupConfigParamForCluster(p *ClusterBackupConfigParam) *BackupConfigParam {
	return &BackupConfigParam{
		BackupBaseUri:         p.BackupBaseUri,
		LogArchiveDestConf:    p.LogArchiveDestConf,
		baseBackupConfigParam: p.baseBackupConfigParam,
	}
}

func NewBackupConfigParamForTenant(p *TenantBackupConfigParam) *BackupConfigParam {
	return &BackupConfigParam{
		DataBaseUri:           p.DataBaseUri,
		ArchiveBaseUri:        p.ArchiveBaseUri,
		LogArchiveDestConf:    p.LogArchiveDestConf,
		baseBackupConfigParam: p.baseBackupConfigParam,
	}
}

type TenantBackupConfigParam struct {
	DataBaseUri    *string `json:"data_base_uri"`
	ArchiveBaseUri *string `json:"archive_base_uri"`
	LogArchiveDestConf
	baseBackupConfigParam
}

type baseBackupConfigParam struct {
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
	baseBackupConfigParam
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
		return nil, fmt.Errorf("log_archive_concurrency must be between %d and %d", constant.LOG_ARCHIVE_CONCURRENCY_LOW, constant.LOG_ARCHIVE_CONCURRENCY_HIGH)
	}

	if p.HaLowThreadScore != nil && (*p.HaLowThreadScore < constant.HA_LOW_THREAD_SCORE_LOW || *p.HaLowThreadScore > constant.HA_LOW_THREAD_SCORE_HIGH) {
		return nil, fmt.Errorf("ha_low_thread_score must be between %d and %d", constant.HA_LOW_THREAD_SCORE_LOW, constant.HA_LOW_THREAD_SCORE_HIGH)
	}

	if p.Binding != nil &&
		(*p.Binding != "" && *p.Binding != constant.BINDING_MODE_OPTIONAL &&
			*p.Binding != constant.BINDING_MODE_MANDATORY) {
		return nil, fmt.Errorf("invalid binding mode: %s, must be %s or %s",
			*p.Binding, constant.BINDING_MODE_OPTIONAL, constant.BINDING_MODE_MANDATORY)
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
			return fmt.Errorf("invalid delete policy: '%s', must be '%s'", p.DeletePolicy.Policy, constant.DELETE_POLICY_DEFAULT)
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
		return fmt.Errorf("piece_switch_interval must be greater than %v", constant.PIECE_SWITCH_INTERVAL_LOW)
	} else if duration > constant.PIECE_SWITCH_INTERVAL_HIGH {
		return fmt.Errorf("piece_switch_interval must be less than %v", constant.PIECE_SWITCH_INTERVAL_HIGH)
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
		return fmt.Errorf("archive_lag_target must be less than %v", constant.ARCHIVE_LAG_TARGET_HIGH)
	}

	if t == constant.PROTOCOL_S3 && duration < constant.ARCHIVE_LAG_TARGET_LOW_FOR_S3 {
		return fmt.Errorf("archive_lag_target must be greater than %v for s3", constant.ARCHIVE_LAG_TARGET_LOW_FOR_S3)
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
		return fmt.Errorf("invalid backup mode: %s, must be %s or %s",
			*p.Mode, constant.BACKUP_MODE_FULL, constant.BACKUP_MODE_INCREMENTAL)
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
		return fmt.Errorf("invalid backup status: '%s', must be '%s'",
			*p.Status, constant.BACKUP_CANCELED)
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
		constant.ARCHIVELOG_STATUS_DOING:
		return nil
	default:
		return fmt.Errorf("invalid archive log status: %s, must be %s or %s",
			*p.Status, constant.ARCHIVELOG_STATUS_STOP, constant.ARCHIVELOG_STATUS_DOING)
	}
}

type BackupOverview struct {
	Statuses []oceanbase.CdbObBackupTask `json:"statuses"`
}

type TenantBackupOverview struct {
	Status oceanbase.CdbObBackupTask `json:"status"`
}
