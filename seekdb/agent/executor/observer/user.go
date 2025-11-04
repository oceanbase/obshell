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

package observer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
	obmodel "github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
	"github.com/oceanbase/obshell/seekdb/param"
	"github.com/oceanbase/obshell/seekdb/utils"
)

func CreateUser(param *param.CreateUserParam) error {
	// check the user name is valid
	if !regexp.MustCompile(constant.USERNAME_PATTERN).MatchString(param.UserName) {
		return errors.Occur(errors.ErrObUserNameInvalid, param.UserName)
	}

	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}

	if param.HostName == "" {
		param.HostName = constant.DEFAULT_HOST
	}
	if err := userService.CreateUser(param.UserName, param.Password, param.HostName); err != nil {
		return errors.Wrapf(err, "create user '%s' failed", param.UserName)
	}

	// Grant privileges.
	if len(param.GlobalPrivileges) != 0 {
		if err := userService.GrantGlobalPrivileges(param.UserName, param.GlobalPrivileges); err != nil {
			return errors.Wrapf(err, "grant global privileges to user '%s' failed", param.UserName)
		}
	}

	for _, dbPrivilege := range param.DbPrivileges {
		if err := tenantService.GrantDbPrivilegesWithHost(db, param.UserName, param.HostName, dbPrivilege); err != nil {
			return errors.Wrapf(err, "grant db privileges to user '%s' failed", param.UserName)
		}
	}

	return nil
}

func DropUser(userName string) error {
	// Check user exist.
	if exist, err := userService.IsUserExist(userName); err != nil {
		return errors.Wrapf(err, "check user '%s' exist failed", userName)
	} else if !exist {
		return nil
	}

	// Drop user.
	if err := userService.DropUser(userName); err != nil {
		return errors.Wrapf(err, "drop user '%s' failed", userName)
	}

	return nil
}

func ListUsers(queryParam *param.ListUsersQueryParam) ([]bo.ObUser, error) {
	users, err := userService.ListUsers()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query user list")
	}
	filteredUsers := make([]obmodel.ObUser, 0)
	for _, user := range users {
		if !utils.ContainsString(constant.OB_EXCLUDED_USERS, user.Name()) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	users = filteredUsers

	globalPrivilegeMap, err := userService.GetGlobalPrivilegesMap()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query global privileges")
	}

	dbPrivilegeMap := make(map[string][]obmodel.MysqlDb)
	allDatabases := make([]obmodel.Database, 0)
	allDatabases, err = tenantService.ListDatabases()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query database list")
	}

	dbPrivileges, err := tenantService.ListDatabasePrivileges()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query database privileges")
	}
	for _, dbPrivilege := range dbPrivileges {
		privileges, ok := dbPrivilegeMap[dbPrivilege.User]
		if !ok {
			privileges = make([]obmodel.MysqlDb, 0)
		}
		privileges = append(privileges, dbPrivilege)
		dbPrivilegeMap[dbPrivilege.User] = privileges
	}

	result := make([]bo.ObUser, 0)
	for _, user := range users {
		// filter out inner users
		obUser := user.ToUserBo()
		isInnerUser := false
		for _, innerUserName := range constant.OB_INNER_USERS {
			if obUser.UserName == innerUserName {
				isInnerUser = true
				break
			}
		}
		if isInnerUser {
			continue
		}
		privileges, ok := dbPrivilegeMap[obUser.UserName]
		if !ok {
			privileges = make([]obmodel.MysqlDb, 0)
		}
		attachMysqlUserInfo(&obUser, globalPrivilegeMap[obUser.UserName], privileges, allDatabases)
		result = append(result, obUser)
	}

	sort.Slice(result, func(i, j int) bool {
		if queryParam.SortBy == "create_time" {
			if queryParam.SortOrder == "asc" {
				return result[i].CreateTime.Before(result[j].CreateTime)
			} else {
				return result[i].CreateTime.After(result[j].CreateTime)
			}
		} else {
			return result[i].UserName < result[j].UserName
		}
	})

	return result, nil
}

