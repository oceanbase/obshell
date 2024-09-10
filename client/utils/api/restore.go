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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/param"
)

func GetTenantRestoreOverview(name string) (res *param.RestoreOverview, err error) {
	uri := constant.URI_TENANT_API_PREFIX + "/" + name + constant.URI_RESTORE + constant.URI_OVERVIEW
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &res)
	if err != nil {
		return nil, err
	}
	log.Infof("%v", res)
	return
}
