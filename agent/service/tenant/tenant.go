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
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func (t *TenantService) TryExecute(sql string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(sql).Error
}

func (t *TenantService) CheckVariables(vars map[string]interface{}) error {
	if len(vars) == 0 {
		return nil
	}
	items := make([]string, 0)
	for k, v := range vars {
		if val, ok := v.(string); ok {
			items = append(items, fmt.Sprintf(k+"= `%v`", val))
		} else {
			items = append(items, fmt.Sprintf(k+"= %v", v))
		}
	}

	err := t.TryExecute("create tenant sys resource_pool_list = ('') set " + strings.Join(items, ","))
	if err == nil {
		// err could not be nil, because tenant sys is exist.
		return errors.Wrap(err, "Unexpected error when validate variables.")
	} else if strings.Contains(err.Error(), "Error 5156") {
		/* The validation of the system variables by the observer
		 * takes place before the tenant is created.'Error 5156
		 * (Tenant already exists)' occuredindicates that the
		 * system variable validation has already passed. */
		return nil
	}
	return err
}

func (t *TenantService) IsTenantExist(name string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}

	var count int64
	err = db.Table(DBA_OB_TENANTS).Where("TENANT_NAME = ?", name).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) GetTenantStatus(name string) (string, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}

	var status string
	err = db.Table(DBA_OB_TENANTS).Select("STATUS").Where("TENANT_NAME = ?", name).Scan(&status).Error
	return status, err
}

func (t *TenantService) IsTenantExistInRecyclebin(name string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}

	var count int64
	err = db.Table(DBA_RECYCLEBIN).Where("(ORIGINAL_NAME = ? OR OBJECT_NAME = ?) AND TYPE = 'TENANT'", name, name).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) GetTenantId(name string) (int, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}

	var id int
	err = db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", name).Scan(&id).Error
	return id, err
}

func (t *TenantService) GetTenantName(id int) (string, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}

	var name string
	err = db.Table(DBA_OB_TENANTS).Select("TENANT_NAME").Where("TENANT_ID = ?", id).Scan(&name).Error
	return name, err
}

func (t *TenantService) CreateTenant(basic string, input []interface{}) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(basic, input...)
	return db.Exec(sql).Error
}

func (t *TenantService) DropTenant(tenantName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_DROP_TENANT, tenantName)).Error
}

func (t *TenantService) RenameTenant(name string, newName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_RENAME_TENANT, name, newName)).Error
}

func (t *TenantService) RecycleTenant(tenantName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_RECYCLE_TENANT, tenantName)).Error
}

func (t *TenantService) FlashbackTenant(name string, newName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_FLASHBACK_TENANT, name, newName)).Error
}

// This function return immediately, but the tenant may not be purged yet.
// There is a column named 'CAN_PURGE' in DBA_RECYCLEBIN
func (t *TenantService) PurgeTenant(name string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_PURGE_TENANT, name)).Error
}

func (t *TenantService) GetRecycledTenant() (overview []oceanbase.DbaRecyclebin, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_RECYCLEBIN).Where("TYPE = 'TENANT'").Scan(&overview).Error
	return
}

func (t *TenantService) GetRecycledTenantOriginalName(objectOrOriginalName string) (originalName string, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}
	err = db.Table(DBA_RECYCLEBIN).
		Select("ORIGINAL_NAME").
		Where("ORIGINAL_NAME = ? or OBJECT_NAME = ?", objectOrOriginalName, objectOrOriginalName).
		Order("OBJECT_NAME DESC").
		Limit(1).
		Scan(&originalName).Error
	return
}

func (t *TenantService) GetRecycledTenantObjectName(objectOrOriginalName string) (objectName string, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}
	err = db.Table(DBA_RECYCLEBIN).
		Select("OBJECT_NAME").
		Where("ORIGINAL_NAME = ? or OBJECT_NAME = ?", objectOrOriginalName, objectOrOriginalName).
		Order("OBJECT_NAME DESC").
		Limit(1).
		Scan(&objectName).Error
	return
}

func (t *TenantService) SetTenantParameters(tenantName string, parameters map[string]interface{}) error {
	if len(parameters) == 0 {
		return nil
	}
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	items := make([]string, 0)
	for k, v := range parameters {
		items = append(items, fmt.Sprintf("`%s` = \"%v\" tenant = `%s`", k, v, tenantName))
	}
	sql := SQL_SET_TENANT_PARAMETER_BASIC + strings.Join(items, ",")
	return db.Exec(sql).Error
}

