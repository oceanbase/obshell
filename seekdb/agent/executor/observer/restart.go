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

package observer

import (
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/param"
)

func CreateRestartDag(p param.ObRestartParam) (*task.DagDetailDTO, error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}

	obState := oceanbase.GetState()

	if p.Terminate {
		if obState != oceanbase.STATE_CONNECTION_AVAILABLE {
			return nil, errors.Occur(errors.ErrAgentOceanbaseUesless) // when restart with terminate, the ob state must be available
		}
	}

	ctx := task.NewTaskContext().SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	builder := task.NewTemplateBuilder(DAG_RESTART_OBSERVER).SetMaintenance(task.GlobalMaintenance())
	if p.Terminate {
		builder.AddTask(newMinorFreezeTask(), false)
	}
	if err := AppendRestartTasks(builder, ctx); err != nil {
		return nil, err
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

// AppendRestartTasks adds a seekdb-only restart sequence to an existing DAG.
// The obshell process is deliberately not restarted: it owns and persists the
// parent DAG while seekdb is stopped and started. This is also used by the
// SeekDB standby switchover flow after the target standby becomes PRIMARY.
func AppendRestartTasks(builder *task.TemplateBuilder, ctx *task.TaskContext) error {
	if exist, err := process.CheckObserverProcess(); err != nil {
		log.Warnf("Check seekdb process failed: %v", err)
	} else if exist {
		pid, err := process.GetObserverPid()
		if err != nil {
			return err
		}
		ctx.SetParam(PARAM_OBSERVER_PID, pid)
		builder.AddTask(newStopObserverTask(), false)
	}

	builder.AddTask(newStartObServerTask(), false)
	return nil
}
