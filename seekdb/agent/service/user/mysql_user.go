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

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
)

type UserService struct {
}

func (t *UserService) IsUserExist(userName string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(DBA_OB_USERS).Where("USER_NAME = ?", userName).Count(&count).Error
	return count == 1, err
}

func (t *UserService) CreateUser(userName, password, hostName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s`@`%s` IDENTIFIED BY '%s'", userName, hostName, strings.ReplaceAll(password, "'", "'\"'\"'"))
	return db.Exec(sql).Error
}

func convertPrivileges(privileges []string) []string {
	realPrivileges := make([]string, 0)
	for _, privilege := range privileges {
		realPrivileges = append(realPrivileges, strings.ReplaceAll(privilege, "_", " "))
	}
	return realPrivileges
}

func (t *UserService) GrantGlobalPrivileges(userName string, privileges []string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	realPrivileges := convertPrivileges(privileges)
	if len(realPrivileges) > 0 {
		sql := fmt.Sprintf("GRANT %s ON *.* TO `%s`", strings.Join(realPrivileges, ", "), userName)
		return db.Exec(sql).Error
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

func (t *UserService) GetGrantedGlobalPrivileges(userName string) ([]string, error) {
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

func (t *UserService) GetGlobalPrivilegesMap() (map[string][]string, error) {
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

func (t *UserService) RevokeGlobalPrivileges(userName string, privileges []string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	realPrivileges := convertPrivileges(privileges)
	if len(realPrivileges) > 0 {
		sql := fmt.Sprintf("REVOKE %s ON *.* FROM `%s`", strings.Join(realPrivileges, ", "), userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *UserService) DropUser(userName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP USER `%s`", userName)
	return db.Exec(sql).Error
}

func (t *UserService) GetUserSessionStats(userName string) ([]oceanbase.SessionStats, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sessionStats := make([]oceanbase.SessionStats, 0)
	result := oceanbaseDb.Table(GV_OB_SESSION).Where("user=?", userName).Select("COUNT(*) as COUNT, STATE").Group("STATE").Scan(&sessionStats)
	return sessionStats, result.Error
}

func (t *UserService) ChangeUserPassword(userName, password string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER USER `%s` IDENTIFIED BY \"%s\"", userName, strings.ReplaceAll(password, "\"", "\\\""))
	return db.Exec(sql).Error
}

func (t *UserService) ListUsers() ([]oceanbase.ObUser, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	users := make([]oceanbase.MysqlUser, 0)
	result := oceanbaseDb.Table(MYSQL_USER).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	obUsers := make([]oceanbase.ObUser, len(users))
	for i := range users {
		obUsers[i] = &users[i]
	}
	return obUsers, nil
}

func (t *UserService) GetUser(userName string) (oceanbase.ObUser, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var user oceanbase.MysqlUser
	result := oceanbaseDb.Table(MYSQL_USER).First(&user, "user=?", userName)
	return user, result.Error
}

func (t *UserService) GetUserCount() (int, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}
	var count int
	result := db.Raw("SELECT COUNT(*) FROM oceanbase.DBA_OB_USERS").Scan(&count)
	return count, result.Error
}
