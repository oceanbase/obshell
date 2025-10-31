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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	obmodel "github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"

	"gorm.io/gorm"
)

func TenantPreCheck(tenantName string, password *string) (*bo.ObTenantPreCheckResult, error) {
	// TODO: password must be not nil...
	isPasswordExists := password == nil
	isConnectable := false
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	isConnectable = (err == nil)
	isEmptyRootPassword, err := IsEmptyRootPassword(tenantName)
	if err != nil {
		return nil, errors.Wrapf(err, "check tenant '%s' password if empty failed", tenantName)
	}

	return &bo.ObTenantPreCheckResult{
		IsConnectable:       isConnectable,
		IsPasswordExists:    isPasswordExists,
		IsEmptyRootPassword: isEmptyRootPassword,
	}, nil
}

func IsEmptyRootPassword(tenantName string) (bool, error) {
	if tenantName == constant.TENANT_SYS {
		return meta.OCEANBASE_PWD == "", nil
	} else {
		mode, err := tenantService.GetTenantMode(tenantName)
		if err != nil {
			return true, err
		}
		if db, err := oceanbase.LoadGormWithTenant(tenantName, "", mode); err != nil {
			if strings.Contains(err.Error(), "Access denied") {
				return false, nil
			} else {
				return true, err
			}
		} else if db != nil {
			oceanbaseDB, _ := db.DB()
			if oceanbaseDB != nil {
				oceanbaseDB.Close()
			}
		}
	}
	return true, nil
}

func GetConnectionWithPasswordAndMode(tenantName string, password *string, mode string) (*gorm.DB, error) {
	if tenantName == constant.TENANT_SYS {
		return oceanbase.GetInstance()
	} else {
		if password != nil {
			return oceanbase.LoadGormWithTenant(tenantName, *password, mode)
		} else {
			return oceanbase.LoadGormWithTenant(tenantName, "", mode)
		}
	}
}

func GetConnectionWithPassword(tenantName string, password *string) (*gorm.DB, error) {
	if tenantName == constant.TENANT_SYS {
		return oceanbase.GetInstance()
	} else {
		// get the tenant type
		mode, err := tenantService.GetTenantMode(tenantName)
		if err != nil {
			return nil, errors.Wrapf(err, "get tenant '%s' mode failed", tenantName)
		}
		if password != nil {
			return oceanbase.LoadGormWithTenant(tenantName, *password, mode)
		} else {
			return oceanbase.LoadGormWithTenant(tenantName, "", mode)
		}
	}
}

func GetConnectionWithTenantInfo(tenantInfo *obmodel.DbaObTenant, password *string) (*gorm.DB, error) {
	if tenantInfo.TenantName == constant.TENANT_SYS {
		return oceanbase.GetInstance()
	}
	if password != nil {
		return oceanbase.LoadGormWithTenant(tenantInfo.TenantName, *password, tenantInfo.Mode)
	} else {
		return oceanbase.LoadGormWithTenant(tenantInfo.TenantName, "", tenantInfo.Mode)
	}
}

func CloseDbConnection(db *gorm.DB) {
	if db == oceanbase.GetRawInstance() {
		return
	}
	if db != nil {
		tempDb, _ := db.DB()
		if tempDb != nil {
			tempDb.Close()
		}
	}
}
