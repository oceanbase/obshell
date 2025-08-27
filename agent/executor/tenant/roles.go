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
	"regexp"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/service/user"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
	"gorm.io/gorm"
)

func ListRoles(name string, password *string) ([]bo.ObRole, error) {
	db, err := getOracleTenantConnection(name, password)
	if err != nil {
		return nil, err
	}

	roles, err := tenantService.ListRoles(db)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, nil
	}

	obRoles := make([]bo.ObRole, len(roles))
	for i := range roles {
		obRoles[i] = roles[i].ToRoleBo()
	}

	if err := attachRoleInfo(db, obRoles); err != nil {
		return nil, err
	}

	return obRoles, nil
}

func attachRoleInfo(db *gorm.DB, obRoles []bo.ObRole) error {
	sysPrivMap, err := user.GetUserService(db).GetGlobalPrivilegesMap()
	if err != nil {
		return err
	}
	grantedRoleMap, err := tenantService.GetGrantedRoleMap(db)
	if err != nil {
		return err
	}
	objectPrivilegesMap, err := tenantService.GetObjectPrivilegesMap(db)
	if err != nil {
		return err
	}
	userGranteesMap, err := tenantService.GetUserGranteesMap(db)
	if err != nil {
		return err
	}
	roleGranteesMap, err := tenantService.GetRoleGranteesMap(db)
	if err != nil {
		return err
	}

	for i := range obRoles {
		// attach global privileges
		obRoles[i].GlobalPrivileges = sysPrivMap[obRoles[i].Role]
		// attach granted roles
		obRoles[i].GrantedRoles = grantedRoleMap[obRoles[i].Role]
		// attach object privileges
		obRoles[i].ObjectPrivileges = aggregateObjectPrivileges(objectPrivilegesMap[obRoles[i].Role])
		// attach user grantees
		obRoles[i].UserGrantees = userGranteesMap[obRoles[i].Role]
		// attach role grantees
		obRoles[i].RoleGrantees = roleGranteesMap[obRoles[i].Role]
	}

	return nil
}

func GetRole(name, role string, password *string) (*bo.ObRole, error) {
	db, err := getOracleTenantConnection(name, password)
	if err != nil {
		return nil, err
	}
	if exists, err := tenantService.IsRoleExist(db, role); err != nil {
		return nil, errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, name)
	} else if !exists {
		return nil, errors.Occur(errors.ErrObRoleNotExists, role)
	}

	roleInfo, err := tenantService.GetRole(db, role)
	if err != nil {
		return nil, err
	}

	obRoles := make([]bo.ObRole, 0)
	obRoles = append(obRoles, roleInfo.ToRoleBo())
	if err := attachRoleInfo(db, obRoles); err != nil {
		return nil, err
	}

	return &obRoles[0], nil
}

func ModifyRole(name, role string, param *param.ModifyRoleParam) error {
	db, err := getOracleTenantConnection(name, param.RootPassword)
	if err != nil {
		return err
	}

	exist, err := tenantService.IsRoleExist(db, role)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, name)
	}
	if !exist {
		return errors.Occur(errors.ErrObRoleNotExists, role)
	}

	grantedRoles, err := tenantService.GetGrantedRole(db, role)
	if err != nil {
		return err
	}

	toGrantRoles, toRevokeRoles := utils.Difference(param.Roles, grantedRoles)

	if len(toGrantRoles) > 0 {
		if err := tenantService.GrantRoles(db, role, toGrantRoles); err != nil {
			return errors.Wrapf(err, "Failed to grant roles to role %s of tenant %s", role, name)
		}
	}

	if len(toRevokeRoles) > 0 {
		if err := tenantService.RevokeRoles(db, role, toRevokeRoles); err != nil {
			return errors.Wrapf(err, "Failed to revoke roles from role %s of tenant %s", role, name)
		}
	}

	return nil
}

