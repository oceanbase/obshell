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
	"github.com/oceanbase/obshell/seekdb/agent/cmd/daemon"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	obmodel "github.com/oceanbase/obshell/seekdb/model/observer"
)

var (
	myAgent *meta.AgentInfoWithIdentity
)

func GetMyAgentInfo() (agent *meta.AgentInfoWithIdentity, err error) {
	if myAgent != nil {
		return myAgent, nil
	}

	status, err := GetMyAgentStatus()
	if err != nil {
		return nil, err
	}
	myAgent = &status.Agent
	return myAgent, nil
}

func GetMyAgentStatus() (status *http.AgentStatus, err error) {
	uri := constant.URI_API_V1 + constant.URI_STATUS
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &status)
	if err != nil {
		return nil, errors.Wrap(err, "get my agent status failed")
	}
	return
}

func GetMyDaemonStatus() (status *daemon.DaemonStatus, err error) {
	err = http.SendGetRequestViaUnixSocket(path.DaemonSocketPath(), constant.URI_API_V1+constant.URI_STATUS, nil, &status)
	return
}

func GetObserverInfo() (observerInfo *obmodel.ObserverInfo, err error) {
	uri := constant.URI_API_V1 + constant.URI_SEEKDB_GROUP + constant.URI_INFO
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &observerInfo)
	return
}