func ChangeUserPassword(userName string, p *param.ChangeUserPasswordParam) error {
	// check current password is correct
	dsConfig := config.NewObMysqlDataSourceConfig().SetPassword(meta.GetOceanbasePwd()).SetParseTime(true)
	if userName == "root" {
		if err := oceanbasedb.LoadOceanbaseInstanceForTest(dsConfig); err != nil {
			return errors.Wrapf(err, "Failed to check if current root password is correct")
		}
	}
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists", userName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}
	err = userService.ChangeUserPassword(userName, p.Password)
	if err != nil {
		return errors.Wrapf(err, "Failed to change password of user %s", userName)
	}
	if userName == "root" {
		if err := secure.VerifyOceanbasePassword(p.Password); err != nil {
			return err
		}
	}
	return nil
}

func GetUserStats(userName string) (*bo.ObUserStats, error) {
	sessionStats, err := tenantService.GetUserSessionStats(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query session stat of user %s", userName)
	}
	var totalCount int64 = 0
	var activeCount int64 = 0
	for _, sessionStat := range sessionStats {
		totalCount += sessionStat.Count
		if strings.ToUpper(sessionStat.State) == "ACTIVE" {
			activeCount += sessionStat.Count
		}
	}
	return &bo.ObUserStats{
		Session: &bo.ObUserSessionStats{
			Total:  totalCount,
			Active: activeCount,
		},
	}, nil
}

func verifyMysqlPrivilege(privileges []string) error {
	for _, privilege := range privileges {
		if !utils.ContainsString(constant.OB_MYSQL_PRIVILEGES, strings.ToUpper(privilege)) {
			return errors.Occur(errors.ErrObUserPrivilegeNotSupported, privilege)
		}
	}
	return nil
}

func ModifyUserGlobalPrivilege(userName string, p *param.ModifyUserGlobalPrivilegeParam) error {
	err := verifyMysqlPrivilege(p.GlobalPrivileges)
	if err != nil {
		return err
	}

	// check if user exist
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists", userName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	grantedGlobalPrivileges, err := userService.GetGrantedGlobalPrivileges(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to get granted global privileges of user %s", userName)
	}

	privilegesToGrant, privilegesToRevoke := utils.Difference(p.GlobalPrivileges, grantedGlobalPrivileges)
	err = userService.GrantGlobalPrivileges(userName, privilegesToGrant)
	if err != nil {
		return errors.Wrapf(err, "Failed to grant privilege to user %s", userName)
	}
	err = userService.RevokeGlobalPrivileges(userName, privilegesToRevoke)
	if err != nil {
		return errors.Wrapf(err, "Failed to revoke privilege from user %s", userName)
	}

	return nil
}

func ModifyUserDbPrivilege(userName string, p *param.ModifyUserDbPrivilegeParam) error {
	for _, dbPrivilege := range p.DbPrivileges {
		err := verifyMysqlPrivilege(dbPrivilege.Privileges)
		if err != nil {
			return err
		}
	}
	obuser, userErr := GetUser(userName)
	if userErr != nil {
		return userErr
	}

	for _, desiredDbPrivilege := range p.DbPrivileges {
		found := false
		for _, currentDbPrivilege := range obuser.DbPrivileges {
			if desiredDbPrivilege.DbName == currentDbPrivilege.DbName {
				privilegesToGrant, privilegesToRevoke := utils.Difference(desiredDbPrivilege.Privileges, currentDbPrivilege.Privileges)
				err := tenantService.GrantDbPrivileges(userName, &param.DbPrivilegeParam{
					DbName:     currentDbPrivilege.DbName,
					Privileges: privilegesToGrant,
				})
				if err != nil {
					return errors.Wrapf(err, "Failed to grant privilege of database %s to user %s", currentDbPrivilege.DbName, userName)
				}
				err = tenantService.RevokeDbPrivileges(userName, &param.DbPrivilegeParam{
					DbName:     currentDbPrivilege.DbName,
					Privileges: privilegesToRevoke,
				})
				if err != nil {
					return errors.Wrapf(err, "Failed to revoke privilege of database %s from user %s", currentDbPrivilege.DbName, userName)
				}
			}
		}
		if !found {
			err := tenantService.GrantDbPrivileges(userName, &desiredDbPrivilege)
			if err != nil {
				return errors.Wrapf(err, "Failed to grant privilege of database %s to user %s", desiredDbPrivilege.DbName, userName)
			}
		}
	}

	for _, currentDbPrivilege := range obuser.DbPrivileges {
		found := false
		for _, desiredDbPrivilege := range p.DbPrivileges {
			if desiredDbPrivilege.DbName == currentDbPrivilege.DbName {
				found = true
				break
			}
		}
		if !found {
			err := tenantService.RevokeDbPrivileges(userName, &param.DbPrivilegeParam{
				DbName:     currentDbPrivilege.DbName,
				Privileges: currentDbPrivilege.Privileges,
			})
			if err != nil {
				return errors.Wrapf(err, "Failed to revoke privilege of database %s from user %s", currentDbPrivilege.DbName, userName)
			}
		}
	}
	return nil
}

