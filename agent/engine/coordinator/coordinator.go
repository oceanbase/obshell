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
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	agentservice "github.com/oceanbase/obshell/agent/service/agent"
	"github.com/oceanbase/obshell/agent/service/coordinator"
)

const (
	// FAULTY means coordinator is not initialized
	FAULTY = iota + 1
	MAINTAINER
	WATCHER
)

var (
	coordinatorService = coordinator.CoordinatorService{}
	agentService       = agentservice.AgentService{}

	// OCS_COORDINATOR is the coordinator of agent and globally Unique
	OCS_COORDINATOR *Coordinator

	// identifiedMap String Map
	identifiedMap = map[int]string{
		FAULTY:     "FAULTY",
		MAINTAINER: "MAINTAINER",
		WATCHER:    "WATCHER",
	}
)

type Maintainer struct {
	MaintainerInfo oceanbase.TaskMaintainer
	agentInfo      meta.AgentInfoInterface
	LifeTime       float64   // How long has it been since the maintainer's active time (seconds)
	LastUpdateTime time.Time // The last time the maintainer was updated (local time)
	ExpirationTime time.Time // The expiration time of the maintainer (local time)
}

type Coordinator struct {
	Maintainer *Maintainer
	identity   int
	isSuspend  bool
	lock       sync.Mutex
	eventChans map[interface{}]*coordinatorEventChan
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		Maintainer: &Maintainer{},
		identity:   FAULTY,
		isSuspend:  false,
		lock:       sync.Mutex{},
		eventChans: make(map[interface{}]*coordinatorEventChan),
	}
}

func (c *Coordinator) Start() error {
	log.Info("coordinator start")
	for {
		if meta.OCS_AGENT.IsScalingInAgent() && c.IsMaintainer() {
			c.Suspend()
			continue
		} else if meta.OCS_AGENT == nil || !meta.OCS_AGENT.IsClusterAgent() {
			time.Sleep(constant.COORDINATOR_MIN_INTERVAL)
			continue
		}
		if c.isSuspend {
			c.Resume()
		}
		if err := c.reconcile(); err != nil {
			log.WithError(err).Error("coordinator reconcile failed")
		}
		c.wait()
	}
}

func (c *Coordinator) wait() {
	duration := constant.COORDINATOR_MIN_INTERVAL
	switch c.identity {
	case MAINTAINER:
		log.Debug("maintainer wait")
		delta := constant.MAINTAINER_UPDATE_INTERVAL - time.Duration(c.Maintainer.GetLifeTime())
		if delta > duration {
			duration = delta
		}
	case WATCHER:
		log.Debug("watcher wait")
		delta := constant.MAINTAINER_MAX_ACTIVE_TIME - time.Duration(c.Maintainer.GetLifeTime())
		if delta > duration {
			duration = delta
		}
	default:
		// No maintainer, wait for constant.COORDINATOR_MIN_INTERVAL
		log.Debug("coordinator identity is faulty")
	}
	log.Info("coordinator wait: ", duration)
	time.Sleep(duration)
	c.addLifeTime(duration.Seconds())
}

func (c *Coordinator) reconcile() error {
	if c.IsMaintainer() {
		return c.maintainerEffector()
	}
	if c.IsWatcher() {
		return c.watcherEffector()
	}
	return c.init()
}

func (c *Coordinator) watcherEffector() error {
	log.Info("watcher effector")
	err := c.getMaintainerbyRpc(c.Maintainer)
	if err != nil {
		if err := c.buildMaintainerByPolling(); err != nil {
			log.WithError(err).Error("try to be maintainer")
			if err := c.tryToBeMaintainer(); err != nil {
				c.removeMaintainer()
				return err
			}
		}
	}
	// If maintainer is still active, do nothing.
	if c.Maintainer.ExpirationTime.After(time.Now()) {
		return nil
	}
	return nil
}

func (c *Coordinator) maintainerEffector() error {
	log.Info("renewal maintainer")
	c.Maintainer.MaintainerInfo.Counter++
	if err := coordinatorService.RenewalMaintainer(c.Maintainer.MaintainerInfo); err == nil {
		c.Maintainer.setLifeTime(0)
		return nil
	}
	return c.init()
}

func (c *Coordinator) init() error {
	log.Info("coordinator initializing...")
	maintainerDO, err := coordinatorService.GetMaintainerFromOb()
	if err != nil {
		log.WithError(err).Error("get maintainer from ob failed")
		if err := c.buildMaintainerByPolling(); err != nil {
			log.WithError(err).Error("try to be maintainer")
			c.removeMaintainer()
		}
		return nil
	}
	if !maintainerDO.IsActive {
		log.Info("maintainer from db is not active")
		return c.tryToBeMaintainer()
	}
	c.setMaintainer(Maintainer{
		MaintainerInfo: maintainerDO.TaskMaintainer,
	}, maintainerDO.Gap)
	return nil
}

func (c *Coordinator) tryToBeMaintainer() error {
	log.Info("try to be maintainer...")
	newMaintainer := oceanbase.TaskMaintainer{
		AgentIp:   meta.OCS_AGENT.GetIp(),
		AgentPort: meta.OCS_AGENT.GetPort(),
		Counter:   1,
	}
	err := coordinatorService.UpdateMaintainerToOb(newMaintainer)
	if err != nil {
		log.WithError(err).Error("try to be maintainer failed")
		return err
	}
	c.setMaintainer(Maintainer{
		MaintainerInfo: newMaintainer,
	}, 0)
	return nil
}

