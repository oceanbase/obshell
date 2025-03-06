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

func (t *TenantService) GrantGlobalPrivileges(db *gorm.DB, userName, hostName string, privilege []string) error {
	sql := fmt.Sprintf("GRANT %s ON *.* TO `%s`@`%s`", strings.Join(privilege, ","), userName, hostName)
	return db.Exec(sql).Error
}

func (t *TenantService) GrantDbPrivileges(db *gorm.DB, userName, hostName string, privilege param.DbPrivilegeParam) error {
	sql := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`@`%s`", strings.Join(privilege.Privileges, ","), privilege.DbName, userName, hostName)
	return db.Exec(sql).Error
}

func (t *TenantService) DropUser(db *gorm.DB, userName string) error {
	sql := fmt.Sprintf("DROP USER `%s`", userName)
	return db.Exec(sql).Error
}
