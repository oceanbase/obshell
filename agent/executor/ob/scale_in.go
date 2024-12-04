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
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/coordinator"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
	log "github.com/sirupsen/logrus"
)

const WAIT_DELETE_SERVER_FINISH_INTERVAL = 3 * time.Second
const WAIT_DELETE_SERVER_FINISH_TIEMS = 1200

func isAllLsMultiPaxosAlive(svrInfo meta.ObserverSvrInfo, infoFunc func(string, ...interface{})) (alive bool, err error) {
	// Get all log infos.
	logInfos, err := obclusterService.GetLogInfosInServer(svrInfo)
	if err != nil {
		return false, errors.New("get all log info failed")
	}
	for _, logStat := range logInfos {
		infoFunc("check multi paxos member alive of tenant '%d', log stream '%d'", logStat.TenantId, logStat.LsId)
		if alive, err := obclusterService.IsLsMultiPaxosAlive(logStat.LsId, logStat.TenantId, svrInfo); err != nil {
			return false, errors.Errorf("check multi paxos member alive failed: %s", err.Error())
		} else if !alive {
			infoFunc("the log stream %d of tenant %d has no majority alive.", logStat.LsId, logStat.TenantId)
			return false, nil
		}
	}
	return true, nil
}

func ClusterScaleIn(param param.ClusterScaleInParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	agentInfo := param.AgentInfo
	server, err := obclusterService.GetOBServerByAgentInfo(agentInfo)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "check server exist failed: %s", err.Error())
	}
	if server == nil {
		return nil, nil
	}
	svrInfo := meta.ObserverSvrInfo{Ip: server.SvrIp, Port: server.SvrPort}

	if param.ForceKill {
		if alive, err := isAllLsMultiPaxosAlive(svrInfo, log.Infof); err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err.Error())
		} else if !alive {
			return nil, errors.Occur(errors.ErrBadRequest, "check multi paxos member alive failed")
		}
	}

	deleteServer := meta.ObserverSvrInfo{Ip: svrInfo.GetIp(), Port: svrInfo.GetPort()}
	agentInfos := []meta.AgentInfo{agentInfo}

	context := task.NewTaskContext().
		SetParam(PARAM_DELETE_SERVER, deleteServer).
		SetParam(PARAM_DELETE_AGENTS, agentInfos)
	builder := task.NewTemplateBuilder(DAG_CLUSTER_SCALE_IN).
		SetMaintenance(task.GlobalMaintenance()).
		AddNode(newDeleteObserverNode(deleteServer, true)).
		AddTask(newSetAgentToScaleInTask(), false)

	if param.ForceKill {
		builder.AddNode(newCheckMultiPaxosMemberAliveNode()).
			AddNode(newTryToInformToKillObserverNode(param.ForceKill, agentInfo)).
			AddNode(newWaitDeleteServerSuccessNode(deleteServer)).
			AddTask(newDeleteAgentTask(), false)
	} else {
		builder.AddNode(newWaitDeleteServerSuccessNode(deleteServer)).
			AddTask(newDeleteAgentTask(), false).
			AddNode(newTryToInformToKillObserverNode(param.ForceKill, agentInfo))
	}

	dag, err := clusterTaskService.CreateDagInstanceByTemplate(builder.Build(), context)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "create dag instance failed: %s", err.Error())
	}

	return task.NewDagDetailDTO(dag), nil
}

type BaseDeleteObserverTask struct {
	task.Task
	server meta.ObserverSvrInfo
}

func newBaseDeleteObserverTask(taskName string) *BaseDeleteObserverTask {
	newTask := &BaseDeleteObserverTask{
		Task: *task.NewSubTask(taskName),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanRollback().
		SetCanPass()
	return newTask
}

func (t *BaseDeleteObserverTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_SERVER, &t.server); err != nil {
		return err
	}

	observer, err := obclusterService.GetOBServer(t.server)
	if err != nil {
		return errors.Errorf("find observer %s failed.", t.server.String())
	}

	if observer == nil {
		t.ExecuteLogf("observer '%s' is already not in cluster.", t.server.String())
		return nil
	}

	if observer.Status != constant.OBSERVER_STATUS_DELETING {
		if err = obclusterService.DeleteServer(t.server); err != nil {
			return errors.Errorf("delete observer %s failed: %s", t.server.String(), err.Error())
		}

	}
	return nil
}

