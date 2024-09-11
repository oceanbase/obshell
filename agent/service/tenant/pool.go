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

	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	model "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func (t *TenantService) CreateResourcePool(name string, unitConfigName string, unitNum int, zonelist []string) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(SQL_CREATE_RESOURCE_POOL, name, unitConfigName, unitNum, strings.Join(zonelist, "','"))
	return db.Exec(sql).Error
}

func (t *TenantService) IsResourcePoolExistAndFreed(name string, unitConfigName string, unitNum int, zoneName string) (bool, error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	unitQuery := db.Table(DBA_OB_UNIT_CONFIGS).Select("UNIT_CONFIG_ID").Where("NAME = ?", unitConfigName)
	err = db.Table(DBA_OB_RESOURCE_POOLS).Where("NAME = ? and UNIT_CONFIG_ID = (?) and UNIT_COUNT = ? and ZONE_LIST = ? and TENANT_ID is NULL", name, unitQuery, unitNum, zoneName).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) DropResourcePool(name string, ifExist bool) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	var sql string
	if ifExist {
		sql = fmt.Sprintf(SQL_DROP_RESOURCE_POOL_IF_EXISTS, name)
	} else {
		sql = fmt.Sprintf(SQL_DROP_RESOURCE_POOL, name)
	}
	return db.Exec(sql).Error
}

func (t *TenantService) GetTenantResourcePoolNames(tenantId int) (pools []string, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_RESOURCE_POOLS).Select("NAME").Where("TENANT_ID = (?)", tenantId).Scan(&pools).Error
	return
}

func (t *TenantService) CheckTenantHasPoolOnZone(tenantId int, zoneName string) (bool, error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	poolQuery := db.Table(DBA_OB_RESOURCE_POOLS).Select("RESOURCE_POOL_ID").Where("TENANT_ID = (?)", tenantId)
	err = db.Table(DBA_OB_UNITS).Where("ZONE = ? AND RESOURCE_POOL_ID IN (?)", zoneName, poolQuery).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) GetTenantResourcePool(tenantId int) (pools []model.DbaObResourcePool, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_RESOURCE_POOLS).Where("TENANT_ID = (?)", tenantId).Scan(&pools).Error
	return
}

func (t *TenantService) GetResourcePoolByName(name string) (pool *model.DbaObResourcePool, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_RESOURCE_POOLS).Where("NAME = (?)", name).Scan(&pool).Error
	return
}

func (t *TenantService) GetAllResourcePool() (pools []model.DbaObResourcePool, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_RESOURCE_POOLS).Scan(&pools).Error
	return
}

func (t *TenantService) AlterResourcePoolList(tenantId int, poolList []string) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	tenantName, err := t.GetTenantName(tenantId)
	if err != nil {
		return err
	}
	resource_pool_list := "\"" + strings.Join(poolList, "\",\"") + "\""
	return db.Exec(fmt.Sprintf(SQL_ALTER_RESOURCE_LIST, tenantName, resource_pool_list)).Error
}

func (t *TenantService) SplitResourcePool(poolName string, poolList []string, zoneList []string) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	zones := "'" + strings.Join(zoneList, "','") + "'"
	pools := "'" + strings.Join(poolList, "','") + "'"
	return db.Exec(fmt.Sprintf(SQL_ALTER_RESOURCE_POOL_SPLIT, poolName, pools, zones)).Error
}

// AlterResourcePoolUnitConfigByZoneName alter resource pool unit config by zone name,
// only support the pool contains only one zone
func (t *TenantService) AlterResourcePoolUnitConfig(poolName string, unitConfigName string) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_ALTER_RESOURCE_POOL_UNIT_CONFIG, poolName, unitConfigName)).Error
}
