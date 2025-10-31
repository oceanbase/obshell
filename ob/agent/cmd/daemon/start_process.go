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
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/process"
)

func (s *Server) Start() (err error) {
	if s.state.IsRunning() {
		return nil
	}

	if !s.casState(constant.STATE_STOPPED, constant.STATE_STARTING) {
		return errors.Occur(errors.ErrCommonUnexpected, "obshell server already started")
	}

	if err = s.startProc(); err != nil {
		s.cleanup()
		return
	}
	log.Info("obshell server started")
	return nil
}

func (s *Server) startProc() (err error) {
	startTimes := 1
	if err = s.startProcWithCount(&startTimes, nil); err != nil {
		return
	}

	for {
		// Service is starting, set its state to running.
		if s.state.IsStarting() {
			s.setState(constant.STATE_RUNNING)
		}

		// Get real state of obshell server via unix socket, and
		// if obshell server is running, break the loop.
		if s.IsRunning() {
			break
		}

		// Wait for obshell server to start or exit.
		time.Sleep(100 * time.Millisecond)

		procState := s.proc.GetState()
		if procState.Exited {
			if err = s.handleProcExited(procState, &startTimes); err != nil {
				log.Error(err)
				return
			}
		}
	}
	return nil
}

func (s *Server) startProcWithCount(count *int, procState *process.ProcState) (err error) {
	s.setState(constant.STATE_STOPPED)

	log.Infof("start the obshell server for the  %d time", *count)
	*count++
	if procState != nil {
		liveTime := procState.EndAt.Sub(procState.StartAt)
		if s.conf.MinLiveTime > 0 && liveTime < s.conf.MinLiveTime {
			log.Warnf("obshell server exited too quickly. live time: %d, MinLiveTime: %d, count: %d", liveTime, s.conf.MinLiveTime, count)
			if *count > s.conf.QuickExitLimit {
				return errors.Occur(errors.ErrCommonUnexpected, "daemon retry limit exceeded")
			}
		}
	}

	s.setState(constant.STATE_STARTING)
	if err = s.startServerProc(); err != nil {
		s.setState(constant.STATE_STOPPED)
		return
	}
	return nil
}

func (s *Server) startServerProc() (err error) {
	s.cleanup()
	log.Info("starting obshell server")
	if err := os.Chdir(global.HomePath); err != nil {
		return err
	}
	if err = s.proc.Start(); err != nil {
		s.cleanup()
		return errors.Wrap(err, "failed to start obshell server")
	}
	if err := s.writePid(); err != nil {
		log.WithError(err).Error("failed to write pid file")
	}
	log.Infof("obshell server pid is %d", s.GetPid())
	return nil
}

func (s *Server) writePid() error {
	return process.WritePid(path.ObshellPidPath(), s.GetPid())
}

func (s *Server) handleProcExited(procState process.ProcState, count *int) (err error) {
	log.Warnf("obshell server exited with code %d, state %v", procState.ExitCode, s.state.GetState())
	switch procState.ExitCode {
	// If process is exited normally or started failed, exit guarding.
	case constant.EXIT_CODE_ERROR_INVAILD_AGENT,
		constant.EXIT_CODE_ERROR_NOT_CLUSTER_AGENT,
		constant.EXIT_CODE_ERROR_IP_NOT_MATCH,
		constant.EXIT_CODE_ERROR_AGENT_START_FAILED,
		constant.EXIT_CODE_ERROR_SERVER_LISTEN,
		constant.EXIT_CODE_ERROR_OB_START_FAILED,
		constant.EXIT_CODE_ERROR_OB_CONN_TIMEOUT,
		constant.EXIT_CODE_ERROR_PERMISSION_DENIED,
		constant.EXIT_CODE_ERROR_OB_PWD_ERROR,
		constant.EXIT_CODE_ERROR_BACKUP_BINARY_FAILED,
		constant.EXIT_CODE_ERROR_TAKE_OVER_FAILED,
		constant.EXIT_CODE_ERROR_EXEC_BINARY_FAILED:
		return fmt.Errorf("obshell server exited with code %d, please check obshell.log for more details", procState.ExitCode)
	default:
		// If process is exited abnormally, restart the process.
		return s.startProcWithCount(count, &procState)
	}
}
