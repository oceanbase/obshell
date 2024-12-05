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
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/cmd"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/global"
	http2 "github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/client/command"
)

func NewDaemonCmd() *cobra.Command {
	opts := &cmd.CommonFlag{}
	daemonCmd := command.NewCommand(&cobra.Command{
		Use:    cmd.CMD_DAEMON,
		Hidden: true,
		Args:   cobra.NoArgs,
	})
	daemonCmd.RunE = func(c *cobra.Command, args []string) (err error) {
		opts.HiddenPassword()
		ocsagentlog.InitLogger(config.DefaultDaemonLoggerConifg())
		global.InitGlobalVariable()
		daemon := newDaemon(opts)
		if err = daemon.Start(); err != nil {
			process.ExitWithMsg(constant.EXIT_CODE_ERROR_DAEMON_START_FAILED, err.Error())
		}
		daemon.ListenSignal()
		return
	}
	cmd.SetCommandFlags(daemonCmd, opts)
	return daemonCmd.Command
}

type Daemon struct {
	agent              meta.AgentInfo
	oldSvrPid          int32
	isTakeOver         int
	localHttpServer    *http.Server
	localRouter        *gin.Engine
	server             *Server
	state              *http2.State
	upgradeMode        bool
	tmpLocalHttpServer *http.Server // Only used for upgrade.
	wg                 sync.WaitGroup
	ch                 chan os.Signal
}

type DaemonStatus struct {
	State        int32        `json:"state"`
	Ready        bool         `json:"ready"`
	Version      string       `json:"version"`
	Pid          int          `json:"pid"`
	Socket       string       `json:"socket"`
	ServerStatus ServerStatus `json:"server"`
	StartAt      int64        `json:"startAt"`
}

func newDaemon(flag *cmd.CommonFlag) *Daemon {
	daemon := &Daemon{
		agent:       flag.AgentInfo,
		oldSvrPid:   flag.OldServerPid,
		isTakeOver:  flag.IsTakeover,
		server:      newObshellServer(flag),
		localRouter: gin.New(),
		state:       http2.NewState(constant.STATE_STOPPED),
		wg:          sync.WaitGroup{},
		ch:          make(chan os.Signal, 1),
	}
	daemon.initRoutes()
	return daemon
}

var startAt = time.Now().UnixNano()

func (d *Daemon) initRoutes() {
	d.localRouter.Use(
		gin.CustomRecovery(common.Recovery),
		common.PostHandlers(),
	)

	v1 := d.localRouter.Group(constant.URI_API_V1)
	v1.GET(constant.URI_STATUS, d.GetStatus)
}

// GetStatus returns the status of the daemon. The relative route is /api/v1/status.
func (d *Daemon) GetStatus(c *gin.Context) {
	ready := d.IsReady()
	status := DaemonStatus{
		State:        d.state.GetState(),
		Ready:        ready,
		Version:      constant.VERSION,
		Pid:          os.Getpid(),
		Socket:       d.getSocketPath(),
		ServerStatus: d.GetServerStatus(),
		StartAt:      startAt,
	}
	common.SendResponse(c, status, nil)
}

func (d *Daemon) getSocketPath() string {
	if d.upgradeMode {
		return path.DaemonSocketTmpPath()
	} else {
		return path.DaemonSocketPath()
	}
}

func (d *Daemon) initLocalServer() {
	server := &http.Server{
		Handler:      d.localRouter,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
	}
	if d.upgradeMode {
		d.tmpLocalHttpServer = server
	} else {
		d.localHttpServer = server
	}
}

func (d *Daemon) GetServerStatus() ServerStatus {
	return d.server.GetStatus()
}

func (d *Daemon) IsReady() bool {
	return d.state.IsRunning() && d.server.IsRunning()
}

func (d *Daemon) setState(state int32) {
	log.Infof("set daemon state to %d", state)
	d.state.SetState(state)
}

func (d *Daemon) casState(old, new int32) bool {
	log.Infof("cas daemon state from %d to %d", old, new)
	return d.state.CasState(old, new)
}
