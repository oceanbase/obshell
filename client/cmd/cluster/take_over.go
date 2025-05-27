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

package cluster

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/cmd/admin"
	"github.com/oceanbase/obshell/agent/cmd/daemon"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/client/global"
	"github.com/oceanbase/obshell/client/lib/http"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
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
	if agentStatus.Agent.IsUnidentified() || agentStatus.Agent.IsTakeOverFollowerAgent() || agentStatus.Agent.IsTakeOverMasterAgent() {
		return true, nil
	}
	return false, nil
}

func getServersForEmecStart(flags *ClusterStartFlags) (servers []string, err error) {
	if getScopeType(&flags.scopeFlags) == ob.SCOPE_ZONE {
		return nil, errors.New("'-z' is not supported for emergency start, please use '-s' or '-a'")
	}

	serversWithRpcPort, err := ob.GetAllServerFromOBConf()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all servers from ob conf")
	}

	return getServersByInputAndConf(flags, serversWithRpcPort)
}

// getServersByInputAndConf takes a ClusterStartFlags structure and a list of server addresses paired with their RPC ports,
func getServersByInputAndConf(flags *ClusterStartFlags, serversWithRpcPort []meta.AgentInfoInterface) (servers []string, err error) {
	if getScopeType(&flags.scopeFlags) == ob.SCOPE_GLOBAL {
		for _, server := range serversWithRpcPort {
			servers = append(servers, server.GetIp())
		}
		return
	}

	// If Server scope is specified, perform detailed validation.
	inputServers := strings.Split(strings.TrimSpace(flags.server), ",")
	for _, inputServer := range inputServers {
		inputServerInfo, err := meta.ConvertAddressToAgentInfo(inputServer)
		if err != nil {
			return nil, errors.Errorf("invalid server '%s'", inputServerInfo)
		}

		// Check if the server with the default port is present in the configuration.
		var found bool
		for _, server := range serversWithRpcPort {
			if server.GetIp() == inputServerInfo.GetIp() {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.Errorf("server %s is not in the ob conf", inputServerInfo.GetIp())
		}
		servers = append(servers, inputServerInfo.GetIp())
	}
	log.Info("servers to start ", servers)
	return
}

func handleTakeoverForStart(flags *ClusterStartFlags) (err error) {
	if err = restartDaemonForTakeover(); err != nil {
		return
	}

	servers, err := getServersForEmecStart(flags)
	if err != nil {
		return
	}
	// not need to check my agent in servers, because it return by self
	if len(servers) > 1 {
		if err = startRemoteAgent(servers, flags.SSHFlags); err != nil {
			return err
		}
	}

	return pollingOBStatus()
}

// startRemoteAgent starts remote agent on the specified servers.
// If multiple servers are on the same host, the function will return an error.
func startRemoteAgent(servers []string, flags SSHFlags) (err error) {
	exist := false // Whether the current agent is in the server list.
	myAgentIp, err := global.MyAgentIp()
	if err != nil {
		return err
	}

	agents := make([]string, 0)
	for _, server := range servers {
		if server == myAgentIp {
			if !exist {
				exist = true
				continue
			}
			return errors.New("multi-server on the same host")
		}
		agents = append(agents, server)
	}

	if len(agents) == 0 {
		return nil
	}

	agentCh := make(chan string)
	errCh := make(chan error)
	subIO := stdio.NewIO()
	subIO.StartLoading("start remote agent")
	for _, agents := range agents {
		go sshStartRemoteAgentForTakeOver(agents, constant.DEFAULT_AGENT_PORT, flags, agentCh, errCh)
	}

	errs := make([]error, 0)
	for count := len(agents); count > 0; count-- {
		select {
		case agent := <-agentCh:
			subIO.LoadStageSuccessf("remote agent on %s started successfully", agent)
		case err := <-errCh:
			errs = append(errs, err)
		}
	}

	subIO.StopLoading()
	if len(errs) > 0 {
		for _, err := range errs {
			subIO.Failed(err.Error())
		}
		return errors.New("failed to start remote agent")
	}

	return nil
}

func sshStartRemoteAgentForTakeOver(server string, agentPort int, sshFlags SSHFlags, agentCh chan string, errCh chan error) {
	stdio.Verbosef("start remote agent on %s", server)

	SSHClient, err := http.NewSSHClient(server, sshFlags.user, sshFlags.port)
	if err != nil {
		errCh <- errors.Wrapf(err, "failed to create ssh config for %s", server)
		return
	}
	SSHClient.SetPassword(sshFlags.password)
	SSHClient.SetPrivateKeyFile(sshFlags.keyfile, sshFlags.passphrase)

	_, err = SSHClient.Connect()
	if err != nil {
		errCh <- errors.Wrapf(err, "failed to connect to %s", server)
		return
	}
	defer SSHClient.Close()

	agentInfo := meta.NewAgentInfo(server, agentPort)
	cmd := fmt.Sprintf(`export OB_ROOT_PASSWORD='%s';%s cluster start -s '%s'`, os.Getenv(constant.OB_ROOT_PASSWORD), path.ObshellBinPath(), agentInfo.String())
	if msg, err := SSHClient.Exec(cmd); err != nil {
		errCh <- errors.Wrapf(err, "failed to start remote agent on %s, error msg: %s", server, string(msg))
		return
	}
	agentCh <- server
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
	case meta.TAKE_OVER_FOLLOWER, meta.TAKE_OVER_MASTER:
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
