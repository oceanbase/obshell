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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/engine/executor"
	"github.com/oceanbase/obshell/seekdb/agent/engine/scheduler"
)

func StartTaskEngine() {
	startLocalTaskEngine()
	// SeekDB mode: all metadata lives in local SQLite, so only the local
	// scheduler is needed.  Starting the cluster scheduler as well would
	// cause two schedulers to compete on the same SQLite DAG table, leading
	// to maintenance-state corruption (ocs_info.status never restored).
}

func startLocalTaskEngine() {
	log.Info("local task engine starting ...")
	if executor.OCS_EXECUTOR_POOL == nil {
		executor.OCS_EXECUTOR_POOL = executor.NewExecutorPool()
		go executor.OCS_EXECUTOR_POOL.Start()
	}
	if scheduler.OCS_LOCAL_SCHEDULER == nil {
		scheduler.OCS_LOCAL_SCHEDULER = scheduler.NewScheduler(true)
		go scheduler.OCS_LOCAL_SCHEDULER.Start()
	}
	log.Info("local task engine started")
}

