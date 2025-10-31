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

package observer

import (
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/param"
)

func GetVariables(filter string) ([]oceanbase.DbaObSysVariable, error) {
	if filter == "" {
		filter = "%"
	}
	variables, err := tenantService.GetVariables(filter)
	if err != nil {
		return nil, err
	}
	return variables, nil
}

func checkVariablesExist(vars map[string]interface{}) error {
	for k, v := range vars {
		if k == "" || v == nil {
			return errors.Occur(errors.ErrObEmptyVariable)
		}
		if exist, err := tenantService.IsVariableExist(k); err != nil {
			return err
		} else if !exist {
			return errors.Occur(errors.ErrObVariableNotExist, k)
		}
	}
	return nil
}

func SetVariables(param param.SetVariablesParam) error {
	for k, v := range param.Variables {
		if k == "" || v == nil {
			return errors.Occur(errors.ErrObEmptyVariable)
		}
	}

	if err := checkVariablesExist(param.Variables); err != nil {
		return err
	}

	transferNumber(param.Variables)

	if err := tenantService.SetVariables(param.Variables); err != nil {
		return err
	}

	return nil
}
