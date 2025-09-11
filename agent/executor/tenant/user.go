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
	"regexp"
	"sort"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	obmodel "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/service/user"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
	"gorm.io/gorm"
)

func CreateUser(tenantName string, param *param.CreateUserParam) error {
	// check the user name is valid
	if !regexp.MustCompile(constant.USERNAME_PATTERN).MatchString(param.UserName) {
		return errors.Occur(errors.ErrObUserNameInvalid, param.UserName)
	}

	tenantInfo, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return errors.Wrapf(err, "get tenant '%s' info failed", tenantName)
	}

	var db *gorm.DB
	defer CloseDbConnection(db)
	db, err = GetConnectionWithTenantInfo(tenantInfo, param.RootPassword)
	if err != nil {
		return errors.Wrapf(err, "Failed to get connection of tenant %s", tenantName)
	}

	if param.HostName == "" {
		param.HostName = constant.DEFAULT_HOST
	}
	userService := user.GetUserService(db)
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
		if tenantInfo.Mode != constant.MYSQL_MODE {
			return errors.Occur(errors.ErrObUserOracleModeNotSupport)
		}
		if err := tenantService.GrantDbPrivilegesWithHost(db, param.UserName, param.HostName, dbPrivilege); err != nil {
			return errors.Wrapf(err, "grant db privileges to user '%s' failed", param.UserName)
		}
	}

	if len(param.Roles) != 0 {
		if tenantInfo.Mode != constant.ORACLE_MODE {
			return errors.Occur(errors.ErrObUserOracleModeNotSupport)
		}
		if err := tenantService.GrantRoles(db, param.UserName, param.Roles); err != nil {
			return errors.Wrapf(err, "grant roles to user '%s' failed", param.UserName)
		}
	}

	return nil
}

func DropUser(tenantName, userName string, param *param.DropUserParam) error {
	db, err := GetConnectionWithPassword(tenantName, param.RootPassword)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	defer CloseDbConnection(db)
	userService := user.GetUserService(db)

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

func ListUsers(tenantName string, password *string, queryParam *param.ListUsersQueryParam) ([]bo.ObUser, error) {
	tenantInfo, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, errors.Wrapf(err, "get tenant '%s' mode failed", tenantName)
	}

	db, err := GetConnectionWithTenantInfo(tenantInfo, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}

	userService := user.GetUserService(db)
	users, err := userService.ListUsers()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query user list of tenant %s", tenantName)
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
		return nil, errors.Wrapf(err, "Failed to query global privileges of tenant %s", tenantName)
	}

	dbPrivilegeMap := make(map[string][]obmodel.MysqlDb)
	allDatabases := make([]obmodel.Database, 0)
	if tenantInfo.Mode == constant.MYSQL_MODE {
		allDatabases, err = tenantService.ListDatabases(db)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query database list of tenant %s", tenantName)
		}

		dbPrivileges, err := tenantService.ListDatabasePrivileges(db)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query database privileges of tenant %s", tenantName)
		}
		for _, dbPrivilege := range dbPrivileges {
			privileges, ok := dbPrivilegeMap[dbPrivilege.User]
			if !ok {
				privileges = make([]obmodel.MysqlDb, 0)
			}
			privileges = append(privileges, dbPrivilege)
			dbPrivilegeMap[dbPrivilege.User] = privileges
		}
	}

	rolePrivMap := make(map[string][]string)
	objectPrivMap := make(map[string][]obmodel.ObjectPrivilege)
	if tenantInfo.Mode == constant.ORACLE_MODE {
		grantedRoleMap, err := tenantService.GetGrantedRoleMap(db)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query role privileges of tenant %s", tenantName)
		}
		rolePrivMap = grantedRoleMap

		objectPrivileges, err := tenantService.GetObjectPrivilegesMap(db)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query object privileges of tenant %s", tenantName)
		}
		objectPrivMap = objectPrivileges
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
		// for mysql user
		if tenantInfo.Mode == constant.MYSQL_MODE {
			attachMysqlUserInfo(tenantName, &obUser, globalPrivilegeMap[obUser.UserName], privileges, allDatabases)
		} else if tenantInfo.Mode == constant.ORACLE_MODE {
			attachOracleUserInfo(tenantName, &obUser, globalPrivilegeMap[obUser.UserName], rolePrivMap[obUser.UserName], objectPrivMap[obUser.UserName])
		}
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

