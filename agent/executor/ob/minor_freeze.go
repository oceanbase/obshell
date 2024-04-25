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

package ob

import (
	"strconv"
	"strings"
	"time"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

const DEFAULT_MINOR_FREEZE_TIMEOUT = 120

type MinorFreezeTask struct {
	task.Task
	scope param.Scope
}

func newMinorFreezeTask() *MinorFreezeTask {
	newTask := &MinorFreezeTask{
		Task: *task.NewSubTask(TASK_NAME_MINOR_FREEZE),
	}
	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel()
	return newTask
}

func (t *MinorFreezeTask) GetAllObServer() (servers []oceanbase.OBServer, err error) {
	switch t.scope.Type {
	case SCOPE_GLOBAL:
		servers, err = obclusterService.GetAllOBServers()
		if err != nil {
			return nil, errors.Wrap(err, "get all servers failed")
		}

	case SCOPE_ZONE:
		for _, zone := range t.scope.Target {
			serversInZone, err := obclusterService.GetOBServersByZone(zone)
			if err != nil {
				return nil, errors.Wrapf(err, "get servers by zone %s failed", zone)
			}
			servers = append(servers, serversInZone...)
		}

	case SCOPE_SERVER:
		for _, server := range t.scope.Target {
			info := strings.Split(server, ":")
			ip := info[0]
			port, _ := strconv.Atoi(info[1])
			server, err := obclusterService.GetOBServerByAgentInfo(ip, port)
			if err != nil {
				return nil, errors.Wrap(err, "get server by agent info failed")
			}
			servers = append(servers, server)
		}
	}

	return
}

func (t *MinorFreezeTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_SCOPE, &t.scope); err != nil {
		return errors.Wrap(err, "get scope failed")
	}

	servers, err := t.GetAllObServer()
	if err != nil {
		return errors.Wrap(err, "get all target observers failed")
	}

	curTs, err := obclusterService.GetUTCTime()
	if err != nil {
		return err
	}
	t.ExecuteLogf("timestamp before minor freeze: %v", curTs)

	if err := obclusterService.MinorFreeze(servers); err != nil {
		return errors.Wrap(err, "minor freeze failed")
	}
	t.ExecuteLogf("minor freeze servers: %v", servers)

	checkOk := make(map[oceanbase.OBServer]bool)
	for count := 0; count < DEFAULT_MINOR_FREEZE_TIMEOUT; count++ {
		time.Sleep(10 * time.Second)
		if ok, err := t.isMinorFreezeOver(servers, curTs, checkOk); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
	return errors.New("minor freeze timeout")
}

func (t *MinorFreezeTask) isMinorFreezeOver(servers []oceanbase.OBServer, curTs time.Time, checkedServer map[oceanbase.OBServer]bool) (bool, error) {
	for _, server := range servers {
		if checkedServer[server] {
			continue
		}
		if checkpointScn, err := obclusterService.IsLsCheckpointAfterTs(server); err != nil {
			return false, errors.Wrap(err, "check minor freeze failed")
		} else if checkpointScn.Equal(time.Time{}) {
			continue
		} else if checkpointScn.After(curTs) {
			t.ExecuteLogf("[server: %s:%d]smallest checkpoint_scn %+v bigger than expired timestamp %+v, check pass ", server.SvrIp, server.SvrPort, checkpointScn, curTs)
			checkedServer[server] = true
			continue
		} else {
			t.ExecuteLogf("[server: %s:%d]smallest checkpoint_scn: %+v smaller than expired timestamp %+v, waiting...", server.SvrIp, server.SvrPort, checkpointScn, curTs)
			return false, nil
		}
	}
	return true, nil
}
