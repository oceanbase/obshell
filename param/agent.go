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
	"time"

	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

type JoinApiParam struct {
	AgentInfo      meta.AgentInfo `json:"agentInfo" binding:"required"`
	ZoneName       string         `json:"zoneName" binding:"required"`
	MasterPassword string         `json:"masterPassword"`
}

type JoinMasterParam struct {
	JoinApiParam JoinApiParam `json:"joinApiParam" binding:"required"`
	HomePath     string       `json:"home_path" binding:"required"`
	Version      string       `json:"version" binding:"required"`
	Os           string       `json:"os" binding:"required"`
	Architecture string       `json:"architecture" binding:"required"`
	PublicKey    string       `json:"public_key" binding:"required"`
	Token        string       `json:"token" binding:"required"`
}

type AllAgentsSyncData struct {
	Maintainer   meta.AgentInfo       `json:"maintainer" binding:"required"`
	AllAgents    []oceanbase.AllAgent `json:"all_agents" binding:"required"`
	LastSyncTime time.Time            `json:"last_sync_time" binding:"required"`
}

type SetAgentPasswordParam struct {
	Password string `json:"password" binding:"required"`
}

type AddTokenParam struct {
	AgentInfo meta.AgentInfo `json:"agentInfo" binding:"required"`
	Token     string         `json:"token" binding:"required"`
}
