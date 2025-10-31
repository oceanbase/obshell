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

package user

import (
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"gorm.io/gorm"
)

type userService interface {
	ListUsers() ([]oceanbase.ObUser, error)
	GetUser(userName string) (oceanbase.ObUser, error)
	CreateUser(userName, password, hostName string) error
	DropUser(userName string) error
	ChangeUserPassword(userName, password string) error
	IsUserExist(userName string) (bool, error)
	GrantGlobalPrivileges(userName string, privileges []string) error
	RevokeGlobalPrivileges(userName string, privileges []string) error
	GetGrantedGlobalPrivileges(userName string) ([]string, error)
	GetGlobalPrivilegesMap() (map[string][]string, error)
	ModifyTenantRootPassword(newPwd string) error
}

func GetUserService(db *gorm.DB) userService {
	mode := db.Dialector.Name()
	if mode == constant.MYSQL_MODE {
		return NewMysqlUserService(db)
	} else {
		return NewOracleUserService(db)
	}
}

var (
	// mysql user table
	DBA_OB_USERS  = "oceanbase.DBA_OB_USERS"
	GV_OB_SESSION = "oceanbase.GV$OB_SESSION"
	MYSQL_USER    = "mysql.user"

	// oracle execute sql
	SQL_SELECT_ORACLE_ALL_USER            = "SELECT USERNAME, CREATED, ACCOUNT_STATUS FROM DBA_USERS"
	SQL_SELECT_ORACLE_USER                = "SELECT USERNAME, CREATED, ACCOUNT_STATUS FROM DBA_USERS WHERE USERNAME = ?"
	SQL_SELECT_SYSTEM_PRIVS               = "SELECT GRANTEE, PRIVILEGE FROM DBA_SYS_PRIVS"
	SQL_ORACLE_ALTER_TENANT_ROOT_PASSWORD = "ALTER USER 'SYS' IDENTIFIED BY \"%s\""

	// mysql execute sql
	SQL_MYSQL_ALTER_TENANT_ROOT_PASSWORD = "ALTER USER root@'%%' IDENTIFIED BY \"%s\""
)
