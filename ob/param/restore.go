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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type RestoreWindowsParam struct {
	RestoreStorageParam
}

type RestoreStorageParam struct {
	DataBackupUri string  `json:"data_backup_uri" binding:"required"`
	ArchiveLogUri *string `json:"archive_log_uri"`
}

type RestoreParam struct {
	RestoreStorageParam

	TenantName string `json:"restore_tenant_name" binding:"required"`

	Timestamp *time.Time `json:"timestamp" time_format:"2006-01-02T15:04:05.000Z07:00"` // time_format does not cause precision loss
	SCN       *int64     `json:"scn"`

	ZoneList          []ZoneParam `json:"zone_list" binding:"required"` // Tenant zone list with unit config.
	PrimaryZone       *string     `json:"primary_zone"`
	Concurrency       *int        `json:"concurrency"`
	HaHighThreadScore *int        `json:"ha_high_thread_score"`
	Decryption        *[]string   `json:"decryption"`

	KmsEncryptInfo *string `json:"kms_encrypt_info"`
}

func (p *RestoreParam) Format() {
	if p.HaHighThreadScore == nil {
		score := constant.HA_LOW_THREAD_SCORE_HIGH
		p.HaHighThreadScore = &score
	}
	if p.ArchiveLogUri == nil || *p.ArchiveLogUri == "" {
		p.ArchiveLogUri = &p.DataBackupUri
	}
	if p.SCN != nil && *p.SCN == 0 {
		p.SCN = nil
	}
	if p.Timestamp != nil && *p.Timestamp == constant.ZERO_TIME {
		p.Timestamp = nil
	}
	if p.PrimaryZone == nil || *p.PrimaryZone == "" ||
		strings.ToUpper(*p.PrimaryZone) == constant.PRIMARY_ZONE_RANDOM {
		primaryZone := constant.PRIMARY_ZONE_RANDOM
		p.PrimaryZone = &primaryZone
	}
}

func (p *RestoreParam) Check() error {
	p.Format()
	if (p.Timestamp != nil && *p.Timestamp != constant.ZERO_TIME) && (p.SCN != nil && *p.SCN != 0) {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "timestamp or scn", "cannot be set at the same time")
	}

	if p.HaHighThreadScore != nil && (*p.HaHighThreadScore < 0 || *p.HaHighThreadScore > 100) {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "ha_high_thread_score", "should be in [0, 100]")
	}
	return nil
}

type RestoreOverview struct {
	bo.RestoreInfo

	RecoverScn        int64  `json:"recover_scn"`
	RecoverScnDisplay string `json:"recover_scn_display"`
	RecoverProgress   string `json:"recover_progress"`
	RestoreProgress   string `json:"restore_progress"`

	BackupClusterVersion int    `json:"backup_cluster_version"`
	LsCount              int    `json:"ls_count"`
	FinishLsCount        int    `json:"finish_ls_count"`
	Comment              string `json:"comment"`
	FinishTimestamp      string `json:"finish_timestamp"`
}

type QueryRestoreTasksParam struct {
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
	CustomPageQuery
	Status       string   `form:"status"`
	ParsedStatus []string `form:"-"`
}

func (p *QueryRestoreTasksParam) Format() {
	if p.Status != "" {
		p.ParsedStatus = strings.Split(strings.ToUpper(p.Status), ",")
	}
	p.CustomPageQuery.Format()
}
