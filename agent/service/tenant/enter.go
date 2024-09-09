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

package tenant

type TenantService struct{}

const (
	GV_OB_PARAMETERS = "oceanbase.GV$OB_PARAMETERS"

	DBA_OB_TENANTS               = "oceanbase.DBA_OB_TENANTS"
	DBA_OB_UNIT_CONFIGS          = "oceanbase.DBA_OB_UNIT_CONFIGS"
	DBA_OB_RESOURCE_POOLS        = "oceanbase.DBA_OB_RESOURCE_POOLS"
	DBA_OB_CLUSTER_EVENT_HISTORY = "oceanbase.DBA_OB_CLUSTER_EVENT_HISTORY"

	CDB_OB_ARCHIVELOG           = "oceanbase.CDB_OB_ARCHIVELOG"
	CDB_OB_BACKUP_DELETE_POLICY = "oceanbase.CDB_OB_BACKUP_DELETE_POLICY"
	CDB_OB_BACKUP_JOBS          = "oceanbase.CDB_OB_BACKUP_JOBS"
	CDB_OB_ARCHIVE_DEST         = "oceanbase.CDB_OB_ARCHIVE_DEST"
	CDB_OB_BACKUP_PARAMETER     = "oceanbase.CDB_OB_BACKUP_PARAMETER"
	CDB_OB_BACKUP_TASKS         = "oceanbase.CDB_OB_BACKUP_TASKS"
	CDB_OB_BACKUP_TASK_HISTORY  = "oceanbase.CDB_OB_BACKUP_TASK_HISTORY"
	CDB_OB_RESTORE_PROGRESS     = "oceanbase.CDB_OB_RESTORE_PROGRESS"
	CDB_OB_RESTORE_HISTORY      = "oceanbase.CDB_OB_RESTORE_HISTORY"
)