func (t *BaseDeleteObserverTask) Rollback() error {
	if retry, err := clusterTaskService.IsRetryTask(t.Task.GetID()); err != nil {
		return err
	} else if retry {
		return nil
	}

	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_SERVER, &t.server); err != nil {
		return err
	}

	observer, err := obclusterService.GetOBServer(t.server)
	if err != nil {
		return errors.Errorf("find observer %s failed.", t.server.String())
	}

	if observer == nil {
		return errors.Errorf("observer '%s' is already not in cluster.", t.server.String())
	}

	if observer.Status == constant.OBSERVER_STATUS_DELETING {
		if err = obclusterService.CancelDeleteServer(t.server); err != nil {
			return errors.Errorf("cancel delete observer %s failed: %s", t.server.String(), err.Error())
		}
	}
	return nil
}

func (t *BaseDeleteObserverTask) Cancel() {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_SERVER, &t.server); err != nil {
		return
	}
	observer, err := obclusterService.GetOBServer(t.server)
	if err != nil {
		return
	}

	if observer == nil {
		t.ExecuteLogf("observer '%s' has already been deleted, no need cancel.", t.server.String())
		return
	}

	if observer.Status == constant.OBSERVER_STATUS_DELETING {
		t.ExecuteLogf("cancel delete observer %s start", t.server.String())
		if err := obclusterService.CancelDeleteServer(t.server); err != nil {
			t.ExecuteErrorLogf("cancel delete observer %s failed: %s", t.server.String(), err.Error())
			return
		}
	} else {
		t.ExecuteLogf("observer %s is not deleting, status: %s, no need cancel.", t.server.String(), observer.Status)
	}
}

type DeleteObserverTask struct {
	BaseDeleteObserverTask
}

func newDeleteObserverTask(server *meta.ObserverSvrInfo) *DeleteObserverTask {
	taskName := fmt.Sprintf(TASK_NAME_DELETE_OBSERVER, server.String())
	newTask := &DeleteObserverTask{
		BaseDeleteObserverTask: BaseDeleteObserverTask{
			Task: *task.NewSubTask(taskName),
		},
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanRollback().
		SetCanPass()
	return newTask
}

// newDeleteObserverNode can create a delete observer node with FAILURE_EXIT_MAINTENANCE.
func newDeleteObserverNode(server meta.ObserverSvrInfo, failureExitMaintenance bool) *task.Node {
	context := task.NewTaskContext().SetParam(PARAM_DELETE_SERVER, server)
	if failureExitMaintenance {
		context.SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	}
	return task.NewNodeWithContext(newDeleteObserverTask(&server), false, context)
}

type SetAgentToScaleInTask struct {
	BaseDeleteObserverTask
	agentInfo []meta.AgentInfo
}

func newSetAgentToScaleInTask() *SetAgentToScaleInTask {
	newTask := &SetAgentToScaleInTask{
		BaseDeleteObserverTask: *newBaseDeleteObserverTask(TASK_NAME_SET_AGENT_TO_SCALING_IN),
	}
	newTask.SetCanCancel().
		SetCanContinue().
		SetCanRetry().
		SetCanPass().
		SetCanRollback()
	return newTask
}

func (t *SetAgentToScaleInTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_AGENTS, &t.agentInfo); err != nil {
		return err
	}

	for _, agent := range t.agentInfo {
		if err := agentService.UpdateAgentIdentity(&agent, meta.SCALING_IN); err != nil {
			return errors.Errorf("update %s identity to 'SCALING IN' failed", agent.String())
		}
	}

	// Wait for agent to update identity to local.
	for {
		t.TimeoutCheck()
		time.Sleep(constant.COORDINATOR_MIN_INTERVAL)
		newMaintainer := true
		for _, agent := range t.agentInfo {
			if coordinator.OCS_COORDINATOR.Maintainer.GetIp() == agent.Ip && coordinator.OCS_COORDINATOR.Maintainer.GetPort() == agent.Port {
				newMaintainer = false
				break
			}
		}
		if newMaintainer {
			break
		}
	}

	t.ExecuteLog("change agent identity to 'SCALING IN' success")
	return nil
}