func (t *TenantService) SetTenantVariables(tenantName string, variables map[string]interface{}) error {
	if len(variables) == 0 {
		return nil
	}
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := SQL_SET_TENANT_VARIABLE_BASIC
	variablesSq := ""
	for k, v := range variables {
		if val, ok := v.(string); ok {
			variablesSq += fmt.Sprintf(", "+k+"= `%v`", val)
		} else {
			variablesSq += fmt.Sprintf(", "+k+"= %v", v)
		}
	}
	sql += variablesSq[1:]
	return db.Exec(fmt.Sprintf(sql, tenantName)).Error
}

func (t *TenantService) LockTenant(name string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_LOCK_TENANT, name)).Error
}

func (t *TenantService) UnlockTenant(name string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_UNLOCK_TENANT, name)).Error
}

func (t *TenantService) GetTenantPrimaryZone(tenantId int) (zone string, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}
	err = db.Table(DBA_OB_TENANTS).Select("PRIMARY_ZONE").Where("TENANT_ID = ?", tenantId).Scan(&zone).Error
	return
}

func (t *TenantService) GetTenantUnitNum(tenantId int) (unitNum int, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}
	err = db.Table(DBA_OB_RESOURCE_POOLS).Select("UNIT_COUNT").Where("TENANT_ID = (?)", tenantId).Scan(&unitNum).Error
	return
}

func (t *TenantService) AlterTenantUnitNum(tenantName string, unitNum int) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_ALTER_TENANT_UNIT_NUM, tenantName, unitNum)).Error
}

func (t *TenantService) ModifyTenantWhitelist(tenantName string, whitelist string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_ALTER_TENANT_WHITELIST, tenantName, whitelist)).Error
}

func (t *TenantService) ModifyTenantRootPassword(tenantName string, oldPwd string, newPwd string) *errors.OcsAgentError {
	tempDb, err := oceanbasedb.LoadGormWithTenant(tenantName, oldPwd)
	if err != nil {
		return errors.Occur(errors.ErrUnexpected, err.Error())
	}
	defer func() {
		db, _ := tempDb.DB()
		if db != nil {
			db.Close()
		}
	}()
	if err = tempDb.Exec(fmt.Sprintf(SQL_ALTER_TENANT_ROOT_PASSWORD, transfer(newPwd))).Error; err != nil {
		return errors.Occurf(errors.ErrUnexpected, "modify tenant root password failed: %s", err.Error())
	}
	return nil
}

func (t *TenantService) AlterTenantPrimaryZone(tenantName string, primaryZone string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_ALTER_TENANT_PRIMARY_ZONE, tenantName, primaryZone)).Error
}

func (s *TenantService) GetTenantByName(name string) (res *oceanbase.DbaObTenant, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where("tenant_name = ? and TENANT_TYPE != ?", name, constant.TENANT_TYPE_META).Scan(&res).Error
	return
}

func (t *TenantService) GetTenantsOverView() (overviews []oceanbase.DbaObTenant, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_TENANTS).
		Where("TENANT_TYPE != 'META' AND IN_RECYCLEBIN = 'NO'").
		Scan(&overviews).
		Error
	return
}

func (t *TenantService) GetTenantParameters(tenantName string, filter string) (parameters []oceanbase.GvObParameter, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	tenantIdQuery := db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", tenantName)
	err = db.Table(GV_OB_PARAMETERS).
		Select("DISTINCT NAME, VALUE, DATA_TYPE, INFO, EDIT_LEVEL").
		Where("NAME LIKE ? AND SCOPE = 'tenant' AND TENANT_ID = (?)", filter, tenantIdQuery).
		Scan(&parameters).Error
	return
}

func (t *TenantService) GetTenantParameter(tenantId int, parameterName string) (parameter *oceanbase.GvObParameter, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(GV_OB_PARAMETERS).Select("DISTINCT NAME, VALUE, DATA_TYPE, INFO, EDIT_LEVEL").
		Where("NAME = ? AND SCOPE = 'tenant' AND TENANT_ID = (?)", parameterName, tenantId).
		Scan(&parameter).Error
	// retry for bad case for virtual table
	if parameter == nil && err == nil {
		err = db.Table(GV_OB_PARAMETERS).Where("NAME = ? AND SCOPE = 'tenant' AND TENANT_ID = (?)", parameterName, tenantId).Scan(&parameter).Error
	}
	return
}

func (t *TenantService) GetTenantVariables(tenantName string, filter string) (variables []oceanbase.CdbObSysVariable, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	tenantIdQuery := db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", tenantName)
	err = db.Table(CDB_OB_SYS_VARIABLES).
		Where("TENANT_ID = (?) AND NAME LIKE ?", tenantIdQuery, filter).Scan(&variables).Error
	return
}

func (t *TenantService) GetTenantVariable(tenantName string, variableName string) (variable *oceanbase.CdbObSysVariable, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	tenantIdQuery := db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", tenantName)
	err = db.Table(CDB_OB_SYS_VARIABLES).
		Where("TENANT_ID = (?) AND NAME = ?", tenantIdQuery, variableName).Scan(&variable).Error
	return
}

