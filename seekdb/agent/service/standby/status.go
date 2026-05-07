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

package standby

import (
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/param"
)

// obStandbyStatus mirrors columns returned by
// SELECT ROLE, LOG_RESTORE_SOURCE, SYNC_SCN FROM oceanbase.__all_virtual_server_stat
type obStandbyStatus struct {
	Role             string `gorm:"column:ROLE"`
	LogRestoreSource string `gorm:"column:LOG_RESTORE_SOURCE"`
	SyncScn          uint64 `gorm:"column:SYNC_SCN"`
	ReadableScn      uint64 `gorm:"column:READABLE_SCN"`
}

// GetLocalStatus queries the local seekdb for role, log_restore_source and
// sync_scn. SyncStatus is intentionally NOT populated here: deriving it
// requires the upstream peer's authoritative sync_scn, which only
// GetFullStatus (in executor layer) has access to after the peer RPC completes.
func (s *StandbyService) GetLocalStatus() (param.LocalStandbyStatus, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return param.LocalStandbyStatus{Role: "unknown"}, err
	}

	var row obStandbyStatus
	err = db.Raw(
		"SELECT ROLE, LOG_RESTORE_SOURCE, SYNC_SCN, READABLE_SCN FROM oceanbase.__all_virtual_server_stat",
	).Scan(&row).Error
	if err != nil {
		return param.LocalStandbyStatus{Role: "unknown"}, err
	}

	var clusterName string
	_ = db.Raw("SELECT VALUE FROM oceanbase.V$OB_PARAMETERS WHERE NAME = ? LIMIT 1", "cluster").Scan(&clusterName).Error

	return param.LocalStandbyStatus{
		Role:             row.Role,
		InstanceName:     clusterName,
		LogRestoreSource: row.LogRestoreSource,
		SyncScn:          row.SyncScn,
		ReadableScn:      row.ReadableScn,
	}, nil
}