func ChangeUserPassword(tenantName, userName string, p *param.ChangeUserPasswordParam) error {
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}
	err = userService.ChangeUserPassword(userName, p.NewPassword)
	if err != nil {
		return errors.Wrapf(err, "Failed to change password of user %s of tenant %s", userName, tenantName)
	}
	return nil
}

func GetUserStats(tenantName, userName string) (*bo.ObUserStats, error) {
	sessionStats, err := tenantService.GetUserSessionStats(tenantName, userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query session stat of user %s of tenant %s", userName, tenantName)
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

func verifyOraclePrivilege(privileges []string) error {
	for _, privilege := range privileges {
		if !utils.ContainsString(constant.OB_ORACLE_PRIVILEGES, strings.ToUpper(privilege)) {
			return errors.Occur(errors.ErrObUserPrivilegeNotSupported, privilege)
		}
	}
	return nil
}

func ModifyUserGlobalPrivilege(tenantName, userName string, p *param.ModifyUserGlobalPrivilegeParam) error {
	tenantInfo, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return errors.Wrapf(err, "get tenant '%s' info failed", tenantName)
	}
	if tenantInfo.Mode == constant.MYSQL_MODE {
		err = verifyMysqlPrivilege(p.GlobalPrivileges)
	} else if tenantInfo.Mode == constant.ORACLE_MODE {
		err = verifyOraclePrivilege(p.GlobalPrivileges)
	}
	if err != nil {
		return err
	}

	db, err := GetConnectionWithTenantInfo(tenantInfo, p.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}

	userService := user.GetUserService(db)
	// check if user exist
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	grantedGlobalPrivileges, err := userService.GetGrantedGlobalPrivileges(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to get granted global privileges of user %s of tenant %s", userName, tenantName)
	}

	privilegesToGrant, privilegesToRevoke := utils.Difference(p.GlobalPrivileges, grantedGlobalPrivileges)
	err = userService.GrantGlobalPrivileges(userName, privilegesToGrant)
	if err != nil {
		return errors.Wrapf(err, "Failed to grant privilege to user %s of tenant %s", userName, tenantName)
	}
	err = userService.RevokeGlobalPrivileges(userName, privilegesToRevoke)
	if err != nil {
		return errors.Wrapf(err, "Failed to revoke privilege from user %s of tenant %s", userName, tenantName)
	}

	return nil
}

func ModifyUserDbPrivilege(tenantName, userName string, p *param.ModifyUserDbPrivilegeParam) error {
	for _, dbPrivilege := range p.DbPrivileges {
		err := verifyMysqlPrivilege(dbPrivilege.Privileges)
		if err != nil {
			return err
		}
	}
	obuser, userErr := GetUser(tenantName, userName, p.RootPassword)
	if userErr != nil {
		return userErr
	}
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}

	for _, desiredDbPrivilege := range p.DbPrivileges {
		found := false
		for _, currentDbPrivilege := range obuser.DbPrivileges {
			if desiredDbPrivilege.DbName == currentDbPrivilege.DbName {
				privilegesToGrant, privilegesToRevoke := utils.Difference(desiredDbPrivilege.Privileges, currentDbPrivilege.Privileges)
				err = tenantService.GrantDbPrivileges(db, userName, &param.DbPrivilegeParam{
					DbName:     currentDbPrivilege.DbName,
					Privileges: privilegesToGrant,
				})
				if err != nil {
					return errors.Wrapf(err, "Failed to grant privilege of database %s to user %s of tenant %s", currentDbPrivilege.DbName, userName, tenantName)
				}
				err = tenantService.RevokeDbPrivileges(db, userName, &param.DbPrivilegeParam{
					DbName:     currentDbPrivilege.DbName,
					Privileges: privilegesToRevoke,
				})
				if err != nil {
					return errors.Wrapf(err, "Failed to revoke privilege of database %s from user %s of tenant %s", currentDbPrivilege.DbName, userName, tenantName)
				}
			}
		}
		if !found {
			err = tenantService.GrantDbPrivileges(db, userName, &desiredDbPrivilege)
			if err != nil {
				return errors.Wrapf(err, "Failed to grant privilege of database %s to user %s of tenant %s", desiredDbPrivilege.DbName, userName, tenantName)
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
			err = tenantService.RevokeDbPrivileges(db, userName, &param.DbPrivilegeParam{
				DbName:     currentDbPrivilege.DbName,
				Privileges: currentDbPrivilege.Privileges,
			})
			if err != nil {
				return errors.Wrapf(err, "Failed to revoke privilege of database %s from user %s of tenant %s", currentDbPrivilege.DbName, userName, tenantName)
			}
		}
	}
	return nil
}

func ModifyUserObjectPrivilege(tenantName, userName string, p *param.ModifyObjectPrivilegeParam) error {
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	if err != nil {
		return err
	}
	defer CloseDbConnection(db)

	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	grantedObjectPrivileges, err := tenantService.GetGrantedObjectPrivileges(db, userName)
	if err != nil {
		return err
	}
	aggGrantedObjectPrivileges := aggregateObjectPrivileges(grantedObjectPrivileges)
	if err := putObjectPrivilege(db, tenantName, userName, aggGrantedObjectPrivileges, p.ObjectPrivileges); err != nil {
		return err
	}

	return nil
}

func PatchUserObjectPrivilege(tenantName, userName string, p *param.ModifyObjectPrivilegeParam) error {
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	if err != nil {
		return err
	}
	defer CloseDbConnection(db)

	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	grantedObjectPrivileges, err := tenantService.GetGrantedObjectPrivileges(db, userName)
	if err != nil {
		return err
	}
	aggGrantedObjectPrivileges := aggregateObjectPrivileges(grantedObjectPrivileges)
	if err := patchObjectPrivilege(db, tenantName, userName, aggGrantedObjectPrivileges, p.ObjectPrivileges); err != nil {
		return err
	}

	return nil
}

func RevokeUserObjectPrivilege(tenantName, userName string, p *param.RevokeObjectPrivilegeParam) error {
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	if err != nil {
		return err
	}
	defer CloseDbConnection(db)
	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	for _, objectPrivilege := range p.ObjectPrivileges {
		err = tenantService.RevokeObjectPrivileges(db, userName, objectPrivilege.Owner, objectPrivilege.ObjectName, objectPrivilege.Privileges)
		if err != nil {
			return errors.Wrapf(err, "Failed to revoke object privileges from user %s of tenant %s", userName, tenantName)
		}
	}

	return nil
}

func GrantUserObjectPrivilege(tenantName, userName string, p *param.GrantObjectPrivilegeParam) error {
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	if err != nil {
		return err
	}
	defer CloseDbConnection(db)
	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	for _, objectPrivilege := range p.ObjectPrivileges {
		err = tenantService.GrantObjectPrivileges(db, userName, objectPrivilege.Owner, objectPrivilege.ObjectName, objectPrivilege.Privileges)
		if err != nil {
			return errors.Wrapf(err, "Failed to grant object privileges to user %s of tenant %s", userName, tenantName)
		}
	}

	return nil
}

func ModifyUserRole(tenantName, userName string, p *param.ModifyRoleParam) error {
	db, err := getOracleTenantConnection(tenantName, p.RootPassword)
	if err != nil {
		return err
	}
	defer CloseDbConnection(db)

	userService := user.GetUserService(db)
	// check if user exist
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s of tenant %s exists", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}

	grantedRoles, err := tenantService.GetGrantedRole(db, userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to get granted roles of user %s of tenant %s", userName, tenantName)
	}

	toGrantRoles, toRevokeRoles := utils.Difference(p.Roles, grantedRoles)

	if err := tenantService.GrantRoles(db, userName, toGrantRoles); err != nil {
		return errors.Wrapf(err, "Failed to grant roles to user %s of tenant %s", userName, tenantName)
	}

	if err := tenantService.RevokeRoles(db, userName, toRevokeRoles); err != nil {
		return errors.Wrapf(err, "Failed to revoke roles from user %s of tenant %s", userName, tenantName)
	}

	return nil
}

func LockUser(tenantName, userName string, password *string) error {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}
	err = tenantService.LockUser(db, userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to lock user %s of tenant %s", userName, tenantName)
	}
	return nil
}

