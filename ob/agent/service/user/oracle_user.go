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
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"gorm.io/gorm"
)

type oracleUserService struct {
	db *gorm.DB
}

func NewOracleUserService(db *gorm.DB) *oracleUserService {
	return &oracleUserService{db: db}
}

func (t *oracleUserService) GetDb() *gorm.DB {
	return t.db
}

func (t *oracleUserService) ListUsers() ([]oceanbase.ObUser, error) {
	var users []oceanbase.OracleUser
	err := t.db.Raw(SQL_SELECT_ORACLE_ALL_USER).Scan(&users).Error
	if err != nil {
		return nil, err
	}

	obUsers := make([]oceanbase.ObUser, len(users))
	for i := range users {
		obUsers[i] = &users[i]
	}
	return obUsers, nil
}

func (t *oracleUserService) GetUser(userName string) (oceanbase.ObUser, error) {
	var user oceanbase.OracleUser
	err := t.db.Raw(SQL_SELECT_ORACLE_USER, userName).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (t *oracleUserService) CreateUser(userName, password, hostName string) error {
	sql := fmt.Sprintf("CREATE USER '%s' IDENTIFIED BY \"%s\"", userName, password) // ' is not allowed in password
	return t.db.Exec(sql).Error
}

func (t *oracleUserService) DropUser(userName string) error {
	sql := fmt.Sprintf("DROP USER '%s' CASCADE", userName)
	return t.db.Exec(sql).Error
}

func (t *oracleUserService) ChangeUserPassword(userName, password string) error {
	sql := fmt.Sprintf("ALTER USER '%s' IDENTIFIED BY \"%s\"", userName, password)
	return t.db.Exec(sql).Error
}

func (t *oracleUserService) IsUserExist(userName string) (bool, error) {
	var count int64
	err := t.db.Raw("SELECT COUNT(*) FROM DBA_USERS WHERE USERNAME = ?", userName).Scan(&count).Error
	return count == 1, err
}

func (t *oracleUserService) GrantGlobalPrivileges(userName string, privileges []string) error {
	if len(privileges) == 0 {
		return nil
	}
	realPrivileges := convertPrivileges(privileges)
	sql := fmt.Sprintf("GRANT %s TO '%s'", strings.Join(realPrivileges, ","), userName)
	return t.db.Exec(sql).Error
}

func (t *oracleUserService) GetGlobalPrivilegesMap() (map[string][]string, error) {
	var globalPrivileges []oceanbase.GlobalPrivilege
	result := t.db.Raw(SQL_SELECT_SYSTEM_PRIVS).Scan(&globalPrivileges)
	globalPrivilegesMap := make(map[string][]string)
	for _, systemPriv := range globalPrivileges {
		globalPrivilegesMap[systemPriv.Grantee] = append(globalPrivilegesMap[systemPriv.Grantee], systemPriv.Privilege)
	}
	for userName, privileges := range globalPrivilegesMap {
		globalPrivilegesMap[userName] = formatGlobalPrivileges(privileges)
	}
	return globalPrivilegesMap, result.Error
}

func (t *oracleUserService) RevokeGlobalPrivileges(userName string, privileges []string) error {
	if len(privileges) == 0 {
		return nil
	}
	realPrivileges := convertPrivileges(privileges)
	sql := fmt.Sprintf("REVOKE %s FROM '%s'", strings.Join(realPrivileges, ","), userName)
	return t.db.Exec(sql).Error
}

func (t *oracleUserService) GetGrantedGlobalPrivileges(userName string) ([]string, error) {
	var globalPrivileges []oceanbase.GlobalPrivilege
	err := t.db.Raw("SELECT GRANTEE, PRIVILEGE FROM DBA_SYS_PRIVS WHERE GRANTEE = ?", userName).Scan(&globalPrivileges).Error
	if err != nil {
		return nil, err
	}

	privileges := make([]string, len(globalPrivileges))
	for i := range globalPrivileges {
		privileges[i] = globalPrivileges[i].Privilege
	}
	return formatGlobalPrivileges(privileges), nil
}

func (t *oracleUserService) ModifyTenantRootPassword(newPwd string) error {
	if err := t.db.Exec(fmt.Sprintf(SQL_ORACLE_ALTER_TENANT_ROOT_PASSWORD, newPwd)).Error; err != nil {
		return errors.Wrap(err, "modify tenant root password failed")
	}
	return nil
}

func formatGlobalPrivileges(privileges []string) []string {
	formatPrivileges := make([]string, 0)
	for _, privilege := range privileges {
		formatPrivileges = append(formatPrivileges, strings.ReplaceAll(strings.ToUpper(privilege), " ", "_"))
	}
	return formatPrivileges
}
