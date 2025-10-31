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

package instance

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/cmd/admin"
	"github.com/oceanbase/obshell/seekdb/agent/cmd/daemon"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/client/global"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	"github.com/oceanbase/obshell/seekdb/client/utils/api"
)

func handleIfInTakeoverProcess() error {
	isInTakeoverProcess, err := isInTakeoverProcess()
	if err != nil {
		return err
	}
	if isInTakeoverProcess {
		return handleTakeover()
	}
	return nil
}

func handleTakeover() (err error) {
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return
	}

	if agentStatus.Agent.IsTakeOverMasterAgent() {
		stdio.Info("The current agent is in the process of taking over. Waiting for the process to complete.")
		for {
			time.Sleep(1 * time.Second)
			agentStatus, err = api.GetMyAgentStatus()
			if err != nil {
				stdio.StopLoading()
				stdio.Error(err.Error())
				os.Exit(1)
			}
			if !agentStatus.Agent.IsTakeOverMasterAgent() {
				return nil
			}
		}
	}
	return nil
}

func isInTakeoverProcess() (res bool, err error) {
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return false, err
	}
	if agentStatus.Agent.IsUnidentified() || agentStatus.Agent.IsTakeOverMasterAgent() {
		return true, nil
	}
	return false, nil
}

func handleTakeoverForStart() (err error) {
	if err = restartDaemonForTakeover(); err != nil {
		return
	}
	return pollingOBStatus()
}

func pollingOBStatus() (err error) {
	log.Info("start to poll and print ob status")
	done := make(chan struct{})
	statusChan := make(chan bool)
	go printOBStatus(done, statusChan)

	timeoutDuration := 10 * time.Minute
	timer := time.NewTimer(timeoutDuration)
	defer timer.Stop()

	for {
		select {
		case err = <-errorCh:
			stdio.LoadErrorf("Failed to get the current obshell status: %v", err)
			close(done)
			os.Exit(1)
		case <-statusChan:
			stdio.LoadSuccess("OB connected successfully!")
			close(done)
			return
		case <-timer.C:
			if !askForContinue() {
				stdio.LoadError("Timeout waiting for takeover, please check obshell.log for more details")
				return
			} else {
				timer.Reset(timeoutDuration)
			}
		}
	}
}

func askForContinue() bool {
	continueTakeover, _ := stdio.Confirm("Already waiting for takeover for 10 minutes, do you need to continue waiting?")
	return continueTakeover
}

func printOBStatus(done chan struct{}, statusChan chan bool) {
	log.Info("start to print ob status")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var daemonStatus *daemon.DaemonStatus
	var err error
	for {
		daemonStatus, err = api.GetMyDaemonStatus()
		stdio.Verbose("get daemon status")
		if err != nil {
			stdio.Verbosef("Failed to get daemon status: %v", err)
			if global.DaemonIsBrandNew() {
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				errorCh <- err
				return
			}
		}

		select {
		case <-done:
			return
		case <-ticker.C:
			if daemonStatus.ServerStatus.Status.OBState == oceanbase.STATE_CONNECTION_AVAILABLE &&
				daemonStatus.ServerStatus.Status.State == constant.STATE_RUNNING {
				statusChan <- true
				return
			}
		default:
			if daemonStatus != nil {
				msg := fmt.Sprintf("OB connection status: %s", oceanbase.OBStateMap[daemonStatus.ServerStatus.Status.OBState])
				stdio.StartOrUpdateLoading(msg)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

var restartFlagForTakeover bool

func restartDaemonForTakeover() (err error) {
	if global.DaemonIsBrandNew() {
		return nil
	}
	if restartFlagForTakeover {
		log.Info("obshell has been restarted for takeover")
		return nil
	}

	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}

	switch agentStatus.Agent.GetIdentity() {
	case meta.TAKE_OVER_MASTER:
		if agentStatus.OBState >= oceanbase.STATE_CONNECTION_RESTRICTED {
			return nil
		}
	case meta.UNIDENTIFIED:
	default:
		return nil
	}

	log.Info("Restarting obshell for takeover")
	admin := admin.NewAdmin(nil)
	if err = admin.RestartDaemon(); err != nil {
		return
	}
	restartFlagForTakeover = true
	return nil
}
