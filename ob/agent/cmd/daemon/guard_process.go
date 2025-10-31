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
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/cmd"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
)

func (s *Server) guard(wg *sync.WaitGroup, ch chan os.Signal) {
	log.Info("daemon starts guarding obshell server")
	var err error
	startTimes := 1
	wg.Add(1)
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			log.Warnf("Panic:\n%s\n%s\n", err, buf[:n])
		}
		s.setState(constant.STATE_STOPPED)
		s.cleanup()
		wg.Done()
		log.Info("daemon stops guarding obshell server, send SIGTERM to daemon")
		ch <- syscall.SIGTERM
	}()

	// Clear root password, avoid to cover sqlite when agent restart.
	syscall.Unsetenv(constant.OB_ROOT_PASSWORD)
	for {
		// Service is starting, set its state to running.
		if s.state.IsStarting() {
			s.setState(constant.STATE_RUNNING)
		}

		svcState := s.state.GetState()
		// Service is stopped by daemon, exit guard.
		if svcState == constant.STATE_STOPPING || svcState == constant.STATE_STOPPED {
			log.Infof("obshell server stopped. state is %v. daemon no longer guard", svcState)
			return
		}

		procState := s.proc.GetState()
		// Process is exited, determine whether to exit guard or restart.
		if procState.Exited {
			if err = s.handleProcExited(procState, &startTimes); err != nil {
				log.WithError(err).Error("guarded process exit")
				return
			}
			continue
		}

		startTimes = 1
		time.Sleep(time.Second)
	}
}

func (s *Server) handleBackupBin() (err error) {
	if !cmd.BackupBinExist() {
		if err = cmd.BackupBin(); err != nil {
			return errors.Wrap(err, "backup binary failed")
		}
	}

	// If an error occurs, the following judgment always holds.
	binVersion, _ := system.GetBinaryVersion(path.ObshellBinPath())
	if pkg.CompareVersion(binVersion, constant.VERSION_RELEASE) == 1 {
		log.Infof("obshell binary version is %s, but my version is %s", binVersion, constant.VERSION_RELEASE)
		cmd := exec.Command(path.ObshellBinPath(), cmd.CMD_ADMIN, cmd.CMD_RESTART,
			fmt.Sprintf("--%s", constant.FLAG_IP), s.agent.GetIp(),
			fmt.Sprintf("--%s", constant.FLAG_PORT), fmt.Sprint(s.agent.GetPort()),
			fmt.Sprintf("--%s", constant.FLAG_PID), fmt.Sprint(s.GetPid()),
		)
		log.Infof("call admin restart: %s", cmd.String())
		if err = cmd.Run(); err != nil {
			return errors.Wrap(err, "restart failed")
		}
	}
	return
}
