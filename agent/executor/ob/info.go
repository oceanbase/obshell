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

package ob

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/coordinator"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

func GetObInfo() (*param.ObInfoResp, *errors.OcsAgentError) {
	identity := meta.OCS_AGENT.GetIdentity()
	switch identity {
	case meta.SINGLE, meta.SCALING_IN:
		resp := &param.ObInfoResp{}
		resp.Agents = append(resp.Agents, *meta.NewAgentInstanceByAgent(meta.OCS_AGENT))
		return resp, nil
	case meta.FOLLOWER:
		// if agent is follower, it will forward the request to master in handler.
		return nil, errors.Occurf(errors.ErrUnexpected, "wrong case: the request should be forwarded to master")
	case meta.MASTER:
		return getAgentInfo()
	case meta.CLUSTER_AGENT, meta.TAKE_OVER_FOLLOWER, meta.TAKE_OVER_MASTER, meta.SCALING_OUT:
		if resp, err := getClusterAndAgentsInfo(); err == nil {
			return resp, nil
		}
		return getAgentInfo()
	default:
		return nil, errors.Occurf(errors.ErrUnexpected, "unknown agent identity %s", identity)
	}
}

func getAgentInfo() (*param.ObInfoResp, *errors.OcsAgentError) {
	agents, err := agentService.GetAllAgentInstances()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	// If err occurs, return empty cluster info.
	clusterID, _ := getClusterID()
	clusterName, _ := getClusterName()

	info := param.ObInfoResp{
		Agents: agents,
		Config: param.ClusterConfig{
			ClusterID:   clusterID,
			ClusterName: clusterName,
		},
	}
	return &info, nil
}

func getClusterAndAgentsInfo() (*param.ObInfoResp, *errors.OcsAgentError) {
	var err error
	agents, _ := agentService.GetAllAgentsFromOB()
	if len(agents) == 0 {
		agents, err = agentService.GetAllAgentInstances()
		if err != nil {
			if maintainer, err := coordinator.GetMaintainer(); err != nil {
				return nil, errors.Occur(errors.ErrUnexpected, errors.Wrap(err, "get maintainer error"))
			} else if !maintainer.IsActive() {
				return nil, errors.Occur(errors.ErrUnexpected, "maintainer is not active")
			} else {
				return sendInfoApiTo(&maintainer)
			}
		}
	}
	clusterConfig, err := getClusterConfig()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	resp := &param.ObInfoResp{
		Agents: agents,
		Config: *clusterConfig,
	}
	return resp, nil
}

func getClusterID() (clusterID int, err error) {
	confVal, err := getClusterConfigByName(constant.OB_PARAM_CLUSTER_ID, constant.CONFIG_CLUSTER_ID)
	if confVal != "" {
		clusterID, err = strconv.Atoi(confVal)
	}
	return
}

func getClusterName() (clusterName string, err error) {
	return getClusterConfigByName(constant.OB_PARAM_CLUSTER_NAME, constant.CONFIG_CLUSTER_NAME)
}

func getClusterConfigByName(serverConfName string, agentConfName string) (value string, err error) {
	if meta.OCS_AGENT.IsClusterAgent() {
		value, err = observerService.GetOBStringParatemerByName(serverConfName)
	}
	if value == "" {
		var config sqlite.ObGlobalConfig
		config, err = observerService.GetObGlobalConfigByName(agentConfName)
		value = config.Value
	}
	return
}

func getClusterConfig() (resp *param.ClusterConfig, err error) {
	servers, err := obclusterService.GetAllOBServers()
	if err != nil {
		return nil, err
	}
	agents, err := agentService.GetAllAgentsDOFromOB()
	if err != nil {
		return nil, err
	}
	obVersion, err := obclusterService.GetObVersion()
	if err != nil {
		return nil, err
	}

	clusterID, err := getClusterID()
	if err != nil {
		return nil, err
	}
	clusterName, err := getClusterName()
	if err != nil {
		return nil, err
	}

	zoneMap := make(map[string][]*param.ServerConfig)
	for _, server := range servers {
		if _, ok := zoneMap[server.Zone]; !ok {
			zoneMap[server.Zone] = make([]*param.ServerConfig, 0)
		}

		if !server.StopTime.IsZero() {
			server.Status = "STOPPED"
		}
		svrInfo := &param.ServerConfig{
			SvrIP:        server.SvrIp,
			SvrPort:      server.SvrPort,
			SqlPort:      server.SqlPort,
			WithRootSvr:  server.WithRs,
			Status:       server.Status,
			BuildVersion: strings.Split(server.BuildVersion, "-")[0],
		}
		for _, agent := range agents {
			if server.SvrIp == agent.Ip && server.SvrPort == agent.RpcPort {
				svrInfo.AgentPort = agent.Port
				break
			}
		}
		zoneMap[server.Zone] = append(zoneMap[server.Zone], svrInfo)
	}

	return &param.ClusterConfig{
		ClusterID:   clusterID,
		ClusterName: clusterName,
		Version:     obVersion,
		ZoneConfig:  zoneMap,
	}, nil
}

