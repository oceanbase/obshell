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
package unit

import (
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/parse"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func DropUnitConfig(name string) *errors.OcsAgentError {
	if exist, err := unitService.IsUnitConfigExist(name); err != nil {
		return errors.Occur(errors.ErrUnexpected, err.Error())
	} else if !exist {
		return nil
	}
	if err := unitService.DropUnit(name); err != nil {
		return errors.Occur(errors.ErrUnexpected, err.Error())
	}
	return nil
}

func validateCreateResourceUnitConfigParams(param param.CreateResourceUnitConfigParams) error {
	if *param.Name == "" {
		return errors.New("Name is empty.")
	}
	if _, pass := parse.CapacityParser(*param.MemorySize); !pass {
		return errors.Errorf("memory_size %s is illegal.", *param.MemorySize)
	}
	if param.LogDiskSize != nil {
		if _, pass := parse.CapacityParser(*param.LogDiskSize); !pass {
			return errors.Errorf("log_disk_size %s is illegal.", *param.LogDiskSize)
		}
	}

	if *param.MaxCpu <= 0 {
		return errors.New("max_cpu should be positive.")
	}

	if param.MinCpu != nil && *param.MinCpu <= 0 {
		return errors.New("min_cpu should be positive.")
	}

	if param.MinIops != nil && *param.MinIops <= 0 {
		return errors.New("min_iops should be positive.")
	}

	if param.MaxIops != nil && *param.MaxIops <= 0 {
		return errors.New("max_iops should be positive.")
	}

	if param.MinCpu != nil && *param.MaxCpu < *param.MinCpu {
		return errors.New("Incorrect arguments to min_cpu, min_cpu is greater than max_cpu.")
	}

	if *param.MaxCpu < constant.RESOURCE_UNIT_CONFIG_CPU_MINE {
		return errors.New("invalid max_cpu value, min value is 1.")
	}

	if param.MinCpu != nil && *param.MinCpu < constant.RESOURCE_UNIT_CONFIG_CPU_MINE {
		return errors.New("invalid min_cpu value, min value is 1.")
	}

	if param.MaxIops != nil && param.MinIops != nil && *param.MaxIops < *param.MinIops {
		return errors.New("Incorrect arguments to min_iops, min_iops is greater than max_iops.")
	}
	return nil
}

func CreateUnitConfig(param param.CreateResourceUnitConfigParams) *errors.OcsAgentError {
	if err := validateCreateResourceUnitConfigParams(param); err != nil {
		return errors.Occur(errors.ErrIllegalArgument, err.Error())
	}

	if exist, err := unitService.IsUnitConfigExist(*param.Name); err != nil {
		return errors.Occur(errors.ErrUnexpected, err.Error())
	} else if exist {
		return errors.Occurf(errors.ErrBadRequest, "Resource unit '%s' already exists.", *param.Name)
	}

	if err := unitService.CreateUnit(param); err != nil {
		return errors.Occur(errors.ErrUnexpected, err.Error())
	}
	return nil
}

func GetAllUnitConfig() ([]oceanbase.DbaObUnitConfig, error) {
	return unitService.GetAllUnitConfig()
}

func GetUnitConfig(name string) (*oceanbase.DbaObUnitConfig, *errors.OcsAgentError) {
	unitConfig, err := unitService.GetUnitConfigByName(name)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	if unitConfig == nil {
		return nil, errors.Occurf(errors.ErrBadRequest, "Resource unit config '%s' is not exist.", name)
	}
	return unitConfig, nil
}