func (c *Coordinator) buildMaintainerByPolling() error {
	log.Info("build maintainer by polling all agents")
	agentList, err := agentService.GetAllAgentsInfo()
	if err != nil {
		log.WithError(err).Error("get all agent from sqlite failed")
		return err
	}
	if len(agentList) < 2 {
		err = errors.New("agent list length less than 2")
		log.Warn(err)
		return err
	}
	log.Debug("agent list: ", agentList)
	for idx := range agentList {
		agent := agentList[idx]
		if meta.OCS_AGENT.Equal(&agent) {
			continue
		}
		if agent.Equal(c.Maintainer) {
			continue
		}

		if err := c.getMaintainerbyRpc(&agent); err != nil {
			log.WithError(err).Warnf("get maintainer from '%s' failed", agent.String())
			continue
		}
		return nil
	}
	return errors.New("can not get maintainer from rpc")
}

func (c *Coordinator) getMaintainerbyRpc(agentInfo meta.AgentInfoInterface) error {
	now := time.Now()
	log.Infof("try get maintainer rpc request from '%s' to '%s' ", meta.OCS_AGENT.String(), agentInfo.String())
	maintainer := Maintainer{}
	if err := secure.SendGetRequest(agentInfo, constant.URI_RPC_V1+constant.URI_MAINTAINER, nil, &maintainer); err != nil {
		return err
	}
	if maintainer.LifeTime > constant.MAINTAINER_MAX_ACTIVE_TIME_SEC {
		return fmt.Errorf("maintainer is not active, life time: %f", maintainer.LifeTime)
	}
	if meta.OCS_AGENT.Equal(&maintainer) {
		return errors.New("maintainer got by rpc is self")
	}
	c.setMaintainer(maintainer, maintainer.LifeTime+float64(time.Since(now).Seconds()))
	return nil
}

// GetMaintainer will get Coordinator Maintainer with LifeTime.
func GetMaintainer() (Maintainer, error) {
	if OCS_COORDINATOR == nil || OCS_COORDINATOR.IsFaulty() {
		return Maintainer{}, errors.New("coordinator is not initialized")
	} else {
		var maintainer = *OCS_COORDINATOR.Maintainer
		maintainer.LifeTime = float64(time.Since(OCS_COORDINATOR.Maintainer.GetLastUpdateTime()).Seconds())
		return maintainer, nil
	}
}

func (m *Maintainer) IsActive() bool {
	return m.GetExpirationTime().After(time.Now())
}

func (m *Maintainer) GetIp() string {
	return m.MaintainerInfo.AgentIp
}

func (m *Maintainer) GetPort() int {
	return m.MaintainerInfo.AgentPort
}

func (m *Maintainer) String() string {
	if m.agentInfo == nil {
		m.agentInfo = meta.NewAgentInfo(m.MaintainerInfo.AgentIp, m.MaintainerInfo.AgentPort)
	}
	return m.agentInfo.String()
}

func (m *Maintainer) GetLifeTime() float64 {
	return m.LifeTime
}

func (m *Maintainer) setLifeTime(lifeTime float64) {
	m.LifeTime = lifeTime
	m.LastUpdateTime = time.Now().Add(-time.Duration(lifeTime*1000000) * time.Microsecond)
	m.ExpirationTime = m.LastUpdateTime.Add(constant.MAINTAINER_MAX_ACTIVE_TIME)
}

func (m *Maintainer) GetLastUpdateTime() time.Time {
	return m.LastUpdateTime
}

func (m *Maintainer) GetExpirationTime() time.Time {
	return m.ExpirationTime
}

func (c *Coordinator) setMaintainer(maintainer Maintainer, lifeTime float64) {
	if !c.isSuspend {
		c.lock.Lock()
		defer c.lock.Unlock()
		if c.isSuspend {
			return
		}

		maintainer.setLifeTime(lifeTime)
		c.Maintainer = &maintainer
		log.Infof("set maintainer: %s, life time: %f, last update time: %s, expiration time: %s", maintainer.String(), maintainer.GetLifeTime(), maintainer.GetLastUpdateTime(), maintainer.GetExpirationTime())
		if meta.OCS_AGENT.Equal(&maintainer) {
			c.setIdendity(MAINTAINER)
		} else {
			c.setIdendity(WATCHER)
		}
	}
}

func (c *Coordinator) removeMaintainer() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.setIdendity(FAULTY)
	c.Maintainer = &Maintainer{}
}

func (c *Coordinator) addLifeTime(delta float64) {
	c.Maintainer.LifeTime += delta
}

func (c *Coordinator) setIdendity(identity int) {
	if c.identity != identity {
		log.Infof("coordinator identity change from %s to %s", identifiedMap[c.identity], identifiedMap[identity])
		c.identity = identity
		c.publish(identity == MAINTAINER)
	}
}

func (c *Coordinator) GetIdentiy() int {
	return c.identity
}

func (c *Coordinator) IsFaulty() bool {
	return c.identity == FAULTY
}

func (c *Coordinator) IsMaintainer() bool {
	return c.identity == MAINTAINER
}

func (c *Coordinator) HasMaintainer() bool {
	return c.Maintainer.GetIp() != "" && c.Maintainer.GetPort() != 0
}

func (c *Coordinator) IsWatcher() bool {
	return c.identity == WATCHER
}

func (c *Coordinator) Suspend() {
	c.isSuspend = true
	c.removeMaintainer()
}

func (c *Coordinator) Resume() {
	c.isSuspend = false
}
