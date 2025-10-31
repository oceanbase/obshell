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
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/utils"
)

type AgentIdentity string

const (
	MASTER           AgentIdentity = "MASTER"
	SINGLE           AgentIdentity = "SINGLE"
	CLUSTER_AGENT    AgentIdentity = "CLUSTER AGENT"
	TAKE_OVER_MASTER AgentIdentity = "TAKE OVER MASTER"
	UNIDENTIFIED     AgentIdentity = "UNIDENTIFIED"

	AUTH_V1 = "v1"
	AUTH_V2 = "v2"
)

type AgentInfoInterface interface {
	GetIp() string
	GetPort() int
	String() string
}

type Agent interface {
	AgentInfoInterface
	GetIdentity() AgentIdentity
	IsMasterAgent() bool
	IsSingleAgent() bool
	IsClusterAgent() bool
	IsTakeover() bool
	IsTakeOverMasterAgent() bool
	IsUnidentified() bool
	GetVersion() string
	GetAgentInfo() AgentInfo
	String() string
	IsIPv6() bool
	GetLocalIp() string
	Equal(other AgentInfoInterface) bool
}

var OCS_AGENT Agent

type AgentInfo struct {
	Ip   string `json:"ip" form:"ip" binding:"required"` // unused for agent start
	Port int    `json:"port" form:"port" binding:"required"`
}

func (agentInfo *AgentInfo) GetIp() string {
	return agentInfo.Ip
}

func (agentInfo *AgentInfo) GetLocalIp() string {
	if agentInfo.IsIPv6() {
		return constant.LOCAL_IP_V6
	}
	return constant.LOCAL_IP
}

func (agentInfo *AgentInfo) GetPort() int {
	return agentInfo.Port
}

func (agentInfo *AgentInfo) IsIPv6() bool {
	return strings.Contains(agentInfo.Ip, ":")
}

func (agentInfo AgentInfo) String() string {
	if agentInfo.IsIPv6() {
		return fmt.Sprintf("[%s]:%d", agentInfo.Ip, agentInfo.Port)
	}
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

type AgentInfoWithIdentity struct {
	AgentInfo
	IdentityDTO
}

type AgentInstance struct {
	AgentInfo
	IdentityDTO
	Version string `json:"version"`
}

type AgentStatus struct {
	Pid          int    `json:"pid"`
	State        int32  `json:"state"`
	StartAt      int64  `json:"startAt"`
	HomePath     string `json:"homePath"`
	OBVersion    string `json:"obVersion"`
	Architecture string `json:"architecture"`
	Type         string `json:"type"` // "seekdb" means the agent is work for seekdb
	AgentInstance
	Security      bool     `json:"security"`
	SupportedAuth []string `json:"supportedAuth"`
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

func (agent *IdentityDTO) IsClusterAgent() bool {
	return agent.Identity == CLUSTER_AGENT
}

func (agent *IdentityDTO) IsTakeOverMasterAgent() bool {
	return agent.Identity == TAKE_OVER_MASTER
}

func (agent *IdentityDTO) IsTakeover() bool {
	return agent.IsTakeOverMasterAgent()
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

func ConvertAddressToAgentInfo(host string) (*AgentInfo, error) {
	if host == "" {
		return nil, errors.Occur(errors.ErrCommonInvalidAddress, host)
	}
	if strings.Contains(host, ".") {
		// If the host contains '.', it might be an IPv4 address, but further validation is needed.
		return convertIPv4ToAgentInfo(host)
	} else {
		// If the host contains '.', it might be an IPv6 address, but further validation is needed.
		return convertIPv6ToAgentInfo(host)
	}
}

func convertIPv4ToAgentInfo(host string) (*AgentInfo, error) {
	var ip string
	var err error
	var port = constant.DEFAULT_AGENT_PORT
	matches := strings.Split(host, ":")
	if len(matches) == 1 {
		return NewAgentInfo(matches[0], constant.DEFAULT_AGENT_PORT), nil
	} else if len(matches) == 2 {
		if port, err = strconv.Atoi(matches[1]); err != nil || !utils.IsValidPortValue(port) {
			return nil, errors.Occur(errors.ErrCommonInvalidPort, matches[1])
		}
		ip = matches[0]
	} else {
		return nil, errors.Occur(errors.ErrCommonInvalidAddress, host)
	}

	ipv4 := net.ParseIP(ip)
	if ipv4 == nil || ipv4.To4() == nil {
		return nil, errors.Occur(errors.ErrCommonInvalidIp, ip)
	}
	return NewAgentInfo(ip, port), nil
}

func convertIPv6ToAgentInfo(host string) (*AgentInfo, error) {
	re := regexp.MustCompile(`(?:\[([0-9a-fA-F:]+)\]|([0-9a-fA-F:]+))(?:\:(\d+))?`)
	matches := re.FindStringSubmatch(host)

	if matches == nil {
		return nil, errors.Occur(errors.ErrCommonInvalidAddress, host)
	}

	var ip string
	var err error
	var port = constant.DEFAULT_AGENT_PORT
	if matches[1] != "" {
		ip = matches[1]
	} else {
		ip = matches[2]
	}

	if matches[3] != "" {
		if port, err = strconv.Atoi(matches[3]); err != nil || !utils.IsValidPortValue(port) {
			return nil, errors.Occur(errors.ErrCommonInvalidPort, matches[3])
		}
	}

	ipv6 := net.ParseIP(ip)
	if ipv6 == nil || ipv6.To4() != nil {
		return nil, errors.Occur(errors.ErrCommonInvalidIp, ip)
	}
	return NewAgentInfo(ip, port), nil
}

func NewAgentInfoByString(info string) *AgentInfo {
	// if err != nil, agent will be nil. So, no need to check err.
	agent, _ := ConvertAddressToAgentInfo(info)
	return agent
}

func NewAgentInfoByInterface(agentInfo AgentInfoInterface) *AgentInfo {
	return &AgentInfo{
		Ip:   agentInfo.GetIp(),
		Port: agentInfo.GetPort(),
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

func NewAgentInstanceByAgent(agent Agent) *AgentInstance {
	return &AgentInstance{
		AgentInfo: *NewAgentInfoByInterface(agent),
		IdentityDTO: IdentityDTO{
			Identity: agent.GetIdentity(),
		},
		Version: agent.GetVersion(),
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

func NewAgentStatus(agent Agent, pid int, state int32, startAt int64, homePath string, obVersion string) *AgentStatus {
	return &AgentStatus{
		Pid:           pid,
		State:         state,
		StartAt:       startAt,
		HomePath:      homePath,
		OBVersion:     obVersion,
		Type:          "seekdb",
		AgentInstance: *NewAgentInstanceByAgent(agent),
		SupportedAuth: []string{AUTH_V2},
	}
}
