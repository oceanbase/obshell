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

import (
	"fmt"
	"strings"

	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
	"gorm.io/gorm"
)

const (
	SQL_SELECT_OBJECT_PRIVS = `SELECT P.GRANTEE, P.OWNER, O.OBJECT_TYPE, O.OBJECT_NAME, P.PRIVILEGE
		FROM DBA_TAB_PRIVS P
		JOIN (
			SELECT OWNER, OBJECT_TYPE, OBJECT_NAME
			FROM DBA_OBJECTS
			WHERE OBJECT_TYPE IN ('TABLE', 'VIEW', 'PROCEDURE')
		) O ON P.OWNER = O.OWNER AND P.TABLE_NAME = O.OBJECT_NAME
		WHERE GRANTEE = ?`
	SQL_SELECT_ALL_OBJECT_PRIVS = `SELECT P.GRANTEE, P.OWNER, O.OBJECT_TYPE, O.OBJECT_NAME, P.PRIVILEGE
		FROM DBA_TAB_PRIVS P
		JOIN (
			SELECT OWNER, OBJECT_TYPE, OBJECT_NAME
			FROM DBA_OBJECTS
			WHERE OBJECT_TYPE IN ('TABLE', 'VIEW', 'PROCEDURE')
		) O ON P.OWNER = O.OWNER AND P.TABLE_NAME = O.OBJECT_NAME`

	SQL_SELECT_ALL_ROLE         = "SELECT ROLE FROM DBA_ROLES"
	SQL_SELECT_ROLE             = "SELECT ROLE FROM DBA_ROLES WHERE ROLE = ?"
	SQL_SELECT_GRANTED_ROLE     = "SELECT GRANTEE, GRANTED_ROLE FROM DBA_ROLE_PRIVS WHERE GRANTEE = ?"
	SQL_SELECT_ALL_GRANTED_ROLE = "SELECT GRANTEE, GRANTED_ROLE FROM DBA_ROLE_PRIVS"
	SQL_SELECT_USER_GRANTEE     = "SELECT GRANTEE, GRANTED_ROLE FROM DBA_ROLE_PRIVS WHERE GRANTEE IN (SELECT USERNAME FROM DBA_USERS)"
	SQL_SELECT_ROLE_GRANTEE     = "SELECT GRANTEE, GRANTED_ROLE FROM DBA_ROLE_PRIVS WHERE GRANTEE IN (SELECT ROLE FROM DBA_ROLES)"

	SQL_GRANT_OBJECT_PRIVILEGE  = "GRANT %s ON \"%s\".\"%s\" TO '%s'"
	SQL_REVOKE_OBJECT_PRIVILEGE = "REVOKE %s ON \"%s\".\"%s\" FROM '%s'"

	SQL_SELECT_ALL_OBJECTS = "SELECT OBJECT_TYPE, OBJECT_NAME, OWNER FROM DBA_OBJECTS WHERE OBJECT_TYPE IN ('TABLE', 'VIEW', 'PROCEDURE') AND OWNER NOT IN ('SYS', 'oceanbase')"
)

