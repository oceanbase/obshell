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

package executor

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
)

var running_task_map = make(map[int64]*Executor)
var task_id_list_lock sync.Mutex

type ReadyTask struct {
	retmoTaskId int64
	subTask     task.ExecutableTask
}

type Executor struct {
	currentTask    *ReadyTask
	waitingQueue   chan int64
	duplicateQueue chan int64
	logChan        chan task.TaskExecuteLogDTO
	ctx            context.Context
	cancel         context.CancelFunc
	taskCancel     context.CancelFunc
	executorPool   *ExecutorPool
}

func NewExecutor(pool *ExecutorPool) *Executor {
	return &Executor{
		executorPool:   pool,
		waitingQueue:   pool.readyQueue,
		duplicateQueue: make(chan int64, QUEUE_SIZE),
	}
}

func (executor *Executor) Start(ctx context.Context) {
	if executor.logChan != nil {
		panic("Executing")
	}

	executor.logChan = make(chan task.TaskExecuteLogDTO)
	executor.ctx, executor.cancel = context.WithCancel(ctx)
	go executor.logCommiter()
	flag := false
	for {
		executor.currentTask = nil
		select {
		case taskID := <-executor.waitingQueue:
			if err := executor.handler(taskID); err != nil {
				log.WithError(err).Warnf("task %d handler error", taskID)
			}
		case taskID := <-executor.duplicateQueue:
			if err := executor.handler(taskID); err != nil {
				log.Warnf("task %d execute error: %s", taskID, err)
			}

		case <-ctx.Done():
			log.Info("Executor stopped")
			flag = true
		}
		// Finish here to avoid executor stop when task is executing.
		executor.finishTask()
		if flag {
			break
		}
	}
}

func (executor *Executor) logCommiter() {
	defer func() {
		executor.logChan = nil
	}()
	for {
		select {
		case executeLog := <-executor.logChan:
			// If task is remote, insert log to remote.
			if !executeLog.IsSync {
				if err := subTaskLogService.InsertLocalToRemote(executeLog); err != nil {
					// Let syncService synchronize log.
					log.Warnf("insert local sub_task_log to remote error: %s", err)
				} else {
					// Set local log sync, then syncService will not synchronize log.
					executeLog.IsSync = true
				}
			}
			subTaskLogService.InsertLocal(executeLog)
		case <-executor.ctx.Done():
			log.Info("Executor stop")
			return
		}
	}
}

// handler will set task state to running and execute task.
func (executor *Executor) handler(taskID int64) error {
	executor.executorPool.RemoveTask(taskID)
	subTask, err := localTaskService.GetSubTaskByTaskID(taskID)
	if err != nil {
		return errors.Wrapf(err, "get task %d error", taskID)
	}

	log.Infof("try to start task %d", subTask.GetID())
	if err := executor.startTask(subTask); err != nil {
		time.Sleep(constant.MAINTAINER_MAX_ACTIVE_TIME)
		executor.executorPool.AddTask(taskID)
		return errors.Wrapf(err, "start task %d error", subTask.GetID())
	}
	log.Infof("start to task %d execute", subTask.GetID())
	if err := executor.executeTask(); err != nil {
		log.WithError(err).Warnf("task %d execute error", subTask.GetID())
	}
	return nil
}

func (executor *Executor) getTaskLock(taskID int64) bool {
	task_id_list_lock.Lock()
	defer task_id_list_lock.Unlock()
	if running_task_map[taskID] != nil {
		return false
	}
	running_task_map[taskID] = executor
	return true
}

func (executor *Executor) freeTaskLock(taskID int64) {
	task_id_list_lock.Lock()
	defer task_id_list_lock.Unlock()
	delete(running_task_map, taskID)
}

func (executor *Executor) startTask(subTask task.ExecutableTask) error {
	startSucceed := false
	taskID := subTask.GetID()
	if !executor.getTaskLock(taskID) {
		return fmt.Errorf("task %d is executed by other executor", taskID)
	}
	defer func() {
		if !startSucceed {
			// If start task failed, free task lock.
			executor.freeTaskLock(taskID)
		}
	}()

	readyTask := ReadyTask{
		retmoTaskId: 0,
		subTask:     subTask,
	}

	// If task is remote, get remote task id.
	if !subTask.IsLocalTask() {
		remoteTaskID, err := localTaskService.GetRemoteTaskIdByLocalTaskId(taskID)
		if err != nil {
			return err
		}
		readyTask.retmoTaskId = remoteTaskID
	}

	// If task is running, check if it can continue.
	if subTask.IsRunning() {
		executor.currentTask = &readyTask
		startSucceed = true
		subTask.SetIsContinue() // set continue flag
		return nil
	} else {
		// If task is remote, start remote task first.
		if !subTask.IsLocalTask() {
			// Try to get remote task from ob.
			remoteSubTask, err := clusterTaskService.GetSubTaskByTaskID(readyTask.retmoTaskId)
			if err != nil {
				log.WithError(err).Errorf("get remote task %d error", readyTask.retmoTaskId)
				return err
			} else {
				// If get remote task success, start remote task.
				log.Debug("get remote task success, start remote task")
				if err = clusterTaskService.StartSubTask(remoteSubTask); err != nil {
					return err
				}
			}
		}

		if err := localTaskService.StartSubTask(subTask); err != nil {
			delete(running_task_map, taskID)
			return err
		}

		executor.currentTask = &readyTask
		startSucceed = true
		return nil
	}
}

