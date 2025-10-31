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
)

func GetParameters(filter string) ([]oceanbase.GvObParameter, error) {
	if filter == "" {
		filter = "%"
	}
	parameters, err := tenantService.GetParameters(filter)
	if err != nil {
		return nil, err
	}

	return parameters, nil
}

func SetParameters(parameters map[string]interface{}) error {
	if err := checkParameters(parameters); err != nil {
		return err
	}

	transferNumber(parameters)
	if err := tenantService.SetParameters(parameters); err != nil {
		return errors.Wrap(err, "set parameters failed")
	}
	return nil
}

// transferNumber transfer float64(Scientific Notation) to int64 or float64
func transferNumber(mp map[string]interface{}) {
	for key, value := range mp {
		if number, ok := value.(float64); ok {
			if number == float64(int(number)) {
				mp[key] = int(number)
			} else {
				mp[key] = float64(number)
			}
		}
	}
}

func checkParameters(parameters map[string]interface{}) error {
	for k, v := range parameters {
		if k == "" || v == nil {
			return errors.Occur(errors.ErrObEmptyParameter)
		}
	}

	for k := range parameters {
		// Check whether the parameter is exist.
		if param, err := tenantService.GetParameter(k); err != nil {
			return err
		} else if param == nil {
			return errors.Occur(errors.ErrObParameterNotExist, k)
		}
	}

	return nil
}
