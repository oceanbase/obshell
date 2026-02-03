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

	// Optimization: Batch get read_only variable for all tenants in one query
	tenantNames := make([]string, 0, len(tenants))
	for i := range tenants {
		tenantNames = append(tenantNames, tenants[i].TenantName)
	}
	readOnlyMap, err := tenantService.GetTenantsVariableByNames(tenantNames, constant.VARIABLE_READ_ONLY)
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
		readOnly := readOnlyMap[tenants[i].TenantName]
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

	// Optimization: Batch get all tenant variables in one query instead of multiple separate queries
	// This reduces database round trips from 5 queries (with subqueries) to 1 query
	variableNames := []string{
		"ob_tcp_invited_nodes",
		constant.VARIABLE_READ_ONLY,
		constant.VARIABLE_LOWER_CASE_TABLE_NAMES,
		constant.VARIABLE_TIME_ZONE,
		"CHARACTER_SET_SERVER",
	}
	variablesMap, err := tenantService.GetTenantVariablesByNames(tenantName, variableNames)
	if err != nil {
		return nil, err
	}

	// Extract variables from map
	whitelist := variablesMap["ob_tcp_invited_nodes"]
	readOnly := variablesMap[constant.VARIABLE_READ_ONLY]
	lowerCaseTableNames := variablesMap[constant.VARIABLE_LOWER_CASE_TABLE_NAMES]
	timeZone := variablesMap[constant.VARIABLE_TIME_ZONE]
	charset := variablesMap["CHARACTER_SET_SERVER"]

	pools := make([]*bo.ResourcePoolWithUnit, 0)
	poolInfos, err := tenantService.GetTenantResourcePool(tenant.TenantID)
	if err != nil {
		return nil, err
	}

	// Optimization: Batch get all unit configs and pool servers in one query each
	if len(poolInfos) > 0 {
		// Collect all unit config IDs and pool IDs
		unitConfigIds := make([]int, 0, len(poolInfos))
		poolIds := make([]int, 0, len(poolInfos))
		for _, poolInfo := range poolInfos {
			unitConfigIds = append(unitConfigIds, poolInfo.UnitConfigId)
			poolIds = append(poolIds, poolInfo.ResourcePoolID)
		}

		// Batch get unit configs
		unitConfigsMap, err := unitService.GetUnitConfigsByIds(unitConfigIds)
		if err != nil {
			return nil, err
		}

		// Batch get pool servers
		serversMap, err := tenantService.GetTenantResourcePoolServersBatch(poolIds)
		if err != nil {
			return nil, err
		}

		// Build pools from batch results
		for _, poolInfo := range poolInfos {
			unitConfig, ok := unitConfigsMap[poolInfo.UnitConfigId]
			if !ok {
				return nil, fmt.Errorf("unit config not found for pool %s", poolInfo.Name)
			}

			observers := serversMap[poolInfo.ResourcePoolID]
			if observers == nil {
				observers = make([]oceanbase.OBServer, 0)
			}

			observerList := make([]string, 0, len(observers))
			for _, observer := range observers {
				observerList = append(observerList, fmt.Sprintf("%s:%d", observer.SvrIp, observer.SvrPort))
			}
			poolWithUnit := bo.ResourcePoolWithUnit{
				Name:       poolInfo.Name,
				Id:         poolInfo.ResourcePoolID,
				ZoneList:   poolInfo.ZoneList,
				ServerList: strings.Join(observerList, ","),
				UnitNum:    poolInfo.UnitNum,
				Unit:       oceanbase.ConvertDbaObUnitConfigToObUnit(unitConfig),
			}
			pools = append(pools, &poolWithUnit)
		}
	}

	var lclOpInterval string
	if err := observerService.GetOBParatemerByName("_lcl_op_interval", &lclOpInterval); err != nil {
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
		Comment:           tenant.Comment,
		Pools:             pools,
		ConnectionStrings: []bo.ObproxyAndConnectionString{connectionStr},
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
	if lclOpInterval != "0ms" && lclOpInterval != "0" {
		tenantInfo.DeadLockDetectionEnabled = true
	}

	return tenantInfo, nil
}