func (t *SetAgentToScaleInTask) Rollback() error {
	if retry, err := clusterTaskService.IsRetryTask(t.Task.GetID()); err != nil {
		return err
	} else if retry {
		return nil
	}

	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_AGENTS, &t.agentInfo); err != nil {
		return err
	}

	for _, agent := range t.agentInfo {
		if err := agentService.UpdateAgentIdentity(&agent, meta.CLUSTER_AGENT); err != nil {
			return errors.Errorf("update %s identity to '%s' failed", agent.String(), meta.CLUSTER_AGENT)
		}
	}

	return nil
}

type WaitDeleteServerSuccessTask struct {
	BaseDeleteObserverTask
}

func newWaitDeleteServerSuccessTask(server *meta.ObserverSvrInfo) *WaitDeleteServerSuccessTask {
	taskName := fmt.Sprintf(TASK_NAME_WAIT_DELETE_SERVER_SUCCESS, server.String())
	newTask := &WaitDeleteServerSuccessTask{
		BaseDeleteObserverTask: BaseDeleteObserverTask{
			Task: *task.NewSubTask(taskName),
		},
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanRollback().
		SetCanPass()
	return newTask
}

func newWaitDeleteServerSuccessNode(server meta.ObserverSvrInfo) *task.Node {
	context := task.NewTaskContext().SetParam(PARAM_DELETE_SERVER, server)
	return task.NewNodeWithContext(
		newWaitDeleteServerSuccessTask(&server),
		false,
		context)
}

func (t *WaitDeleteServerSuccessTask) Execute() error {
	// Delete observer if it is not deleted or deleting.
	if err := t.BaseDeleteObserverTask.Execute(); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_SERVER, &t.server); err != nil {
		return err
	}

	// Check if observer is deleted from oceanbase.DBA_OB_SERVER.
	for i := 0; i < WAIT_DELETE_SERVER_FINISH_TIEMS; i++ { // Wait 1 hour for timeout.
		t.TimeoutCheck()
		time.Sleep(WAIT_DELETE_SERVER_FINISH_INTERVAL)
		if observer, err := obclusterService.GetOBServer(t.server); err != nil {
			return errors.Errorf("find observer %s failed: ", t.server.String())
		} else if observer == nil {
			t.ExecuteLogf("observer '%s' has been deleted successfully", t.server.String())
			return nil
		} else if observer.Status == constant.OBSERVER_STATUS_DELETING {
			continue
		} else {
			return errors.Errorf("delete observer %s failed: status is %s", t.server.String(), observer.Status)
		}
	}
	return errors.Errorf("delete observer %s timeout", t.server.String())
}

type DeleteAgentTask struct {
	task.Task
	agents []meta.AgentInfo
}

func newDeleteAgentTask() *DeleteAgentTask {
	newTask := &DeleteAgentTask{
		Task: *task.NewSubTask(TASK_NAME_DELETE_AGENTS),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanPass()
	return newTask
}

func (t *DeleteAgentTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_AGENTS, &t.agents); err != nil {
		return err
	}

	for _, agent := range t.agents {
		if err := agentService.DeleteAgentInOB(&agent); err != nil {
			return errors.Wrapf(err, "delete agent %s failed", agent.String())
		}
	}
	t.ExecuteLog("delete agents success")
	return nil
}

type CheckMultiPaxosMemberAliveTask struct {
	BaseDeleteObserverTask
}

func newCheckMultiPaxosMemberAliveTask() *CheckMultiPaxosMemberAliveTask {
	newTask := &CheckMultiPaxosMemberAliveTask{
		BaseDeleteObserverTask: *newBaseDeleteObserverTask(TASK_NAME_CHECK_MULTI_PAXOS_MEMBER_ALIVE),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanRollback().
		SetCanPass()
	return newTask
}

func newCheckMultiPaxosMemberAliveNode() *task.Node {
	return task.NewNodeWithContext(newCheckMultiPaxosMemberAliveTask(), false, nil)
}

func (t *CheckMultiPaxosMemberAliveTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_SERVER, &t.server); err != nil {
		return err
	}
	if alive, err := isAllLsMultiPaxosAlive(t.server, t.ExecuteInfoLogf); err != nil {
		return err
	} else if !alive {
		return errors.Errorf("check multi paxos member alive failed")
	}
	return nil
}

