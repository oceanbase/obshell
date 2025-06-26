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
	"os/signal"
	"syscall"
	"time"

	proc "github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
)

func (d *Daemon) ListenSignal() {
	signal.Notify(d.ch, syscall.SIGTERM, syscall.SIGINT)
	sig := <-d.ch
	log.Infof("signal '%s' received. exiting...", sig.String())
	if err := d.stop(); err != nil {
		log.Error(err)
	}
}

func (d *Daemon) stop() error {
	if d.state.IsStopped() {
		return nil
	}
	log.Info("stop daemon")
	if !d.casState(constant.STATE_RUNNING, constant.STATE_STOPPING) {
		return errors.Occur(errors.ErrCommonUnexpected, "daemon is not running")
	}

	svc := d.server
	state := svc.state
	log.Infof("stop obshell server, state is %v", state)
	if !state.IsStopped() {
		if err := svc.Stop(); err != nil {
			log.WithError(err).Error("stop obshell server failed")
		}
	}

	d.setState(constant.STATE_STOPPED)
	d.cleanup()
	d.wg.Wait()
	if d.localHttpServer != nil {
		d.localHttpServer.Close()
	}
	return nil
}

func (d *Daemon) cleanup() {
	log.Info("cleanup daemon")
	sockPath := path.DaemonSocketPath()
	if http.IsSocketFile(sockPath) {
		log.Infof("socket %s is a socket", sockPath)
		if !http.SocketCanConnect("unix", sockPath, time.Second) {
			log.Infof("socket %s is not connectable, remove it", sockPath)
			os.Remove(sockPath)
		}
	}
	if !IsDaemonRunning() {
		os.Remove(path.DaemonPidPath())
	}
}

func IsDaemonRunning() bool {
	pid, err := process.GetDaemonPid()
	if err != nil {
		// If an error occurs, it can be assumed that the daemon process no longer exists or,
		// even if it does, it is not under the control of the current user.
		return false
	}
	if pidInfo, err := proc.NewProcess(pid); err != nil {
		return false
	} else {
		if name, err := pidInfo.Name(); err == nil && name != constant.PROC_OBSHELL {
			return false
		}
	}
	return true
}
