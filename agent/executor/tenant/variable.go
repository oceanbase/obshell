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
	"regexp"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func isUnkonwnTimeZoneErr(err error) bool {
	return err != nil && err.Error() == "unknown time zone"
}

func GetTenantVariables(tenantName string, filter string) ([]oceanbase.CdbObSysVariable, *errors.OcsAgentError) {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return nil, err
	}
	if filter == "" {
		filter = "%"
	}
	variables, err := tenantService.GetTenantVariables(tenantName, filter)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	return variables, nil
}

func GetTenantVariable(tenantName string, variableName string) (*oceanbase.CdbObSysVariable, *errors.OcsAgentError) {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return nil, err
	}
	variable, err := tenantService.GetTenantVariable(tenantName, variableName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	if variable == nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, "variable not found")
	}
	return variable, nil
}

func SetTenantVariables(tenantName string, variables map[string]interface{}) *errors.OcsAgentError {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return err
	}
	for k, v := range variables {
		if k == "" || v == nil {
			return errors.Occur(errors.ErrIllegalArgument, "variable name or value is empty")
		}
	}
	transferNumber(variables)
	if err := tenantService.SetTenantVariables(tenantName, variables); err != nil {
		if errors.IsUnkonwnTimeZoneErr(err) {
			if value, exist := variables[constant.VARIABLE_TIME_ZONE]; exist {
				return timeZoneErrorReporter(value, err)
			}
		}
		return errors.Occur(errors.ErrBadRequest, err)
	}

	return nil
}

func timeZoneErrorReporter(timeZone interface{}, err error) *errors.OcsAgentError {
	if v, ok := timeZone.(string); ok {
		pattern := `^[A-Za-z]+/[A-Za-z]+$`
		re := regexp.MustCompile(pattern)
		if re.MatchString(v) {
			if empty, _ := tenantService.IsTimeZoneTableEmpty(); empty {
				return errors.Occur(errors.ErrBadRequest, errors.Wrapf(err, "Please check whether the sys tenat has been import time zone info"))
			}
		}
	}
	return errors.Occur(errors.ErrBadRequest, err)
}
