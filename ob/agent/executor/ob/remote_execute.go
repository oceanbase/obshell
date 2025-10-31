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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
)

// RemoteExecutableTask is only supplied for master.
type RemoteExecutableTask struct {
	task.Task
	rollbackTaskName string
	remoteDag        task.DagDetailDTO
	uri              string
	method           string
	params           interface{}
	maxRetry         int
	retryFlag        bool
	inited           bool
}

func newRemoteExecutableTask(name string) *RemoteExecutableTask {
	return &RemoteExecutableTask{
		Task: *task.NewSubTask(name),
	}
}

func (t *RemoteExecutableTask) initial(uri string, method string, params interface{}, maxRetryTimes ...int) {
	t.inited = true
	t.uri = uri
	t.method = method
	t.params = params
	t.rollbackTaskName = t.GetName()
	if len(maxRetryTimes) > 0 && maxRetryTimes[0] > 0 {
		t.maxRetry = maxRetryTimes[0]
	} else {
		t.maxRetry = DEFAULT_REMOTE_REQUEST_RETRY_TIMES
	}
}

func (t *RemoteExecutableTask) retmoteExecute() error {
	if !t.inited {
		return errors.Occur(errors.ErrCommonUnexpected, "task not inited")
	}

	operator, err := localTaskService.GetNodeOperatorBySubTaskId(t.GetID())
	if err != nil {
		return errors.Wrapf(err, "get node operator by sub task id %d failed", t.GetID())
	}

	t.retryFlag = operator == task.RETRY
	t.ExecuteLogf("retry flag: %v", t.retryFlag)
	if t.IsContinue() || t.retryFlag {
		if err := t.getAgentLastMaintainDag(); err != nil {
			return err
		}
		if t.remoteDag.DagID != 0 {
			if t.retryFlag {
				if t.remoteDag.Name == t.GetName() && t.remoteDag.State == task.FAILED_STR {
					return t.handlerCurrentDag()
				}
			} else {
				if t.remoteDag.Name != t.GetName() && t.remoteDag.State != task.SUCCEED_STR {
					return errors.Occur(errors.ErrAgentUnderMaintenanceDag, t.GetExecuteAgent().String(), t.remoteDag.GenericID, t.remoteDag.Name)
				}
				return t.handlerCurrentDag()
			}
		}
	}
	if err := t.request(); err != nil {
		return err
	}
	return t.watchRemoteDag()
}

func (t *RemoteExecutableTask) remoteRollback() error {
	if !t.inited {
		return errors.Occur(errors.ErrCommonUnexpected, "task not inited")
	}

	if err := t.getAgentLastMaintainDag(); err != nil {
		return err
	}
	if t.remoteDag.DagID == 0 { // Dag have never been executed, need not rollback.
		t.ExecuteLog("agent never execute maintain task, no need rollback")
		return nil
	}

	id, ok := t.GetLocalData(PARAM_REMOTE_ID).(string)
	if !ok || t.remoteDag.GenericID != id {
		if t.remoteDag.Name == t.rollbackTaskName {
			return t.handlerRollbackTask()
		} else if t.remoteDag.State != task.SUCCEED_STR {
			return errors.Occur(errors.ErrAgentUnderMaintenanceDag, t.GetExecuteAgent().String(), t.remoteDag.GenericID, t.remoteDag.Name)
		}
		return t.sendRpcToRollback()
	}

	if t.remoteDag.Operator == task.ROLLBACK_STR {
		if t.remoteDag.State == task.FAILED_STR {
			t.ExecuteLogf("remote task %s rollback failed, retry", id)
			if err := t.operatorRemote(task.ROLLBACK_STR); err != nil {
				return err
			}
		}
		return t.watchRemoteDag()
	}

	switch t.remoteDag.State {
	case task.SUCCEED_STR: // The remote dag has been completed and successful. Only create the corresponding rollback task, instead of directly operating the original task for rollback.
		return t.sendRpcToRollback()
	case task.RUNNING_STR: // The remote dag is running, cancel it
		t.ExecuteLogf("remote task %s is running, cancel it", id)
		if err := t.operatorRemote(task.CANCEL_STR); err != nil {
			return err
		}
	}

	if err := t.operatorRemote(task.ROLLBACK_STR); err != nil {
		return err
	}
	return t.watchRemoteDag()
}

func (t *RemoteExecutableTask) sendRpcToRollback() error {
	t.ExecuteLog("remote task is succeed, send rpc to rollback")
	if err := t.request(); err != nil {
		return err
	}
	return t.watchRemoteDag()
}

func (t *RemoteExecutableTask) handlerRollbackTask() error {
	switch t.remoteDag.State {
	case task.SUCCEED_STR:
		t.ExecuteLogf("remote task %s rollback succeed", t.remoteDag.GenericID)
	case task.FAILED_STR:
		if err := t.operatorRemote(task.RETRY_STR); err != nil {
			return err
		}
	}
	return t.watchRemoteDag()
}

