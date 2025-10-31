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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

func checkZoneCanBeDeleted(zoneName string) error {
	exist, err := obclusterService.HasUnitInZone(zoneName)
	if err != nil {
		return errors.Wrap(err, "get ob units on zone failed")
	}
	if exist {
		return errors.Occur(errors.ErrObZoneNotEmpty, zoneName)
	}

	// check if there is other zone stopped
	if exist, err := obclusterService.HasOtherStopTask(zoneName); err != nil {
		return errors.Wrap(err, "check if has other stop task failed")
	} else if exist {
		return errors.Occur(errors.ErrObServerStoppedInMultiZone)
	}
	return nil
}

func DeleteZone(zoneName string) (*task.DagDetailDTO, error) {
	zone, err := obclusterService.GetZone(zoneName)
	if err != nil {
		return nil, errors.Wrap(err, "check zone exist failed")
	}
	if zone == nil {
		return nil, nil
	}
	if err := checkZoneCanBeDeleted(zoneName); err != nil {
		return nil, err
	}

	observers, err := obclusterService.GetOBServersByZone(zoneName)
	if err != nil {
		return nil, errors.Wrapf(err, "get observer in zone '%s'failed", zoneName)
	}
	agentInfos, err := agentService.GetAgentInfoByZoneFromOB(zoneName)
	if err != nil {
		return nil, errors.Wrap(err, "get agents by zone from ob failed")
	}

	for _, agent := range agentInfos {
		if meta.OCS_AGENT.Equal(&agent) {
			return nil, errors.Occur(errors.ErrObZoneDeleteSelf, zoneName)
		}
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
		return nil, err
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
		return errors.Wrapf(err, "check zone '%s' exist failed", t.zoneName)
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
		return errors.Wrapf(err, "check zone '%s' exist failed", t.zoneName)
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
