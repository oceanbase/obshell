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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
)

func (a *Agent) run() (err error) {
	engine.StartTaskEngine()

	if err = a.runServer(); err != nil {
		return errors.Wrap(err, "run local server failed")
	}

	if err = a.restoreSecure(); err != nil {
		return errors.Wrap(err, "restore secure failed")
	}

	if err := a.startConnenctModule(); err != nil {
		return errors.Wrap(err, "statr oceanbase connect module failed")
	}

	a.handleOBMeta()
	return nil
}

func (a *Agent) runServer() (err error) {
	if err = a.runLocalServer(); err != nil {
		return errors.Wrap(err, "run local server failed")
	}

	go func() {
		select {
		case <-a.startChan:
			if err = a.runTcpServer(); err != nil {
				log.Errorf("run tcp server failed, err: %v", err)
				process.ExitWithFailure(constant.EXIT_CODE_ERROR_SERVER_LISTEN, fmt.Sprintf("run tcp server failed, err: %v", err))
			}
		}
	}()
	return nil
}

func (a *Agent) restoreSecure() (err error) {
	// Restore private key from sqlite.
	log.Info("restore secure info")
	err = secure.RestoreKey()
	if err != nil {
		log.WithError(err).Info("restore secure info failed")
		log.Info("reinit secure")
		if err = secure.New(); err != nil {
			log.WithError(err).Error("reinit secure failed")
			return err
		}
	}

	log.Info("restore secure info successed, check password of root@sys in sqlite")
	err = secure.LoadOceanbasePassword(a.GetRootPassword())
	if err != nil {
		log.WithError(err).Info("check password of root@sys in sqlite failed")
		if !meta.OCS_AGENT.IsClusterAgent() {
			process.ExitWithFailure(constant.EXIT_CODE_ERROR_NOT_CLUSTER_AGENT, "check password of root@sys in sqlite failed: not cluster agent")
		}
	} else {
		log.Info("check password of root@sys in sqlite successed")
	}

	log.Info("check agent password from sqlite")
	err = secure.LoadAgentPassword()
	if err != nil {
		log.WithError(err).Error("check agent password from sqlite failed")
	}
	return nil
}

// runLocalServer runs local server which is used to receive request by unix socket.
// If agent is upgrade mode , it will close the tmp unix socket server.
func (a *Agent) runLocalServer() (err error) {
	if err = a.server.ListenUnixSocket(); err != nil {
		return
	}

	if a.upgradeMode {
		log.Infof("close tmp socket server on %s", a.tmpSocketPath)
		a.tmpServer.LocalHttpServer.Close()
	}

	a.server.RunLocalServer()
	return nil
}

func (a *Agent) runTcpServer() (err error) {
	if err = a.server.ListenTcpSocket(); err != nil {
		return
	}

	if a.upgradeMode {
		if _, err = a.server.TcpListener.SyscallConn(); err != nil {
			return
		}
	}

	a.server.RunTcpServer()
	return nil
}

func (a *Agent) startConnenctModule() (err error) {
	defer func() {
		if err == nil && !meta.OCS_AGENT.IsUnidentified() {
			a.Server.startChan <- true
		}
	}()

	oceanbase.Init()
	if meta.OCS_AGENT.IsUnidentified() {
		return a.handleUnidentified()
	} else if a.NeedStartOB() {
		if err = CheckAndStartOBServer(); err != nil {
			process.ExitWithFailure(constant.EXIT_CODE_ERROR_OB_START_FAILED, fmt.Sprintf("start observer via flag failed, err: %v", err))
		}
	}
	return nil
}

func (a *Agent) handleOBMeta() {
	if a.upgradeMode {
		// if agent is in upgrade mode, it will handle OB meta in upgrade process.
		return
	}

	if !meta.OCS_AGENT.IsClusterAgent() {
		// Only cluster agent need handle OB meta.
		// other agent will handle OB meta when it become cluster agent.
		return
	}

	go ob.HandleOBMeta()
}

func (a *Agent) NeedStartOB() bool {
	if a.CommonFlag.NeedStartOB && (meta.OCS_AGENT.IsClusterAgent() || meta.OCS_AGENT.IsTakeover()) {
		return true
	}
	return false
}

func CheckAndStartOBServer() error {
	exist, err := process.CheckObserverProcess()
	if err != nil {
		return errors.Wrap(err, "check observer process failed ")
	}

	if !exist {
		log.Info("observer process has started, but not running, start it")
		// If the config file exists and the observer process is not running,
		// the observer process needs to be started.
		if err := ob.SafeStartObserver(nil); err != nil {
			process.ExitWithFailure(constant.EXIT_CODE_ERROR_OB_START_FAILED, fmt.Sprintf("start observer process failed, err: %v", err))
		}
	}

	// Permission Verification.
	observerUid, err := process.GetObserverUid()
	if err != nil {
		return errors.Wrap(err, "get observer uid failed")
	}
	if process.Uid() != 0 && process.Uid() != observerUid {
		process.ExitWithFailure(constant.EXIT_CODE_ERROR_OB_START_FAILED, "the user of obshell has no permission to start observer")
	}
	return nil
}

func (a *Agent) handleUnidentified() (err error) {
	if err = CheckAndStartOBServer(); err != nil {
		return errors.Wrap(err, "unidetified agent check and start observer failed")
	}

	log.Infof("take over flag is %d", a.IsTakeover)
	if a.IsTakeover == 0 {
		agentService.BeSingleAgent()
	} else {
		waitDbConnectInit()
		go a.handleTakeOverOrRebuild()
	}
	return nil
}

func waitDbConnectInit() {
	var err error
	for {
		if _, err = oceanbase.GetInstance(); err == nil {
			return
		}

		if !oceanbase.IsConnecting() && oceanbase.HasAttemptedConnection() {
			if _, err = oceanbase.GetInstance(); err != nil {
				if oceanbase.IsInitPasswordError() {
					process.ExitWithFailure(constant.EXIT_CODE_ERROR_OB_PWD_ERROR, oceanbase.GetLastInitError().Error())
				} else {
					log.WithError(err).Error("get ob connection failed")
					return
				}
			}
		}
		time.Sleep(constant.GET_INSTANCE_RETRY_INTERVAL * time.Second)
	}
}
