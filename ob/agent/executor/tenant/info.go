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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

func GetTenantsOverView(mode string) ([]oceanbase.TenantOverview, error) {
	if mode != "" && mode != constant.MYSQL_MODE && mode != constant.ORACLE_MODE {
		return nil, errors.Occur(errors.ErrObTenantModeNotSupported, mode)
	}
	tenants, err := tenantService.GetTenantsOverViewByMode(mode)
	if err != nil {
		return nil, err
	}
	tenantOverviews := make([]oceanbase.TenantOverview, 0)
	for i := range tenants {
		connectionStr := bo.ObproxyAndConnectionString{
			Type: constant.OB_CONNECTION_TYPE_DIRECT,
		}
		if tenants[i].Mode == constant.ORACLE_MODE {
			connectionStr.ConnectionString = fmt.Sprintf("obclient -h%s -P%d -uSYS@%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, tenants[i].TenantName)
		} else {
			connectionStr.ConnectionString = fmt.Sprintf("obclient -h%s -P%d -uroot@%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, tenants[i].TenantName)
		}
		connectionStrs := make([]bo.ObproxyAndConnectionString, 0)
		connectionStrs = append(connectionStrs, connectionStr)
		readOnly, err := tenantService.GetTenantVariable(tenants[i].TenantName, constant.VARIABLE_READ_ONLY)
		if err != nil {
			return nil, err
		}
		if readOnly != nil {
			tenants[i].ReadOnly = (readOnly.Value == "1")
		}
		tenantOverviews = append(tenantOverviews, oceanbase.TenantOverview{
			DbaObTenant:       tenants[i],
			ConnectionStrings: connectionStrs,
		})
	}
	return tenantOverviews, nil
}

func GetTenantInfo(tenantName string) (*bo.TenantInfo, error) {
	tenant, err := checkTenantExist(tenantName)
	if err != nil {
		return nil, err
	}

	whitelist, err := tenantService.GetTenantVariable(tenantName, "ob_tcp_invited_nodes")
	if err != nil {
		return nil, err
	}

	pools := make([]*bo.ResourcePoolWithUnit, 0)
	poolInfos, err := tenantService.GetTenantResourcePool(tenant.TenantID)
	if err != nil {
		return nil, err
	}

	for _, poolInfo := range poolInfos {
		unitConfig, err := unitService.GetUnitConfigById(poolInfo.UnitConfigId)
		if err != nil {
			return nil, err
		}
		poolWithUnit := bo.ResourcePoolWithUnit{
			Name:     poolInfo.Name,
			Id:       poolInfo.ResourcePoolID,
			ZoneList: poolInfo.ZoneList,
			UnitNum:  poolInfo.UnitNum,
			Unit:     oceanbase.ConvertDbaObUnitConfigToObUnit(unitConfig),
		}
		pools = append(pools, &poolWithUnit)
	}

	readOnly, err := tenantService.GetTenantVariable(tenantName, constant.VARIABLE_READ_ONLY)
	if err != nil {
		return nil, err
	}

	lowerCaseTableNames, err := tenantService.GetTenantVariable(tenantName, constant.VARIABLE_LOWER_CASE_TABLE_NAMES)
	if err != nil {
		return nil, err
	}

	timeZone, err := tenantService.GetTenantVariable(tenantName, constant.VARIABLE_TIME_ZONE)
	if err != nil {
		return nil, err
	}

	connectionStr := bo.ObproxyAndConnectionString{
		Type: constant.OB_CONNECTION_TYPE_DIRECT,
		// the host may be not in the tcp_invited_nodes
	}
	if tenant.Mode == constant.ORACLE_MODE {
		connectionStr.ConnectionString = fmt.Sprintf("obclient -h%s -P%d -uSYS@%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, tenantName)
	} else {
		connectionStr.ConnectionString = fmt.Sprintf("obclient -h%s -P%d -uroot@%s -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, tenantName)
	}

	tenantInfo := &bo.TenantInfo{
		Name:              tenant.TenantName,
		Id:                tenant.TenantID,
		CreatedTime:       tenant.CreatedTime,
		Mode:              tenant.Mode,
		Status:            tenant.Status,
		Locked:            tenant.Locked,
		PrimaryZone:       tenant.PrimaryZone,
		Locality:          tenant.Locality,
		InRecyclebin:      tenant.InRecyclebin,
		Pools:             pools,
		ConnectionStrings: []bo.ObproxyAndConnectionString{connectionStr},
	}

	charset, err := tenantService.GetTenantVariable(tenantName, "CHARACTER_SET_SERVER")
	if err != nil {
		return nil, err
	}
	collatoinMap, err := obclusterService.GetCollationMap()
	if err != nil {
		return nil, err
	}
	if charset != nil && collatoinMap != nil {
		id, _ := strconv.Atoi(charset.Value)
		collation, ok := collatoinMap[id]
		if ok {
			tenantInfo.Collation = collation.Collation
			tenantInfo.Charset = collation.Charset
		}
	}
	if whitelist != nil {
		tenantInfo.Whitelist = whitelist.Value
	}
	if readOnly != nil {
		tenantInfo.ReadOnly = (readOnly.Value == "1")
	}
	if timeZone != nil {
		tenantInfo.TimeZone = timeZone.Value
	}
	if lowerCaseTableNames != nil {
		tenantInfo.LowercaseTableNames = lowerCaseTableNames.Value
	}
	return tenantInfo, nil
}