func ModifyRoleGlobalPrivilege(name, role string, param *param.ModifyRoleGlobalPrivilegeParam) error {
	db, err := getOracleTenantConnection(name, param.RootPassword)
	if err != nil {
		return err
	}

	exist, err := tenantService.IsRoleExist(db, role)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, name)
	}
	if !exist {
		return errors.Occur(errors.ErrObRoleNotExists, role)
	}

	userService := user.GetUserService(db)
	grantedGlobalPrivileges, err := userService.GetGrantedGlobalPrivileges(role)
	if err != nil {
		return err
	}

	toGrantRoles, toRevokeRoles := utils.Difference(param.GlobalPrivileges, grantedGlobalPrivileges)

	if err := userService.GrantGlobalPrivileges(role, toGrantRoles); err != nil {
		return errors.Wrapf(err, "Failed to grant roles to role %s of tenant %s", role, name)
	}

	if err := userService.RevokeGlobalPrivileges(role, toRevokeRoles); err != nil {
		return errors.Wrapf(err, "Failed to revoke roles from role %s of tenant %s", role, name)
	}

	return nil
}

func PatchRoleObjectPrivilege(name, role string, param *param.ModifyObjectPrivilegeParam) error {
	db, err := getOracleTenantConnection(name, param.TenantRootPasswordParam.RootPassword)
	if err != nil {
		return err
	}

	exist, err := tenantService.IsRoleExist(db, role)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, name)
	}
	if !exist {
		return errors.Occur(errors.ErrObRoleNotExists, role)
	}

	grantedObjectPrivileges, err := tenantService.GetGrantedObjectPrivileges(db, role)
	if err != nil {
		return err
	}
	aggGrantedObjectPrivileges := aggregateObjectPrivileges(grantedObjectPrivileges)

	if err := patchObjectPrivilege(db, name, role, aggGrantedObjectPrivileges, param.ObjectPrivileges); err != nil {
		return err
	}

	return nil
}

func ModifyRoleObjectPrivilege(tenantName, role string, param *param.ModifyObjectPrivilegeParam) error {
	db, err := getOracleTenantConnection(tenantName, param.TenantRootPasswordParam.RootPassword)
	if err != nil {
		return err
	}

	exist, err := tenantService.IsRoleExist(db, role)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, tenantName)
	}
	if !exist {
		return errors.Occur(errors.ErrObRoleNotExists, role)
	}

	grantedObjectPrivileges, err := tenantService.GetGrantedObjectPrivileges(db, role)
	if err != nil {
		return err
	}
	aggGrantedObjectPrivileges := aggregateObjectPrivileges(grantedObjectPrivileges)

	if err := putObjectPrivilege(db, tenantName, role, aggGrantedObjectPrivileges, param.ObjectPrivileges); err != nil {
		return err
	}

	return nil
}

func patchObjectPrivilege(db *gorm.DB, tenantName, role string, currentObjectPrivileges []bo.ObjectPrivilege, targetObjectPrivileges []param.ObjectPrivilegeParam) error {
	for _, objectPrivilege := range targetObjectPrivileges {
		found := false
		for _, currentObjectPrivilege := range currentObjectPrivileges {
			if currentObjectPrivilege.Object.Owner == objectPrivilege.Owner && currentObjectPrivilege.Object.Name == objectPrivilege.ObjectName {
				toGrantPrivileges, toRevokePrivileges := utils.Difference(objectPrivilege.Privileges, currentObjectPrivilege.Privileges)
				if err := tenantService.GrantObjectPrivileges(db, role, objectPrivilege.Owner, objectPrivilege.ObjectName, toGrantPrivileges); err != nil {
					return errors.Wrapf(err, "Failed to grant object privileges to role %s of tenant %s", role, tenantName)
				}
				if err := tenantService.RevokeObjectPrivileges(db, role, objectPrivilege.Owner, objectPrivilege.ObjectName, toRevokePrivileges); err != nil {
					return errors.Wrapf(err, "Failed to revoke object privileges from role %s of tenant %s", role, tenantName)
				}
				found = true
				break
			}
		}
		if !found {
			err := tenantService.GrantObjectPrivileges(db, role, objectPrivilege.Owner, objectPrivilege.ObjectName, objectPrivilege.Privileges)
			if err != nil {
				return errors.Wrapf(err, "Failed to grant object privileges to %s of tenant %s", role, tenantName)
			}
		}
	}
	return nil
}