func UnlockUser(tenantName, userName string, password *string) error {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObUserNotExists, userName)
	}
	err = tenantService.UnlockUser(db, userName)
	if err != nil {
		return errors.Wrapf(err, "Failed to unlock user %s of tenant %s", userName, tenantName)
	}
	return nil
}

func GetUser(tenantName, userName string, password *string) (*bo.ObUser, error) {
	tenantInfo, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, errors.Wrapf(err, "get tenant '%s' info failed", tenantName)
	}
	db, err := GetConnectionWithTenantInfo(tenantInfo, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}

	userService := user.GetUserService(db)
	exist, err := userService.IsUserExist(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to check if user %s exists in tenant %s", userName, tenantName)
	}
	if !exist {
		return nil, errors.Occur(errors.ErrObUserNotExists, userName)
	}
	user, err := userService.GetUser(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query user %s of tenant %s", userName, tenantName)
	}
	obUser := user.ToUserBo()

	globalPrivileges, err := userService.GetGrantedGlobalPrivileges(userName)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to query global privileges of user %s of tenant %s", userName, tenantName)
	}

	if tenantInfo.Mode == constant.MYSQL_MODE {
		allDatabases, err := tenantService.ListDatabases(db)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query database list of tenant %s", tenantName)
		}
		dbPrivileges, err := tenantService.ListDatabasePrivilegesOfUser(db, userName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query database privileges of tenant %s", tenantName)
		}
		attachMysqlUserInfo(tenantName, &obUser, globalPrivileges, dbPrivileges, allDatabases)
	} else if tenantInfo.Mode == constant.ORACLE_MODE {
		roles, err := tenantService.GetGrantedRole(db, userName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query roles of user %s of tenant %s", userName, tenantName)
		}
		obUser.GrantedRoles = roles
		objectPrivileges, err := tenantService.GetObjectPrivilegesMap(db)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to query object privileges of user %s of tenant %s", userName, tenantName)
		}
		attachOracleUserInfo(tenantName, &obUser, globalPrivileges, roles, objectPrivileges[userName])
	}

	return &obUser, nil
}

