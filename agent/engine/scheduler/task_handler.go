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

package scheduler

import (
	"fmt"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/executor"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
)

func (s *Scheduler) advanceTask(node *task.Node) (isFinished bool, isSucceed bool, err error) {
	if _, err = s.service.GetSubTasks(node); err != nil {
		return false, false, errors.Wrap(err, "get sub tasks error")
	}

	// Empty readyTasks not means node is finished, because there could still be ongoing tasks.
	var readyTasks []task.ExecutableTask
	log.withScheduler(s).Infof("advanceTask: node %d operator %d", node.GetID(), node.GetOperator())
	switch node.GetOperator() {
	case task.RUN:
		readyTasks, isFinished, isSucceed, err = s.runHandler(node)
	case task.ROLLBACK:
		readyTasks, isFinished, isSucceed, err = s.rollbackHandler(node)
	case task.RETRY:
		readyTasks, isFinished, isSucceed, err = s.retryHandler(node)
	case task.CANCEL:
		readyTasks, isFinished, isSucceed, err = s.cancelHandler(node)
	default:
		return false, false, fmt.Errorf("node %d unknown operator %d", node.GetID(), node.GetOperator())
	}
	log.withScheduler(s).Infof("ready Task num %d, isFinished %t, isSucceed %t", len(readyTasks), isFinished, isSucceed)
	if err == nil && len(readyTasks) > 0 {
		for _, subTask := range readyTasks {
			log.withScheduler(s).Infof("ready sub task %d operator %d", subTask.GetID(), subTask.GetOperator())
			if s.isLocal {
				executor.OCS_EXECUTOR_POOL.AddTask(subTask.GetID())
			} else {
				s.createAndRunClusterTask(node, subTask)
			}
		}
	}
	return
}

func (s *Scheduler) createAndRunClusterTask(node *task.Node, subTask task.ExecutableTask) {
	remoteTask := s.createRemoteTask(node, subTask)
	if err := s.runSubTask(remoteTask); err != nil {
		log.withScheduler(s).Info("run sub task error: ", err)
	}
}

func (s *Scheduler) runHandler(node *task.Node) ([]task.ExecutableTask, bool, bool, error) {
	subTasks := node.GetSubTasks()
	readyTasks := make([]task.ExecutableTask, 0)
	isFinished := true
	isSucceed := true
	for _, subTask := range subTasks {
		log.withScheduler(s).Infof("sub task %d state %d", subTask.GetID(), subTask.GetState())
		switch subTask.GetState() {
		case task.PENDING:
			isFinished = false
			if err := s.setSubTaskRunReady(node, subTask); err != nil {
				return nil, isFinished, isSucceed, errors.Wrap(err, "set sub task ready error")
			}
			readyTasks = append(readyTasks, subTask)
		case task.READY:
			isFinished = false
			readyTasks = append(readyTasks, subTask)
		case task.RUNNING:
			isTimeout, err := s.runningSubTaskHandler(subTask)
			if err != nil {
				return nil, isFinished, isSucceed, errors.Wrap(err, "running sub task handler error")
			}
			if isTimeout {
				isSucceed = false
			} else {
				isFinished = false
			}
		case task.SUCCEED:
			continue
		case task.FAILED:
			isSucceed = false
		}
	}
	return readyTasks, isFinished, isSucceed, nil
}

func (s *Scheduler) rollbackHandler(node *task.Node) ([]task.ExecutableTask, bool, bool, error) {
	subTasks := node.GetSubTasks()
	readyTasks := make([]task.ExecutableTask, 0)
	isFinished := true
	isSucceed := true

	for _, subTask := range subTasks {
		if subTask.IsRun() {
			if subTask.IsPending() {
				// Task has not been executed, no rollback is required.
				continue
			}
			if subTask.IsReady() {
				// Task has not been executed, so no rollback is necessary.
				// However, the status needs to be updated to prevent it from being started.
				s.service.FinishSubTask(subTask, task.PASS)
			}
		}
		if !subTask.IsRollback() || subTask.GetStartTime().Before(node.GetStartTime()) {
			isFinished = false
			if err := s.setSubTaskRollbackReady(node, subTask); err != nil {
				return nil, isFinished, isSucceed, errors.Wrap(err, "set sub task rollback error")
			}
			readyTasks = append(readyTasks, subTask)
			continue
		}

		// Handler rollback sub task.
		switch subTask.GetState() {
		case task.PENDING:
			return nil, isFinished, isSucceed, fmt.Errorf("sub task %d is pending", subTask.GetID())
		case task.READY:
			isFinished = false
			readyTasks = append(readyTasks, subTask)
		case task.RUNNING:
			isTimeout, err := s.runningSubTaskHandler(subTask)
			if err != nil {
				return nil, isFinished, isSucceed, errors.Wrap(err, "running sub task handler error")
			}
			if isTimeout {
				isSucceed = false
			} else {
				isFinished = false
			}
		case task.SUCCEED:
			continue
		case task.FAILED:
			isSucceed = false
		}
	}
	return readyTasks, isFinished, isSucceed, nil
}

