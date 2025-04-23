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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	obmodel "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
	"gorm.io/gorm"
)

func CreateUser(tenantName string, param *param.CreateUserParam) *errors.OcsAgentError {
	var db *gorm.DB
	defer CloseDbConnection(db)
	var err error
	db, err = GetConnectionWithPassword(tenantName, param.RootPassword)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get connection of tenant %s, error: %s", tenantName, err.Error())
	}

	if param.HostName == "" {
		param.HostName = constant.DEFAULT_HOST
	}

	// Create user.
	if err := tenantService.CreateUser(db, param.UserName, param.Password, param.HostName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "create user '%s' failed: %s", param.UserName, err.Error())
	}

	// Grant privileges.
	if len(param.GlobalPrivileges) != 0 {
		if err := tenantService.GrantGlobalPrivilegesWithHost(db, param.UserName, param.HostName, param.GlobalPrivileges); err != nil {
			return errors.Occurf(errors.ErrUnexpected, "grant global privileges to user '%s' failed: %s", param.UserName, err.Error())
		}
	}

	for _, dbPrivilege := range param.DbPrivileges {
		if err := tenantService.GrantDbPrivilegesWithHost(db, param.UserName, param.HostName, dbPrivilege); err != nil {
			return errors.Occurf(errors.ErrUnexpected, "grant db privileges to user '%s' failed: %s", param.UserName, err.Error())
		}
	}

	return nil
}

func DropUser(tenantName, userName string, param *param.DropUserParam) *errors.OcsAgentError {
	var db *gorm.DB
	defer CloseDbConnection(db)
	var err error
	db, err = GetConnectionWithPassword(tenantName, param.RootPassword)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get connection of tenant %s, error: %s", tenantName, err.Error())
	}

	// Check user exist.
	if exist, err := tenantService.IsUserExist(db, userName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "check user '%s' exist failed", userName)
	} else if !exist {
		return nil
	}

	// Drop user.
	if err := tenantService.DropUser(db, userName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "drop user '%s' failed: %s", userName, err.Error())
	}

	return nil
}

func ListUsers(tenantName string, password *string) ([]bo.ObUser, *errors.OcsAgentError) {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	users, err := tenantService.ListUsers(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query user list of tenant %s, err: %s", tenantName, err.Error())
	}
	allDatabases, err := tenantService.ListDatabases(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query database list of tenant %s, err: %s", tenantName, err.Error())
	}
	dbPrivileges, err := tenantService.ListDatabasePrivileges(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query database privileges of tenant %s, err: %s", tenantName, err.Error())
	}
	dbPrivilegeMap := make(map[string][]obmodel.MysqlDb)
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
		isInnerUser := false
		for _, innerUserName := range constant.OB_INNER_USERS {
			if user.User == innerUserName {
				isInnerUser = true
				break
			}
		}
		if isInnerUser {
			continue
		}
		obUser := &bo.ObUser{
			UserName: user.User,
			IsLocked: strings.HasPrefix(strings.ToUpper(user.AccountLocked), "Y"),
		}
		privileges, ok := dbPrivilegeMap[user.User]
		if !ok {
			privileges = make([]obmodel.MysqlDb, 0)
		}
		attachInfo(tenantName, obUser, &user, privileges, allDatabases)
		result = append(result, *obUser)
	}
	return result, nil
}

func ChangeUserPassword(tenantName, userName string, p *param.ChangeUserPasswordParam) *errors.OcsAgentError {
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	err = tenantService.ChangeUserPassword(db, userName, p.NewPassword)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to change password of user %s of tenant %s, err: %s", userName, tenantName, err.Error())
	}
	return nil
}

func GetUserStats(tenantName, userName string, password *string) (*bo.ObUserStats, *errors.OcsAgentError) {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	sessionStats, err := tenantService.GetUserSessionStats(db, userName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query session stat of user %s of tenant %s, err: %s", userName, tenantName, err.Error())
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

func verifyPrivilege(privileges []string) error {
	for _, privilege := range privileges {
		isPrivilegeValid := false
		for _, availablePrivilege := range constant.OB_MYSQL_PRIVILEGES {
			if strings.ToUpper(privilege) == availablePrivilege {
				isPrivilegeValid = true
				break
			}
		}
		if !isPrivilegeValid {
			return errors.Errorf("unsupported privilege %s", privilege)
		}
	}
	return nil
}

func ModifyUserGlobalPrivilege(tenantName, userName string, p *param.ModifyUserGlobalPrivilegeParam) *errors.OcsAgentError {
	err := verifyPrivilege(p.GlobalPrivileges)
	if err != nil {
		return errors.Occurf(errors.ErrBadRequest, "Found unsupported privilege, err: %s", err.Error())
	}
	obuser, userErr := GetUser(tenantName, userName, p.RootPassword)
	if userErr != nil {
		return userErr
	}
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	privilegesToGrant, privilegesToRevoke := utils.Difference(p.GlobalPrivileges, obuser.GlobalPrivileges)
	err = tenantService.GrantGlobalPrivileges(db, userName, privilegesToGrant)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to grant privilege to user %s of tenant %s, err: %s", userName, tenantName, err.Error())
	}
	err = tenantService.RevokeGlobalPrivileges(db, userName, privilegesToRevoke)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to revoke privilege from user %s of tenant %s, err: %s", userName, tenantName, err.Error())
	}
	return nil
}

