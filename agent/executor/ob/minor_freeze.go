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
	"time"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
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
			serverInfo, err := meta.ConvertAddressToAgentInfo(server)
			if err != nil {
				return nil, errors.Wrap(err, "convert address to agent info failed")
			}
			server, err := obclusterService.GetOBServerByAgentInfo(*serverInfo)
			if err != nil {
				return nil, errors.Wrap(err, "get server by agent info failed")
			}
			if server != nil {
				servers = append(servers, *server)
			}
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

	checkpointScns, err := obclusterService.GetServerCheckpointScn(servers)
	if err != nil {
		return errors.Wrap(err, "get server checkpoint_scn failed")
	}
	t.ExecuteLogf("checkpoint_scn before minor freeze: %v", checkpointScns)

	if err := obclusterService.MinorFreeze(servers); err != nil {
		return errors.Wrap(err, "minor freeze failed")
	}
	t.ExecuteLogf("minor freeze servers: %v", servers)

	checkOk := make(map[oceanbase.OBServer]bool)
	for count := 0; count < DEFAULT_MINOR_FREEZE_TIMEOUT; count++ {
		t.TimeoutCheck()
		time.Sleep(10 * time.Second)
		if ok, err := t.isMinorFreezeOver(servers, checkpointScns, checkOk); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
	return errors.New("minor freeze timeout")
}

func (t *MinorFreezeTask) isMinorFreezeOver(servers []oceanbase.OBServer, oldCheckpointScn map[oceanbase.OBServer]uint64, checkedServer map[oceanbase.OBServer]bool) (bool, error) {
	for _, server := range servers {
		if checkedServer[server] {
			continue
		}
		if checkpointScn, err := obclusterService.IsLsCheckpointAfterTs(server); err != nil {
			return false, errors.Wrap(err, "check minor freeze failed")
		} else if checkpointScn == 0 {
			// checkpoint_scn is 0, means there is no ls in this server
			continue
		} else if checkpointScn > oldCheckpointScn[server] {
			t.ExecuteLogf("[server: %s]smallest checkpoint_scn %+v bigger than expired timestamp %+v, check pass ", meta.NewAgentInfo(server.SvrIp, server.SvrPort).String(), checkpointScn, oldCheckpointScn[server])
			checkedServer[server] = true
			continue
		} else {
			t.ExecuteLogf("[server: %s]smallest checkpoint_scn: %+v smaller than expired timestamp %+v, waiting...", meta.NewAgentInfo(server.SvrIp, server.SvrPort).String(), checkpointScn, oldCheckpointScn[server])
			return false, nil
		}
	}
	return true, nil
}