func (s *Scheduler) retryHandler(node *task.Node) ([]task.ExecutableTask, bool, bool, error) {
	subTasks := node.GetSubTasks()
	readyTasks := make([]task.ExecutableTask, 0)
	isFinished := true
	isSucceed := true
	for _, subTask := range subTasks {
		switch subTask.GetState() {
		case task.PENDING:
			isFinished = false
			if err := s.setSubTaskRunReady(node, subTask); err != nil {
				return nil, isFinished, isSucceed, errors.Wrap(err, "set sub task ready error")
			}
			readyTasks = append(readyTasks, subTask)
		case task.READY:
			isFinished = false
			readyTasks = append(readyTasks, subTask)
		case task.RUNNING:
			isTimeout, err := s.runningSubTaskHandler(subTask)
			if err != nil {
				return nil, isFinished, isSucceed, errors.Wrap(err, "running sub task handler error")
			}
			if isTimeout {
				isSucceed = false
			} else {
				isFinished = false
			}
		case task.SUCCEED:
			if subTask.GetStartTime().After(node.GetStartTime()) {
				if subTask.IsRollback() {
					isFinished = false
					if err := s.setSubTaskRunReady(node, subTask); err != nil {
						return nil, isFinished, isSucceed, errors.Wrap(err, "set sub task ready error")
					}
					readyTasks = append(readyTasks, subTask)
				}
			}
		case task.FAILED:
			if subTask.GetStartTime().After(node.GetStartTime()) {
				isSucceed = false
			} else {
				if err := s.setSubTaskRollbackReady(node, subTask); err != nil {
					return nil, isFinished, isSucceed, errors.Wrap(err, "set sub task rollback error")
				}
				isFinished = false
				readyTasks = append(readyTasks, subTask)
			}
		}
	}
	return readyTasks, isFinished, isSucceed, nil
}

func (s *Scheduler) cancelHandler(node *task.Node) ([]task.ExecutableTask, bool, bool, error) {
	subTasks := node.GetSubTasks()
	readyTasks := make([]task.ExecutableTask, 0)
	isFinished := true
	for _, subTask := range subTasks {
		switch subTask.GetState() {
		case task.PENDING:
		case task.READY:
			if err := s.cancelSubTask(node, subTask); err != nil {
				return nil, isFinished, false, errors.Wrap(err, "set sub task failed error")
			}
		case task.RUNNING:
			if _, err := s.runningSubTaskHandler(subTask); err != nil {
				return nil, isFinished, false, errors.Wrap(err, "running sub task handler error")
			}
			if err := s.cancelSubTask(node, subTask); err != nil {
				return nil, isFinished, false, errors.Wrap(err, "set sub task failed error")
			}
		}
	}
	return readyTasks, isFinished, false, nil
}

func (s *Scheduler) updateExecuterAgent(node *task.Node, subTask task.ExecutableTask) error {
	if node.GetNodeType() == task.NORMAL {
		ctx := node.GetContext()
		agents := ctx.GetParam(task.EXECUTE_AGENTS)
		if agents == nil { // Not specified execute agent.
			log.withScheduler(s).Infof("subtask %d update executer agent %s:%d to %s:%d\n", subTask.GetID(), subTask.GetExecuteAgent().Ip, subTask.GetExecuteAgent().Port, meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort())
			subTask.SetExecuteAgent(*meta.NewAgentInfoByInterface(meta.OCS_AGENT))
		}
	}
	return nil
}

func (s *Scheduler) createRemoteTask(node *task.Node, subTask task.ExecutableTask) *task.RemoteTask {
	return task.NewRemoteTask(node.GetTaskType().Name(), subTask.GetID(), subTask.GetName(), subTask.GetContext(),
		subTask.GetState(), subTask.GetOperator(), subTask.CanCancel(), subTask.CanContinue(), subTask.CanPass(),
		subTask.CanRetry(), subTask.CanRollback(), subTask.GetExecuteTimes(), subTask.GetExecuteAgent(),
		subTask.GetStartTime(), subTask.GetEndTime())
}

