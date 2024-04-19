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

	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/cmd/daemon"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/meta"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
	"github.com/oceanbase/obshell/param"
)

var (
	myAgent          *meta.AgentInfoWithIdentity
	localTaskService = taskservice.NewLocalTaskService()
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

func GetAllAgentInfo() (agents []meta.AgentInfo, err error) {
	obInfo, err := GetObInfo()
	if err != nil {
		return nil, errors.Wrap(err, "get all agent info failed")
	}
	for _, agent := range obInfo.Agents {
		agents = append(agents, agent.AgentInfo)
	}
	return
}

func GetClusterConfig() (clusterConfig *param.ClusterConfig, err error) {
	obInfo, err := GetObInfo()
	if err != nil {
		return nil, errors.Wrap(err, "get cluster config failed")
	}
	clusterConfig = &obInfo.Config
	return
}

func GetObInfo() (obInfo *param.ObInfoResp, err error) {
	uri := constant.URI_OB_API_PREFIX + constant.URI_INFO
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &obInfo)
	if err != nil {
		return nil, errors.Wrap(err, "get ob info failed")
	}
	return
}

func GetAllAgentsStatus() (status map[string]http.AgentStatus, err error) {
	uri := constant.URI_AGENTS_API_PREFIX + constant.URI_STATUS
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &status)
	return
}

func GetAllLastAgentMaintainDag() (dags []*task.DagDetailDTO, err error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_MAINTAIN + constant.URI_AGENTS_GROUP
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &dags)
	return
}

func GetAllMainAndMaintainDag() (mainDags, maintainDags []*task.DagDetailDTO, err error) {
	stdio.Verbose("get all main dags and maintenance dags")
	mainDagIDMap := make(map[string]bool)
	allDags, err := GetAllLastAgentMaintainDag()
	if err != nil {
		return
	}

	for _, dag := range allDags {
		stdio.Verbosef("handle %s", dag.GenericID)
		if dag.IsSucceed() {
			stdio.Verbosef("dag %s is succeed", dag.GenericID)
			continue
		}
		id, isEmecTypeDag := IsEmecTypeDag(dag)
		if isEmecTypeDag {
			if !mainDagIDMap[id] {
				mainDagIDMap[id] = true

				mainDag, err := GetDagDetail(id)
				if err != nil {
					log.WithError(err).Errorf("get main dag %s detail failed", id)
					continue
				}
				if !mainDag.IsSucceed() {
					mainDags = append(mainDags, mainDag)
				}
			}
			continue
		}

		if !dag.IsFinished() {
			maintainDags = append(maintainDags, dag)
		}
	}
	return
}
