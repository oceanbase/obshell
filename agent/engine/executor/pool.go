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

	mapset "github.com/deckarep/golang-set"
	log "github.com/sirupsen/logrus"
)

const (
	QUEUE_SIZE = 10
	WORKER_NUM = 8
)

var OCS_EXECUTOR_POOL *ExecutorPool

type ExecutorPool struct {
	waitingQueue chan int64
	readyQueue   chan int64
	readySet     mapset.Set
	executors    []*Executor
	cancel       context.CancelFunc
}

func NewExecutorPool() *ExecutorPool {
	pool := &ExecutorPool{
		readySet:     mapset.NewSet(),
		waitingQueue: make(chan int64, QUEUE_SIZE),
		readyQueue:   make(chan int64, QUEUE_SIZE),
	}
	for i := 0; i < WORKER_NUM; i++ {
		pool.executors = append(pool.executors, NewExecutor(pool))
	}
	return pool
}

func (pool *ExecutorPool) AddTask(taskID int64) {
	if pool.readySet.Contains(taskID) {
		log.Infof("task %d is already in ExecutorPool", taskID)
		return
	}
	log.Infof("add task %d to ExecutorPool", taskID)
	pool.readySet.Add(taskID)
	pool.waitingQueue <- taskID
}

func (pool *ExecutorPool) recoverLocalTask() {
	subTasks, err := localTaskService.GetAllUnfinishedSubTasks()
	if err != nil {
		panic(err)
	}
	for _, subTask := range subTasks {
		pool.AddTask(subTask.GetID())
	}
}

func (pool *ExecutorPool) Start() {
	if pool.cancel != nil {
		panic("ExecutorPool is running")
	}

	pool.recoverLocalTask()
	ctx, cancel := context.WithCancel(context.Background())
	pool.cancel = cancel
	for _, executor := range pool.executors {
		go executor.Start(ctx)
	}
	flag := false
	for {
		select {
		case taskID := <-pool.waitingQueue:
			executor := running_task_map[taskID]
			if executor != nil { // task is running
				// If the task is being executed, the newly acquired task cannot be discarded.
				// The task needs to be added to the executor's duplicateQueue so that it can continue to be executed later.
				log.Infof("task %d is running, add it to duplicate queue", taskID)
				executor.duplicateQueue <- taskID
			} else {
				pool.readyQueue <- taskID
			}
		case <-ctx.Done():
			log.Info("ExecutorPool stopped")
			flag = true
		}
		if flag {
			break
		}
	}

	for _, executor := range pool.executors {
		executor.Stop()
	}
}

func (pool *ExecutorPool) Stop() {
	if pool.cancel != nil {
		log.Info("ExecutorPool stopping")
		pool.cancel()
	} else {
		log.Info("ExecutorPool is not running")
	}
}
