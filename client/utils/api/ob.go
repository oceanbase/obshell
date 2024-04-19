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
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
)

func CheckOBMaintenance() (bool, error) {
	stdio.Verbose("check ob maintenance")
	dag, err := GetLastOBMaintainDag()
	if err != nil {
		return false, errors.Errorf("get last maintain dag failed: %s", err)
	}
	if dag == nil || dag.IsSucceed() {
		return true, nil
	}
	return false, nil
}

func GetLastOBMaintainDag() (dag *task.DagDetailDTO, err error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_MAINTAIN + constant.URI_OB_GROUP
	if err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &dag); err != nil {
		if errors.IsTaskNotFoundErr(err) {
			stdio.Verbose("last agent maintain dag not found")
			return nil, nil
		}
		return nil, err
	}
	stdio.Verbosef("last agent maintain dag %s", dag.GenericID)
	return dag, nil
}
