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

package bo

import (
	"time"
)

type ArchiveLogTask struct {
	TenantID   int       `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	RoundID    int       `json:"round_id"`
	Status     string    `json:"status"`
	Path       string    `json:"path"`
	Delay      float64   `json:"delay"` // seconds
	Checkpoint time.Time `json:"checkpoint"`
	StartTime  time.Time `json:"start_time"`
}

type TenantBackupInfo struct {
	TenantId                    int        `json:"tenant_id"`
	LastestArchiveLogCheckpoint *time.Time `json:"lastest_archive_log_checkpoint,omitempty"`
	LastestDataBackupTime       *time.Time `json:"lastest_data_backup_time,omitempty"`
	ArchiveLogDelay             float64    `json:"archive_log_delay,omitempty"`
}

type BackupDestInfo struct {
	TenantID       int    `json:"tenant_id"`
	ArchiveBaseUri string `json:"archive_base_uri"`
	DataBaseUri    string `json:"data_base_uri"`
}

type BackupJob struct {
	TenantID       int64      `json:"tenant_id"`
	JobID          int64      `json:"job_id"`
	BackupSetID    int64      `json:"backup_set_id"`
	PlusArchivelog string     `json:"plus_archivelog"`
	BackupType     string     `json:"backup_type"`
	StartTimestamp *time.Time `json:"start_timestamp"`
	EndTimestamp   *time.Time `json:"end_timestamp,omitempty"`
	Status         string     `json:"status"`
	Result         int64      `json:"result"`
	Comment        string     `json:"comment"`
	Description    string     `json:"description"`
	Path           string     `json:"path"`
}

type CustomPage struct {
	TotalElements uint64 `json:"total_elements"`
	TotalPages    uint64 `json:"total_pages"`
	Size          uint64 `json:"size"`
	Number        uint64 `json:"number"`
}

// PaginatedBackupJobResponse 用于 Swagger 文档的具体类型定义
type PaginatedBackupJobResponse struct {
	Contents []BackupJob `json:"contents"`
	Page     CustomPage  `json:"page"`
}