func ModifyUserDbPrivilege(tenantName, userName string, p *param.ModifyUserDbPrivilegeParam) *errors.OcsAgentError {
	for _, dbPrivilege := range p.DbPrivileges {
		err := verifyPrivilege(dbPrivilege.Privileges)
		if err != nil {
			return errors.Occurf(errors.ErrBadRequest, "Found unsupported privilege for db %s, err: %s", dbPrivilege.DbName, err.Error())
		}
	}
	obuser, userErr := GetUser(tenantName, userName, p.RootPassword)
	if userErr != nil {
		return userErr
	}
	db, err := GetConnectionWithPassword(tenantName, p.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
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
					return errors.Occurf(errors.ErrUnexpected, "Failed to grant privilege of database %s to user %s of tenant %s, err: %s", currentDbPrivilege.DbName, userName, tenantName, err.Error())
				}
				err = tenantService.RevokeDbPrivileges(db, userName, &param.DbPrivilegeParam{
					DbName:     currentDbPrivilege.DbName,
					Privileges: privilegesToRevoke,
				})
				if err != nil {
					return errors.Occurf(errors.ErrUnexpected, "Failed to revoke privilege of database %s from user %s of tenant %s, err: %s", currentDbPrivilege.DbName, userName, tenantName, err.Error())
				}
			}
		}
		if !found {
			err = tenantService.GrantDbPrivileges(db, userName, &desiredDbPrivilege)
			if err != nil {
				return errors.Occurf(errors.ErrUnexpected, "Failed to grant privilege of database %s to user %s of tenant %s, err: %s", desiredDbPrivilege.DbName, userName, tenantName, err.Error())
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
				return errors.Occurf(errors.ErrUnexpected, "Failed to revoke privilege of database %s from user %s of tenant %s, err: %s", currentDbPrivilege.DbName, userName, tenantName, err.Error())
			}
		}
	}
	return nil
}

func LockUser(tenantName, userName string, password *string) *errors.OcsAgentError {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	err = tenantService.LockUser(db, userName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to lock user %s of tenant %s, err: %s", userName, tenantName, err.Error())
	}
	return nil
}

func UnlockUser(tenantName, userName string, password *string) *errors.OcsAgentError {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	err = tenantService.UnlockUser(db, userName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to lock user %s of tenant %s, err: %s", userName, tenantName, err.Error())
	}
	return nil
}

func GetUser(tenantName, userName string, password *string) (*bo.ObUser, *errors.OcsAgentError) {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	user, err := tenantService.GetUser(db, userName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query user %s of tenant %s, err: %s", userName, tenantName, err.Error())
	}
	allDatabases, err := tenantService.ListDatabases(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query database list of tenant %s, err: %s", tenantName, err.Error())
	}
	dbPrivileges, err := tenantService.ListDatabasePrivilegesOfUser(db, userName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to query database privileges of tenant %s, err: %s", tenantName, err.Error())
	}

	obUser := &bo.ObUser{
		UserName: user.User,
		IsLocked: strings.HasPrefix(strings.ToUpper(user.AccountLocked), "Y"),
	}
	attachInfo(tenantName, obUser, user, dbPrivileges, allDatabases)
	return obUser, nil
}

func attachInfo(tenantName string, obUser *bo.ObUser, user *obmodel.MysqlUser, dbPrivileges []obmodel.MysqlDb, allDatabases []obmodel.Database) {
	globalPrivileges := extractGlobalPrivileges(user)
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
	connectionStr := bo.ObproxyAndConnectionString{
		Type:             constant.OB_CONNECTION_TYPE_DIRECT,
		ConnectionString: fmt.Sprintf("obclient -h%s -P%d -u%s@%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, obUser.UserName, tenantName),
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

func extractGlobalPrivileges(user *obmodel.MysqlUser) []string {
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
