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

package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/api/web"
	"github.com/oceanbase/obshell/seekdb/agent/cmd"
	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	ocsagentlog "github.com/oceanbase/obshell/seekdb/agent/log"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/client/command"
)

func NewServerCmd() *cobra.Command {
	opts := &cmd.CommonFlag{}
	serverCmd := command.NewCommand(&cobra.Command{
		Use:    cmd.CMD_SERVER,
		Hidden: true,
		Args:   cobra.NoArgs,
	})
	serverCmd.RunE = func(c *cobra.Command, args []string) (err error) {
		opts.HiddenPassword()
		ocsagentlog.InitLogger(config.DefaultAgentLoggerConifg())
		ocsagentlog.SetDBLoggerLevel(ocsagentlog.Error)
		log.SetLevel(log.DebugLevel)
		server := newAgent(opts)
		server.start()
		return nil
	}
	cmd.SetCommandFlags(serverCmd, opts)
	return serverCmd.Command
}

type Agent struct {
	*cmd.CommonFlag
	Server

	upgradeMode  bool
	obHasStarted bool
}

type Server struct {
	server        *web.Server
	tmpServer     web.Server // Only used for upgrade.
	tmpSocketPath string
	startChan     chan bool // Start listening on the HTTP server to handle incoming web requests.
}

func newAgent(flag *cmd.CommonFlag) *Agent {
	return &Agent{
		CommonFlag: flag,
	}
}

func (a *Agent) start() {
	if err := a.init(); err != nil {
		log.WithError(err).Error("initialize failed")
		process.ExitWithError(constant.EXIT_CODE_ERROR_AGENT_START_FAILED, errors.Wrap(err, "initialize failed"))
	}
	if err := a.run(); err != nil {
		log.WithError(err).Error("run failed")
		process.ExitWithError(constant.EXIT_CODE_ERROR_AGENT_START_FAILED, errors.Wrap(err, "run failed"))
	}
	a.cleanup()
	a.wait()
}

// cleanup is only  used for upgrade.
func (a *Agent) cleanup() {
	if a.upgradeMode {
		log.Infof("remove old ocsagent socket %s", a.tmpSocketPath)
		os.Remove(a.tmpSocketPath)
		log.Info("set old ocsagent pid '0'")
		a.OldServerPid = 0
	}
}

func (a *Agent) wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		log.Infof("obshell server received '%s' signal. exiting...", sig.String())
		a.server.Stop()
		sqliteDb, _ := sqlite.GetSqliteInstance()
		db, _ := sqliteDb.DB()
		db.Close()
		process.ExitWithMsg(constant.EXIT_CODE_NOTIFY_SIGNAL, fmt.Sprintf("obshell server received '%s' signal.", sig.String()))
	}
}
