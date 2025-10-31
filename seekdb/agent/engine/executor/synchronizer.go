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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
)

var OCS_SYNCHRONIZER *Synchronizer

type Synchronizer struct {
	taskLogSynchronizer *taskLogSynchronizer
	taskSynchronizer    *taskSynchronizer
	cancel              context.CancelFunc
	clear               context.CancelFunc
	nextChan            chan bool
}

func NewSynchronizer() *Synchronizer {
	return &Synchronizer{
		taskLogSynchronizer: newTaskLogSynchronizer(),
		taskSynchronizer:    newTaskSynchronizer(),
		nextChan:            make(chan bool, 1),
	}
}

func (synchronizer *Synchronizer) Start() {
	if synchronizer.cancel != nil || synchronizer.clear != nil {
		panic("synchronizer is running")
	}

	log.Info("synchronizer started")
	ctx, cancel := context.WithCancel(context.Background())
	synchronizer.cancel = cancel
	synchronizer.nextChan <- true
	for {
		select {
		case <-ctx.Done():
			log.Info("synchronizer stopped")
			synchronizer.release()
			return
		case <-synchronizer.nextChan:
			synchronizer.Sync()
		}
		time.Sleep(constant.SYNC_INTERVAL)
	}
}

func (synchronizer *Synchronizer) Sync() {
	synchronizer.taskLogSynchronizer.sync()
	synchronizer.taskSynchronizer.sync()
	synchronizer.nextChan <- true
}

func (synchronizer *Synchronizer) Stop() *context.Context {
	if synchronizer.cancel != nil {
		log.Info("synchronizer stopping")
		ctx, cancel := context.WithCancel(context.Background())
		synchronizer.clear = cancel

		synchronizer.cancel()
		synchronizer.cancel = nil
		return &ctx
	} else {
		log.Info("synchronizer is not running")
	}
	return nil
}

func (synchronizer *Synchronizer) release() {
	if synchronizer.clear != nil {
		synchronizer.taskLogSynchronizer = newTaskLogSynchronizer()
		synchronizer.taskSynchronizer = newTaskSynchronizer()
		synchronizer.clear()
	} else {
		log.Info("synchronizer is not stopping")
	}
}

func (synchronizer *Synchronizer) Restart() {
	ctx := synchronizer.Stop()
	if ctx != nil {
		<-(*ctx).Done()
		synchronizer.Start()
	}
}
