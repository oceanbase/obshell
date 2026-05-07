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

package standby

import (
	"fmt"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/param"
)

// Activate validates pre-conditions and creates the Activate DAG in one call.
func Activate() (*task.DagDetailDTO, error) {
	if err := CheckActivatePreConditions(); err != nil {
		return nil, err
	}
	return CreateActivateDag()
}

// CreateActivateDag builds and enqueues an Activate DAG for unilateral promotion.
// Flow: ActivatePreCheck → ActivateNode → CleanupMeta
func CreateActivateDag() (*task.DagDetailDTO, error) {
	builder := task.NewTemplateBuilder(constant.DAG_ACTIVATE)
	builder.
		AddTask(newActivatePreCheckTask(), false).
		AddTask(newActivateNodeTask(), false).
		AddTask(newActivateCleanupTask(), false).
		SetMaintenance(task.GlobalMaintenance())

	ctx := task.NewTaskContext().
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)

	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

// CheckActivatePreConditions validates that Activate (failover) can proceed:
//  1. Local role must be STANDBY.
//  2. If a valid UPSTREAM peer exists and is reachable (any role), reject and
//     suggest Switchover. The upstream may already be STANDBY after a prior
//     Switchover, but activating while it is reachable risks data divergence.
func CheckActivatePreConditions() error {
	localStatus, err := standbyService.GetLocalStatus()
	if err != nil {
		return fmt.Errorf("failed to query local seekdb status: %w", err)
	}
	if localStatus.Role != "STANDBY" {
		return errors.Occur(errors.ErrStandbyActivateLocalNotStandby, localStatus.Role)
	}

	upstream, err := standbyService.GetUpstreamPeer()
	if err != nil || upstream == nil {
		// No UPSTREAM peer configured — nothing to probe, allow Activate.
		return nil
	}

	// Probe upstream obshell AND seekdb via standby status RPC.
	// This call queries the remote seekdb's __all_virtual_server_stat,
	// so it fails when either the obshell or seekdb is down.
	// Reject Activate regardless of the upstream's current role: after a
	// prior Switchover the upstream may already be STANDBY, but it is still
	// alive and replicating, so activating this node would risk data divergence.
	var peerResp param.StandbyStatusResp
	if probeErr := callPeerRpcStandbyStatus(upstream.PeerHost, upstream.PeerObshellPort, &peerResp); probeErr == nil {
		return errors.Occur(errors.ErrStandbyUpstreamStillHealthy,
			upstream.PeerHost, upstream.PeerObshellPort)
	}

	// Upstream not reachable — allow Activate.
	return nil
}

// ActivatePreCheckTask validates that Activate (failover) can proceed:
//  1. Local role must be STANDBY.
//  2. If a valid UPSTREAM peer exists and is reachable, reject (suggest Switchover).
type ActivatePreCheckTask struct {
	task.Task
}

func newActivatePreCheckTask() *ActivatePreCheckTask {
	t := &ActivatePreCheckTask{
		Task: *task.NewSubTask(constant.TASK_ACTIVATE_PRECHECK),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *ActivatePreCheckTask) Execute() error {
	t.ExecuteLog("Checking Activate pre-conditions (local role and upstream liveness)")
	if err := CheckActivatePreConditions(); err != nil {
		return err
	}
	t.ExecuteLog("Activate pre-check passed")
	return nil
}

// ActivateNodeTask executes ALTER SYSTEM ACTIVATE STANDBY, promoting the local
// standby to primary without coordination with the original primary.
type ActivateNodeTask struct {
	task.Task
}

func newActivateNodeTask() *ActivateNodeTask {
	t := &ActivateNodeTask{
		Task: *task.NewSubTask(constant.TASK_ACTIVATE_NODE),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *ActivateNodeTask) Execute() error {
	t.ExecuteLog("Executing ALTER SYSTEM ACTIVATE STANDBY")
	return standbyService.ActivateStandby()
}

// ActivateCleanupTask runs post-activate metadata cleanup:
//  1. Deletes the old UPSTREAM peer record (primary is no longer reachable).
//  2. Best-effort notifies the old upstream to delete its local pair record.
type ActivateCleanupTask struct {
	task.Task
}

func newActivateCleanupTask() *ActivateCleanupTask {
	t := &ActivateCleanupTask{
		Task: *task.NewSubTask(constant.TASK_ACTIVATE_CLEANUP),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *ActivateCleanupTask) Execute() error {
	t.ExecuteLog("Deleting old UPSTREAM peer record")
	upstream, err := standbyService.GetUpstreamPeer()
	if err == nil && upstream != nil {
		// DeletePairRecord clears log_restore_source (harmless for a newly-promoted
		// PRIMARY) and returns the deleted peer so we can notify it.
		peer, delErr := standbyService.DeletePairRecord(param.PairDeleteParam{
			PeerHost:        upstream.PeerHost,
			PeerObshellPort: upstream.PeerObshellPort,
		})
		if delErr != nil {
			t.ExecuteLogf("Warning: failed to delete upstream peer record: %v", delErr)
		} else if peer != nil {
			notifyPeerDeletePair(*peer)
		}
	}

	return nil
}