func putObjectPrivilege(db *gorm.DB, tenantName, name string, currentObjectPrivileges []bo.ObjectPrivilege, targetObjectPrivileges []param.ObjectPrivilegeParam) error {
	if err := patchObjectPrivilege(db, tenantName, name, currentObjectPrivileges, targetObjectPrivileges); err != nil {
		return err
	}

	for _, currentObjectPrivilege := range currentObjectPrivileges {
		found := false
		for _, objectPrivilege := range targetObjectPrivileges {
			if currentObjectPrivilege.Object.Owner == objectPrivilege.Owner && currentObjectPrivilege.Object.Name == objectPrivilege.ObjectName {
				found = true
				break
			}
		}
		if !found {
			if err := tenantService.RevokeObjectPrivileges(db, name, currentObjectPrivilege.Object.Owner, currentObjectPrivilege.Object.Name, currentObjectPrivilege.Privileges); err != nil {
				return errors.Wrapf(err, "Failed to revoke object privileges from %s of tenant %s", name, tenantName)
			}
		}
	}
	return nil
}

func RevokeRoleObjectPrivilege(name, role string, param *param.RevokeObjectPrivilegeParam) error {
	db, err := getOracleTenantConnection(name, param.TenantRootPasswordParam.RootPassword)
	if err != nil {
		return err
	}

	exist, err := tenantService.IsRoleExist(db, role)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, name)
	}
	if !exist {
		return errors.Occur(errors.ErrObRoleNotExists, role)
	}

	for _, objectPrivilege := range param.ObjectPrivileges {
		err = tenantService.RevokeObjectPrivileges(db, role, objectPrivilege.Owner, objectPrivilege.ObjectName, objectPrivilege.Privileges)
		if err != nil {
			return errors.Wrapf(err, "Failed to revoke object privileges from role %s of tenant %s", role, name)
		}
	}

	return nil
}

func GrantRoleObjectPrivilege(name, role string, param *param.GrantObjectPrivilegeParam) error {
	db, err := getOracleTenantConnection(name, param.TenantRootPasswordParam.RootPassword)
	if err != nil {
		return err
	}

	exist, err := tenantService.IsRoleExist(db, role)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, name)
	}
	if !exist {
		return errors.Occur(errors.ErrObRoleNotExists, role)
	}

	for _, objectPrivilege := range param.ObjectPrivileges {
		err = tenantService.GrantObjectPrivileges(db, role, objectPrivilege.Owner, objectPrivilege.ObjectName, objectPrivilege.Privileges)
		if err != nil {
			return errors.Wrapf(err, "Failed to grant object privileges to role %s of tenant %s", role, name)
		}
	}

	return nil
}

func CreateRole(name string, param *param.CreateRoleParam) error {
	db, err := getOracleTenantConnection(name, param.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return err
	}

	// check the role name is valid, only alphanumeric characters and underscores are allowed
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z_0-9]{1,29}$`).MatchString(param.RoleName) {
		return errors.Occur(errors.ErrObRoleNameInvalid, param.RoleName)
	}

	if err := tenantService.CreateRole(db, param.RoleName); err != nil {
		return errors.Wrapf(err, "Failed to create role %s of tenant %s", param.RoleName, name)
	}

	// Grant global privileges.
	if len(param.GlobalPrivileges) != 0 {
		if err := user.GetUserService(db).GrantGlobalPrivileges(param.RoleName, param.GlobalPrivileges); err != nil {
			return errors.Wrapf(err, "Failed to grant global privileges to role %s of tenant %s", param.RoleName, name)
		}
	}

	// Grant roles.
	if len(param.Roles) != 0 {
		if err := tenantService.GrantRoles(db, param.RoleName, param.Roles); err != nil {
			return errors.Wrapf(err, "Failed to grant roles to role %s of tenant %s", param.RoleName, name)
		}
	}

	return nil
}

func DropRole(tenantName, role string, param *param.DropRoleParam) error {
	db, err := getOracleTenantConnection(tenantName, param.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return err
	}

	if exists, err := tenantService.IsRoleExist(db, role); err != nil {
		return errors.Wrapf(err, "Failed to check if role %s exists in tenant %s", role, tenantName)
	} else if !exists {
		return nil
	}

	if err := tenantService.DropRole(db, role); err != nil {
		return errors.Wrapf(err, "Failed to drop role %s of tenant %s", role, tenantName)
	}

	return nil
}

func getOracleTenantConnection(name string, rootPassword *string) (*gorm.DB, error) {
	tenantInfo, err := tenantService.GetTenantByName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get tenant '%s' info failed", name)
	}
	if tenantInfo.Mode != constant.ORACLE_MODE {
		return nil, errors.Occur(errors.ErrObUserOracleModeNotSupport)
	}

	db, err := GetConnectionWithTenantInfo(tenantInfo, rootPassword)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get db connection of tenant %s", name)
	}
	return db, nil
}