func (s *Scheduler) runSubTask(subTask *task.RemoteTask) error {
	agentInfo := subTask.ExecuterAgent
	if agentInfo.Equal(meta.OCS_AGENT) {
		taskMapInstance, err := localTaskService.GetTaskMappingByRemoteTaskId(subTask.TaskID)
		if err != nil {
			return errors.Wrapf(err, "get task mapping by remote task id %d error", subTask.TaskID)
		}
		if taskMapInstance != nil {
			if subTask.ExecuteTimes != taskMapInstance.ExecuteTimes {
				if subTask.ExecuteTimes < taskMapInstance.ExecuteTimes {
					return fmt.Errorf("remote task %d execute times `%d` is less than task mapping execute times `%d`", subTask.TaskID, subTask.ExecuteTimes, taskMapInstance.ExecuteTimes)
				}
				if err = localTaskService.UpdateLocalTaskInstanceByRemoteTask(subTask); err != nil {
					return errors.Wrapf(err, "update local task instance by remote task %d error", subTask.TaskID)
				}
			}
			executor.OCS_EXECUTOR_POOL.AddTask(taskMapInstance.LocalTaskId)
		} else {
			localTaskId, err := localTaskService.CreateLocalTaskInstanceByRemoteTask(subTask)
			if err != nil {
				return errors.Wrapf(err, "create local task instance by remote task %d error", subTask.TaskID)
			}
			executor.OCS_EXECUTOR_POOL.AddTask(localTaskId)
		}
	} else {
		if err := s.sendRunSubTaskRpc(subTask); err != nil {
			return errors.Wrapf(err, "send run sub task rpc to %s:%d error", agentInfo.Ip, agentInfo.Port)
		}
	}
	return nil
}

func (s *Scheduler) sendRunSubTaskRpc(subTask *task.RemoteTask) error {
	log.withScheduler(s).Infof("send run sub task %d to %s:%d", subTask.TaskID, subTask.ExecuterAgent.Ip, subTask.ExecuterAgent.Port)
	return secure.SendPostRequest(&subTask.ExecuterAgent, constant.URI_TASK_RPC_PREFIX+constant.URI_SUB_TASK, subTask, nil)
}

func (s *Scheduler) cancelSubTask(node *task.Node, subTask task.ExecutableTask) error {
	if err := s.service.SetSubTaskFailed(subTask, "sub task cancelled"); err != nil {
		return err
	}

	log.withScheduler(s).Infof("handle cancel sub task %d, is local task: %t", subTask.GetID(), subTask.IsLocalTask())
	if subTask.IsLocalTask() {
		executor.OCS_EXECUTOR_POOL.CancelTask(subTask.GetID())
	} else {
		if !subTask.IsPending() {
			agentInfo := subTask.GetExecuteAgent()
			if agentInfo.Equal(meta.OCS_AGENT) {
				taskMapInstance, err := localTaskService.GetTaskMappingByRemoteTaskId(subTask.GetID())
				if err != nil {
					return errors.Wrapf(err, "get task mapping by remote task id %d error", subTask.GetID())
				} else if taskMapInstance == nil {
					return fmt.Errorf("task mapping by remote task id %d not found", subTask.GetID())
				}
				executor.OCS_EXECUTOR_POOL.CancelTask(taskMapInstance.LocalTaskId)
			} else {
				return s.sendCancelSubTaskRpc(node, subTask)
			}
		}
	}
	return nil
}

func (s *Scheduler) sendCancelSubTaskRpc(node *task.Node, subTask task.ExecutableTask) error {
	remoteTask := s.createRemoteTask(node, subTask)
	agent := subTask.GetExecuteAgent()
	return secure.SendDeleteRequest(&agent, constant.URI_TASK_RPC_PREFIX+constant.URI_SUB_TASK, remoteTask, nil)
}

func (s *Scheduler) runningSubTaskHandler(subTask task.ExecutableTask) (bool, error) {
	isTimeout, err := s.checkSubTaskTimeout(subTask)
	if err != nil {
		return false, errors.Wrap(err, "check sub task timeout error")
	}
	if isTimeout {
		err = s.service.SetSubTaskFailed(subTask, "sub task timeout")
		if err != nil {
			return false, errors.Wrap(err, "set sub task timeout error")
		}
	}
	return isTimeout, nil
}

func (s *Scheduler) checkSubTaskTimeout(subtask task.ExecutableTask) (isTimeout bool, err error) {
	var nowTime time.Time
	if s.isLocal {
		nowTime = time.Now()
	} else {
		nowTime, err = global.TIME.ObNow()
		if err != nil {
			err = errors.Wrap(err, "get now time error")
			return
		}
	}
	timeout := subtask.GetTimeout()
	startTime := subtask.GetStartTime()
	isTimeout = nowTime.Sub(startTime) > timeout
	return
}

func (s *Scheduler) setSubTaskRunReady(node *task.Node, subTask task.ExecutableTask) error {
	if subTask.IsRun() && subTask.IsReady() {
		return nil
	}
	subTask.SetContext(node.GetContext())
	s.updateExecuterAgent(node, subTask)
	err := s.service.SetSubTaskReady(subTask, task.RUN)
	if err != nil {
		return errors.Wrap(err, "update task error")
	}
	return nil
}

func (s *Scheduler) setSubTaskRollbackReady(node *task.Node, subTask task.ExecutableTask) error {
	if subTask.IsRollback() && subTask.IsReady() {
		return nil
	}
	s.updateExecuterAgent(node, subTask)
	err := s.service.SetSubTaskReady(subTask, task.ROLLBACK)
	if err != nil {
		return errors.Wrap(err, "set sub task rollback error")
	}
	return nil
}
