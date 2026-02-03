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
	"strconv"
	"strings"

	obdriver "github.com/oceanbase/go-oceanbase-driver"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	bo "github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
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
		return errors.Occur(errors.ErrCommonUnexpected, "Unexpected error when validate variables.")
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
			if number, err := strconv.Atoi(val); err == nil {
				variablesSq += fmt.Sprintf(", "+k+"= %v", number)
			} else if float, err := strconv.ParseFloat(val, 64); err == nil {
				variablesSq += fmt.Sprintf(", "+k+"= %v", float)
			} else {
				variablesSq += fmt.Sprintf(", "+k+"= `%v`", val)
			}
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

func (t *TenantService) IsTenantLocked(name string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(DBA_OB_TENANTS).Where("TENANT_NAME = ? AND LOCKED = 'YES'", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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

func (t *TenantService) SetTenantVariablesWithTenant(tenantName, password string, variables map[string]interface{}) error {
	if len(variables) == 0 {
		return nil
	}
	mode, err := t.GetTenantMode(tenantName)
	if err != nil {
		return err
	}
	tempDb, err := oceanbasedb.LoadGormWithTenant(tenantName, password, mode)
	if err != nil {
		return err
	}
	defer func() {
		db, _ := tempDb.DB()
		if db != nil {
			db.Close()
		}
	}()
	variablesSql := ""
	for k, v := range variables {
		if val, ok := v.(string); ok {
			if mode == constant.ORACLE_MODE {
				variablesSql += fmt.Sprintf(", GLOBAL "+k+"= '%v'", val)
			} else {
				variablesSql += fmt.Sprintf(", GLOBAL "+k+"= `%v`", val)
			}
		} else {
			variablesSql += fmt.Sprintf(", GLOBAL "+k+"= %v", v)
		}
	}
	sqlText := fmt.Sprintf("SET %s", variablesSql[1:])
	return tempDb.Exec(sqlText).Error
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

func (t *TenantService) GetTenantsOverViewByMode(mode string) (overviews []oceanbase.DbaObTenant, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	if mode == "" {
		err = db.Table(DBA_OB_TENANTS).
			Where("TENANT_TYPE != 'META' AND IN_RECYCLEBIN = 'NO'").
			Scan(&overviews).
			Error
		return
	}
	err = db.Table(DBA_OB_TENANTS).
		Where("TENANT_TYPE != 'META' AND IN_RECYCLEBIN = 'NO' AND COMPATIBILITY_MODE = ?", mode).
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

func (s *TenantService) GetDinstinctParameterValue(parameterName string) ([]string, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var values []string
	err = oceanbaseDb.Model(oceanbase.GvObParameter{}).Where("NAME = ?", parameterName).Distinct().Pluck("VALUE", &values).Error
	return values, err
}

func (t *TenantService) GetTenantVariables(tenantName string, filter string) (variables []oceanbase.CdbObSysVariable, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	tenantIdQuery := db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", tenantName)
	err = db.Table(CDB_OB_SYS_VARIABLES).
		Where("TENANT_ID = (?) AND NAME LIKE ?", tenantIdQuery, filter).Scan(&variables).Error
	if err != nil {
		return nil, err
	}

	return
}

func (t *TenantService) IsVariableExist(variableName string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(CDB_OB_SYS_VARIABLES).Where("NAME = ?", variableName).Where("TENANT_ID = ?", constant.TENANT_SYS_ID).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) GetTenantVariable(tenantName string, variableName string) (variable *oceanbase.CdbObSysVariable, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	tenantIdQuery := db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", tenantName)
	err = db.Table(CDB_OB_SYS_VARIABLES).
		Where("TENANT_ID = (?) AND NAME = ?", tenantIdQuery, variableName).Scan(&variable).Error
	if err != nil {
		return nil, err
	}

	return
}

// GetTenantVariablesByNames batch gets multiple tenant variables by names in one query
// This is more efficient than calling GetTenantVariable multiple times
func (t *TenantService) GetTenantVariablesByNames(tenantName string, variableNames []string) (variablesMap map[string]*oceanbase.CdbObSysVariable, err error) {
	if len(variableNames) == 0 {
		return make(map[string]*oceanbase.CdbObSysVariable), nil
	}

	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	// Get tenant_id once
	tenantIdQuery := db.Table(DBA_OB_TENANTS).Select("TENANT_ID").Where("TENANT_NAME = ?", tenantName)

	// Query all variables in one go
	var variables []oceanbase.CdbObSysVariable
	err = db.Table(CDB_OB_SYS_VARIABLES).
		Where("TENANT_ID = (?) AND NAME IN ?", tenantIdQuery, variableNames).Scan(&variables).Error
	if err != nil {
		return nil, err
	}

	// Build map for easy lookup
	variablesMap = make(map[string]*oceanbase.CdbObSysVariable)
	for i := range variables {
		variablesMap[variables[i].Name] = &variables[i]
	}

	return
}

// GetTenantsVariableByNames batch gets a specific variable for multiple tenants in one query
// This is more efficient than calling GetTenantVariable multiple times for different tenants
func (t *TenantService) GetTenantsVariableByNames(tenantNames []string, variableName string) (variablesMap map[string]*oceanbase.CdbObSysVariable, err error) {
	if len(tenantNames) == 0 {
		return make(map[string]*oceanbase.CdbObSysVariable), nil
	}

	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	// Use JOIN to get tenant variables in a single query
	var results []struct {
		TenantName string `gorm:"column:TENANT_NAME"`
		Name       string `gorm:"column:NAME"`
		Value      string `gorm:"column:VALUE"`
		Info       string `gorm:"column:INFO"`
	}
	err = db.Table(DBA_OB_TENANTS+" t").
		Select("t.TENANT_NAME, v.NAME, v.VALUE, v.INFO").
		Joins("JOIN "+CDB_OB_SYS_VARIABLES+" v ON t.TENANT_ID = v.TENANT_ID").
		Where("t.TENANT_NAME IN ? AND v.NAME = ?", tenantNames, variableName).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Build map for easy lookup by tenant name
	variablesMap = make(map[string]*oceanbase.CdbObSysVariable)
	for i := range results {
		variablesMap[results[i].TenantName] = &oceanbase.CdbObSysVariable{
			Name:  results[i].Name,
			Value: results[i].Value,
			Info:  results[i].Info,
		}
	}

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

func (t *TenantService) GetTenantObserverList(tenantId int) (servers []oceanbase.OBServer, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sql := "SELECT s.svr_ip, s.svr_port FROM oceanbase.DBA_OB_UNITS u " +
		"JOIN oceanbase.DBA_OB_SERVERS s ON u.svr_ip = s.svr_ip AND u.svr_port = s.svr_port " +
		"WHERE u.tenant_id = ? "

	err = db.Raw(sql, tenantId).Scan(&servers).Error
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
			return nil, errors.Occur(errors.ErrObTenantLocalityFormatUnexpected, locality)
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

func (s *TenantService) GetAllNotMetaTenantIdToNameMap() (res map[int]string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var tenants []oceanbase.DbaObTenant
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Where(" TENANT_TYPE != ? ", constant.TENANT_TYPE_META).Scan(&tenants).Error
	if err != nil {
		return nil, err
	}
	res = make(map[int]string)
	for _, tenant := range tenants {
		res[tenant.TenantID] = tenant.TenantName
	}
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

func (s *TenantService) CheckModuleData(tenantName string, moduleName string) (pass bool, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}

	sql := fmt.Sprintf("alter system check module data module=%s tenant=%s", moduleName, tenantName)
	err = oceanbaseDb.Exec(sql).Error
	if err != nil {
		if dbErr, ok := err.(*obdriver.MySQLError); ok && dbErr.Number == 4025 {
			pass = false
			err = nil
		}
	} else {
		pass = true
	}
	return
}

func (s *TenantService) LoadModuleData(tenantName string, moduleName string) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}

	return oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET SESSION ob_query_timeout=1000000000").Error; err != nil {
			return err
		}

		sql := fmt.Sprintf("alter system load module data module=%s tenant=%s", moduleName, tenantName)
		return tx.Exec(sql).Error
	})
}

func (s *TenantService) GetTenantCompaction(tenantId int) (compaction *oceanbase.CdbObMajorCompaction, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Model(oceanbase.CdbObMajorCompaction{}).Where("tenant_id = ?", tenantId).Scan(&compaction).Error
	return
}

func (s *TenantService) GetAllMajorCompactions() (compactions []oceanbase.CdbObMajorCompaction, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Model(oceanbase.CdbObMajorCompaction{}).Scan(&compactions).Error
	return
}

func (s *TenantService) TenantMajorCompaction(tenantName string) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec(fmt.Sprintf("ALTER SYSTEM MAJOR FREEZE TENANT = %s", tenantName)).Error
}

