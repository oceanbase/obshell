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
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

type ClusterBoostrapTask struct {
	task.Task
	zoneOrder []string
	zoneRS    map[string]meta.AgentInfo
	unRS      map[string][]meta.AgentInfo
}

func newClusterBoostrapTask() *ClusterBoostrapTask {
	newTask := &ClusterBoostrapTask{
		Task: *task.NewSubTask(TASK_NAME_BOOTSTRAP),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *ClusterBoostrapTask) Execute() error {
	if err := t.getParams(); err != nil {
		return err
	}
	if err := loadOceanbaseInstanceWithoutDBName(t); err != nil {
		return err
	}
	if !t.IsContinue() {
		cmd, err := t.generateBootstrapCmd()
		if err != nil {
			return err
		}
		t.ExecuteLogf("bootstrap cmd: %s", cmd)
		if err = t.execBootstrap(cmd); err != nil {
			return errors.Wrap(err, "bootstrap failed")
		}
	}
	return t.addServers()
}

func (t *ClusterBoostrapTask) Rollback() error {
	t.ExecuteLog("clear db Instance")
	oceanbase.ClearInstance()
	return nil
}

func (t *ClusterBoostrapTask) getParams() (err error) {
	ctx := t.GetContext()
	if err = ctx.GetDataWithValue(PARAM_ZONE_ORDER, &t.zoneOrder); err != nil {
		return
	}
	if err = ctx.GetDataWithValue(PARAM_ZONE_RS, &t.zoneRS); err != nil {
		return
	}
	if err = ctx.GetDataWithValue(PARAM_UNRS, &t.unRS); err != nil {
		return
	}
	return
}

func (t *ClusterBoostrapTask) generateBootstrapCmd() (string, error) {
	bootstrapCmd := "ALTER SYSTEM BOOTSTRAP "
	list := make([]string, 0, len(t.zoneOrder))
	for _, zone := range t.zoneOrder {
		observerInfo, ok := t.zoneRS[zone]
		if !ok {
			return "", errors.Occurf(errors.ErrCommonUnexpected, "zone %s has no rs", zone)
		}
		agent := meta.NewAgentInfo(observerInfo.Ip, observerInfo.Port)
		list = append(list, fmt.Sprintf("ZONE '%s' SERVER '%s'", zone, agent.String()))
	}
	bootstrapCmd = bootstrapCmd + strings.Join(list, ", ")
	return bootstrapCmd, nil
}

func loadOceanbaseInstanceWithoutDBName(t task.ExecutableTask) error {
	t.ExecuteLog("try to connect to observer")
	if oceanbase.HasOceanbaseInstance() {
		if _, err := oceanbase.GetRestrictedInstance(); err == nil {
			return nil
		}
	}
	return oceanbase.LoadOceanbaseInstance(config.NewObMysqlDataSourceConfig().SetDBName(""))
}

func (t *ClusterBoostrapTask) execBootstrap(cmd string) error {
	return obclusterService.Bootstrap(cmd)
}

func (t *ClusterBoostrapTask) addServers() error {
	for zone, serverList := range t.unRS {
		for _, server := range serverList {
			agent := meta.NewAgentInfo(server.Ip, server.Port)
			sql := fmt.Sprintf("ALTER SYSTEM ADD SERVER '%s' ZONE '%s'", agent.String(), zone)
			t.ExecuteInfoLogf("add server: %s", sql)
			if err := obclusterService.ExecuteSqlWithoutIdentityCheck(sql); err != nil {
				return err
			}
		}
	}
	return nil
}
