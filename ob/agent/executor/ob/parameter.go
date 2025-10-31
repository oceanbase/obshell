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

package ob

import (
	"fmt"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/param"
	"github.com/oceanbase/obshell/ob/utils"
)

func GetAllParameters() ([]bo.ClusterParameter, error) {
	obParameters, err := obclusterService.GetAllUnhiddenParameters()
	if err != nil {
		return nil, err
	}

	tenantIdToNameMap, err := tenantService.GetAllNotMetaTenantIdToNameMap()
	if err != nil {
		return nil, err
	}

	parametersMap := make(map[string]*bo.ClusterParameter)
	for _, obParameter := range obParameters {
		if _, exists := parametersMap[obParameter.Name]; !exists {
			parametersMap[obParameter.Name] = &bo.ClusterParameter{
				Name:         obParameter.Name,
				Scope:        obParameter.Scope,
				EditLevel:    obParameter.EditLevel,
				DefaultValue: obParameter.DefaultValue,
				Section:      obParameter.Section,
				DataType:     obParameter.DataType,
				Info:         obParameter.Info,
				ServerValue:  make([]bo.ObParameterValue, 0),
			}
		}

		var tenantName string
		if obParameter.Scope == PARAMETER_SCOPE_TENANT {
			var ok bool
			if tenantName, ok = tenantIdToNameMap[obParameter.TenantId]; !ok {
				continue
			}
		}

		// Set server value
		parametersMap[obParameter.Name].ServerValue = append(parametersMap[obParameter.Name].ServerValue, bo.ObParameterValue{
			SvrIp:      obParameter.SvrIp,
			SvrPort:    obParameter.SvrPort,
			Zone:       obParameter.Zone,
			TenantId:   obParameter.TenantId,
			TenantName: tenantName,
			Value:      obParameter.Value,
		})

		if !utils.ContainsString(parametersMap[obParameter.Name].Values, obParameter.Value) {
			parametersMap[obParameter.Name].Values = append(parametersMap[obParameter.Name].Values, obParameter.Value)
		}
		parametersMap[obParameter.Name].IsSingleValue = len(parametersMap[obParameter.Name].Values) == 1

		// Set tenant value
		if obParameter.Scope == PARAMETER_SCOPE_TENANT {
			if len(parametersMap[obParameter.Name].TenantValue) == 0 {
				parametersMap[obParameter.Name].TenantValue = make([]bo.TenantParameterValue, 0)
			}
			parametersMap[obParameter.Name].TenantValue = append(parametersMap[obParameter.Name].TenantValue, bo.TenantParameterValue{
				TenantId:   obParameter.TenantId,
				TenantName: tenantName,
				Value:      obParameter.Value,
			})
		}
	}

	res := make([]bo.ClusterParameter, 0, len(parametersMap))
	for _, parameter := range parametersMap {
		res = append(res, *parameter)
	}

	return res, nil

}

func SetObclusterParameters(params []param.SetSingleObclusterParameterParam) error {
	if len(params) == 0 {
		return nil
	}

	for _, param := range params {
		if err := checkSetSingleObclusterParameterParam(param); err != nil {
			return err
		}
	}

	for _, param := range params {
		setParameterParams := buildSetParameterParam(param)
		for _, setParameterParam := range setParameterParams {
			if err := obclusterService.SetParameter(setParameterParam); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildSetParameterParam(setSingleObclusterParameterParam param.SetSingleObclusterParameterParam) []param.SetParameterParam {
	setParameterParams := make([]param.SetParameterParam, 0)
	if setSingleObclusterParameterParam.Scope == PARAMETER_SCOPE_CLUSTER {
		if len(setSingleObclusterParameterParam.Zones) != 0 {
			for _, zone := range setSingleObclusterParameterParam.Zones {
				setParameterParams = append(setParameterParams, param.SetParameterParam{
					Name:  setSingleObclusterParameterParam.Name,
					Value: setSingleObclusterParameterParam.Value,
					Zone:  zone,
				})
			}
		} else if len(setSingleObclusterParameterParam.Servers) != 0 {
			for _, server := range setSingleObclusterParameterParam.Servers {
				setParameterParams = append(setParameterParams, param.SetParameterParam{
					Name:   setSingleObclusterParameterParam.Name,
					Value:  setSingleObclusterParameterParam.Value,
					Server: server,
				})
			}
		} else {
			setParameterParams = append(setParameterParams, param.SetParameterParam{
				Name:  setSingleObclusterParameterParam.Name,
				Value: setSingleObclusterParameterParam.Value,
			})
		}
	} else if setSingleObclusterParameterParam.Scope == PARAMETER_SCOPE_TENANT {
		if setSingleObclusterParameterParam.AllUserTenant {
			setParameterParams = append(setParameterParams, param.SetParameterParam{
				Name:   setSingleObclusterParameterParam.Name,
				Value:  setSingleObclusterParameterParam.Value,
				Tenant: "ALL_USER",
			})
		} else if len(setSingleObclusterParameterParam.Tenants) != 0 {
			for _, tenant := range setSingleObclusterParameterParam.Tenants {
				setParameterParams = append(setParameterParams, param.SetParameterParam{
					Name:   setSingleObclusterParameterParam.Name,
					Value:  setSingleObclusterParameterParam.Value,
					Tenant: tenant,
				})
			}
		} else {
			setParameterParams = append(setParameterParams, param.SetParameterParam{
				Name:  setSingleObclusterParameterParam.Name,
				Value: setSingleObclusterParameterParam.Value,
			})
		}
	}
	return setParameterParams
}

func checkSetSingleObclusterParameterParam(param param.SetSingleObclusterParameterParam) error {
	if param.Scope != PARAMETER_SCOPE_CLUSTER && param.Scope != PARAMETER_SCOPE_TENANT {
		return errors.Occur(errors.ErrObParameterScopeInvalid, param.Scope)
	}
	if len(param.Zones) != 0 && len(param.Servers) != 0 {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "zones or servers", "zones and servers cannot be set at the same time")
	}
	if param.Scope == PARAMETER_SCOPE_TENANT {
		if len(param.Tenants) != 0 && param.AllUserTenant {
			return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "tenants or all_user_tenant", "parameter tenants and all_user_tenant cannot be set at the same time")
		}
		// if len(param.Tenants) == 0 && !param.AllUserTenant, set for sys tenant.
	} else if param.Scope == PARAMETER_SCOPE_CLUSTER {
		if len(param.Tenants) != 0 || param.AllUserTenant {
			return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "tenants or all_user_tenant", fmt.Sprintf("parameter tenants and all_user_tenant cannot be set when scope is %s", PARAMETER_SCOPE_CLUSTER))
		}
	}
	return nil
}
