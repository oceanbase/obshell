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
	"github.com/oceanbase/obshell/param"
)

func RenameTenant(param param.RenameTenantParam) *errors.OcsAgentError {
	if err := checkTenantName(*param.NewName); err != nil {
		return errors.Occur(errors.ErrIllegalArgument, err.Error())
	}

	if _, err := checkTenantExistAndStatus(param.Name); err != nil {
		return err
	}

	if err := tenantService.RenameTenant(param.Name, *param.NewName); err != nil {
		return errors.Occur(errors.ErrBadRequest, err)
	}
	return nil
}