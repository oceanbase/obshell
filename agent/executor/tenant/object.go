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
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

func ListObjects(tenantName string, password *string) ([]bo.DbaObjectBo, error) {
	tenantInfo, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, errors.Wrapf(err, "get tenant '%s' info failed", tenantName)
	}
	db, err := GetConnectionWithTenantInfo(tenantInfo, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, err
	}

	objects, err := tenantService.ListObjects(db)
	if err != nil {
		return nil, err
	}

	obObjects := make([]bo.DbaObjectBo, len(objects))
	for i := range objects {
		obObjects[i] = objects[i].ToDbObjectBo()
	}

	return obObjects, nil
}
