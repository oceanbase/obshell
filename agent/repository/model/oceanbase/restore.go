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

import (
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

type RestoreInfo struct {
	TenantId          int64  `json:"tenant_id" gorm:"column:TENANT_ID"`
	JobID             int64  `json:"job_id" gorm:"column:JOB_ID"`
	RestoreTenantName string `json:"restore_tenant_name" gorm:"column:RESTORE_TENANT_NAME"`
	RestoreTenantId   int64  `json:"restore_tenant_id" gorm:"column:RESTORE_TENANT_ID"`
	BackupTenantName  string `json:"backup_tenant_name" gorm:"column:BACKUP_TENANT_NAME"`
	BackupTenantId    int64  `json:"backup_tenant_id" gorm:"column:BACKUP_TENANT_ID"`
	BackupClusterName string `json:"backup_cluster_name" gorm:"column:BACKUP_CLUSTER_NAME"`

	RestoreOption     string `json:"restore_option" gorm:"column:RESTORE_OPTION"`
	RestoreScn        int64  `json:"restore_scn" gorm:"column:RESTORE_SCN"`
	RestoreScnDisplay string `json:"restore_scn_display" gorm:"column:RESTORE_SCN_DISPLAY"`

	Status         string `json:"status" gorm:"column:STATUS"`
	StartTimestamp string `json:"start_timestamp" gorm:"column:START_TIMESTAMP"`

	BackupSetList   string `json:"backup_set_list" gorm:"column:BACKUP_SET_LIST"`
	BackupPieceList string `json:"backup_piece_list" gorm:"column:BACKUP_PIECE_LIST"`

	BackupDest string `json:"backup_dest" gorm:"column:BACKUP_DEST"` // NOTICE: this field contains ak, don't return it!!!

	TabletCount        int64  `json:"tablet_count" gorm:"column:TABLET_COUNT"`
	FinishTabletCount  int64  `json:"finish_tablet_count" gorm:"column:FINISH_TABLET_COUNT"`
	TotalBytes         int64  `json:"total_bytes" gorm:"column:TOTAL_BYTES"`
	TotalBytesDisplay  string `json:"total_bytes_display" gorm:"column:TOTAL_BYTES_DISPLAY"`
	FinishBytes        int64  `json:"finish_bytes" gorm:"column:FINISH_BYTES"`
	FinishBytesDisplay string `json:"finish_bytes_display" gorm:"column:FINISH_BYTES_DISPLAY"`
	Description        string `json:"description" gorm:"column:DESCRIPTION"`
}

func (i *RestoreInfo) ToBO() *bo.RestoreInfo {
	res := &bo.RestoreInfo{
		TenantId:          i.TenantId,
		JobID:             i.JobID,
		RestoreTenantName: i.RestoreTenantName,
		RestoreTenantId:   i.RestoreTenantId,
		BackupTenantName:  i.BackupTenantName,
		BackupTenantId:    i.BackupTenantId,
		BackupClusterName: i.BackupClusterName,

		RestoreOption:     i.RestoreOption,
		RestoreScnDisplay: i.RestoreScnDisplay,
		RestoreScn:        i.RestoreScn,

		Status:         i.Status,
		StartTimestamp: i.StartTimestamp,

		BackupSetList:   i.BackupSetList,
		BackupPieceList: i.BackupPieceList,

		TabletCount:        i.TabletCount,
		FinishTabletCount:  i.FinishTabletCount,
		TotalBytes:         i.TotalBytes,
		TotalBytesDisplay:  i.TotalBytesDisplay,
		FinishBytes:        i.FinishBytes,
		FinishBytesDisplay: i.FinishBytesDisplay,
		Description:        i.Description,
	}
	return res
}

type CdbObRestoreProgress struct {
	RestoreInfo

	RecoverScn        int64  `json:"recover_scn" gorm:"column:RECOVER_SCN"`
	RecoverScnDisplay string `json:"recover_scn_display" gorm:"column:RECOVER_SCN_DISPLAY"`
	RecoverProgress   string `json:"recover_progress" gorm:"column:RECOVER_PROGRESS"`
	RestoreProgress   string `json:"restore_progress" gorm:"column:RESTORE_PROGRESS"`
}

func (i *CdbObRestoreProgress) ToBO() *bo.CdbObRestoreProgress {
	res := &bo.CdbObRestoreProgress{
		RestoreInfo: *i.RestoreInfo.ToBO(),

		RecoverScn:        i.RecoverScn,
		RecoverScnDisplay: i.RecoverScnDisplay,
		RecoverProgress:   i.RecoverProgress,
		RestoreProgress:   i.RestoreProgress,
	}
	return res
}

type CdbObRestoreHistory struct {
	RestoreInfo

	BackupClusterVersion int    `json:"backup_cluster_version" gorm:"column:BACKUP_CLUSTER_VERSION"`
	LsCount              int    `json:"ls_count" gorm:"column:LS_COUNT"`
	FinishLsCount        int    `json:"finish_ls_count" gorm:"column:FINISH_LS_COUNT"`
	Comment              string `json:"comment" gorm:"column:COMMENT"`
	FinishTimestamp      string `json:"finish_timestamp" gorm:"column:FINISH_TIMESTAMP"`
}

func (i *CdbObRestoreHistory) ToBO() *bo.CdbObRestoreHistory {
	res := &bo.CdbObRestoreHistory{
		RestoreInfo: *i.RestoreInfo.ToBO(),

		BackupClusterVersion: i.BackupClusterVersion,
		LsCount:              i.LsCount,
		FinishLsCount:        i.FinishLsCount,
		Comment:              i.Comment,
		FinishTimestamp:      i.FinishTimestamp,
	}
	return res
}
