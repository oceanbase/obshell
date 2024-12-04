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
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/api/web"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/agent"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/executor/pool"
	"github.com/oceanbase/obshell/agent/executor/recyclebin"
	"github.com/oceanbase/obshell/agent/executor/script"
	"github.com/oceanbase/obshell/agent/executor/tenant"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
	agentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/secure"
	agentservice "github.com/oceanbase/obshell/agent/service/agent"
)

var agentService = agentservice.AgentService{}

// init will initialize all modules,
// including log, global, sqlite, agent, httpServer, and task registration.
func (a *Agent) init() (err error) {
	a.initLogger()
	global.InitGlobalVariable()
	a.isOBhasStarted()
	a.isUpgradeMode()

	if err = a.initSqlite(); err != nil {
		return errors.Wrap(err, "init sqlite failed")
	}

	if err = secure.Init(); err != nil {
		return errors.Wrap(err, "secure init failed")
	}

	if err = a.initAgent(); err != nil {
		return errors.Wrap(err, "init agent failed")
	}

	if err = a.preCheckForUpgrade(); err != nil {
		return errors.Wrap(err, "pre check for upgrade failed")
	}

	a.initServer()
	a.initTask()
	return nil
}

// initLogger initializes logger module with default config
func (a *Agent) initLogger() {
	log.Info("initialize logger")
	agentlog.InitLogger(config.DefaultAgentLoggerConifg())
}

// initSqlite loads the sqlite instance and migrate the tables when necessary,
// and set the agent to the running state
func (a *Agent) initSqlite() (err error) {
	log.Info("initialize sqlite")
	if err = sqlite.LoadSqliteInstance(); err != nil {
		return err
	}
	if err = sqlite.MigrateSqliteTables(a.upgradeMode); err != nil {
		return errors.Wrap(err, "migrate sqlite tables failed")
	}
	return nil
}

// initServerForUpgrade will only start the unix socket service  When upgrading.
func (a *Agent) initServerForUpgrade() error {
	log.Info("init local server [upgrade mode]")
	serverConfig := config.ServerConfig{
		Ip:          "0.0.0.0",
		Port:        meta.OCS_AGENT.GetPort(),
		Address:     fmt.Sprintf("0.0.0.0:%d", meta.OCS_AGENT.GetPort()),
		RunDir:      path.RunDir(),
		UpgradeMode: true,
	}
	a.server = web.NewServerOnlyLocal(config.DebugMode, serverConfig)
	socketListener, err := a.server.NewUnixListener()
	if err != nil {
		return err
	}
	a.tmpSocketPath = a.server.SocketPath()
	a.tmpServer = *a.server
	a.server.UnixListener = socketListener
	a.server.RunLocalServer()
	return nil
}

// WaitServerProcKilled will wait for the old agent to exit When upgrading.
// If the old agent does not exit within 10 minutes, an error will be returned.
func WaitServerProcKilled(pid int32) error {
	log.Infof("wait %d killed", pid)
	for i := 0; i < constant.AGENT_START_TIMEOUT; i++ {
		exist, err := process.CheckProcessExist(pid)
		if err != nil {
			return errors.Wrap(err, "check server proc failed")
		}
		if !exist {
			log.Infof("%d killed", pid)
			time.Sleep(5 * time.Second)
			return nil
		}
		log.Infof("%d still exist, wait 1s", pid)
		time.Sleep(time.Second)
	}
	return errors.New("wait obshell server killed timeout")
}

// initAgent will get the final agent info based on meta ,incoming configuration, and default value.
func (a *Agent) initAgent() (err error) {
	log.Info("initialize agent")
	if err = agentService.InitAgent(); err != nil {
		return errors.Wrap(err, "init agent failed")
	}
	log.Infof("meta from sqlite is %s", meta.OCS_AGENT)

	if a.obHasStarted {
		if !a.upgradeMode {
			if err = ob.LoadOBConfigFromConfigFile(); err != nil {
				log.WithError(err).Error("load ob config from config file failed")
				process.ExitWithFailure(constant.EXIT_CODE_ERROR_IP_NOT_MATCH, fmt.Sprintf("load ob config from config file failed: %v\n", err))
			}
		}
	} else if meta.OCS_AGENT.IsUnidentified() {
		// Error must be nil.
		agentService.BeSingleAgent()
	}

	a.checkAgentInfo()

	// Update agent info if necessary.
	if err = a.updateAgent(); err != nil {
		return err
	}

	log.Info("initialize agent status")
	if err = agentService.InitializeAgentStatus(); err != nil {
		return err
	}

	log.Info("update base info")
	return agentService.UpdateBaseInfo()
}

