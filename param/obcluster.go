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

package param

import (
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

type ObClusterConfigParams struct {
	ClusterId   *int    `json:"clusterId"`
	ClusterName *string `json:"clusterName"`
	RsList      *string `json:"rsList"`
	RootPwd     *string `json:"rootPwd"`
}

// ObInfoResp is the response of ob/info
type ObInfoResp struct {
	Agents []meta.AgentInstance `json:"agent_info"`
	Config ClusterConfig        `json:"obcluster_info"`
}

type ClusterConfig struct {
	ClusterID   int                        `json:"id"`
	ClusterName string                     `json:"name"`
	Version     string                     `json:"version"`
	ZoneConfig  map[string][]*ServerConfig `json:"topology"`
}

type ServerConfig struct {
	SvrIP        string `json:"svr_ip"`
	SvrPort      int    `json:"svr_port"`
	SqlPort      int    `json:"sql_port"`
	AgentPort    int    `json:"agent_port"`
	WithRootSvr  string `json:"with_rootserver"`
	Status       string `json:"status"`
	BuildVersion string `json:"build_version"`
}

type RestoreParams struct {
	Params []oceanbase.ObParameters `json:"params" binding:"required"`
}
