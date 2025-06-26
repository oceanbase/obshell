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

package admin

import (
	"syscall"
	"time"

	proc "github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/cmd"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/lib/system"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
)

func newStopCmd() *cobra.Command {
	stopCmd := command.NewCommand(&cobra.Command{
		Use:    cmd.CMD_STOP,
		Hidden: true,
		Args:   cobra.NoArgs,
	})
	stopCmd.RunE = func(c *cobra.Command, args []string) (err error) {
		c.SilenceUsage = true
		ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
		admin := NewAdmin(nil)
		return admin.StopDaemon()
	}
	return stopCmd.Command
}

func (a *Admin) StopDaemon() (err error) {
	if !system.IsFileExist(path.DaemonPidPath()) {
		log.Info("daemon process's pid file not exist")
		return nil
	}

	pid, err := process.GetDaemonPid()
	if err != nil {
		return err
	}
	if err = a.stopDaemonByPid(pid); err != nil {
		return err
	}
	return nil
}

func (a *Admin) stopDaemonByPid(pid int32) (err error) {
	p, err := proc.NewProcess(pid)
	if err != nil {
		log.Info("daemon process not running")
		return nil
	}
	if err = p.SendSignal(syscall.SIGTERM); err != nil {
		return errors.Wrap(err, "failed to send SIGTERM to daemon process")
	}

	for i := 0; i < WAIT_DAEMON_TIME_LIMIT; i++ {
		if _, err := proc.NewProcess(pid); err != nil {
			log.Info("daemon stopped")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.Occur(errors.ErrCommonUnexpected, "wait for daemon process exit timeout")
}
