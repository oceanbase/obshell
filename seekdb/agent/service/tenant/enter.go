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
	DBA_OB_USERS     = "oceanbase.DBA_OB_USERS"
	DBA_OB_DATABASES = "oceanbase.DBA_OB_DATABASES"

	DBA_OB_SYS_VARIABLES = "oceanbase.DBA_OB_SYS_VARIABLES"

	GV_OB_PARAMETERS = "oceanbase.GV$OB_PARAMETERS"
	GV_OB_SERVERS    = "oceanbase.GV$OB_SERVERS"
	GV_OB_SESSION    = "oceanbase.GV$OB_SESSION"

	MYSQL_TIME_ZONE = "mysql.time_zone"
	MYSQL_DB        = "mysql.db"

	INFOMATION_SCHEMA_COLLATIONS = "information_schema.collations"
)

const (
	SQL_SET_PARAMETER_BASIC = "ALTER SYSTEM SET "
	SQL_SET_VARIABLE_BASIC  = "SET GLOBAL "

	// parameters and variables
	SQL_ALTER_TENANT_WHITELIST = "SET GLOBAL ob_tcp_invited_nodes = `%s`"
)
