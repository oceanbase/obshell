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

package oceanbase

import "time"

type CdbObBackupDeletePolicy struct {
	TenantID       int64  `gorm:"column:TENANT_ID"`
	PolicyName     string `gorm:"column:POLICY_NAME"`
	RecoveryWindow string `gorm:"column:RECOVERY_WINDOW"`
}

type CdbOBArchivelog struct {
	tenantID int    `gorm:"column:TENANT_ID"`
	Status   string `gorm:"column:STATUS"`
	Path     string `gorm:"column:PATH"`
}

type CdbObBackupTask struct {
	TenantID              int64     `gorm:"column:TENANT_ID" json:"tenant_id"`
	TaskID                int64     `gorm:"column:TASK_ID" json:"task_id"`
	JobID                 int64     `gorm:"column:JOB_ID" json:"job_id"`
	Incarnation           int64     `gorm:"column:INCARNATION" json:"incarnation"`
	BackupSetID           int64     `gorm:"column:BACKUP_SET_ID" json:"backup_set_id"`
	StartTimestamp        time.Time `gorm:"column:START_TIMESTAMP" json:"start_timestamp"`
	EndTimestamp          time.Time `gorm:"column:END_TIMESTAMP" json:"end_timestamp"`
	Status                string    `gorm:"column:STATUS" json:"status"`
	StartScn              int64     `gorm:"column:START_SCN" json:"start_scn"`
	EndScn                int64     `gorm:"column:END_SCN" json:"end_scn"`
	UserLsStartScn        int64     `gorm:"column:USER_LS_START_SCN" json:"user_ls_start_scn"`
	EncryptionMode        string    `gorm:"column:ENCRYPTION_MODE" json:"encryption_mode"`
	InputBytes            int64     `gorm:"column:INPUT_BYTES" json:"input_bytes"`
	OutputBytes           int64     `gorm:"column:OUTPUT_BYTES" json:"output_bytes"`
	OutputRateBytes       float64   `gorm:"column:OUTPUT_RATE_BYTES" json:"output_rate_bytes"`
	ExtraMetaBytes        int64     `gorm:"column:EXTRA_META_BYTES" json:"extra_meta_bytes"`
	TabletCount           int64     `gorm:"column:TABLET_COUNT" json:"tablet_count"`
	FinishTabletCount     int64     `gorm:"column:FINISH_TABLET_COUNT" json:"finish_tablet_count"`
	MacroBlockCount       int64     `gorm:"column:MACRO_BLOCK_COUNT" json:"macro_block_count"`
	FinishMacroBlockCount int64     `gorm:"column:FINISH_MACRO_BLOCK_COUNT" json:"finish_macro_block_count"`
	FileCount             int64     `gorm:"column:FILE_COUNT" json:"file_count"`
	MetaTurnID            int64     `gorm:"column:META_TURN_ID" json:"meta_turn_id"`
	DataTurnID            int64     `gorm:"column:DATA_TURN_ID" json:"data_turn_id"`
	Result                int64     `gorm:"column:RESULT" json:"result"`
	Comment               string    `gorm:"column:COMMENT" json:"comment"`
	Path                  string    `gorm:"column:PATH" json:"path"`
}
