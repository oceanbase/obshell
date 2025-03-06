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
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/param"
	"gorm.io/gorm"
)

func CreateUser(tenantName string, param param.CreateUserParam) *errors.OcsAgentError {
	if exist, err := tenantService.IsTenantExist(tenantName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "check tenant '%s' exist failed", tenantName)
	} else if !exist {
		return errors.Occurf(errors.ErrBadRequest, "Tenant '%s' not exists.", tenantName)
	}

	var db *gorm.DB
	var err error
	if tenantName == constant.TENANT_SYS {
		db, err = oceanbase.GetInstance()
		if err != nil {
			return errors.Occurf(errors.ErrUnexpected, "get oceanbase instance failed")
		}
	} else {
		defer func() {
			if db != nil {
				tempDb, _ := db.DB()
				if tempDb != nil {
					tempDb.Close()
				}
			}
		}()
		db, err = oceanbase.LoadGormWithTenant(tenantName, param.RootPassword)
		if err != nil {
			return errors.Occurf(errors.ErrUnexpected, "load gorm with tenant '%s' failed", tenantName)
		}
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
		if err := tenantService.GrantGlobalPrivileges(db, param.UserName, param.HostName, param.GlobalPrivileges); err != nil {
			return errors.Occurf(errors.ErrUnexpected, "grant global privileges to user '%s' failed: %s", param.UserName, err.Error())
		}
	}

	for _, dbPrivilege := range param.DbPrivileges {
		if err := tenantService.GrantDbPrivileges(db, param.UserName, param.HostName, dbPrivilege); err != nil {
			return errors.Occurf(errors.ErrUnexpected, "grant db privileges to user '%s' failed: %s", param.UserName, err.Error())
		}
	}

	return nil
}

func DropUser(tenantName, userName, rootPassword string) *errors.OcsAgentError {
	if exist, err := tenantService.IsTenantExist(tenantName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "check tenant '%s' exist failed", tenantName)
	} else if !exist {
		return errors.Occurf(errors.ErrBadRequest, "Tenant '%s' not exists.", tenantName)
	}

	var db *gorm.DB
	var err error
	if tenantName == constant.TENANT_SYS {
		db, err = oceanbase.GetInstance()
		if err != nil {
			return errors.Occurf(errors.ErrUnexpected, "get oceanbase instance failed")
		}
	} else {
		defer func() {
			if db != nil {
				tempDb, _ := db.DB()
				if tempDb != nil {
					tempDb.Close()
				}
			}
		}()
		db, err = oceanbase.LoadGormWithTenant(tenantName, rootPassword)
		if err != nil {
			return errors.Occurf(errors.ErrUnexpected, "load gorm with tenant '%s' failed", tenantName)
		}
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
