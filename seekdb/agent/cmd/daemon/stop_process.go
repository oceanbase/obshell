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
	"os"
	"strconv"
	"time"

	proc "github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/lib/system"
)

const (
	waitObshellStopTimeout = 10 * time.Second
	waitObshellKillTimeout = 5 * time.Second
)

func (s *Server) Stop() (err error) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	if s.state.IsStopped() {
		return nil
	}

	s.setState(constant.STATE_STOPPING) // State may be running, staring or stopping.
	log.Info("stopping obshell")

	if err = s.proc.Stop(); err != nil {
		log.WithError(err).Warn("failed to send TERM signal to obshell")
	}

	if !s.proc.WaitForExit(waitObshellStopTimeout) {
		log.Warn("obshell did not exit after TERM signal, try KILL it")
		if err = s.proc.Kill(); err != nil && s.proc.IsRunning() {
			return errors.Wrap(err, "failed to kill obshell")
		}
		if !s.proc.WaitForExit(waitObshellKillTimeout) {
			return errors.Occur(errors.ErrCommonUnexpected, "wait for obshell process exit timeout")
		}
	}

	s.setState(constant.STATE_STOPPED)
	s.cleanup()
	return nil
}

func (s *Server) cleanup() {
	socketPath := s.getSocketPath()
	if http.IsSocketFile(socketPath) {
		if !http.SocketCanConnect("unix", socketPath, time.Second) {
			log.Infof("remove obshell socket file %s", socketPath)
			os.Remove(socketPath)
		}
	}

	if system.IsFileExist(path.ObshellPidPath()) {
		obshellPidStr, err := process.GetObshellPidStr()
		if err != nil {
			log.WithError(err).Error("failed to get obshell pid")
			return
		}
		obshellPid, err := strconv.Atoi(obshellPidStr)
		if err != nil {
			log.Infof("unexpect obshell pid: %s, delete directly", obshellPidStr)
			os.Remove(path.ObshellPidPath())
		}

		log.Info("the stopped obshell pid is ", obshellPid)

		if pidInfo, err := proc.NewProcess(int32(obshellPid)); err != nil {
			os.Remove(path.ObshellPidPath())
		} else {
			if name, err := pidInfo.Name(); err == nil && name != constant.PROC_OBSHELL {
				os.Remove(path.ObshellPidPath())
			}
		}
	}
}
