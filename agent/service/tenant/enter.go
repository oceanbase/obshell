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

import "strings"

type TenantService struct{}

const (
	DBA_OB_TENANTS               = "oceanbase.DBA_OB_TENANTS"
	DBA_OB_UNITS                 = "oceanbase.DBA_OB_UNITS"
	DBA_OB_RESOURCE_POOLS        = "oceanbase.DBA_OB_RESOURCE_POOLS"
	DBA_OB_TENANT_JOBS           = "oceanbase.DBA_OB_TENANT_JOBS"
	DBA_OB_UNIT_CONFIGS          = "oceanbase.DBA_OB_UNIT_CONFIGS"
	DBA_OB_CLUSTER_EVENT_HISTORY = "oceanbase.DBA_OB_CLUSTER_EVENT_HISTORY"
	DBA_RECYCLEBIN               = "oceanbase.DBA_RECYCLEBIN"
	DBA_OB_USERS                 = "oceanbase.DBA_OB_USERS"
	DBA_OB_DATABASES             = "oceanbase.DBA_OB_DATABASES"
	DBA_OBJECTS                  = "oceanbase.DBA_OBJECTS"

	CDB_OB_SYS_VARIABLES        = "oceanbase.CDB_OB_SYS_VARIABLES"
	DBA_OB_SYS_VARIABLES        = "oceanbase.DBA_OB_SYS_VARIABLES"
	CDB_OB_ARCHIVELOG           = "oceanbase.CDB_OB_ARCHIVELOG"
	CDB_OB_ARCHIVELOG_SUMMARY   = "oceanbase.CDB_OB_ARCHIVELOG_SUMMARY"
	CDB_OB_BACKUP_DELETE_POLICY = "oceanbase.CDB_OB_BACKUP_DELETE_POLICY"
	CDB_OB_BACKUP_JOBS          = "oceanbase.CDB_OB_BACKUP_JOBS"
	CDB_OB_BACKUP_JOB_HISTORY   = "oceanbase.CDB_OB_BACKUP_JOB_HISTORY"
	CDB_OB_ARCHIVE_DEST         = "oceanbase.CDB_OB_ARCHIVE_DEST"
	CDB_OB_BACKUP_PARAMETER     = "oceanbase.CDB_OB_BACKUP_PARAMETER"
	CDB_OB_BACKUP_TASKS         = "oceanbase.CDB_OB_BACKUP_TASKS"
	CDB_OB_BACKUP_TASK_HISTORY  = "oceanbase.CDB_OB_BACKUP_TASK_HISTORY"
	CDB_OB_RESTORE_PROGRESS     = "oceanbase.CDB_OB_RESTORE_PROGRESS"
	CDB_OB_RESTORE_HISTORY      = "oceanbase.CDB_OB_RESTORE_HISTORY"

	GV_OB_PARAMETERS = "oceanbase.GV$OB_PARAMETERS"
	GV_OB_SERVERS    = "oceanbase.GV$OB_SERVERS"
	GV_OB_SESSION    = "oceanbase.GV$OB_SESSION"

	MYSQL_TIME_ZONE = "mysql.time_zone"
	MYSQL_USER      = "mysql.user"
	MYSQL_DB        = "mysql.db"

	INFOMATION_SCHEMA_COLLATIONS = "information_schema.collations"
)

const (
	// tenant sql
	SQL_CREATE_TENANT_BASIC = "CREATE TENANT `%s` resource_pool_list=(%s)"
	SQL_DROP_TENANT         = "DROP TENANT IF EXISTS `%s` FORCE"
	SQL_RECYCLE_TENANT      = "set session recyclebin=1; DROP TENANT `%s`"
	SQL_RENAME_TENANT       = "ALTER TENANT `%s` RENAME GLOBAL_NAME TO `%s`"
	SQL_FLASHBACK_TENANT    = "FLASHBACK TENANT `%s` TO BEFORE DROP RENAME TO `%s`"
	SQL_PURGE_TENANT        = "PURGE TENANT `%s`"

	SQL_ALTER_RESOURCE_LIST        = "ALTER TENANT `%s` RESOURCE_POOL_LIST=(%s)"
	SQL_ALTER_TENANT_LOCALITY      = "ALTER TENANT `%s` LOCALITY = \"%s\""
	SQL_ALTER_TENANT_UNIT_NUM      = "ALTER RESOURCE TENANT `%s` UNIT_NUM = %d"
	SQL_ALTER_TENANT_PRIMARY_ZONE  = "ALTER TENANT `%s` PRIMARY_ZONE = `%s`"
	SQL_ALTER_TENANT_ROOT_PASSWORD = "ALTER USER root@'%%' IDENTIFIED BY \"%s\""

	SQL_LOCK_TENANT   = "ALTER TENANT `%s` LOCK"
	SQL_UNLOCK_TENANT = "ALTER TENANT `%s` UNLOCK"

	SQL_SET_TENANT_PARAMETER_BASIC = "ALTER SYSTEM SET "
	SQL_SET_TENANT_VARIABLE_BASIC  = "ALTER TENANT `%s` SET VARIABLES "

	// resource pool sql
	SQL_CREATE_RESOURCE_POOL = "CREATE RESOURCE POOL `%s` UNIT = `%s`, UNIT_NUM = %d, ZONE_LIST = ('%s')"

	SQL_DROP_RESOURCE_POOL_IF_EXISTS = "DROP RESOURCE POOL IF EXISTS `%s`"
	SQL_DROP_RESOURCE_POOL           = "DROP RESOURCE POOL `%s`"

	SQL_ALTER_RESOURCE_POOL_SPLIT       = "ALTER RESOURCE POOL `%s` SPLIT INTO (%s) ON (%s)"
	SQL_ALTER_RESOURCE_POOL_UNIT_CONFIG = "ALTER RESOURCE POOL `%s` UNIT = `%s`"

	// parameters and variables
	SQL_ALTER_TENANT_WHITELIST = "ALTER TENANT `%s` SET VARIABLES ob_tcp_invited_nodes = `%s`"
)

func transfer(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	return str
}
