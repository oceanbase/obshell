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

package daemon

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/cmd"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
)

type Server struct {
	agent      meta.AgentInfo
	upradeMode bool
	oldPid     int32
	conf       ServerConfig
	proc       *process.Process
	done       chan struct{}
	state      *http.State
}

type ServerStatus struct {
	Status http.AgentStatus
	Socket string `json:"socket"`
	EndAt  int64  `json:"endAt"`
}

type ServerConfig struct {
	MinLiveTime    time.Duration
	QuickExitLimit int
}

func newObshellServer(flag *cmd.CommonFlag) *Server {
	args := getServerArgs(flag)
	log.Info(fmt.Sprintf("start obshell server with args: %v", args))
	proc := process.NewProcess(process.ProcessConfig{
		Program:     path.ObshellBinPath(),
		Args:        args,
		LogFilePath: path.ObshellStdPath(),
	})
	return &Server{
		agent: flag.AgentInfo,
		done:  nil,
		state: http.NewState(constant.STATE_STOPPED),
		proc:  proc,
		conf: ServerConfig{
			MinLiveTime:    time.Second * 3,
			QuickExitLimit: 10,
		}}
}

func getServerArgs(flag *cmd.CommonFlag) []string {
	args := []string{constant.PROC_OBSHELL_SERVER}
	return append(args, flag.GetArgs()...)
}

// GetStatus get status of obshell server via unix socket
func (s *Server) GetStatus() (res ServerStatus) {
	procState := s.proc.GetState()
	res.EndAt = procState.EndAt.UnixNano()
	if s.IsProcRunning() {
		if ret, err := s.getRealStatus(); err == nil {
			res.Status = ret
			s.agent = ret.Agent.AgentInfo
			res.Socket = s.getSocketPath()
			return
		}
	}
	res.Status = http.AgentStatus{
		State:   constant.STATE_STOPPED,
		Pid:     procState.Pid,
		StartAt: procState.StartAt.UnixNano(),
	}
	return
}

func (s *Server) IsProcRunning() bool {
	return s.proc.IsRunning()
}

func (s *Server) getRealStatus() (res http.AgentStatus, err error) {
	if s.upradeMode {
		err = http.SendGetRequestViaUnixSocket(path.ObshellTmpSocketPath(), constant.URI_API_V1+constant.URI_STATUS, nil, &res)
	} else {
		err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), constant.URI_API_V1+constant.URI_STATUS, nil, &res)
	}
	return
}

func (s *Server) getSocketPath() string {
	if !s.upradeMode {
		return path.ObshellSocketPath()
	}
	return path.ObshellTmpSocketPath()
}

func (s *Server) GetPid() int {
	return s.GetStatus().Status.Pid
}

func (s *Server) GetRealState() int32 {
	return s.GetStatus().Status.State
}

func (s *Server) IsRunning() bool {
	status := s.GetStatus()
	log.Infof("obshell server state: %d, agent is unidentified: %v", status.Status.State, status.Status.Agent.IsUnidentified())
	return status.Status.State == constant.STATE_RUNNING
}

func (s *Server) casState(old, new int32) bool {
	log.Infof("cas obshell server state from %d to %d", old, new)
	return s.state.CasState(old, new)
}

func (s *Server) setState(state int32) {
	log.Infof("set obshell server state to %d", state)
	s.state.SetState(state)
}