func (t *TenantService) GetTenantActiveAgent(tenantName string) (agent *meta.AgentInfo, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sql := "SELECT a.ip AS ip, a.port AS port FROM oceanbase.DBA_OB_UNITS u " +
		"JOIN oceanbase.DBA_OB_SERVERS s ON u.svr_ip = s.svr_ip AND u.svr_port = s.svr_port " +
		"JOIN ocs.all_agent a ON s.svr_ip = a.ip AND s.svr_port = a.rpc_port " +
		"WHERE u.tenant_id = (select tenant_id from oceanbase.DBA_OB_TENANTS where tenant_name = ?) AND s.status = 'ACTIVE' LIMIT 1"
	err = db.Raw(sql, tenantName).Scan(&agent).Error
	return
}

func (t *TenantService) GetTenantActiveServer(tenantName string) (server *oceanbase.OBServer, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sql := "SELECT s.* FROM oceanbase.DBA_OB_UNITS u " +
		"JOIN oceanbase.DBA_OB_SERVERS s ON u.svr_ip = s.svr_ip AND u.svr_port = s.svr_port " +
		"WHERE u.tenant_id = (select tenant_id from oceanbase.DBA_OB_TENANTS where tenant_name = ?) AND s.status = 'ACTIVE' LIMIT 1"
	err = db.Raw(sql, tenantName).Scan(&server).Error
	return
}

func (t *TenantService) IsTenantActiveAgent(tenantName string, ip string, rpcPort int) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	sql := "SELECT count(*) FROM oceanbase.DBA_OB_UNITS u " +
		"JOIN oceanbase.DBA_OB_SERVERS s ON u.svr_ip = s.svr_ip AND u.svr_port = s.svr_port " +
		"WHERE u.tenant_id = (select tenant_id from oceanbase.DBA_OB_TENANTS where tenant_name = ?) AND s.status = 'ACTIVE' and s.svr_ip = ? and s.svr_port = ?"
	var count int
	err = db.Raw(sql, tenantName, ip, rpcPort).Scan(&count).Error
	return count > 0, err
}

func (t *TenantService) GetObServerCapacityByZone(zone string) (servers []oceanbase.ObServerCapacity, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(GV_OB_SERVERS).Where("ZONE = ?", zone).Scan(&servers).Error
	return
}

func (t *TenantService) IsTimeZoneTableEmpty() (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(MYSQL_TIME_ZONE).Count(&count).Error
	return count == 0, err
}

func ParseLocalityToReplicaInfoMap(locality string) (map[string]string, error) {
	replicaInfoMap := make(map[string]string)
	parts := strings.Split(locality, ",")
	for _, part := range parts {
		if strings.Contains(part, "@") {
			segments := strings.Split(part, "@")
			if len(segments) == 2 {
				arr := strings.Split(segments[0], "{")
				if len(arr) == 2 {
					replicaInfoMap[strings.TrimSpace(segments[1])] = strings.TrimSpace(arr[0])
				} else {
					replicaInfoMap[strings.TrimSpace(segments[1])] = strings.TrimSpace(segments[0])
				}
			}
		} else {
			return nil, errors.Occur(errors.ErrUnexpected, "Invalid locality format.")
		}
	}
	return replicaInfoMap, nil
}

// GetTenantReplicaInfoMap return a map, key is zone, value is replica type
func (t *TenantService) GetTenantReplicaInfoMap(tenantId int) (map[string]string, error) {
	locality, err := t.GetTenantLocality(tenantId)
	if err != nil {
		return nil, errors.Wrap(err, "Get tenant locality failed.")
	}
	return ParseLocalityToReplicaInfoMap(locality)
}

func (s *TenantService) GetAllUserTenants() (res []oceanbase.DbaObTenant, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where(" TENANT_TYPE = ? ", constant.TENANT_TYPE_USER).Scan(&res).Error
	return
}

func (s *TenantService) GetTenantByID(id int) (res *oceanbase.DbaObTenant, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where("tenant_id = ?", id).Scan(&res).Error
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

func (s *TenantService) GetUnitConfigByName(name string) (res *oceanbase.DbaObUnitConfig, err error) {
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

func (s *TenantService) GetResourcePoolsByName(poolName string) (res *oceanbase.DbaObResourcePool, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_RESOURCE_POOLS).Where("name = ?", poolName).Scan(&res).Error
	return
}

func (s *TenantService) GetResourcePoolsByTenantID(tenantID int) (res []oceanbase.DbaObResourcePool, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_RESOURCE_POOLS).Where("TENANT_ID = ?", tenantID).Scan(&res).Error
	return
}

func (s *TenantService) GetResourcePoolsNameByTenantID(tenantID int) (res []string, err error) {
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
