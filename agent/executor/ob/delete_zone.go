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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
)

func checkZoneCanBeDeleted(zoneName string) *errors.OcsAgentError {
	exist, err := obclusterService.HasUnitInZone(zoneName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "get ob units on zone failed: %s", err.Error())
	}
	if exist {
		return errors.Occurf(errors.ErrBadRequest, "The zone '%s' is not empty and can not be deleted", zoneName)
	}

	// check if there is other zone stopped
	if exist, err := obclusterService.HasOtherStopTask(zoneName); err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Check if has other stop task failed: %s", err.Error())
	} else if exist {
		return errors.Occur(errors.ErrBadRequest, "cannot stop server or stop zone in multiple zones")
	}
	return nil
}

func DeleteZone(zoneName string) (*task.DagDetailDTO, *errors.OcsAgentError) {
	zone, err := obclusterService.GetZone(zoneName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "check zone exist failed: %s", err.Error())
	}
	if zone == nil {
		return nil, nil
	}
	if err := checkZoneCanBeDeleted(zoneName); err != nil {
		return nil, err
	}

	observers, err := obclusterService.GetOBServersByZone(zoneName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "get observer in zone '%s'failed: %s", zoneName, err.Error())
	}
	agentInfos, err := agentService.GetAgentInfoByZoneFromOB(zoneName)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "get agents by zone from ob failed: %s", err.Error())
	}

	builder := task.NewTemplateBuilder(DAG_DELETE_ZONE).SetMaintenance(task.GlobalMaintenance())
	context := task.NewTaskContext().SetParam(PARAM_ZONE, zoneName).SetParam(PARAM_ZONE_REGION, zone.Region)
	if len(agentInfos) != 0 {
		builder.AddTask(newSetAgentToScaleInTask(), false)
		context.SetParam(PARAM_DELETE_AGENTS, agentInfos)
	}
	builder.AddNode(newStopZoneNodeForDelete(zoneName))

	if len(observers) != 0 {
		for _, server := range observers {
			svrInfo := meta.ObserverSvrInfo{Ip: server.SvrIp, Port: server.SvrPort}
			builder.AddNode(newDeleteObserverNode(svrInfo, false)).
				AddNode(newWaitDeleteServerSuccessNode(svrInfo))
		}
	}

	builder.AddTask(newDeleteZoneTask(zoneName), false)

	if len(agentInfos) != 0 {
		builder.AddTask(newDeleteAgentTask(), false).
			AddTask(newTryToInformToKillObserversTask(), false)
	}

	dag, err := clusterTaskService.CreateDagInstanceByTemplate(builder.Build(), context)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "create dag instance failed: %s", err.Error())
	}
	return task.NewDagDetailDTO(dag), nil
}

type DeleteZoneTask struct {
	task.Task
	zoneName string
	region   string
}

func newDeleteZoneTask(zoneName string) *DeleteZoneTask {
	newTask := &DeleteZoneTask{
		Task: *task.NewSubTask(fmt.Sprintf(TASK_NAME_DELETE_ZONE, zoneName)),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanPass()
	return newTask
}

func (t *DeleteZoneTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_ZONE, &t.zoneName); err != nil {
		return err
	}

	exist, err := obclusterService.IsZoneExistInOB(t.zoneName)
	if err != nil {
		return fmt.Errorf("check zone '%s' exist failed: %s", t.zoneName, err.Error())
	}
	if exist {
		return obclusterService.DeleteZone(t.zoneName)
	}
	return nil
}

func (t *DeleteZoneTask) Rollback() error {
	if err := t.GetContext().GetParamWithValue(PARAM_ZONE, &t.zoneName); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_ZONE_REGION, &t.region); err != nil {
		return err
	}
	exist, err := obclusterService.IsZoneExistInOB(t.zoneName)
	if err != nil {
		return errors.Errorf("check zone '%s' exist failed: %s", t.zoneName, err.Error())
	}
	if !exist {
		return obclusterService.AddZoneInRegion(t.zoneName, t.region)
	}
	return nil
}

type TryToInformToKillObserversTask struct {
	task.Task
	agents []meta.AgentInfo
}

func newTryToInformToKillObserversTask() *TryToInformToKillObserversTask {
	newTask := &TryToInformToKillObserversTask{
		Task: *task.NewSubTask(TASK_NAME_INFORM_TO_KILL_OBSERVERS),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanPass()
	return newTask
}

func (t *TryToInformToKillObserversTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_AGENTS, &t.agents); err != nil {
		return err
	}

	// Try to inform all agent to kill observer.
	for _, agent := range t.agents {
		var dagDetailDTO task.DagDetailDTO
		if err := secure.SendDeleteRequest(&agent, constant.URI_OBSERVER_RPC_PREFIX, nil, &dagDetailDTO); err != nil {
			// Igonre any error because this task won't influence deleting zone.
			t.ExecuteWarnLogf("inform agent %s to kill observer failed: %v", agent.String(), err)
		}
	}
	return nil
}
