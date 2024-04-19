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
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/client/global"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/agent/cmd"
	"github.com/oceanbase/obshell/agent/cmd/admin"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/meta"
)

var (
	startOnce = sync.Once{}
	statusCh  = make(chan int32)
	errorCh   = make(chan error)
)

func CheckAndStartDaemon(needBeCluster ...bool) error {
	statusCh, errCh := AsyncCheckAndStartDaemon(needBeCluster...)
	waitMin := time.After(1 * time.Minute)
	for {
		select {
		case status := <-statusCh:
			if status == constant.STATE_RUNNING {
				return nil
			}
		case err := <-errCh:
			return err
		case <-waitMin:
			status, err := api.GetMyAgentStatus()
			if err != nil {
				return errors.Wrap(err, "get my agent status failed")
			}
			if status.Agent.IsUnidentified() {
				return errors.New("Cluster not taken over. Run 'obshell cluster start -a' to start it.")
			}
		}
	}
}

func AsyncCheckAndStartDaemon(needBeCluster ...bool) (<-chan int32, <-chan error) {
	go startOnce.Do(func() {
		var preStatus int32
		go checkAndStartDaemon(needBeCluster...)
		for {
			status, err := api.GetMyAgentStatus()
			if err == nil {
				if status.State != preStatus {
					log.Info("agent status changed: ", status.State)
					preStatus = status.State
					statusCh <- preStatus
					if preStatus == constant.STATE_RUNNING {
						break
					}
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
	return statusCh, errorCh
}

func checkAndStartDaemon(needBeCluster ...bool) {
	if http.SocketIsActive(path.DaemonSocketPath()) {
		return
	}
	global.SetDaemonIsBrandNew(true)

	log.Info("daemon's unix socket is not active")
	if err := startDaemon(needBeCluster...); err != nil {
		errorCh <- errors.Wrap(err, "start daemon failed")
		return
	}

	ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
	if err := handleIfInTakeoverProcess(); err != nil {
		errorCh <- err
		return
	}
}

func startDaemon(needBeCluster ...bool) error {
	stdio.Print("Detected that obshell is not running. Starting obshell for you now.")
	stdio.StartLoading("Starting the obshell.")
	ip, err := global.MyAgentIp()
	if err != nil {
		stdio.LoadFailed("start obshell failed!")
		return errors.Wrap(err, "get my agent ip failed")
	}

	flag := &cmd.CommonFlag{
		AgentInfo: meta.AgentInfo{
			Ip: ip,
		},
		IsTakeover: 1,
	}
	if len(needBeCluster) > 0 && needBeCluster[0] {
		flag.IsTakeover = 0
		flag.NeedBeCluster = needBeCluster[0]
	}

	admin := admin.NewAdmin(flag)
	if err = admin.StartDaemon(); err != nil {
		stdio.LoadFailed("start obshell failed!")
		return errors.Wrap(err, "start daemon failed")
	}
	stdio.LoadSuccess("obshell started successfully!")
	return nil
}
