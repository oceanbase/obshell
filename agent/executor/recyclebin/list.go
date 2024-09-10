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

import (
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func ListRecyclebinTenant() ([]oceanbase.DbaRecyclebin, *errors.OcsAgentError) {
	tenants, err := tenantService.GetRecycledTenant()
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "List recyclebin's tenants failed: %s", err.Error())
	}
	return tenants, nil
}