func attachMysqlUserInfo(tenantName string, obUser *bo.ObUser, globalPrivileges []string, dbPrivileges []obmodel.MysqlDb, allDatabases []obmodel.Database) {
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
	attachUserConnectInfo(tenantName, obUser, constant.MYSQL_MODE)
}

func attachOracleUserInfo(tenantName string, obUser *bo.ObUser, systemPriv []string, rolePriv []string, objectPriv []obmodel.ObjectPrivilege) {
	if len(systemPriv) > 0 {
		obUser.GlobalPrivileges = systemPriv
	}
	if len(rolePriv) > 0 {
		obUser.GrantedRoles = rolePriv
	}
	if len(objectPriv) > 0 {
		objectPrivileges := aggregateObjectPrivileges(objectPriv)
		obUser.ObjectPrivileges = objectPrivileges
	}
	attachUserConnectInfo(tenantName, obUser, constant.ORACLE_MODE)
}

func aggregateObjectPrivileges(objectList []obmodel.ObjectPrivilege) []bo.ObjectPrivilege {
	objectPrivilegesMap := make(map[string]*[]string, 0)
	objectMap := make(map[string]bo.DbaObjectBo, 0)
	for _, objectPriv := range objectList {
		key := fmt.Sprintf("%s.%s.%s", objectPriv.Type, objectPriv.Owner, objectPriv.Name)
		if _, ok := objectPrivilegesMap[key]; !ok {
			objectPrivilegesMap[key] = &[]string{objectPriv.Privilege}
		} else {
			*objectPrivilegesMap[key] = append(*objectPrivilegesMap[key], objectPriv.Privilege)
		}
		if _, ok := objectMap[key]; !ok {
			objectMap[key] = bo.DbaObjectBo{
				Type:     objectPriv.Type,
				Name:     objectPriv.Name,
				Owner:    objectPriv.Owner,
				FullName: fmt.Sprintf("%s.%s", objectPriv.Owner, objectPriv.Name),
			}
		}
	}
	objectPrivileges := make([]bo.ObjectPrivilege, 0)
	for key, privileges := range objectPrivilegesMap {
		objectPrivileges = append(objectPrivileges, bo.ObjectPrivilege{Object: objectMap[key], Privileges: *privileges})
	}
	return objectPrivileges
}

func attachUserConnectInfo(tenantName string, obUser *bo.ObUser, mode string) {
	connectionStr := bo.ObproxyAndConnectionString{
		Type:             constant.OB_CONNECTION_TYPE_DIRECT,
		ConnectionString: fmt.Sprintf("obclient -h%s -P%d -u%s@%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, obUser.UserName, tenantName),
	}

	// if userName is oracle mode and contains lowercase letter, add double quotes to userName
	if mode == constant.ORACLE_MODE {
		for _, char := range obUser.UserName {
			if char >= 'a' && char <= 'z' {
				connectionStr.ConnectionString = fmt.Sprintf("obclient -h%s -P%d -u'\"%s\"@%s' -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, obUser.UserName, tenantName)
				break
			}
		}

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
