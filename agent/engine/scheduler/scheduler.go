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
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/coordinator"
	agentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/service/task"
)

const (
	LOCAL_SCHEDULER_TRACE_ID   = "LS00000000000000"
	CLUSTER_SCHEDULER_TRACE_ID = "CS00000000000000"
)

var (
	OCS_SCHEDULER       *Scheduler
	OCS_LOCAL_SCHEDULER *Scheduler
)

type Scheduler struct {
	ctx         context.Context
	coordinator *coordinator.Coordinator
	cancel      context.CancelFunc
	isLocal     bool
	service     task.TaskServiceInterface
}

func NewScheduler(coordinator *coordinator.Coordinator, isLocal bool) *Scheduler {
	s := &Scheduler{
		coordinator: coordinator,
		isLocal:     isLocal,
	}

	ctx := context.Background()
	if isLocal {
		s.service = task.NewLocalTaskService()
		ctx = context.WithValue(ctx, agentlog.TraceIdKey{}, LOCAL_SCHEDULER_TRACE_ID)
	} else {
		s.service = task.NewClusterTaskService()
		ctx = context.WithValue(ctx, agentlog.TraceIdKey{}, CLUSTER_SCHEDULER_TRACE_ID)
	}
	s.ctx = ctx
	return s
}

func (s *Scheduler) Start() {
	// Local scheduler start directly.
	if s.isLocal {
		go s.run(context.Background())
		return
	}

	for {
		isMaintainer := <-s.coordinator.GetEventChan()
		if isMaintainer && s.coordinator.IsMaintainer() {
			ctx := context.Background()
			go s.run(ctx)
		} else if !isMaintainer && !s.coordinator.IsMaintainer() {
			s.stop()
		}
	}
}

func (s *Scheduler) run(ctx context.Context) {
	if s.cancel != nil {
		log.withScheduler(s).Warn("s is running")
		return
	}
	log.withScheduler(s).Info("scheduler starting")
	_, s.cancel = context.WithCancel(ctx)
	for s.cancel != nil {
		duration := s.handle()
		time.Sleep(duration)
	}
	log.withScheduler(s).Info("scheduler stopped")
}

func (s *Scheduler) handle() time.Duration {
	defer func() {
		err := recover()
		if err != nil {
			log.withScheduler(s).Errorf("s handle panic: %v", err)
		}
	}()
	dags, err := s.service.GetAllUnfinishedDagInstance()
	if err != nil {
		log.withScheduler(s).Errorf("get all unfinished dag instance from db error: %s", err)
		return constant.SCHEDULER_WAIT_TIME
	}
	for _, dag := range dags {
		if err := s.advanceDag(dag); err != nil {
			log.withScheduler(s).Errorf("advance dag `%d` error: %s", dag.GetID(), err)
		}
	}
	return constant.SCHEDULER_INTERVAL
}

func (s *Scheduler) stop() {
	if s.cancel != nil {
		log.withScheduler(s).Info("scheduler stopping")
		s.cancel()
		s.cancel = nil
	}
}

type schedulerLogger struct {
}

func (logger *schedulerLogger) withScheduler(s *Scheduler) *logrus.Entry {
	return logrus.WithContext(s.ctx)
}
