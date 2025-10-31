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
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/cmd"
	"github.com/oceanbase/obshell/ob/agent/cmd/daemon"
	"github.com/oceanbase/obshell/ob/agent/config"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/process"
	ocsagentlog "github.com/oceanbase/obshell/ob/agent/log"
	"github.com/oceanbase/obshell/ob/client/command"
)

func newStartCmd() *cobra.Command {
	opts := &cmd.CommonFlag{}
	startCmd := command.NewCommand(&cobra.Command{
		Use:    cmd.CMD_START,
		Hidden: true,
		Args:   cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			opts.HiddenPassword()
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			admin := NewAdmin(opts)
			if err := admin.StartDaemon(); err != nil {
				process.ExitWithFailure(constant.EXIT_CODE_ERROR_ADMIN_START_FAILED, err.Error())
			}
		},
	})
	cmd.SetCommandFlags(startCmd, opts)
	return startCmd.Command
}

func (a *Admin) StartDaemon() (err error) {
	ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
	global.InitGlobalVariable()

	if daemon.IsDaemonRunning() {
		log.Info("daemon process is running")
		return nil
	}

	if err = a.makeWorkDir(); err != nil {
		return errors.Wrap(err, "failed to make all dir")
	}

	log.Info("start daemon process")
	daemonProc := a.newDaemonProc()
	log.Info("change work dir to ", global.HomePath)
	if err := os.Chdir(global.HomePath); err != nil {
		return err
	}
	defer func() {
		daemonProc.SwitchToLogMode()
	}()

	if err = daemonProc.Start(); err != nil {
		log.WithError(err).Error("failed to start daemon process")
		return
	}

	for {
		if !daemonProc.IsRunning() {
			daemonProc.SwitchToLogMode()
			process.Exit(constant.EXIT_CODE_ERROR_ADMIN_START_FAILED)
		}
		if ready, _ := a.isDaemonReady(); ready {
			log.Info("daemon process started successfully")
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (a *Admin) makeWorkDir() (err error) {
	for _, dir := range []string{path.RunDir(), path.LogDir()} {
		log.Infof("create dir %s", dir)
		if err = os.MkdirAll(dir, 0755); err != nil {
			log.WithError(err).Errorf("failed to create dir %s", dir)
			return
		}
	}
	return nil
}

func (a *Admin) newDaemonProc() *process.Process {
	args := a.getDaemonArgs()
	return process.NewProcess(process.ProcessConfig{
		Program:     path.DaemonBinPath(),
		Args:        args,
		LogFilePath: path.DaemonLogPath(),
	})
}

func (a *Admin) getDaemonArgs() []string {
	args := []string{constant.PROC_OBSHELL_DAEMON}
	if a.flags == nil {
		return args
	}
	return append(args, a.flags.GetArgs()...)
}

func (a *Admin) isDaemonReady() (res bool, err error) {
	log.Info("admin get daemon status")
	status := daemon.DaemonStatus{}
	err = http.SendGetRequestViaUnixSocket(a.getDaemonSocketPath(), constant.URI_API_V1+constant.URI_STATUS, nil, &status)
	if err != nil {
		log.WithError(err).Warn("failed to get daemon status")
		return false, err
	}
	return status.Ready, nil
}

func (a *Admin) getDaemonSocketPath() string {
	if a.upgradeMode {
		return path.DaemonSocketTmpPath()
	} else {
		return path.DaemonSocketPath()
	}
}
