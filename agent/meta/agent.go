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

package meta

import (
	"fmt"
	"strconv"
	"strings"
)

type AgentIdentity string

const (
	MASTER             AgentIdentity = "MASTER"
	FOLLOWER           AgentIdentity = "FOLLOWER"
	SINGLE             AgentIdentity = "SINGLE"
	CLUSTER_AGENT      AgentIdentity = "CLUSTER AGENT"
	TAKE_OVER_MASTER   AgentIdentity = "TAKE OVER MASTER"
	TAKE_OVER_FOLLOWER AgentIdentity = "TAKE OVER FOLLOWER"
	SCALING_OUT        AgentIdentity = "SCALING OUT"
	UNIDENTIFIED       AgentIdentity = "UNIDENTIFIED"
)

type AgentInfoInterface interface {
	GetIp() string
	GetPort() int
	String() string
}

type AgentInfoWithZoneInterface interface {
	AgentInfoInterface
	GetZone() string
}

type Agent interface {
	AgentInfoWithZoneInterface
	GetIdentity() AgentIdentity
	IsMasterAgent() bool
	IsSingleAgent() bool
	IsFollowerAgent() bool
	IsClusterAgent() bool
	IsTakeover() bool
	IsTakeOverFollowerAgent() bool
	IsTakeOverMasterAgent() bool
	IsScalingOutAgent() bool
	IsUnidentified() bool
	GetVersion() string
	GetAgentInfo() AgentInfo
	String() string
	Equal(other AgentInfoInterface) bool
}

var OCS_AGENT Agent

type AgentInfo struct {
	Ip   string `json:"ip" form:"ip" binding:"required"`
	Port int    `json:"port" form:"port" binding:"required"`
}

func (agentInfo *AgentInfo) GetIp() string {
	return agentInfo.Ip
}

func (agentInfo *AgentInfo) GetPort() int {
	return agentInfo.Port
}

func (agentInfo *AgentInfo) String() string {
	return fmt.Sprintf("%s:%d", agentInfo.Ip, agentInfo.Port)
}

func (a *AgentInfo) Equal(agent AgentInfoInterface) bool {
	return a.GetIp() == agent.GetIp() && a.GetPort() == agent.GetPort()
}

type PublicKeyDTO struct {
	PublicKey string `json:"public_key" form:"public_key" binding:"required"`
}

type ZoneDTO struct {
	Zone string `json:"zone" binding:"required"`
}

type IdentityDTO struct {
	Identity AgentIdentity `json:"identity" binding:"required"`
}

type AgentSecret struct {
	AgentInfo
	PublicKeyDTO
}

type AgentInfoWithZone struct {
	AgentInfo
	ZoneDTO
}

type AgentInfoWithIdentity struct {
	AgentInfo
	IdentityDTO
}

type AgentInstance struct {
	AgentInfo
	ZoneDTO
	IdentityDTO
	Version string `json:"version"`
}

type AgentStatus struct {
	Pid      int    `json:"pid"`
	State    int32  `json:"state"`
	StartAt  int64  `json:"startAt"`
	HomePath string `json:"homePath"`
	AgentInstance
}

func (agent *ZoneDTO) GetZone() string {
	return agent.Zone
}

func (agent *IdentityDTO) GetIdentity() AgentIdentity {
	return agent.Identity
}

func (agent *IdentityDTO) IsMasterAgent() bool {
	return agent.Identity == MASTER
}

func (agent *IdentityDTO) IsSingleAgent() bool {
	return agent.Identity == SINGLE
}

func (agent *IdentityDTO) IsFollowerAgent() bool {
	return agent.Identity == FOLLOWER
}

func (agent *IdentityDTO) IsClusterAgent() bool {
	return agent.Identity == CLUSTER_AGENT
}

func (agent *IdentityDTO) IsTakeOverMasterAgent() bool {
	return agent.Identity == TAKE_OVER_MASTER
}

func (agent *IdentityDTO) IsTakeOverFollowerAgent() bool {
	return agent.Identity == TAKE_OVER_FOLLOWER
}

func (agent *IdentityDTO) IsTakeover() bool {
	return agent.IsTakeOverMasterAgent() || agent.IsTakeOverFollowerAgent()
}

