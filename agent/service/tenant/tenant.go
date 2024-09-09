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

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func (s *TenantService) GetAllUserTenants() (res []oceanbase.DbaOBTenants, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where(" TENANT_TYPE = ? ", constant.TENANT_TYPE_USER).Scan(&res).Error
	return
}

func (s *TenantService) GetTenantByID(id int) (res *oceanbase.DbaOBTenants, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where("tenant_id = ?", id).Scan(&res).Error
	return
}

func (s *TenantService) GetTenantByName(name string) (res *oceanbase.DbaOBTenants, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where("tenant_name = ? and TENANT_TYPE = ? ", name, constant.TENANT_TYPE_USER).Scan(&res).Error
	return
}

func (s *TenantService) IsMetaTenantStatusNormal(tenantName string) (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}

	var id int
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Select("tenant_id").Where("tenant_name = ? and TENANT_TYPE = ? ", tenantName, constant.TENANT_TYPE_USER).Scan(&id).Error
	if err != nil || id == 0 {
		return false, err
	}

	metaTenantName := fmt.Sprintf("META$%d", id)
	var status string
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Select("STATUS").Where("tenant_name = ? and TENANT_TYPE = ?", metaTenantName, constant.TENANT_TYPE_META).Scan(&status).Error
	return strings.ToUpper(status) == constant.TENANT_STATUS_NORMAL, err
}

func (s *TenantService) GetUnitConfigByName(name string) (res *oceanbase.DbaObUnitConfigs, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_UNIT_CONFIGS).Where("name = ?", name).Scan(&res).Error
	return
}

func (s *TenantService) DeleteTenant(name string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("DROP TENANT %s FORCE", name)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) CreateResourcePool(poolName, unitConfigName string, unitNum int, zoneList []string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	for i, zone := range zoneList {
		zoneList[i] = fmt.Sprintf("'%s'", zone)
	}
	sql := fmt.Sprintf("CREATE RESOURCE POOL %s UNIT = %s, UNIT_NUM = %d, ZONE_LIST = (%s)", poolName, unitConfigName, unitNum, strings.Join(zoneList, ","))
	err = oceanbaseDb.Exec(sql).Error
	return
}

func (s *TenantService) GetResourcePoolByName(poolName string) (res *oceanbase.DbaObResourcePools, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_RESOURCE_POOLS).Where("name = ?", poolName).Scan(&res).Error
	return
}

func (s *TenantService) GetResourcePoolsByTenantID(tenantID int64) (res []oceanbase.DbaObResourcePools, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_RESOURCE_POOLS).Where("TENANT_ID = ?", tenantID).Scan(&res).Error
	return
}

func (s *TenantService) GetResourcePoolsNameByTenantID(tenantID int64) (res []string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_RESOURCE_POOLS).Select("NAME").Where("TENANT_ID = ?", tenantID).Scan(&res).Error
	return
}

func (s *TenantService) DeleteResourcePool(name string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("DROP RESOURCE POOL %s", name)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) ActiveTenant(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM ACTIVATE STANDBY TENANT %s", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) GetTenantRole(tenantName string) (role string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Select("TENANT_ROLE").Where("tenant_name = ? and TENANT_TYPE = ?", tenantName, constant.TENANT_TYPE_USER).Scan(&role).Error
	return
}

func (s *TenantService) Upgrade(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM RUN UPGRADE JOB 'UPGRADE_ALL' TENANT = %s", tenantName)
	err = oceanbaseDb.Exec(sql).Error
	return
}

func (s *TenantService) GetUpgradeJobHistoryCount(tenantName string) (count int64, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	var version string
	err = oceanbaseDb.Raw("SELECT ob_version();").Scan(&version).Error
	if err != nil {
		return
	}
	log.Infof("ob version is %s", version)

	var id int
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Select("tenant_id").Where("TENANT_NAME = ? and TENANT_TYPE = ?", tenantName, constant.TENANT_TYPE_USER).Scan(&id).Error
	if err != nil || id == 0 {
		return 0, err
	}

	err = oceanbaseDb.Table(DBA_OB_CLUSTER_EVENT_HISTORY).Where("event = 'UPGRADE_ALL' AND value3 = ? AND value5 = ?", version, id).Count(&count).Error
	return
}
