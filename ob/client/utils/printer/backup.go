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

const (
	TENANT_ID           = "TENANT_ID"
	JOB_ID              = "JOB_ID"
	STATUS              = "STATUS"
	START_TIMESTAMP     = "START_TIMESTAMP"
	TABLET_COUNT        = "TABLET_COUNT"
	FINISH_TABLET_COUNT = "FINISH_TABLET_COUNT"
	COMMENT             = "COMMENT"

	TASK_ID                  = "TASK_ID"
	INCARNATION              = "INCARNATION"
	BACKUP_SET_ID            = "BACKUP_SET_ID"
	END_TIMESTAMP            = "END_TIMESTAMP"
	START_SCN                = "START_SCN"
	END_SCN                  = "END_SCN"
	USER_LS_START_SCN        = "USER_LS_START_SCN"
	ENCRYPTION_MODE          = "ENCRYPTION_MODE"
	INPUT_BYTES              = "INPUT_BYTES"
	OUTPUT_BYTES             = "OUTPUT_BYTES"
	OUTPUT_RATE_BYTES        = "OUTPUT_RATE_BYTES"
	EXTRA_META_BYTES         = "EXTRA_META_BYTES"
	MACRO_BLOCK_COUNT        = "MACRO_BLOCK_COUNT"
	FINISH_MACRO_BLOCK_COUNT = "FINISH_MACRO_BLOCK_COUNT"
	FILE_COUNT               = "FILE_COUNT"
	META_TURN_ID             = "META_TURN_ID"
	DATA_TURN_ID             = "DATA_TURN_ID"
	RESULT                   = "RESULT"
	PATH                     = "PATH"
)

func PrintDetailedClusterBackupOverview(overview *param.BackupOverview) {
	headers := []string{TENANT_ID, TASK_ID, JOB_ID, INCARNATION, BACKUP_SET_ID, START_TIMESTAMP, END_TIMESTAMP, STATUS, START_SCN, END_SCN, USER_LS_START_SCN, ENCRYPTION_MODE, INPUT_BYTES, OUTPUT_BYTES, OUTPUT_RATE_BYTES, EXTRA_META_BYTES, TABLET_COUNT, FINISH_TABLET_COUNT, MACRO_BLOCK_COUNT, FINISH_MACRO_BLOCK_COUNT, FILE_COUNT, META_TURN_ID, DATA_TURN_ID, RESULT, COMMENT, PATH}
	data := [][]string{}
	for _, status := range overview.Statuses {
		data = append(data, []string{
			fmt.Sprint(status.TenantID),
			fmt.Sprint(status.TaskID),
			fmt.Sprint(status.JobID),
			fmt.Sprint(status.Incarnation),
			fmt.Sprint(status.BackupSetID),
			fmt.Sprint(status.StartTimestamp),
			fmt.Sprint(status.EndTimestamp),
			status.Status,
			fmt.Sprint(status.StartScn),
			fmt.Sprint(status.EndScn),
			fmt.Sprint(status.UserLsStartScn),
			status.EncryptionMode,
			fmt.Sprint(status.InputBytes),
			fmt.Sprint(status.OutputBytes),
			fmt.Sprint(status.OutputRateBytes),
			fmt.Sprint(status.ExtraMetaBytes),
			fmt.Sprint(status.TabletCount),
			fmt.Sprint(status.FinishTabletCount),
			fmt.Sprint(status.MacroBlockCount),
			fmt.Sprint(status.FinishMacroBlockCount),
			fmt.Sprint(status.FileCount),
			fmt.Sprint(status.MetaTurnID),
			fmt.Sprint(status.DataTurnID),
			fmt.Sprint(status.Result),
			status.Comment,
			status.Path,
		})
	}
	stdio.PrintTable(headers, data)
}

func PrintClusterBackupOverview(overview *param.BackupOverview) {
	headers := []string{TENANT_ID, TASK_ID, JOB_ID, BACKUP_SET_ID, START_TIMESTAMP, END_TIMESTAMP, STATUS, ENCRYPTION_MODE, PATH}
	data := [][]string{}
	for _, status := range overview.Statuses {
		data = append(data, []string{
			fmt.Sprint(status.TenantID),
			fmt.Sprint(status.TaskID),
			fmt.Sprint(status.JobID),
			fmt.Sprint(status.BackupSetID),
			fmt.Sprint(status.StartTimestamp),
			fmt.Sprint(status.EndTimestamp),
			status.Status,
			status.EncryptionMode,
			status.Path,
		})
	}
	stdio.PrintTable(headers, data)

}

func PrintDetailedTenantBackupOverview(overview *param.TenantBackupOverview) {
	data := [][]string{
		{TASK_ID, fmt.Sprint(overview.Status.TaskID)},
		{JOB_ID, fmt.Sprint(overview.Status.JobID)},
		{INCARNATION, fmt.Sprint(overview.Status.Incarnation)},
		{BACKUP_SET_ID, fmt.Sprint(overview.Status.BackupSetID)},
		{START_TIMESTAMP, fmt.Sprint(overview.Status.StartTimestamp)},
		{END_TIMESTAMP, fmt.Sprint(overview.Status.EndTimestamp)},
		{STATUS, overview.Status.Status},
		{START_SCN, fmt.Sprint(overview.Status.StartScn)},
		{END_SCN, fmt.Sprint(overview.Status.EndScn)},
		{USER_LS_START_SCN, fmt.Sprint(overview.Status.UserLsStartScn)},
		{ENCRYPTION_MODE, overview.Status.EncryptionMode},
		{INPUT_BYTES, fmt.Sprint(overview.Status.InputBytes)},
		{OUTPUT_BYTES, fmt.Sprint(overview.Status.OutputBytes)},
		{OUTPUT_RATE_BYTES, fmt.Sprint(overview.Status.OutputRateBytes)},
		{EXTRA_META_BYTES, fmt.Sprint(overview.Status.ExtraMetaBytes)},
		{TABLET_COUNT, fmt.Sprint(overview.Status.TabletCount)},
		{FINISH_TABLET_COUNT, fmt.Sprint(overview.Status.FinishTabletCount)},
		{MACRO_BLOCK_COUNT, fmt.Sprint(overview.Status.MacroBlockCount)},
		{FINISH_MACRO_BLOCK_COUNT, fmt.Sprint(overview.Status.FinishMacroBlockCount)},
		{FILE_COUNT, fmt.Sprint(overview.Status.FileCount)},
		{META_TURN_ID, fmt.Sprint(overview.Status.MetaTurnID)},
		{DATA_TURN_ID, fmt.Sprint(overview.Status.DataTurnID)},
		{RESULT, fmt.Sprint(overview.Status.Result)},
		{COMMENT, overview.Status.Comment},
		{PATH, overview.Status.Path},
	}
	stdio.PrintTable(nil, data)
}

func PrintTenantBackupOverview(overview *param.TenantBackupOverview) {
	data := [][]string{
		{TASK_ID, fmt.Sprint(overview.Status.TaskID)},
		{JOB_ID, fmt.Sprint(overview.Status.JobID)},
		{BACKUP_SET_ID, fmt.Sprint(overview.Status.BackupSetID)},
		{START_TIMESTAMP, fmt.Sprint(overview.Status.StartTimestamp)},
		{END_TIMESTAMP, fmt.Sprint(overview.Status.EndTimestamp)},
		{STATUS, overview.Status.Status},
		{ENCRYPTION_MODE, overview.Status.EncryptionMode},
		{PATH, overview.Status.Path},
	}
	stdio.PrintTable(nil, data)
}