func (agent *IdentityDTO) IsScalingOutAgent() bool {
	return agent.Identity == SCALING_OUT
}

func (agent *IdentityDTO) IsUnidentified() bool {
	return agent.Identity == UNIDENTIFIED
}

func (agent *AgentInstance) GetVersion() string {
	return agent.Version
}

func (agent *AgentInstance) GetAgentInfo() AgentInfo {
	return agent.AgentInfo
}

func NewAgentInfo(ip string, port int) *AgentInfo {
	return &AgentInfo{
		Ip:   ip,
		Port: port,
	}
}

func NewAgentInfoByString(info string) *AgentInfo {
	portIndex := strings.LastIndex(info, ":")
	if portIndex == -1 {
		return nil
	}

	ip := info[:portIndex]
	port, err := strconv.Atoi(info[portIndex+1:])
	if err != nil {
		return nil
	}
	return &AgentInfo{
		Ip:   ip,
		Port: port,
	}
}

func NewAgentInfoByInterface(agentInfo AgentInfoInterface) *AgentInfo {
	return &AgentInfo{
		Ip:   agentInfo.GetIp(),
		Port: agentInfo.GetPort(),
	}
}

func NewAgentWithZone(ip string, port int, zone string) *AgentInfoWithZone {
	return &AgentInfoWithZone{
		AgentInfo: *NewAgentInfo(ip, port),
		ZoneDTO: ZoneDTO{
			Zone: zone,
		},
	}
}

func NewAgentWithZoneByAgentInfo(agentInfo AgentInfoInterface, zone string) *AgentInfoWithZone {
	return &AgentInfoWithZone{
		AgentInfo: *NewAgentInfoByInterface(agentInfo),
		ZoneDTO: ZoneDTO{
			Zone: zone,
		},
	}
}

func NewAgentInfoWithIdentity(ip string, port int, identity AgentIdentity) *AgentInfoWithIdentity {
	return &AgentInfoWithIdentity{
		AgentInfo: *NewAgentInfo(ip, port),
		IdentityDTO: IdentityDTO{
			Identity: identity,
		},
	}
}

func NewAgentInstance(ip string, port int, zone string, identity AgentIdentity, version string) *AgentInstance {
	return &AgentInstance{
		AgentInfo: *NewAgentInfo(ip, port),
		ZoneDTO: ZoneDTO{
			Zone: zone,
		},
		IdentityDTO: IdentityDTO{
			Identity: identity,
		},
		Version: version,
	}
}
func NewAgentInstanceByAgentInfo(agentInfo AgentInfoInterface, zone string, identity AgentIdentity, version string) *AgentInstance {
	return &AgentInstance{
		AgentInfo: *NewAgentInfoByInterface(agentInfo),
		ZoneDTO: ZoneDTO{
			Zone: zone,
		},
		IdentityDTO: IdentityDTO{
			Identity: identity,
		},
		Version: version,
	}
}

func NewAgentInstanceByAgent(agent Agent) *AgentInstance {
	return &AgentInstance{
		AgentInfo: *NewAgentInfoByInterface(agent),
		ZoneDTO: ZoneDTO{
			Zone: agent.GetZone(),
		},
		IdentityDTO: IdentityDTO{
			Identity: agent.GetIdentity(),
		},
		Version: agent.GetVersion(),
	}
}

func NewAgentSecret(ip string, port int, publicKey string) *AgentSecret {
	return &AgentSecret{
		AgentInfo: *NewAgentInfo(ip, port),
		PublicKeyDTO: PublicKeyDTO{
			PublicKey: publicKey,
		},
	}
}

func NewAgentSecretByAgentInfo(agent AgentInfoInterface, publicKey string) *AgentSecret {
	return &AgentSecret{
		AgentInfo: *NewAgentInfoByInterface(agent),
		PublicKeyDTO: PublicKeyDTO{
			PublicKey: publicKey,
		},
	}
}

func NewAgentStatus(agent Agent, pid int, state int32, startAt int64, homePath string) *AgentStatus {
	return &AgentStatus{
		Pid:           pid,
		State:         state,
		StartAt:       startAt,
		HomePath:      homePath,
		AgentInstance: *NewAgentInstanceByAgent(agent),
	}
}