func sendInfoApiTo(agent meta.AgentInfoInterface) (resp *param.ObInfoResp, agentErr *errors.OcsAgentError) {
	err := secure.SendGetRequest(agent, constant.URI_OB_API_PREFIX+constant.URI_INFO, nil, &resp)
	if err != nil {
		return resp, errors.Occur(errors.ErrUnexpected, err)
	}
	return resp, nil
}

func IsValidScope(s *param.Scope) (err error) {
	s.Format()
	switch s.Type {
	case SCOPE_GLOBAL:
		return nil
	case SCOPE_ZONE:
		if len(s.Target) == 0 {
			return errors.New("zone scope must have target")
		}
		for _, zone := range s.Target {
			exist, err := obclusterService.IsZoneExist(zone)
			if err != nil {
				return err
			}
			if !exist {
				return errors.Errorf("zone '%s' not exist", zone)
			}
		}
		return nil
	case SCOPE_SERVER:
		if len(s.Target) == 0 {
			return errors.New("server scope must have target")
		}
		for _, server := range s.Target {
			info := strings.Split(server, ":")
			if len(info) != 2 {
				return errors.Errorf("invalid server '%s'", server)
			}
			port, err := strconv.Atoi(info[1])
			if err != nil {
				return errors.Errorf("invalid server '%s' port '%s'", server, info[1])
			}
			agentInfo := meta.AgentInfo{Ip: info[0], Port: port}
			exist, err := agentService.IsAgentExist(&agentInfo)
			if err != nil {
				return err
			}
			if !exist {
				return errors.Errorf("server '%s' not exist", server)
			}
		}
		return nil
	default:
		return errors.Errorf("unknow scope type '%v'", s.Type)
	}
}

func ScopeOnlySelf(s *param.Scope) bool {
	s.Format()
	return s.Type == SCOPE_SERVER && len(s.Target) == 1 && s.Target[0] == meta.OCS_AGENT.String()
}

func GetObAgents() (agents []meta.AgentInfo, err error) {
	switch meta.OCS_AGENT.GetIdentity() {
	case meta.SINGLE:
		agents = append(agents, meta.OCS_AGENT.GetAgentInfo())
		return
	case meta.FOLLOWER:
		// if agent is follower, it will forward the request to master in handler.
		return nil, errors.Errorf("wrong case: the request should be forwarded to master")
	case meta.MASTER:
		agents, err = agentService.GetAllAgentsInfo()
		return
	case meta.CLUSTER_AGENT:
		agents, err = agentService.GetAllAgentsInfoFromOB()
		if len(agents) == 0 {
			agents, err = agentService.GetAllAgentsInfo()
		}
		return
	case meta.TAKE_OVER_FOLLOWER, meta.TAKE_OVER_MASTER, meta.SCALING_OUT:
		serversWithRpcPort, err := GetAllServerFromOBConf()
		if err != nil {
			return nil, errors.Wrap(err, "get servers from ob.conf failed")
		}

		for _, server := range serversWithRpcPort {
			agents = append(agents, meta.AgentInfo{Ip: server[0], Port: meta.OCS_AGENT.GetPort()})
		}
	}
	return

}

func GetAllServerFromOBConf() (serversWithRpcPort [][2]string, err error) {
	f := path.ObConfigPath()
	log.Info("get conf from ", f)
	file, err := os.Open(f)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err = scanner.Err(); err != nil {
		return
	}
	re := regexp.MustCompile("\x00*([_a-zA-Z]+)=(.*)")

	var servers, items []string
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		if match[1] == ETC_KEY_ALL_SERVER_LIST {
			servers = strings.Split(match[2], ",")
			for _, server := range servers {
				items = strings.Split(server, ":")
				if len(items) != 2 {
					return nil, errors.Errorf("invalid server '%s'", server)
				}
				_, err = strconv.Atoi(items[1])
				if err != nil {
					return nil, errors.Wrapf(err, "invalid server '%s' port '%s'", server, items[1])
				}
				serversWithRpcPort = append(serversWithRpcPort, [2]string{items[0], items[1]})
			}
			log.Infof("get servers from ob.conf %v", serversWithRpcPort)
			return
		}
	}
	return nil, fmt.Errorf("not found %s in ob.conf", ETC_KEY_ALL_SERVER_LIST)
}
