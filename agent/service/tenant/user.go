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

	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
	"gorm.io/gorm"
)

func (t *TenantService) IsUserExist(db *gorm.DB, userName string) (bool, error) {
	var count int64
	err := db.Table(DBA_OB_USERS).Where("USER_NAME = ?", userName).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) CreateUser(db *gorm.DB, userName, password, hostName string) error {
	sql := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s`@`%s` IDENTIFIED BY '%s'", userName, hostName, strings.ReplaceAll(password, "'", "'\"'\"'"))
	return db.Exec(sql).Error
}

func (t *TenantService) GrantGlobalPrivilegesWithHost(db *gorm.DB, userName, hostName string, privileges []string) error {
	realPrivileges := convertPrivileges(privileges)
	sql := fmt.Sprintf("GRANT %s ON *.* TO `%s`@`%s`", strings.Join(realPrivileges, ","), userName, hostName)
	return db.Exec(sql).Error
}

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

func (t *TenantService) GrantGlobalPrivileges(db *gorm.DB, userName string, privileges []string) error {
	realPrivileges := convertPrivileges(privileges)
	if len(realPrivileges) > 0 {
		sql := fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(realPrivileges, ", "), userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *TenantService) GrantDbPrivileges(db *gorm.DB, userName string, param *param.DbPrivilegeParam) error {
	if param != nil && len(param.Privileges) > 0 {
		realPrivileges := convertPrivileges(param.Privileges)
		sql := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`", strings.Join(realPrivileges, ", "), param.DbName, userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *TenantService) RevokeGlobalPrivileges(db *gorm.DB, userName string, privileges []string) error {
	realPrivileges := convertPrivileges(privileges)
	if len(realPrivileges) > 0 {
		sql := fmt.Sprintf("REVOKE %s ON *.* FROM `%s`", strings.Join(realPrivileges, ", "), userName)
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

func (t *TenantService) DropUser(db *gorm.DB, userName string) error {
	sql := fmt.Sprintf("DROP USER `%s`", userName)
	return db.Exec(sql).Error
}

func (t *TenantService) LockUser(db *gorm.DB, userName string) error {
	sql := fmt.Sprintf("ALTER USER `%s` ACCOUNT LOCK", userName)
	return db.Exec(sql).Error
}

func (t *TenantService) GetUserSessionStats(db *gorm.DB, userName string) ([]oceanbase.SessionStats, error) {
	sessionStats := make([]oceanbase.SessionStats, 0)
	result := db.Table(GV_OB_SESSION).Where("user=?", userName).Select("COUNT(*) as COUNT, STATE").Group("STATE").Scan(&sessionStats)
	return sessionStats, result.Error
}

func (t *TenantService) ChangeUserPassword(db *gorm.DB, userName, password string) error {
	sql := fmt.Sprintf("ALTER USER `%s` IDENTIFIED BY '%s'", userName, password)
	return db.Exec(sql).Error
}

func (t *TenantService) UnlockUser(db *gorm.DB, userName string) error {
	sql := fmt.Sprintf("ALTER USER `%s` ACCOUNT UNLOCK", userName)
	return db.Exec(sql).Error
}

func (t *TenantService) ListUsers(db *gorm.DB) ([]oceanbase.MysqlUser, error) {
	users := make([]oceanbase.MysqlUser, 0)
	result := db.Table(MYSQL_USER).Find(&users)
	return users, result.Error
}

func (t *TenantService) GetUser(db *gorm.DB, userName string) (*oceanbase.MysqlUser, error) {
	var user oceanbase.MysqlUser
	result := db.Table(MYSQL_USER).First(&user, "user=?", userName)
	return &user, result.Error
}
