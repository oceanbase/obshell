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

package recyclebin

import "github.com/oceanbase/obshell/agent/errors"

func FlashbackTenant(name string, newName *string) error {
	objectName, err := tenantService.GetRecycledTenantObjectName(name)
	if err != nil {
		return errors.Wrapf(err, "Check tenant '%s' exist in recyclebin failed", name)
	} else if objectName == "" {
		return errors.Occur(errors.ErrObRecyclebinTenantNotExist, name)
	}

	var tenantName string
	if newName == nil {
		// check name is object_name or original_name
		originalName, err := tenantService.GetRecycledTenantOriginalName(objectName)
		if err != nil {
			return errors.Wrapf(err, "Get original name of tenant '%s' failed", name)
		}
		tenantName = originalName
	} else {
		tenantName = *newName
	}

	// check if tenantName is valid
	if exist, err := tenantService.IsTenantExist(tenantName); err != nil {
		return errors.Wrapf(err, "Check tenant '%s' exist failed", tenantName)
	} else if exist {
		return errors.Occur(errors.ErrObTenantExisted, tenantName)
	}

	if err := tenantService.FlashbackTenant(objectName, tenantName); err != nil {
		return errors.Wrapf(err, "Flashback tenant '%s' failed", name)
	}
	return nil
}
