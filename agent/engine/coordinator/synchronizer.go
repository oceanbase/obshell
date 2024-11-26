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

package coordinator

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

var OCS_AGENT_SYNCHRONIZER *AgentSynchronizer

type AllAgentsSyncData struct {
	// AllAgents is a map containing all the agents in the cluster.
	// It is used to maintain the agent information within the cluster.
	AllAgents map[meta.AgentInfo]oceanbase.AllAgent

	// LastSyncTime is the last time the all_agent was synchronized
	LastSyncTime time.Time
}

type AgentSynchronizer struct {
	coordinator *Coordinator
	AllAgentsSyncData

	// when the agent is maintainer, it need to notice AllAgentsSyncData to synchronize the all_agents.
	// allAgentsLastSyncTime is used to record the last synchronization time of AllAgentsSyncData
	allAgentsLastSyncTime map[meta.AgentInfo]time.Time
	// cancel is used to stop the agent synchronizer handler
	cancel context.CancelFunc

	// any agent synchronizerHandler can only be executed once when the first Start is called
	once sync.Once

	// newAllAgentsSyncData is used to record the AllAgentsSyncData that needs to be updateAllAgentsToSqlite
	newAllAgentsSyncData *AllAgentsSyncData
	// lock is used to protect newAllAgentsSyncData
	lock sync.Mutex
}

func NewAllAgentsSyncData(allAgents map[meta.AgentInfo]oceanbase.AllAgent, lastSyncTime time.Time) *AllAgentsSyncData {
	return &AllAgentsSyncData{
		AllAgents:    allAgents,
		LastSyncTime: lastSyncTime,
	}
}

func ConvertToAllAgentsSyncDataParam(maintainer meta.AgentInfo, data *AllAgentsSyncData) param.AllAgentsSyncData {
	allAgents := make([]oceanbase.AllAgent, 0, len(data.AllAgents))
	for _, agent := range data.AllAgents {
		allAgents = append(allAgents, agent)
	}
	return param.AllAgentsSyncData{
		Maintainer:   maintainer,
		AllAgents:    allAgents,
		LastSyncTime: data.LastSyncTime,
	}
}

func ConvertToAllAgentsSyncData(data param.AllAgentsSyncData) *AllAgentsSyncData {
	allAgents := make(map[meta.AgentInfo]oceanbase.AllAgent, len(data.AllAgents))
	for _, agent := range data.AllAgents {
		allAgents[*meta.NewAgentInfo(agent.Ip, agent.Port)] = agent
	}
	return NewAllAgentsSyncData(allAgents, data.LastSyncTime)
}

func NewAgentSynchronizer(coordinator *Coordinator) *AgentSynchronizer {
	return &AgentSynchronizer{
		coordinator:           coordinator,
		allAgentsLastSyncTime: make(map[meta.AgentInfo]time.Time),
		once:                  sync.Once{},
		lock:                  sync.Mutex{},
	}
}

func (as *AgentSynchronizer) Start() {
	go as.once.Do(as.synchronizerHandler)

	eventChan := as.coordinator.Subscribe(as)
	defer eventChan.Close()
	for {
		isMaintainer := <-eventChan.Listen()
		if isMaintainer && as.coordinator.IsMaintainer() {
			ctx := context.Background()
			go as.run(ctx)
		} else if !isMaintainer && !as.coordinator.IsMaintainer() {
			as.stop()
		}
	}
}

func (as *AgentSynchronizer) Update(data *AllAgentsSyncData) {
	as.lock.Lock()
	defer as.lock.Unlock()
	if as.newAllAgentsSyncData == nil || data.LastSyncTime.After(as.newAllAgentsSyncData.LastSyncTime) {
		as.newAllAgentsSyncData = data
	}
}

func (as *AgentSynchronizer) synchronizerHandler() {
	for {
		if as.newAllAgentsSyncData != nil && as.LastSyncTime.Before(as.newAllAgentsSyncData.LastSyncTime) {
			as.updateAllAgentsToSqlite(as.newAllAgentsSyncData)
		}
		time.Sleep(1 * time.Second)
	}
}

func (as *AgentSynchronizer) run(ctx context.Context) {
	if as.cancel != nil {
		log.Warn("agent synchronizer already running")
		return
	}

	log.Info("agent synchronizer starting")
	_, as.cancel = context.WithCancel(ctx)
	for as.cancel != nil {
		duration := as.handle()
		time.Sleep(duration)
	}
	log.Info("agent synchronizer stopped")
}

func (as *AgentSynchronizer) stop() {
	if as.cancel != nil {
		log.Info("agent synchronizer stopping")
		as.cancel()
		as.cancel = nil
	}
}