func (t *RemoteExecutableTask) getRemoteDagURI() string {
	return fmt.Sprintf("%s%s/%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, t.remoteDag.GenericID)
}

func (t *RemoteExecutableTask) operatorRemote(operator string) error {
	agent := t.GetExecuteAgent()
	uri := t.getRemoteDagURI()
	params := &task.DagOperator{
		DagDetailDTO: t.remoteDag,
		Operator:     operator,
	}
	t.ExecuteLogf("send operator %s request to %s", operator, uri)
	for count := 0; count < t.maxRetry; count++ {
		if resp, err := secure.SendPostRequestAndReturnResponse(&agent, uri, params, nil); resp != nil && resp.IsError() {
			return errors.Occur(errors.ErrAgentRPCRequestError, http.POST, uri, agent.String(), resp.Error())
		} else if err != nil {
			t.ExecuteLogf("send %s request failed: %v [%d/%d]", operator, err, count, t.maxRetry)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	t.remoteDag.State = task.PENDING_STR
	return nil
}

func (t *RemoteExecutableTask) handlerCurrentDag() error {
	var operator string
	if t.retryFlag {
		operator = task.RETRY_STR
	} else {
		operator = task.OPERATOR_MAP[t.GetOperator()]
	}

	if t.remoteDag.Operator == operator && t.remoteDag.State != task.FAILED_STR {
		return t.watchRemoteDag()
	}

	if err := t.operatorRemote(operator); err != nil {
		return err
	}
	return t.watchRemoteDag()
}

func (t *RemoteExecutableTask) request() error {
	agent := t.GetExecuteAgent()
	for count := 0; count < t.maxRetry; count++ {
		if resp, err := secure.SendRequestAndReturnResponse(&agent, t.uri, t.method, t.params, &t.remoteDag); err != nil {
			t.ExecuteWarnLogf("request %s failed: %v [%d/%d]", t.uri, err, count, t.maxRetry)
			time.Sleep(1 * time.Second)
			continue
		} else if resp != nil && resp.IsError() {
			return errors.Occur(errors.ErrAgentRPCRequestError, t.method, t.uri, agent.String(), resp.Error())
		}
		return nil
	}
	return errors.Occur(errors.ErrAgentRPCRequestFailed, t.method, t.uri, agent.String(), fmt.Sprintf("retry %d times", t.maxRetry))
}

func (t *RemoteExecutableTask) watchRemoteDag() error {
	if t.remoteDag.GenericDTO == nil {
		t.ExecuteLog("do noting")
		return nil
	}

	t.SetLocalData(PARAM_REMOTE_ID, t.remoteDag.GenericID)
	agent := t.GetExecuteAgent()
	uri := t.getRemoteDagURI()
	params := &param.TaskQueryParams{ShowDetails: constant.PTR_TRUE}
	for count := 0; count < t.maxRetry; {
		if t.remoteDag.State == task.SUCCEED_STR || t.remoteDag.State == task.FAILED_STR {
			if t.remoteDag.Nodes == nil {
				// get dag detail
				params = nil
			} else {
				if t.retryFlag && t.remoteDag.Operator == task.ROLLBACK_STR && t.remoteDag.State == task.SUCCEED_STR {
					t.ExecuteInfoLog("rollback remote task succeed, retry")
				} else {
					return t.getResult()
				}
			}
		}
		time.Sleep(1 * time.Second)
		if resp, err := secure.SendGetRequestAndReturnResponse(&agent, uri, params, &t.remoteDag); resp != nil && resp.IsError() {
			return errors.Occur(errors.ErrAgentRPCRequestError, http.GET, uri, agent.String(), resp.Error())
		} else if err != nil {
			count += 1
			t.ExecuteWarnLogf("watch dag failed, count %d, err: %v", count, err)
			continue
		}
		count = 0
		t.ExecuteInfoLogf("remote task %s %s running [%d/%d]", t.remoteDag.GenericID, t.remoteDag.Name, t.remoteDag.Stage, t.remoteDag.MaxStage)
	}
	return errors.Occur(errors.ErrAgentRPCRequestFailed, http.GET, uri, agent.String(), fmt.Sprintf("retry %d times", t.maxRetry))
}

func (t *RemoteExecutableTask) getResult() error {
	for _, node := range t.remoteDag.Nodes {
		for _, task := range node.SubTasks {
			for _, log := range task.TaskLogs {
				t.ExecuteLogf("task %s: %s", task.Name, log)
			}
		}
	}
	if t.remoteDag.State == task.FAILED_STR {
		return errors.Occur(errors.ErrTaskRemoteTaskFailed, t.remoteDag.GenericID, t.remoteDag.Name)
	}
	return nil
}

func (t *RemoteExecutableTask) getAgentLastMaintainDag() error {
	agent := t.GetExecuteAgent()
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_MAINTAIN + constant.URI_AGENT_GROUP
	for count := 0; count < 30; count++ {
		if resp, err := secure.SendGetRequestAndReturnResponse(&agent, uri, nil, &t.remoteDag); resp != nil && resp.IsError() {
			return errors.Occur(errors.ErrAgentRPCRequestError, http.GET, uri, agent.String(), fmt.Sprintf("get current maintain dag failed: %v", resp.Error()))
		} else if err != nil {
			t.ExecuteWarnLogf("get current maintain dag failed, err: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	return nil
}
