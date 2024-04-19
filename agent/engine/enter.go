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

package engine

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/coordinator"
	"github.com/oceanbase/obshell/agent/engine/executor"
	"github.com/oceanbase/obshell/agent/engine/scheduler"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

func StartTaskEngine() {
	startLocalTaskEngine()
	go startClusterTaskEngine()
}

func startLocalTaskEngine() {
	log.Info("local task engine starting ...")
	if executor.OCS_EXECUTOR_POOL == nil {
		executor.OCS_EXECUTOR_POOL = executor.NewExecutorPool()
		go executor.OCS_EXECUTOR_POOL.Start()
	}
	if scheduler.OCS_LOCAL_SCHEDULER == nil {
		scheduler.OCS_LOCAL_SCHEDULER = scheduler.NewScheduler(coordinator.OCS_COORDINATOR, true)
		go scheduler.OCS_LOCAL_SCHEDULER.Start()
	}
	log.Info("local task engine started")
}

func startClusterTaskEngine() {
	for !meta.OCS_AGENT.IsClusterAgent() {
		time.Sleep(time.Second)
	}

	log.Info("cluster task engine starting ...")
	for {
		if db, _ := oceanbase.GetOcsInstance(); db != nil {
			log.Info("start cluster task engine")
			if coordinator.OCS_COORDINATOR == nil {
				coordinator.OCS_COORDINATOR = coordinator.NewCoordinator()
				go coordinator.OCS_COORDINATOR.Start()
			}
			if scheduler.OCS_SCHEDULER == nil {
				scheduler.OCS_SCHEDULER = scheduler.NewScheduler(coordinator.OCS_COORDINATOR, false)
				go scheduler.OCS_SCHEDULER.Start()
			}

			if executor.OCS_SYNCHRONIZER == nil {
				executor.OCS_SYNCHRONIZER = executor.NewSynchronizer(coordinator.OCS_COORDINATOR)
				go executor.OCS_SYNCHRONIZER.Start()
			}
			break
		}
		time.Sleep(time.Second)
	}
	log.Info("cluster task engine started")
}
