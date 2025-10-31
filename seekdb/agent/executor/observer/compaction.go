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
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
)

func GetCompaction() (*bo.TenantCompaction, error) {
	tenantCompaction, err := tenantService.GetCompaction()
	if err != nil {
		return nil, err
	}
	return tenantCompaction.ToBO(), nil
}

func MajorCompaction() error {
	tenantCompaction, err := tenantService.GetCompaction()
	if err != nil {
		return err
	}
	if tenantCompaction.Status != "IDLE" {
		return errors.Occur(errors.ErrObTenantCompactionStatusNotIdle, tenantCompaction.Status)
	}

	err = tenantService.MajorCompaction()
	if err != nil {
		return err
	}
	return nil
}

func ClearCompactionError() error {
	if err := tenantService.ClearCompactionError(); err != nil {
		return err
	}
	return nil
}