func createRemoteTask(remoteTaskId int64, subTask task.ExecutableTask) *task.RemoteTask {
	structName := reflect.TypeOf(subTask).Elem().Name()
	return task.NewRemoteTask(structName, remoteTaskId, subTask.GetName(), subTask.GetContext(),
		subTask.GetState(), subTask.GetOperator(), subTask.CanCancel(), subTask.CanContinue(), subTask.CanPass(),
		subTask.CanRetry(), subTask.CanRollback(), subTask.GetExecuteTimes(), subTask.GetExecuteAgent(),
		subTask.GetStartTime(), subTask.GetEndTime())
}

func (executor *Executor) finishTask() {
	readyTask := executor.currentTask
	if readyTask == nil {
		return
	}

	subTask := readyTask.subTask
	taskID := subTask.GetID()
	defer func() {
		executor.freeTaskLock(taskID)
		executor.currentTask = nil
	}()

	if !subTask.IsFinished() {
		executor.CancelTask()
	}

	log.Infof("finishing local task %d", taskID)
	if err := localTaskService.FinishSubTask(readyTask.subTask, readyTask.subTask.GetState()); err != nil {
		log.WithError(err).Errorf("finish local task %d failed", taskID)
	} else if readyTask.retmoTaskId != 0 {
		log.Infof("finishing remote task %d", readyTask.retmoTaskId)
		if err := finishRemoteTaskByService(readyTask.retmoTaskId, readyTask.subTask); err != nil {
			log.WithError(err).Infof("finish remote task %d. Wait Sync", readyTask.retmoTaskId)
		} else {
			// Finish remote task success, set task mapping sync.
			if err = localTaskService.SetTaskMappingSync(readyTask.retmoTaskId, subTask.GetExecuteTimes()); err != nil {
				log.Warnf("set task mapping sync error: %s", err)
			}
			log.Infof("the remote task %d has been successfully finished", readyTask.retmoTaskId)
		}
	}
	log.Infof("finish task %d end", taskID)
}

// executeTask will always set the status of the subtask to "finished", regardless of whether the subtask ends normally or not.
func (executor *Executor) executeTask() (err error) {
	log.Infof("try to execute task %d", executor.currentTask.subTask.GetID())
	subTask := executor.currentTask.subTask
	finished := make(chan bool, 1)
	defer func() {
		subTask.Finish(err)
		// When task is finished set log channel to nil, so that task will exit when set log
		subTask.SetLogChannel(nil)
	}()

	after := time.After(subTask.GetTimeout())
	ctx, cancel := context.WithCancel(context.Background())
	executor.taskCancel = cancel

	// Execute task.
	go func() {
		defer func() {
			err1 := recover()
			if err1 != nil {
				err = fmt.Errorf("task %d panic: %s", subTask.GetID(), err1)
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				log.Warnf("task %d %s Execute Panic:\n%s\n%s\n", subTask.GetID(), subTask.GetName(), err1, buf[:n])
			}
			finished <- true
		}()

		subTask.SetLogChannel(executor.logChan)

		if subTask.IsContinue() && !subTask.CanContinue() {
			// Task was unexpectedly interrupted and cannot continue.
			err = errors.New("task unexpectedly interrupted and cannot continue")
		} else if subTask.IsRollback() {
			log.Infof("execute task %d, rollback", subTask.GetID())
			err = subTask.Rollback()
		} else {
			log.Infof("execute task %d, execute", subTask.GetID())
			err = subTask.Execute()
		}
	}()

	select {
	case <-finished:
		return
	case <-after:
		err = fmt.Errorf("task %d timeout", subTask.GetID())
	case <-ctx.Done():
		log.Infof("task %d cancel", subTask.GetID())
		subTask.SetOperator(task.CANCEL)
		go executor.handleCancelTask(finished, subTask) // Execute cancel function.
	}
	return
}

func (executor *Executor) handleCancelTask(finished chan bool, subTask task.ExecutableTask) {
	<-finished // Wait for the goroutine to exit.
	log.Infof("run task %d cancel function", subTask.GetID())
	go func() {
		defer func() {
			finished <- true
		}()
		subTask.SetOperator(task.RUN)
		subTask.SetLogChannel(executor.logChan)
		subTask.Cancel()
	}()

	after := time.After(subTask.GetTimeout())
	select {
	case <-finished:
		log.Infof("execute task %d cancel function success", subTask.GetID())
	case <-after:
		log.Warnf("execute task %d cancel function timeout", subTask.GetID())
	case <-executor.ctx.Done():
		log.Infof("executor stop, interrupt task %d cancel function", subTask.GetID())
	}
	subTask.SetLogChannel(nil)
}

func (executor *Executor) CancelTask() {
	if executor.currentTask != nil {
		subTask := executor.currentTask.subTask
		subTask.SetOperator(task.CANCEL)
		executor.taskCancel()
	}
}

func (executor *Executor) Stop() {
	if executor.cancel != nil {
		log.Info("Executor stopping")
		executor.CancelTask() // Stop task
		executor.cancel()     // Stop executor
	} else {
		log.Info("Executor is not running")
	}
}