func (as *AgentSynchronizer) handle() time.Duration {
	expirationTime := as.coordinator.Maintainer.GetExpirationTime()
	if data := as.GetAllAgentsSyncDataIfChanged(); data != nil {
		as.broadcast(data)
	} else if as.AllAgentsSyncData.AllAgents != nil {
		as.broadcast(&as.AllAgentsSyncData)
	}

	delta := time.Until(expirationTime)
	if delta < 0 {
		return 0
	}
	return delta
}

func (as *AgentSynchronizer) GetAllAgentsSyncDataIfChanged() *AllAgentsSyncData {
	agentsDO, err := agentService.GetAllAgentsDOFromOB()
	if err != nil {
		log.WithError(err).Error("get all agents from ob failed")
		return nil
	}

	newAllAgents := map[meta.AgentInfo]oceanbase.AllAgent{}
	for _, agentDO := range agentsDO {
		agent := *meta.NewAgentInfo(agentDO.Ip, agentDO.Port)
		newAllAgents[agent] = agentDO
	}

	if len(newAllAgents) != len(as.AllAgents) {
		return NewAllAgentsSyncData(newAllAgents, as.coordinator.Maintainer.GetLastUpdateTime())
	}

	for agent, newAgentDO := range newAllAgents {
		if agentDO, ok := as.AllAgents[agent]; !ok || agentDO != newAgentDO {
			return NewAllAgentsSyncData(newAllAgents, as.coordinator.Maintainer.GetLastUpdateTime())
		}
	}
	return nil // no change
}

func (as *AgentSynchronizer) broadcast(data *AllAgentsSyncData) {
	for agent := range data.AllAgents {
		if _, ok := as.allAgentsLastSyncTime[agent]; !ok {
			as.allAgentsLastSyncTime[agent] = time.Time{}
		}
	}

	var wg sync.WaitGroup
	var succeedChan = make(chan meta.AgentInfo, len(as.allAgentsLastSyncTime))
	for agent, lastSyncTime := range as.allAgentsLastSyncTime {
		if lastSyncTime.Before(data.LastSyncTime) {
			wg.Add(1)
			log.Infof("notice agent %s to synchronize all_agents", agent.String())
			go func(agent meta.AgentInfo, data *AllAgentsSyncData) {
				defer wg.Done()
				if err := as.notice(agent, data); err != nil {
					log.WithError(err).Errorf("notice agent %s failed", agent.String())
				} else {
					succeedChan <- agent
				}
			}(agent, data)
		}
	}

	wg.Wait()
	close(succeedChan)
	for agent := range succeedChan {
		as.allAgentsLastSyncTime[agent] = data.LastSyncTime
	}
}

func (as *AgentSynchronizer) notice(agent meta.AgentInfo, data *AllAgentsSyncData) error {
	if agent.Equal(meta.OCS_AGENT) {
		as.Update(data)
		return nil
	} else {
		uri := constant.URI_RPC_V1 + constant.URI_MAINTAINER + constant.URI_UPDATE
		return secure.SendPostRequest(&agent, uri, ConvertToAllAgentsSyncDataParam(meta.OCS_AGENT.GetAgentInfo(), data), nil)
	}
}

func (as *AgentSynchronizer) updateAllAgentsToSqlite(data *AllAgentsSyncData) {
	var allAgents map[meta.AgentInfo]oceanbase.AllAgent
	oldAllAgents, err := agentService.GetAllAgentsDO()
	if err != nil {
		log.WithError(err).Error("agent synchronizer get all agents from sqlite failed")
		return
	}

	allAgents = make(map[meta.AgentInfo]oceanbase.AllAgent)
	for _, agentDO := range oldAllAgents {
		agent := *meta.NewAgentInfo(agentDO.Ip, agentDO.Port)
		allAgents[agent] = agentService.ConvertToOBAgentDO(agentDO)
	}

	succeed := true
	newAllAgents := data.AllAgents
	for agent, newAgentDO := range newAllAgents {
		if agentDO, ok := allAgents[agent]; !ok || agentDO != newAgentDO {
			log.Infof("sync agent: %s(%s) by coordinator", agent.String(), newAgentDO.Identity)
			if err := agentService.SyncAgentDOToSqlite(newAgentDO); err != nil {
				succeed = false
				log.WithError(err).Errorf("sync agent %s(%s) to sqlite failed", agent.String(), newAgentDO.Identity)
			}
		}
	}

	for agent := range allAgents {
		if agentDO, ok := newAllAgents[agent]; !ok {
			log.Infof("remove agent: %s(%s) by coordinator", agent.String(), agentDO.Identity)
			if err := agentService.RemoveAgent(&agent); err != nil {
				succeed = false
				log.WithError(err).Errorf("remove agent %s(%s) by coordinator failed", agent.String(), agentDO.Identity)
			}
		}
	}

	// Make sure c.allAgents is the latest
	as.AllAgents = newAllAgents
	if succeed {
		as.LastSyncTime = data.LastSyncTime
	}
}