func (s *TenantService) ClearTenantCompactionError(tenantName string) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec(fmt.Sprintf("ALTER SYSTEM CLEAR MERGE ERROR TENANT = %s", tenantName)).Error
}

func (s *TenantService) GetSlowSqlRank(top int, startTime int64, endTime int64) (res []bo.TenantSlowSqlCount, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sql :=
		"select tenant_id, tenant_name, count(distinct db_id, sql_id) as count " +
			" from oceanbase.GV$OB_SQL_AUDIT where" +
			" char_length(tenant_name) != 0 and elapsed_time > ? " +
			" and (request_time + elapsed_time) > ? and (request_time + elapsed_time) < ?" +
			" and tenant_name NOT LIKE '%$%'" +
			" group by tenant_name" +
			" order by count desc limit ?"
	err = oceanbaseDb.Raw(sql, constant.SLOW_SQL_THRESHOLD, startTime, endTime, top).Find(&res).Error
	return
}

func (s *TenantService) GetTenantDataDiskUsageMap() (dataDiskUsageMap map[int]int64, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sql := "select coalesce(t1.tenant_id, -1) as tenant_id, sum(data_disk_in_use) as data_disk_in_use from (select t1.unit_id, t1.svr_ip, t1.svr_port, t2.tenant_id, t1.data_disk_in_use from (select  unit_id, svr_ip, svr_port, sum(data_disk_in_use) as data_disk_in_use from oceanbase.gv$ob_units  group by unit_id ) t1 join oceanbase.dba_ob_units t2 on t1.unit_id = t2.unit_id) t1 join oceanbase.dba_ob_tenants t2 on t1.tenant_id = t2.tenant_id where tenant_type <>'meta' group by tenant_id"
	var results []struct {
		TenantId      int   `gorm:"column:tenant_id"`
		DataDiskInUse int64 `gorm:"column:data_disk_in_use"`
	}
	err = oceanbaseDb.Raw(sql).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	dataDiskUsageMap = make(map[int]int64)
	for _, result := range results {
		dataDiskUsageMap[result.TenantId] = result.DataDiskInUse
	}
	return dataDiskUsageMap, nil
}

