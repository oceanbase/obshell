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
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/parse"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func DropUnitConfig(name string) error {
	if exist, err := unitService.IsUnitConfigExist(name); err != nil {
		return err
	} else if !exist {
		return nil
	}
	if err := unitService.DropUnit(name); err != nil {
		return err
	}
	return nil
}

func validateCreateResourceUnitConfigParams(param param.CreateResourceUnitConfigParams) error {
	if *param.Name == "" {
		return errors.Occur(errors.ErrObResourceUnitConfigNameEmpty)
	}
	if _, pass := parse.CapacityParser(*param.MemorySize); !pass {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "memory_size", *param.MemorySize)
	}
	if param.LogDiskSize != nil {
		if _, pass := parse.CapacityParser(*param.LogDiskSize); !pass {
			return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "log_disk_size", *param.LogDiskSize)
		}
	}

	if *param.MaxCpu <= 0 {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "max_cpu", "max_cpu should be positive.")
	}

	if param.MinCpu != nil && *param.MinCpu <= 0 {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "min_cpu", "min_cpu should be positive.")
	}

	if param.MinIops != nil && *param.MinIops <= 0 {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "min_iops", "min_iops should be positive.")
	}

	if param.MaxIops != nil && *param.MaxIops <= 0 {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "max_iops", "max_iops should be positive.")
	}

	if param.MinCpu != nil && *param.MaxCpu < *param.MinCpu {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "min_cpu", "min_cpu should not be greater than max_cpu.")
	}

	if *param.MaxCpu < constant.RESOURCE_UNIT_CONFIG_CPU_MINE {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "max_cpu", "min value is 1.")
	}

	if param.MinCpu != nil && *param.MinCpu < constant.RESOURCE_UNIT_CONFIG_CPU_MINE {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "min_cpu", "min value is 1.")
	}

	if param.MaxIops != nil && param.MinIops != nil && *param.MaxIops < *param.MinIops {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "min_iops", "min_iops should not be greater than max_iops.")
	}
	return nil
}

func CreateUnitConfig(param param.CreateResourceUnitConfigParams) error {
	if err := validateCreateResourceUnitConfigParams(param); err != nil {
		return err
	}

	if exist, err := unitService.IsUnitConfigExist(*param.Name); err != nil {
		return err
	} else if exist {
		return errors.Occur(errors.ErrObResourceUnitConfigExisted, *param.Name)
	}

	if err := unitService.CreateUnit(param); err != nil {
		return err
	}
	return nil
}

func GetAllUnitConfig() ([]oceanbase.DbaObUnitConfig, error) {
	return unitService.GetAllUnitConfig()
}

func GetUnitConfig(name string) (*oceanbase.DbaObUnitConfig, error) {
	unitConfig, err := unitService.GetUnitConfigByName(name)
	if err != nil {
		return nil, err
	}
	if unitConfig == nil {
		return nil, errors.Occur(errors.ErrObResourceUnitConfigNotExist, name)
	}
	return unitConfig, nil
}
