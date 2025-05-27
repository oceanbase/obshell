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
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/cmd/server"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
)

func (d *Daemon) Start() (err error) {
	if d.isForUpgrade() {
		log.Info("start daemon for upgrade")
		d.upgradeMode = true
		d.server.upradeMode = true
		d.server.oldPid = d.oldSvrPid
	}
	return d.start()
}

func (d *Daemon) start() (err error) {
	if d.state.IsRunning() {
		return nil
	}

	d.cleanup()
	log.Info("start daemon")
	if err = d.writePid(); err != nil {
		return
	}
	d.setState(constant.STATE_STARTING)

	socketListener, err := d.newSocketListener()
	if err != nil {
		return
	}
	go d.startSocket(socketListener)
	go d.ListenSignal()

	s := d.server
	defer s.proc.SwitchToLogMode()
	if err = s.Start(); err != nil {
		log.WithError(err).Error("start obshell server failed")
		return
	}

	if d.upgradeMode {
		if err = server.WaitServerProcKilled(d.oldSvrPid); err != nil {
			log.WithError(err).Error("wait old obshell server exit failed")
			return
		}

		d.upgradeMode = false
		socketListener, err = d.newSocketListener()
		if err != nil {
			return
		}
		go d.startSocket(socketListener)

		log.Info("close tmp local http server")
		d.tmpLocalHttpServer.Close()
		s.upradeMode = false
	}

	s.done = make(chan struct{})
	go s.guard(&d.wg, d.ch)
	return nil
}

func (d *Daemon) newSocketListener() (socketListener *net.UnixListener, err error) {
	d.initLocalServer()
	socketPath := d.getSocketPath()
	log.Infof("start socket server on %s", socketPath)
	socketListener, err = http.NewSocketListener(socketPath)
	return
}

func (d *Daemon) startSocket(socketListener *net.UnixListener) {
	if d.upgradeMode {
		go d.tmpLocalHttpServer.Serve(socketListener)
	} else {
		go func() {
			defer func() {
				d.setState(constant.STATE_STOPPED)
			}()
			err := d.localHttpServer.Serve(socketListener)
			if err != nil && d.state.IsStarting() {
				log.WithError(err).Error("daemon serve on socket listener failed")
				process.ExitWithFailure(constant.EXIT_CODE_ERROR_SERVER_LISTEN, fmt.Sprintf("daemon serve on socket listener failed: %s\n", err))
			} else {
				d.setState(constant.STATE_RUNNING)
			}
		}()
	}
	time.Sleep(1 * time.Second)
	d.setState(constant.STATE_RUNNING)
}

func (d *Daemon) writePid() (err error) {
	pid := os.Getpid()
	log.Info("obshell daemon pid is ", pid)
	return process.WritePid(path.DaemonPidPath(), pid)
}

func (d *Daemon) isForUpgrade() bool {
	if d.oldSvrPid == 0 || d.agent.GetIp() == "" || d.agent.GetPort() == 0 {
		log.Info("daemon is not for upgrade")
		return false
	}
	var status http.AgentStatus
	err := http.SendGetRequestViaUnixSocket(path.ObshellSocketBakPath(), constant.URI_API_V1+constant.URI_STATUS, nil, &status)
	if err != nil {
		log.WithError(err).Error("failed to get obshell status")
		return false
	}
	return status.Pid == int(d.oldSvrPid) && status.Agent.Equal(&d.agent)
}