func (t *TenantService) GrantDbPrivilegesWithHost(db *gorm.DB, userName, hostName string, privilege param.DbPrivilegeParam) error {
	if len(privilege.Privileges) > 0 {
		realPrivileges := convertPrivileges(privilege.Privileges)
		sql := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`@`%s`", strings.Join(realPrivileges, ","), privilege.DbName, userName, hostName)
		return db.Exec(sql).Error
	}
	return nil
}

func convertPrivileges(privileges []string) []string {
	realPrivileges := make([]string, 0)
	for _, privilege := range privileges {
		realPrivileges = append(realPrivileges, strings.ReplaceAll(privilege, "_", " "))
	}
	return realPrivileges
}

func (t *TenantService) GrantDbPrivileges(db *gorm.DB, userName string, param *param.DbPrivilegeParam) error {
	if param != nil && len(param.Privileges) > 0 {
		realPrivileges := convertPrivileges(param.Privileges)
		sql := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`", strings.Join(realPrivileges, ", "), param.DbName, userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *TenantService) RevokeDbPrivileges(db *gorm.DB, userName string, param *param.DbPrivilegeParam) error {
	if param != nil && len(param.Privileges) > 0 {
		realPrivileges := convertPrivileges(param.Privileges)
		sql := fmt.Sprintf("REVOKE %s ON %s.* FROM `%s`", strings.Join(realPrivileges, ", "), param.DbName, userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *TenantService) LockUser(db *gorm.DB, userName string) error {
	sql := fmt.Sprintf("ALTER USER '%s' ACCOUNT LOCK", userName)
	return db.Exec(sql).Error
}

func (t *TenantService) GetUserSessionStats(tenantName, userName string) ([]oceanbase.SessionStats, error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, err
	}
	sessionStats := make([]oceanbase.SessionStats, 0)
	result := oceanbaseDb.Table(GV_OB_SESSION).Where("user=? and tenant=?", userName, tenantName).Select("COUNT(*) as COUNT, STATE").Group("STATE").Scan(&sessionStats)
	return sessionStats, result.Error
}

func (t *TenantService) UnlockUser(db *gorm.DB, userName string) error {
	sql := fmt.Sprintf("ALTER USER '%s' ACCOUNT UNLOCK", userName)
	return db.Exec(sql).Error
}

func (t *TenantService) ListObjects(db *gorm.DB) ([]oceanbase.DbaObject, error) {
	var objects []oceanbase.DbaObject
	err := db.Raw(SQL_SELECT_ALL_OBJECTS).Scan(&objects).Error
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (t *TenantService) GetObjectPrivilegesMap(db *gorm.DB) (map[string][]oceanbase.ObjectPrivilege, error) {
	var objectPrivileges []oceanbase.ObjectPrivilege
	result := db.Raw(SQL_SELECT_ALL_OBJECT_PRIVS).Scan(&objectPrivileges)
	objectPrivilegesMap := make(map[string][]oceanbase.ObjectPrivilege)
	for _, objectPriv := range objectPrivileges {
		objectPrivilegesMap[objectPriv.Grantee] = append(objectPrivilegesMap[objectPriv.Grantee], objectPriv)
	}
	return objectPrivilegesMap, result.Error
}

func (t *TenantService) GetGrantedObjectPrivileges(db *gorm.DB, userName string) ([]oceanbase.ObjectPrivilege, error) {
	var objectPrivileges []oceanbase.ObjectPrivilege
	result := db.Raw(SQL_SELECT_OBJECT_PRIVS, userName).Scan(&objectPrivileges)
	return objectPrivileges, result.Error
}

func (t *TenantService) GrantObjectPrivileges(db *gorm.DB, name string, objectOwner, objectName string, privileges []string) error {
	if len(privileges) == 0 {
		return nil
	}
	sql := fmt.Sprintf(SQL_GRANT_OBJECT_PRIVILEGE, strings.Join(privileges, ","), objectOwner, objectName, name)
	return db.Exec(sql).Error
}

func (t *TenantService) RevokeObjectPrivileges(db *gorm.DB, name string, objectOwner, objectName string, privileges []string) error {
	if len(privileges) == 0 {
		return nil
	}
	sql := fmt.Sprintf(SQL_REVOKE_OBJECT_PRIVILEGE, strings.Join(privileges, ","), objectOwner, objectName, name)
	return db.Exec(sql).Error
}

func (t *TenantService) CreateRole(db *gorm.DB, roleName string) error {
	sql := fmt.Sprintf("CREATE ROLE '%s'", roleName)
	return db.Exec(sql).Error
}

func (t *TenantService) IsRoleExist(db *gorm.DB, roleName string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM DBA_ROLES WHERE ROLE = ?", roleName).Scan(&count).Error
	return count == 1, err
}

func (t *TenantService) DropRole(db *gorm.DB, roleName string) error {
	sql := fmt.Sprintf("DROP ROLE '%s'", roleName)
	return db.Exec(sql).Error
}

func (t *TenantService) GetRole(db *gorm.DB, roleName string) (*oceanbase.Role, error) {
	var role oceanbase.Role
	result := db.Raw(SQL_SELECT_ROLE, roleName).Scan(&role)
	return &role, result.Error
}

func (t *TenantService) ListRoles(db *gorm.DB) ([]oceanbase.Role, error) {
	var roles []oceanbase.Role
	result := db.Raw(SQL_SELECT_ALL_ROLE).Scan(&roles)
	return roles, result.Error
}

func (t *TenantService) GrantRoles(db *gorm.DB, name string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	sql := fmt.Sprintf("GRANT '%s' TO '%s'", strings.Join(roles, "','"), name)
	return db.Exec(sql).Error
}

func (t *TenantService) RevokeRoles(db *gorm.DB, name string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	sql := fmt.Sprintf("REVOKE '%s' FROM '%s'", strings.Join(roles, "','"), name)
	return db.Exec(sql).Error
}

// GetGrantedRoleMap get the map of granted roles, the key is the user name, the value is the list of granted roles
func (t *TenantService) GetGrantedRoleMap(db *gorm.DB) (map[string][]string, error) {
	var rolePrivileges []oceanbase.RolePrivilege
	err := db.Raw(SQL_SELECT_ALL_GRANTED_ROLE).Scan(&rolePrivileges).Error
	if err != nil {
		return nil, err
	}
	rolePrivilegesMap := make(map[string][]string)
	for _, rolePriv := range rolePrivileges {
		rolePrivilegesMap[rolePriv.Grantee] = append(rolePrivilegesMap[rolePriv.Grantee], rolePriv.GrantedRole)
	}
	return rolePrivilegesMap, nil
}

func (t *TenantService) GetGrantedRole(db *gorm.DB, roleName string) ([]string, error) {
	var rolePrivileges []oceanbase.RolePrivilege
	result := db.Raw(SQL_SELECT_GRANTED_ROLE, roleName).Scan(&rolePrivileges)
	grantedRoles := make([]string, 0)
	for _, rolePriv := range rolePrivileges {
		grantedRoles = append(grantedRoles, rolePriv.GrantedRole)
	}
	return grantedRoles, result.Error
}

// GetUserGranteesMap get the map of user grantees, the key is the user name, the value is the list of grantees
func (t *TenantService) GetUserGranteesMap(db *gorm.DB) (map[string][]string, error) {
	var userGrantees []oceanbase.RolePrivilege
	err := db.Raw(SQL_SELECT_USER_GRANTEE).Scan(&userGrantees).Error
	if err != nil {
		return nil, err
	}
	userGranteesMap := make(map[string][]string)
	for _, userGrantee := range userGrantees {
		userGranteesMap[userGrantee.GrantedRole] = append(userGranteesMap[userGrantee.GrantedRole], userGrantee.Grantee)
	}
	return userGranteesMap, nil
}

// GetRoleGranteesMap get the map of role grantees, the key is the role name, the value is the list of grantees
func (t *TenantService) GetRoleGranteesMap(db *gorm.DB) (map[string][]string, error) {
	var roleGrantees []oceanbase.RolePrivilege
	err := db.Raw(SQL_SELECT_ROLE_GRANTEE).Scan(&roleGrantees).Error
	if err != nil {
		return nil, err
	}
	roleGranteesMap := make(map[string][]string)
	for _, roleGrantee := range roleGrantees {
		roleGranteesMap[roleGrantee.GrantedRole] = append(roleGranteesMap[roleGrantee.GrantedRole], roleGrantee.Grantee)
	}
	return roleGranteesMap, nil
}