type TryToInformToKillObserverTask struct {
	BaseDeleteObserverTask
	agent            meta.AgentInfo
	forceKill        bool
	originalStateInt int64
}

func newTryToInformToKillObserverNode(forceKill bool, agentInfo meta.AgentInfo) *task.Node {
	context := task.NewTaskContext().
		SetParam(PARAM_FORCE_KILL, forceKill).
		SetParam(PARAM_DELETE_AGENT, agentInfo)
	return task.NewNodeWithContext(newTryToInformToKillObserverTask(), false, context)
}

func newTryToInformToKillObserverTask() *TryToInformToKillObserverTask {
	newTask := &TryToInformToKillObserverTask{
		BaseDeleteObserverTask: *newBaseDeleteObserverTask(TASK_NAME_INFORM_TO_KILL_OBSERVER),
	}
	newTask.SetCanContinue().
		SetCanRetry().
		SetCanCancel().
		SetCanPass().
		SetCanRollback()
	return newTask
}

func (t *TryToInformToKillObserverTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_AGENT, &t.agent); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_FORCE_KILL, &t.forceKill); err != nil {
		return err
	}

	var state http.AgentStatus
	if err := secure.SendGetRequest(&t.agent, constant.URI_API_V1+constant.URI_STATUS, nil, &state); err != nil {
		// the agent is not alive, no need to kill observer by rpc.
		return nil
	}

	t.ExecuteLogf("agent %s ob state: %d", t.agent.String(), state.OBState)
	if state.OBState <= oceanbase.STATE_PROCESS_NOT_RUNNING {
		return nil
	}

	t.GetContext().SetData(PARAM_OBSERVER_STATE, state.OBState)

	// Try to inform all agent to kill observer.
	var dagDetailDTO task.DagDetailDTO
	if err := secure.SendDeleteRequest(&t.agent, constant.URI_OBSERVER_RPC_PREFIX, nil, &dagDetailDTO); err != nil {
		if t.forceKill {
			return errors.Errorf("inform agent %s to kill observer failed: %v", t.agent.String(), err)
		} else {
			// Igonre any error because this task won't influence deleting server.
			t.ExecuteWarnLogf("inform agent %s to kill observer failed: %v", t.agent.String(), err)
		}
	}
	return nil
}

func (t *TryToInformToKillObserverTask) Rollback() error {
	if retry, err := clusterTaskService.IsRetryTask(t.Task.GetID()); err != nil {
		return err
	} else if retry {
		return nil
	}

	if err := t.GetContext().GetParamWithValue(PARAM_DELETE_AGENT, &t.agent); err != nil {
		return err
	}

	if t.GetContext().GetData(PARAM_OBSERVER_STATE) == nil {
		return nil
	}
	if err := t.GetContext().GetDataWithValue(PARAM_OBSERVER_STATE, &t.originalStateInt); err != nil {
		return err
	}

	if t.originalStateInt < oceanbase.STATE_CONNECTION_AVAILABLE {
		// no need to launch target observer.
		return nil
	}
	t.ExecuteLogf("target server original state: %d", t.originalStateInt)

	// launch target observer.
	var dag *task.DagDetailDTO
	if err := secure.SendPostRequest(&t.agent, constant.URI_OBSERVER_RPC_PREFIX, nil, &dag); err != nil {
		return errors.Errorf("rollback kill observer of agent '%s' failed: %v", t.agent.String(), err)
	}
	if dag != nil && dag.GenericDTO != nil {
		// Watch the dag until it is finished.
		for {
			t.TimeoutCheck()
			if err := secure.SendGetRequest(&t.agent, fmt.Sprintf("%s%s/%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, dag.GenericID), nil, dag); err != nil {
				return errors.Errorf("watch launch observer dag %s failed: %v", dag.GenericID, err)
			}
			if dag.IsSucceed() {
				return nil
			}
			if dag.IsFailed() {
				return errors.Errorf("launch observer dag %s failed", dag.GenericID)
			}
			time.Sleep(WAIT_START_OBSERVER_INTERVAL)
		}
	}

	return nil
}
