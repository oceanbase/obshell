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
	"strings"

	"github.com/oceanbase/obshell/ob/agent/meta"
)

type DeployTaskParams struct {
	Dirs map[string]string `json:"dirs" binding:"required"`
}

type StartTaskParams struct {
	Config      map[string]string `json:"config" binding:"required"`
	HealthCheck bool              `json:"healthCheck"`
}

type StopTaskParams struct {
	Force bool `json:"force" binding:"required"`
}

type SyncAgentParams struct {
	Password string `json:"password" binding:"required"`
}

type ObServerConfigParams struct {
	ObServerConfig map[string]string `json:"observerConfig" binding:"required"`
	Restart        bool              `json:"restart"`
	Scope          Scope             `json:"scope" binding:"required"`
}

type ScaleOutParam struct {
	AgentInfo meta.AgentInfo    `json:"agentInfo" binding:"required"`
	ObConfigs map[string]string `json:"obConfigs" binding:"required"`
	Zone      string            `json:"zone" binding:"required"`
}

type ClusterScaleOutParam struct {
	ScaleOutParam
	TargetAgentPassword string `json:"targetAgentPassword"`
}

type LocalScaleOutParam struct {
	ScaleOutParam
	TargetVersion                string            `json:"targetVersion"`
	AllAgents                    []meta.AgentInfo  `json:"allAgents" binding:"required"`
	Dirs                         map[string]string `json:"dirs" binding:"required"`
	CoordinateDagId              string            `json:"coordinateDagId" binding:"required"`
	RootPwd                      string            `json:"rootPwd" binding:"required"`
	Uuid                         string            `json:"uuid" binding:"required"`
	ParamExpectDeployNextStage   int               `json:"paramExpectDeployNextStage" binding:"required"`
	ParamExpectStartNextStage    int               `json:"paramExpectStartNextStage" binding:"required"`
	ParamExpectRollbackNextStage int               `json:"paramExpectRollbackNextStage" binding:"required"`
}

type ClusterScaleInParam struct {
	AgentInfo meta.AgentInfo `json:"agent_info" binding:"required"`
	ForceKill bool           `json:"force_kill"` // default to false
}

type ObInitParam struct {
	ImportScript      bool   `json:"import_script"`
	CreateProxyroUser bool   `json:"create_proxyro_user"`
	ProxyroPassword   string `json:"proxyro_password"`
}

type ObStopParam struct {
	Scope             Scope             `json:"scope" binding:"required"`
	Force             bool              `json:"force"`
	Terminate         bool              `json:"terminate"`
	ForcePassDagParam ForcePassDagParam `json:"forcePassDag"`
}

type ForcePassDagParam struct {
	ID []string `json:"id"`
}

type StartObParam struct {
	Scope             Scope             `json:"scope" binding:"required"`
	ForcePassDagParam ForcePassDagParam `json:"forcePassDag"`
}

type ObVersion struct {
	Version string `json:"version" binding:"required"`
	Release string `json:"release" binding:"required"`
}

type UpgradeCheckParam struct {
	Version    string `json:"version" binding:"required"`
	Release    string `json:"release" binding:"required"`
	UpgradeDir string `json:"upgradeDir" `
}

type ObUpgradeParam struct {
	UpgradeCheckParam
	Mode         string `json:"mode" binding:"required"`
	FreezeServer bool   `json:"freeze_server"`
}

type Scope struct {
	Type       string   `json:"type"`
	Target     []string `json:"target"`
	isFormated bool
}

func (s *Scope) Format() {
	if !s.isFormated {
		s.Type = strings.ToUpper(s.Type)
		s.Target = unique(s.Target)
		s.isFormated = true
	}
}

func unique(s []string) []string {
	m := make(map[string]bool)
	for _, v := range s {
		m[v] = true
	}
	var result []string
	for key := range m {
		result = append(result, key)
	}
	return result
}
