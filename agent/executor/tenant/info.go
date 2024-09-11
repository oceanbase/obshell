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
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func GetTenantsOverView() ([]oceanbase.DbaObTenant, *errors.OcsAgentError) {
	tenants, err := tenantService.GetTenantsOverView()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	return tenants, nil
}

func GetTenantInfo(tenantName string) (*bo.TenantInfo, *errors.OcsAgentError) {
	tenant, ocsErr := checkTenantExistAndStatus(tenantName)
	if ocsErr != nil {
		return nil, ocsErr
	}

	whitelist, err := tenantService.GetTenantVariable(tenantName, "ob_tcp_invited_nodes")
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	if whitelist == nil {
		return nil, errors.Occur(errors.ErrUnexpected, nil)
	}

	pools := make([]*bo.ResourcePoolWithUnit, 0)
	poolInfos, err := tenantService.GetTenantResourcePool(tenant.TenantID)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}

	for _, poolInfo := range poolInfos {
		unitConfig, err := unitService.GetUnitConfigById(poolInfo.UnitConfigId)
		if err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err.Error())
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

	return &bo.TenantInfo{
		Name:         tenant.TenantName,
		Id:           tenant.TenantID,
		CreatedTime:  tenant.CreatedTime,
		Mode:         tenant.Mode,
		Status:       tenant.Status,
		Locked:       tenant.Locked,
		PrimaryZone:  tenant.PrimaryZone,
		Locality:     tenant.Locality,
		InRecyclebin: tenant.InRecyclebin,
		Whitelist:    whitelist.Value,
		Pools:        pools,
	}, nil
}