func (a *Agent) updateAgent() (err error) {
	if meta.OCS_AGENT.Equal(&a.AgentInfo) {
		return nil
	}

	switch meta.OCS_AGENT.GetIdentity() {
	case meta.UNIDENTIFIED:
		if meta.OCS_AGENT.GetIp() != a.AgentInfo.Ip {
			process.ExitWithFailure(constant.EXIT_CODE_ERROR_IP_NOT_MATCH, fmt.Sprintf("agent ip not match, input is %s, meta is %s", a.AgentInfo.Ip, meta.OCS_AGENT.GetIp()))
		}
		fallthrough
	case meta.SINGLE:
		err = agentService.UpdateAgentInfo(&a.AgentInfo)
	default:
		err = fmt.Errorf("agent info not equal, input is %v, meta is %v", a.AgentInfo, meta.OCS_AGENT)
	}

	return
}

func (a *Agent) checkAgentInfo() {
	log.Info("check agent info")

	// Fill agent ip.
	if a.AgentInfo.Ip == "" && meta.OCS_AGENT.GetIp() != "" {
		a.AgentInfo.Ip = meta.OCS_AGENT.GetIp()
	}

	// Fill agent port.
	if a.AgentInfo.GetPort() == 0 {
		// While port is empty and agent is single, set port to default value.
		// If agent is not single, it must have port. Otherwise, there will be an error
		if meta.OCS_AGENT.GetPort() == 0 && (meta.OCS_AGENT.IsSingleAgent() || meta.OCS_AGENT.IsUnidentified()) {
			a.AgentInfo.Port = constant.DEFAULT_AGENT_PORT
		}
		if meta.OCS_AGENT.GetPort() != 0 {
			a.AgentInfo.Port = meta.OCS_AGENT.GetPort()
		}
	}

	// If agent ip or port is empty, exit.
	if a.AgentInfo.Ip == "" || a.AgentInfo.Port == 0 {
		log.Error("agent info is invalid")
		process.ExitWithFailure(constant.EXIT_CODE_ERROR_INVAILD_AGENT, fmt.Sprintf("agent info is invalid: %v", a.AgentInfo))
	}

	if a.NeedBeCluster && !meta.OCS_AGENT.IsClusterAgent() {
		process.ExitWithFailure(constant.EXIT_CODE_ERROR_NOT_CLUSTER_AGENT, "obshell need to be cluster. Please do takeover first.")
	}
}

// initServer will only initialize the Server and will not start the service.
func (a *Agent) initServer() {
	log.Info("init server")
	serverConfig := config.ServerConfig{
		Ip:      "0.0.0.0",
		Port:    meta.OCS_AGENT.GetPort(),
		Address: fmt.Sprintf("0.0.0.0:%d", meta.OCS_AGENT.GetPort()),
		RunDir:  path.RunDir(),
	}
	log.Infof("server config is %v", serverConfig)
	a.server = web.NewServer(config.DebugMode, serverConfig)
	a.startChan = make(chan bool, 1)
}

// initTask will register tasks
func (a *Agent) initTask() {
	ob.RegisterObStartTask()
	ob.RegisterObStopTask()
	ob.RegisterObInitTask()
	ob.RegisterObScaleOutTask()
	ob.RegisterObScaleInTask()
	ob.RegisterUpgradeTask()
	ob.RegisterBackupTask()
	ob.RegisterRestoreTask()
	agent.RegisterAgentTask()
	tenant.RegisterTenantTask()
	recyclebin.RegisterRecyclebinTask()
	task.RegisterTaskType(script.ImportScriptForTenantTask{})
	pool.RegisterPoolTask()
}

// Check if the ob config file exists.
func (a *Agent) isOBhasStarted() bool {
	// If an error occurs, it's assumed that OB is not started.
	a.obHasStarted, _ = ob.HasStarted()
	return a.obHasStarted
}

func (a *Agent) isUpgradeMode() bool {
	log.Info("Check if obshell is in upgrade mode.")
	if a.OldServerPid != 0 {
		// If the old agent is running in the same directory as the new agent,
		// it is considered an upgrade.
		cwdDir, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", a.OldServerPid))
		if err != nil {
			return false
		}
		log.Infof("the cwd of %d is %s", a.OldServerPid, cwdDir)
		if global.HomePath == cwdDir {
			log.Info("The obshell is in upgrade mode.")
			a.upgradeMode = true
			// Unset root password env to avoid cover sqlite when upgrade (agent restart)
			syscall.Unsetenv(constant.OB_ROOT_PASSWORD)
		}
	}
	return a.upgradeMode
}

// preCheckForUpgrade will initialize the unix socket service,
// and check if the old server has been killed.
func (a *Agent) preCheckForUpgrade() (err error) {
	if a.upgradeMode {
		if err = a.initServerForUpgrade(); err != nil {
			return err
		}
		if err = WaitServerProcKilled(a.OldServerPid); err != nil {
			return err
		}
	}
	return nil
}
