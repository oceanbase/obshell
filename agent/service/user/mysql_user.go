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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"gorm.io/gorm"
)

type mysqlUserService struct {
	db *gorm.DB
}

func NewMysqlUserService(db *gorm.DB) *mysqlUserService {
	return &mysqlUserService{db: db}
}

func (t *mysqlUserService) GetDb() *gorm.DB {
	return t.db
}

func (t *mysqlUserService) IsUserExist(userName string) (bool, error) {
	var count int64
	err := t.db.Table(DBA_OB_USERS).Where("USER_NAME = ?", userName).Count(&count).Error
	return count == 1, err
}

func (t *mysqlUserService) CreateUser(userName, password, hostName string) error {
	sql := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s`@`%s` IDENTIFIED BY '%s'", userName, hostName, strings.ReplaceAll(password, "'", "'\"'\"'"))
	return t.db.Exec(sql).Error
}

func convertPrivileges(privileges []string) []string {
	realPrivileges := make([]string, 0)
	for _, privilege := range privileges {
		if strings.ToUpper(privilege) == constant.OB_ORACLE_PRIVILEGE_PURGE_DBA_RECYCLEBIN {
			realPrivileges = append(realPrivileges, "PURGE DBA_RECYCLEBIN")
			continue
		}
		realPrivileges = append(realPrivileges, strings.ReplaceAll(privilege, "_", " "))
	}
	return realPrivileges
}

func (t *mysqlUserService) GrantGlobalPrivileges(userName string, privileges []string) error {
	realPrivileges := convertPrivileges(privileges)
	if len(realPrivileges) > 0 {
		sql := fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(realPrivileges, ", "), userName)
		return t.db.Exec(sql).Error
	}
	return nil
}

func extractGlobalPrivileges(user *oceanbase.MysqlUser) []string {
	globalPrivileges := make([]string, 0)
	if strings.HasPrefix(strings.ToUpper(user.AlterPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_ALTER)
	}
	if strings.HasPrefix(strings.ToUpper(user.CreatePriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_CREATE)
	}
	if strings.HasPrefix(strings.ToUpper(user.DeletePriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_DELETE)
	}
	if strings.HasPrefix(strings.ToUpper(user.DropPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_DROP)
	}
	if strings.HasPrefix(strings.ToUpper(user.InsertPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_INSERT)
	}
	if strings.HasPrefix(strings.ToUpper(user.SelectPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_SELECT)
	}
	if strings.HasPrefix(strings.ToUpper(user.UpdatePriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_UPDATE)
	}
	if strings.HasPrefix(strings.ToUpper(user.IndexPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_INDEX)
	}
	if strings.HasPrefix(strings.ToUpper(user.CreateViewPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_CREATE_VIEW)
	}
	if strings.HasPrefix(strings.ToUpper(user.ShowViewPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_SHOW_VIEW)
	}
	if strings.HasPrefix(strings.ToUpper(user.CreateUserPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_CREATE_USER)
	}
	if strings.HasPrefix(strings.ToUpper(user.ProcessPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_PROCESS)
	}
	if strings.HasPrefix(strings.ToUpper(user.SuperPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_SUPER)
	}
	if strings.HasPrefix(strings.ToUpper(user.ShowDbPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_SHOW_DATABASES)
	}
	if strings.HasPrefix(strings.ToUpper(user.GrantPriv), "Y") {
		globalPrivileges = append(globalPrivileges, constant.OB_MYSQL_PRIVILEGE_GRANT_OPTION)
	}
	return globalPrivileges
}

func (t *mysqlUserService) GetGrantedGlobalPrivileges(userName string) ([]string, error) {
	user, err := t.GetUser(userName)
	if err != nil {
		return nil, err
	}
	mysqlUser, ok := user.(oceanbase.MysqlUser)
	if !ok {
		return nil, nil
	}
	return extractGlobalPrivileges(&mysqlUser), nil
}

func (t *mysqlUserService) GetGlobalPrivilegesMap() (map[string][]string, error) {
	users, err := t.ListUsers()
	if err != nil {
		return nil, err
	}
	globalPrivilegesMap := make(map[string][]string)
	for _, user := range users {
		globalPrivileges := extractGlobalPrivileges(user.(*oceanbase.MysqlUser))
		globalPrivilegesMap[user.ToUserBo().UserName] = globalPrivileges
	}
	return globalPrivilegesMap, nil
}

func (t *mysqlUserService) RevokeGlobalPrivileges(userName string, privileges []string) error {
	realPrivileges := convertPrivileges(privileges)
	if len(realPrivileges) > 0 {
		sql := fmt.Sprintf("REVOKE %s ON *.* FROM `%s`", strings.Join(realPrivileges, ", "), userName)
		return t.db.Exec(sql).Error
	}
	return nil
}

func (t *mysqlUserService) DropUser(userName string) error {
	sql := fmt.Sprintf("DROP USER `%s`", userName)
	return t.db.Exec(sql).Error
}

func (t *mysqlUserService) GetUserSessionStats(db *gorm.DB, userName string) ([]oceanbase.SessionStats, error) {
	sessionStats := make([]oceanbase.SessionStats, 0)
	result := db.Table(GV_OB_SESSION).Where("user=?", userName).Select("COUNT(*) as COUNT, STATE").Group("STATE").Scan(&sessionStats)
	return sessionStats, result.Error
}

func (t *mysqlUserService) ChangeUserPassword(userName, password string) error {
	sql := fmt.Sprintf("ALTER USER `%s` IDENTIFIED BY \"%s\"", userName, strings.ReplaceAll(password, "\"", "\\\""))
	return t.db.Exec(sql).Error
}

func (t *mysqlUserService) ListUsers() ([]oceanbase.ObUser, error) {
	users := make([]oceanbase.MysqlUser, 0)
	result := t.db.Table(MYSQL_USER).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	obUsers := make([]oceanbase.ObUser, len(users))
	for i := range users {
		obUsers[i] = &users[i]
	}
	return obUsers, nil
}

func (t *mysqlUserService) GetUser(userName string) (oceanbase.ObUser, error) {
	var user oceanbase.MysqlUser
	result := t.db.Table(MYSQL_USER).First(&user, "user=?", userName)
	return user, result.Error
}

func (t *mysqlUserService) ModifyTenantRootPassword(newPwd string) error {
	if err := t.db.Exec(fmt.Sprintf(SQL_MYSQL_ALTER_TENANT_ROOT_PASSWORD, strings.ReplaceAll(newPwd, "\"", "\\\""))).Error; err != nil {
		return errors.Wrap(err, "modify tenant root password failed")
	}
	return nil
}