func (s *TenantService) GetTenantMode(tenantName string) (string, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}
	var mode string
	err = oceanbaseDb.Table(DBA_OB_TENANTS).Select("COMPATIBILITY_MODE").Where("tenant_name = ?", tenantName).Scan(&mode).Error
	return mode, err
}

func (s *TenantService) GetUnfreshedTenants() ([]string, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	sql := "SELECT DISTINCT tenant_name FROM (SELECT tenant_id, tenant_name, svr_ip, svr_port FROM oceanbase.DBA_OB_TENANTS, oceanbase.DBA_OB_SERVERS) AS ts" +
		" LEFT JOIN oceanbase.GV$OB_SERVER_SCHEMA_INFO AS os ON ts.tenant_id = os.tenant_id AND ts.svr_ip = os.svr_ip AND ts.svr_port = os.svr_port" +
		" WHERE refreshed_schema_version IS NULL OR refreshed_schema_version <= 1 OR refreshed_schema_version % 8 != 0"
	var tenants []string
	err = oceanbaseDb.Raw(sql).Scan(&tenants).Error
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *TenantService) GetMajorCompactionTenantCount() (int, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, errors.Wrap(err, "Get major compaction tenant count failed")
	}
	var count int64
	err = oceanbaseDb.Model(oceanbase.CdbObMajorCompaction{}).Where("STATUS != 'IDLE'").Count(&count).Error
	return int(count), err
}
