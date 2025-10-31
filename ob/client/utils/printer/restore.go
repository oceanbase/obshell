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

package printer

import (
	"fmt"

	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/param"
)

type RestoreOverview struct {
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

type RestoreInfo struct {
	TenantId          int64  `json:"tenant_id" gorm:"column:TENANT_ID"`
	JobID             int64  `json:"job_id" gorm:"column:JOB_ID"`
	RestoreTenantName string `json:"restore_tenant_name" gorm:"column:RESTORE_TENANT_NAME"`
	RestoreTenantId   int64  `json:"restore_tenant_id" gorm:"column:RESTORE_TENANT_ID"`
	BackupTenantName  string `json:"backup_tenant_name" gorm:"column:BACKUP_TENANT_NAME"`
	BackupTenantId    int64  `json:"backup_tenant_id" gorm:"column:BACKUP_TENANT_ID"`
	BackupClusterName string `json:"backup_cluster_name" gorm:"column:BACKUP_CLUSTER_NAME"`
	BackupDest        string `json:"backup_dest" gorm:"column:BACKUP_DEST"`

	RestoreOption     string `json:"restore_option" gorm:"column:RESTORE_OPTION"`
	RestoreScn        int64  `json:"restore_scn" gorm:"column:RESTORE_SCN"`
	RestoreScnDisplay string `json:"restore_scn_display" gorm:"column:RESTORE_SCN_DISPLAY"`

	Status         string `json:"status" gorm:"column:STATUS"`
	StartTimestamp string `json:"start_timestamp" gorm:"column:START_TIMESTAMP"`

	BackupSetList   string `json:"backup_set_list" gorm:"column:BACKUP_SET_LIST"`
	BackupPieceList string `json:"backup_piece_list" gorm:"column:BACKUP_PIECE_LIST"`

	TabletCount        int64  `json:"tablet_count" gorm:"column:TABLET_COUNT"`
	TotalBytes         int64  `json:"total_bytes" gorm:"column:TOTAL_BYTES"`
	Description        string `json:"description" gorm:"column:DESCRIPTION"`
	FinishTabletCount  int64  `json:"finish_tablet_count" gorm:"column:FINISH_TABLET_COUNT"`
	FinishBytes        int64  `json:"finish_bytes" gorm:"column:FINISH_BYTES"`
	FinishBytesDisplay string `json:"finish_bytes_display" gorm:"column:FINISH_BYTES_DISPLAY"`
	TotalBytesDisplay  string `json:"total_bytes_display" gorm:"column:TOTAL_BYTES_DISPLAY"`
}

const (
	RESTORE_TENANT_NAME    = "RESTORE_TENANT_NAME"
	RESTORE_TENANT_ID      = "RESTORE_TENANT_ID"
	BACKUP_TENANT_NAME     = "BACKUP_TENANT_NAME"
	BACKUP_TENANT_ID       = "BACKUP_TENANT_ID"
	BACKUP_CLUSTER_NAME    = "BACKUP_CLUSTER_NAME"
	BACKUP_DEST            = "BACKUP_DEST"
	RESTORE_OPTION         = "RESTORE_OPTION"
	RESTORE_SCN            = "RESTORE_SCN"
	RESTORE_SCN_DISPLAY    = "RESTORE_SCN_DISPLAY"
	BACKUP_SET_LIST        = "BACKUP_SET_LIST"
	BACKUP_PIECE_LIST      = "BACKUP_PIECE_LIST"
	TOTAL_BYTES            = "TOTAL_BYTES"
	DESCRIPTION            = "DESCRIPTION"
	FINISH_BYTES           = "FINISH_BYTES"
	FINISH_BYTES_DISPLAY   = "FINISH_BYTES_DISPLAY"
	TOTAL_BYTES_DISPLAY    = "TOTAL_BYTES_DISPLAY"
	RECOVER_SCN            = "RECOVER_SCN"
	RECOVER_SCN_DISPLAY    = "RECOVER_SCN_DISPLAY"
	RECOVER_PROGRESS       = "RECOVER_PROGRESS"
	RESTORE_PROGRESS       = "RESTORE_PROGRESS"
	BACKUP_CLUSTER_VERSION = "BACKUP_CLUSTER_VERSION"
	LS_COUNT               = "LS_COUNT"
	FINISH_LS_COUNT        = "FINISH_LS_COUNT"
	FINISH_TIMESTAMP       = "FINISH_TIMESTAMP"
)

func PrintDetailedTenantRestoreOverview(overview *param.RestoreOverview) {
	data := [][]string{
		{TENANT_ID, fmt.Sprint(overview.TenantId)},
		{JOB_ID, fmt.Sprint(overview.JobID)},
		{RESTORE_TENANT_NAME, overview.RestoreTenantName},
		{RESTORE_TENANT_ID, fmt.Sprint(overview.RestoreTenantId)},
		{BACKUP_TENANT_NAME, overview.BackupTenantName},
		{BACKUP_TENANT_ID, fmt.Sprint(overview.BackupTenantId)},
		{BACKUP_CLUSTER_NAME, overview.BackupClusterName},
		{STATUS, overview.Status},
		{START_TIMESTAMP, overview.StartTimestamp},
		{RESTORE_OPTION, overview.RestoreOption},
		{RESTORE_SCN, fmt.Sprint(overview.RestoreScn)},
		{RESTORE_SCN_DISPLAY, overview.RestoreScnDisplay},
		{BACKUP_SET_LIST, overview.BackupSetList},
		{BACKUP_PIECE_LIST, overview.BackupPieceList},
		{TABLET_COUNT, fmt.Sprint(overview.TabletCount)},
		{TOTAL_BYTES, fmt.Sprint(overview.TotalBytes)},
		{DESCRIPTION, overview.Description},
		{FINISH_TABLET_COUNT, fmt.Sprint(overview.FinishTabletCount)},
		{FINISH_BYTES, fmt.Sprint(overview.FinishBytes)},
		{FINISH_BYTES_DISPLAY, overview.FinishBytesDisplay},
		{TOTAL_BYTES_DISPLAY, overview.TotalBytesDisplay},
		{RECOVER_SCN, fmt.Sprint(overview.RecoverScn)},
		{RECOVER_SCN_DISPLAY, overview.RecoverScnDisplay},
		{RECOVER_PROGRESS, overview.RecoverProgress},
		{BACKUP_CLUSTER_VERSION, fmt.Sprint(overview.BackupClusterVersion)},
		{LS_COUNT, fmt.Sprint(overview.LsCount)},
		{FINISH_LS_COUNT, fmt.Sprint(overview.FinishLsCount)},
		{FINISH_TIMESTAMP, overview.FinishTimestamp},
		{COMMENT, overview.Comment},
	}
	stdio.PrintTable(nil, data)
}

func PrintTenantRestoreOverview(overview *param.RestoreOverview) {
	data := [][]string{
		{TENANT_ID, fmt.Sprint(overview.TenantId)},
		{JOB_ID, fmt.Sprint(overview.JobID)},
		{RESTORE_TENANT_NAME, overview.RestoreTenantName},
		{BACKUP_TENANT_NAME, overview.BackupTenantName},
		{BACKUP_CLUSTER_NAME, overview.BackupClusterName},
		{STATUS, overview.Status},
		{START_TIMESTAMP, overview.StartTimestamp},
		{RESTORE_OPTION, overview.RestoreOption},
		{RESTORE_SCN_DISPLAY, overview.RestoreScnDisplay},
		{BACKUP_SET_LIST, overview.BackupSetList},
		{BACKUP_PIECE_LIST, overview.BackupPieceList},
		{RECOVER_SCN_DISPLAY, overview.RecoverScnDisplay},
		{BACKUP_CLUSTER_VERSION, fmt.Sprint(overview.BackupClusterVersion)},
		{FINISH_TIMESTAMP, overview.FinishTimestamp},
	}
	stdio.PrintTable(nil, data)
}
