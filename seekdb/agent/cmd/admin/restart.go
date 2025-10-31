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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/cmd"
	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/lib/system"
	ocsagentlog "github.com/oceanbase/obshell/seekdb/agent/log"
	"github.com/oceanbase/obshell/seekdb/client/command"
)

func newRestartCmd() *cobra.Command {
	opts := &cmd.CommonFlag{}
	restartCmd := command.NewCommand(&cobra.Command{
		Use:    cmd.CMD_RESTART,
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) (err error) {
			opts.HiddenPassword()
			c.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			opts.BaseDir = path.AgentDir()
			admin := NewAdmin(opts)
			return admin.RestartDaemon()
		},
	})
	cmd.SetCommandFlags(restartCmd, opts)
	return restartCmd.Command
}

func (a *Admin) RestartDaemon() (err error) {
	if a.isForUpgrade() {
		a.upgradeMode = true
		return a.restartDaemonForUpgrade()
	} else {
		a.agent = nil
		a.oldServerPid = 0
		return a.restartDaemon()
	}
}

func (a *Admin) isForUpgrade() bool {
	if a.oldServerPid == 0 || a.agent == nil {
		log.Info("admin is not for upgrade")
		return false
	}
	var status http.AgentStatus
	err := http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), constant.URI_API_V1+constant.URI_STATUS, nil, &status)
	if err != nil {
		log.WithError(err).Error("failed to get status")
		return false
	}
	return status.Pid == int(a.oldServerPid) && status.Agent.Equal(a.agent)
}

func (a *Admin) restartDaemon() (err error) {
	log.Info("restart daemon")
	if err = a.StopDaemon(); err != nil {
		return errors.Wrap(err, "failed to stop daemon")
	}

	if err = a.StartDaemon(); err != nil {
		return errors.Wrap(err, "failed to start daemon")
	}
	return nil
}

func (a *Admin) restartDaemonForUpgrade() (err error) {
	log.Info("restart daemon for upgrade")

	if err = a.backup(); err != nil {
		return errors.Wrap(err, "failed to backup pid")
	}

	log.Info("start new daemon process")
	if err = a.StartDaemon(); err != nil {
		if err := a.StopDaemon(); err != nil {
			log.WithError(err).Error("failed to stop daemon")
		}
		a.restore()
		return errors.Wrap(err, "failed to start daemon")
	}

	log.Info("stop old daemon process")
	if err = a.stopDaemonByPid(a.daemonPid); err != nil {
		log.WithError(err).Error("failed to stop daemon")
		return err
	}

	a.cleanupBackup()
	return nil
}

func (a *Admin) backup() (err error) {
	a.daemonPid, err = backupPid(path.DaemonPidPath(), path.DaemonPidBakPath())
	if err != nil {
		return errors.Wrap(err, "failed to backup daemon pid")
	}
	a.oldServerPid, err = backupPid(path.ObshellPidPath(), path.ObshellPidBakPath())
	if err != nil {
		return errors.Wrap(err, "failed to backup obshell pid")
	}
	if err = a.backupSocket(path.DaemonSocketPath(), path.DaemonSocketBakPath()); err != nil {
		return errors.Wrap(err, "failed to backup daemon socket")
	}
	if err = a.backupSocket(path.ObshellSocketPath(), path.ObshellSocketBakPath()); err != nil {
		return errors.Wrap(err, "failed to backup obshell socket")
	}
	return nil
}

func backupPid(pidPath, bakPath string) (pid int32, err error) {
	log.Info("backup pidfile ", pidPath)
	if !system.IsFileExist(pidPath) {
		return 0, nil
	}

	if pid, err = process.GetPid(pidPath); err != nil {
		os.Remove(pidPath)
		return 0, nil
	}
	if err = os.Rename(pidPath, bakPath); err != nil {
		return 0, err
	}
	return pid, nil
}

func (a *Admin) backupSocket(src, dest string) (err error) {
	log.Info("backup socket ", src)
	if !system.IsFileExist(src) {
		return nil
	}
	if err := os.Rename(src, dest); err != nil {
		return err
	}
	return nil
}

func (a *Admin) restore() {
	if a.daemonPid != 0 {
		if err := os.Rename(path.DaemonPidBakPath(), path.DaemonPidPath()); err != nil {
			log.WithError(err).Error("failed to restore daemon pid")
		}
	}
	if a.oldServerPid != 0 {
		if err := os.Rename(path.ObshellPidBakPath(), path.ObshellPidPath()); err != nil {
			log.WithError(err).Error("failed to restore obshell pid")
		}
	}

	if err := os.Rename(path.DaemonSocketBakPath(), path.DaemonSocketPath()); err != nil {
		log.WithError(err).Error("failed to restore daemon socket")
	}
	if err := os.Rename(path.ObshellSocketBakPath(), path.ObshellSocketPath()); err != nil {
		log.WithError(err).Error("failed to restore obshell socket")
	}
}

func (a *Admin) cleanupBackup() {
	var err error
	if err = os.Remove(path.DaemonPidBakPath()); err != nil {
		log.WithError(err).Error("failed to remove daemon pid bak")
	}
	if err = os.Remove(path.ObshellPidBakPath()); err != nil {
		log.WithError(err).Error("failed to remove obshell pid bak")
	}
	if err = os.Remove(path.DaemonSocketBakPath()); err != nil {
		log.WithError(err).Error("failed to remove daemon socket bak")
	}
	if err = os.Remove(path.ObshellSocketBakPath()); err != nil {
		log.WithError(err).Error("failed to remove obshell socket bak")
	}
}
