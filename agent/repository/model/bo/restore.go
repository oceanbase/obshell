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

type RestoreInfo struct {
	TenantId          int64  `json:"tenant_id"`
	JobID             int64  `json:"job_id"`
	RestoreTenantName string `json:"restore_tenant_name"`
	RestoreTenantId   int64  `json:"restore_tenant_id"`
	BackupTenantName  string `json:"backup_tenant_name"`
	BackupTenantId    int64  `json:"backup_tenant_id"`
	BackupClusterName string `json:"backup_cluster_name"`

	RestoreOption     string `json:"restore_option"`
	RestoreScn        int64  `json:"restore_scn"`
	RestoreScnDisplay string `json:"restore_scn_display"`

	Status         string `json:"status"`
	StartTimestamp string `json:"start_timestamp"`

	BackupSetList   string `json:"backup_set_list"`
	BackupPieceList string `json:"backup_piece_list"`

	BackupDataUri string `json:"backup_data_uri"`
	BackupLogUri  string `json:"backup_log_uri"`

	TabletCount        int64  `json:"tablet_count"`
	FinishTabletCount  int64  `json:"finish_tablet_count"`
	TotalBytes         int64  `json:"total_bytes"`
	TotalBytesDisplay  string `json:"total_bytes_display"`
	FinishBytes        int64  `json:"finish_bytes"`
	FinishBytesDisplay string `json:"finish_bytes_display"`
	Description        string `json:"description"`
}

type CdbObRestoreProgress struct {
	RestoreInfo

	RecoverScn        int64  `json:"recover_scn"`
	RecoverScnDisplay string `json:"recover_scn_display"`
	RecoverProgress   string `json:"recover_progress"`
	RestoreProgress   string `json:"restore_progress"`
}

type CdbObRestoreHistory struct {
	RestoreInfo

	BackupClusterVersion int    `json:"backup_cluster_version"`
	LsCount              int    `json:"ls_count"`
	FinishLsCount        int    `json:"finish_ls_count"`
	Comment              string `json:"comment"`
	FinishTimestamp      string `json:"finish_timestamp"`
}

type RestoreTask struct {
	RestoreInfo

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

type PaginatedRestoreTaskResponse struct {
	Contents []RestoreTask `json:"contents"`
	Page     CustomPage    `json:"page"`
}
