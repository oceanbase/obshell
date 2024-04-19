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

package http

import (
	"sync/atomic"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
)

// Status info api response
type AgentStatus struct {
	Agent meta.AgentInfoWithIdentity `json:"agent"`

	State            int32  `json:"state"`   // service state
	Version          string `json:"version"` // service version
	Pid              int    `json:"pid"`     // service pid
	StartAt          int64  `json:"startAt"` // timestamp when service started
	Port             int    `json:"port"`    // Ports process occupied ports
	OBState          int    `json:"obState"`
	UnderMaintenance bool   `json:"underMaintenance"`
}

type State struct {
	state int32
}

func NewState(state int32) *State {
	return &State{
		state: state,
	}
}

func (s *State) SetState(state int32) {
	s.state = state
}

func (s *State) GetState() int32 {
	return s.state
}

func (s *State) CasState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&(s.state), old, new)
}

func (s *State) IsStarting() bool {
	return s.state == constant.STATE_STARTING
}

func (s *State) IsRunning() bool {
	return s.state == constant.STATE_RUNNING
}

func (s *State) IsStopping() bool {
	return s.state == constant.STATE_STOPPING
}

func (s *State) IsStopped() bool {
	return s.state == constant.STATE_STOPPED
}