func LockUser(userName string) error {
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists", userName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}
	err = tenantService.LockUser(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to lock user %s", userName)
	}
	return nil
}

func UnlockUser(userName string) error {
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists", userName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}
	err = tenantService.UnlockUser(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to unlock user %s", userName)
	}
	return nil
}

func GetUser(userName string) (*bo.ObUser, error) {
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to check if user %s exists", userName)
	}
	if !exist {
		return nil, errors.Occur(errors.ErrObUserNotExists, userName)
	}
	user, err := userService.GetUser(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query user %s", userName)
	}
	obUser := user.ToUserBo()

	globalPrivileges, err := userService.GetGrantedGlobalPrivileges(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query global privileges of user %s", userName)
	}

	allDatabases, err := tenantService.ListDatabases()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query database list of user %s", userName)
	}
	dbPrivileges, err := tenantService.ListDatabasePrivilegesOfUser(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query database privileges of user %s", userName)
	}
	attachMysqlUserInfo(&obUser, globalPrivileges, dbPrivileges, allDatabases)

	return &obUser, nil
}

func attachMysqlUserInfo(obUser *bo.ObUser, globalPrivileges []string, dbPrivileges []obmodel.MysqlDb, allDatabases []obmodel.Database) {
	obUser.GlobalPrivileges = globalPrivileges
	databasePrivileges := extractDatabasePrivileges(dbPrivileges)
	obUser.DbPrivileges = databasePrivileges
	allDatabaseNames := make([]string, 0)
	privilegedDatabaseNames := make([]string, 0)
	for _, database := range allDatabases {
		allDatabaseNames = append(allDatabaseNames, database.Name)
	}
	for _, databasePrivilege := range databasePrivileges {
		privilegedDatabaseNames = append(privilegedDatabaseNames, databasePrivilege.DbName)
	}
	if len(globalPrivileges) > 0 {
		obUser.AccessibleDatabases = allDatabaseNames
	} else {
		obUser.AccessibleDatabases = privilegedDatabaseNames
	}
	attachUserConnectInfo(obUser)
}

func attachUserConnectInfo(obUser *bo.ObUser) {
	connectionStr := bo.ObproxyAndConnectionString{
		Type:             constant.OB_CONNECTION_TYPE_DIRECT,
		ConnectionString: fmt.Sprintf("obclient -h%s -P%d -u%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, obUser.UserName),
	}

	connectionStrs := make([]bo.ObproxyAndConnectionString, 0)
	connectionStrs = append(connectionStrs, connectionStr)
	obUser.ConnectionStrings = connectionStrs
}

func extractDatabasePrivileges(dbPrivileges []obmodel.MysqlDb) []bo.DbPrivilege {
	result := make([]bo.DbPrivilege, 0)
	for _, dbPrivilege := range dbPrivileges {
		privileges := make([]string, 0)
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.AlterPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_ALTER)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.CreatePriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_CREATE)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.DeletePriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_DELETE)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.DropPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_DROP)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.InsertPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_INSERT)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.SelectPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_SELECT)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.UpdatePriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_UPDATE)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.IndexPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_INDEX)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.CreateViewPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_CREATE_VIEW)
		}
		if strings.HasPrefix(strings.ToUpper(dbPrivilege.ShowViewPriv), "Y") {
			privileges = append(privileges, constant.OB_MYSQL_PRIVILEGE_SHOW_VIEW)
		}
		if len(privileges) > 0 {
			result = append(result, bo.DbPrivilege{
				DbName:     dbPrivilege.Db,
				Privileges: privileges,
			})
		}
	}
	return result
}
