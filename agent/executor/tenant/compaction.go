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
	"sort"
	"time"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

func GetTenantCompaction(tenantName string) (*bo.TenantCompaction, *errors.OcsAgentError) {
	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Get tenant '%s' failed.", tenantName)
	}
	if tenant == nil {
		return nil, errors.Occurf(errors.ErrBadRequest, "Tenant '%s' is not exist.", tenantName)
	}
	tenantCompaction, err := tenantService.GetTenantCompaction(tenant.TenantID)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, err.Error())
	}
	return tenantCompaction.ToBO(), nil
}

func TenantMajorCompaction(tenantName string) *errors.OcsAgentError {
	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Get tenant '%s' failed.", tenantName)
	}
	if tenant == nil {
		return errors.Occurf(errors.ErrBadRequest, "Tenant '%s' is not exist.", tenantName)
	}

	tenantCompaction, err := tenantService.GetTenantCompaction(tenant.TenantID)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, err.Error())
	}
	if tenantCompaction.Status != "IDLE" {
		return errors.Occurf(errors.ErrIllegalArgument, "Tenant '%s' is in '%s' status, operation not allowed", tenantName, tenantCompaction.Status)
	}

	err = tenantService.TenantMajorCompaction(tenantName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, err.Error())
	}
	return nil
}

func ClearTenantCompactionError(tenantName string) *errors.OcsAgentError {
	if err := tenantService.ClearTenantCompactionError(tenantName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, err.Error())
	}
	return nil
}

func GetTopCompactions(top int) ([]bo.TenantCompactionHistory, *errors.OcsAgentError) {
	tenantCompactions, err := tenantService.GetAllMajorCompactions()
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, err.Error())
	}
	tenantIdToNameMap, err := tenantService.GetAllNotMetaTenantIdToNameMap()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	tenantCompactionHistories := make([]bo.TenantCompactionHistory, 0)
	for _, tenantCompaction := range tenantCompactions {
		tenantName, ok := tenantIdToNameMap[tenantCompaction.TenantId]
		if !ok {
			continue
		}
		tenantCompactionHistory := bo.TenantCompactionHistory{
			TenantId:       tenantCompaction.TenantId,
			StartTime:      tenantCompaction.StartTime,
			LastFinishTime: tenantCompaction.LastFinishTime,
			Status:         tenantCompaction.Status,
			TenantName:     tenantName,
		}
		if !tenantCompactionHistory.StartTime.After(tenantCompactionHistory.LastFinishTime) {
			tenantCompactionHistory.CostTime = int64(tenantCompactionHistory.LastFinishTime.Sub(tenantCompactionHistory.StartTime) / time.Second)
		} else {
			timeNow, err := obclusterService.GetCurrentTimestamp()
			if err != nil {
				return nil, errors.Occurf(errors.ErrUnexpected, err.Error())
			}
			tenantCompactionHistory.CostTime = int64(timeNow.Sub(tenantCompactionHistory.StartTime) / time.Second)
		}
		tenantCompactionHistories = append(tenantCompactionHistories, tenantCompactionHistory)
	}
	// sort by cost time
	sort.Slice(tenantCompactionHistories, func(i, j int) bool {
		return tenantCompactionHistories[i].CostTime > tenantCompactionHistories[j].CostTime
	})
	if len(tenantCompactionHistories) > top {
		tenantCompactionHistories = tenantCompactionHistories[:top]
	}
	return tenantCompactionHistories, nil
}
