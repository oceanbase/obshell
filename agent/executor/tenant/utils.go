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
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/service/tenant"

	"gorm.io/gorm"
)

func TenantPreCheck(tenantName string) (*bo.ObTenantPreCheckResult, *errors.OcsAgentError) {
	isPasswordExists := false
	isConnectable := false
	passwordMap := tenant.GetPasswordMap()
	_, isPasswordExists = passwordMap.Get(tenantName)
	// if isPasswordExists {
	db, err := GetConnection(tenantName)
	defer CloseDbConnection(db)
	isConnectable = (err == nil)
	// }
	isEmptyRootPassword, err := IsEmptyRootPassword(tenantName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "check tenant '%s' password if empty failed", tenantName)
	}

	return &bo.ObTenantPreCheckResult{
		IsConnectable:       isConnectable,
		IsPasswordExists:    isPasswordExists,
		IsEmptyRootPassword: isEmptyRootPassword,
	}, nil
}

func GetConnection(tenantName string) (*gorm.DB, error) {
	if tenantName == constant.TENANT_SYS {
		return oceanbase.GetInstance()
	} else {
		passwordMap := tenant.GetPasswordMap()
		password, _ := passwordMap.Get(tenantName)
		return oceanbase.LoadGormWithTenant(tenantName, password)
	}
}

func IsEmptyRootPassword(tenantName string) (bool, error) {
	if tenantName == constant.TENANT_SYS {
		return meta.OCEANBASE_PWD == "", nil
	} else {
		if err := oceanbase.LoadGormWithTenantForTest(tenantName, ""); err != nil {
			if strings.Contains(err.Error(), "Access denied") {
				return false, nil
			} else {
				return true, err
			}
		}
	}
	return true, nil
}

func GetConnectionWithPassword(tenantName, password string) (*gorm.DB, error) {
	if tenantName == constant.TENANT_SYS {
		return oceanbase.GetInstance()
	} else {
		return oceanbase.LoadGormWithTenant(tenantName, password)
	}
}

func CloseDbConnection(db *gorm.DB) {
	sysDb, err := oceanbase.GetInstance()
	if err == nil && db == sysDb {
		return
	}
	if db != nil {
		tempDb, _ := db.DB()
		if tempDb != nil {
			tempDb.Close()
		}
	}
}
