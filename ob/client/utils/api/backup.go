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

package api

import (
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/param"
)

func GetClusterBackupOverview() (res *param.BackupOverview, err error) {
	uri := constant.URI_OBCLUSTER_API_PREFIX + constant.URI_BACKUP + constant.URI_OVERVIEW
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &res)
	if err != nil {
		return nil, err
	}
	return
}

func GetTenantBackupOverview(name string) (res *param.TenantBackupOverview, err error) {
	uri := constant.URI_TENANT_API_PREFIX + "/" + name + constant.URI_BACKUP + constant.URI_OVERVIEW
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &res)
	if err != nil {
		return nil, err
	}
	return
}
