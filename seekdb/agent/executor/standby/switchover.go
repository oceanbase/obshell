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
	"github.com/oceanbase/obshell/seekdb/agent/executor/observer"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/param"
)

// Switchover validates pre-conditions and creates the Switchover DAG in one call.
func Switchover(p param.SwitchoverParam) (*task.DagDetailDTO, error) {
	if err := CheckSwitchoverPreConditions(p.PeerHost, p.PeerObshellPort, p.DelayThresholdSeconds); err != nil {
		return nil, err
	}
	return CreateSwitchoverDag(p)
}

// CheckSwitchoverPreConditions validates that Switchover can proceed:
//  1. Local role must be PRIMARY.
//  2. Peer role must be STANDBY and reachable.
//  3. Replication lag must be within threshold.
func CheckSwitchoverPreConditions(peerHost string, peerObshellPort int, delayThreshold int) error {
	if delayThreshold <= 0 {
		delayThreshold = constant.DefaultSwitchoverDelayThresholdSeconds
	}

	localStatus, err := standbyService.GetLocalStatus()
	if err != nil {
		return errors.Occur(errors.ErrCommonUnexpected, err.Error())
	}
	if localStatus.Role != "PRIMARY" {
		return errors.Occur(errors.ErrStandbySwitchoverLocalNotPrimary, localStatus.Role)
	}
	if localStatus.PendingRole != "INVALID" || localStatus.SwitchoverStatus != "NORMAL" {
		return fmt.Errorf("local role transition is not stable: role=%s pending_role=%s switchover_status=%s",
			localStatus.Role, localStatus.PendingRole, localStatus.SwitchoverStatus)
	}

	var peerResp param.StandbyStatusResp
	if err := callPeerRpcStandbyStatus(peerHost, peerObshellPort, &peerResp); err != nil {
		return errors.Occur(errors.ErrStandbySwitchoverPeerUnreachable, peerHost, peerObshellPort)
	}
	if peerResp.Local.Role != "STANDBY" {
		return errors.Occur(errors.ErrStandbySwitchoverPeerNotStandby, peerResp.Local.Role)
	}
	if peerResp.Local.PendingRole != "INVALID" || peerResp.Local.SwitchoverStatus != "NORMAL" {
		return fmt.Errorf("peer role transition is not stable: role=%s pending_role=%s switchover_status=%s",
			peerResp.Local.Role, peerResp.Local.PendingRole, peerResp.Local.SwitchoverStatus)
	}

	if localStatus.SyncScn > peerResp.Local.SyncScn {
		lagScn := localStatus.SyncScn - peerResp.Local.SyncScn
		lagSeconds := lagScn / 1_000_000_000
		if lagSeconds > uint64(delayThreshold) {
			return errors.Occur(errors.ErrStandbySwitchoverLagExceedsThreshold, lagSeconds, delayThreshold)
		}
	}

	return nil
}

// CreateSwitchoverDag builds and enqueues a Switchover DAG.
// Flow: PreCheck → SetLogRestoreSource → PrepareOldPrimary → PromoteNewPrimary
// → RestartOldPrimary → FinalizeOldPrimary → PostCheck
func CreateSwitchoverDag(p param.SwitchoverParam) (*task.DagDetailDTO, error) {
	_, err := standbyService.GetPeerByAddr(p.PeerHost, p.PeerObshellPort)
	if err != nil {
		return nil, errors.Occur(errors.ErrStandbyPeerNotFound, p.PeerHost, p.PeerObshellPort)
	}

	delayThreshold := p.DelayThresholdSeconds
	if delayThreshold <= 0 {
		delayThreshold = constant.DefaultSwitchoverDelayThresholdSeconds
	}

	ctx := task.NewTaskContext().
		SetParam(constant.PARAM_STANDBY_PEER_HOST, p.PeerHost).
		SetParam(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, p.PeerObshellPort).
		SetParam(constant.PARAM_SWITCHOVER_DELAY_THRESHOLD, delayThreshold).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)

	builder := task.NewTemplateBuilder(constant.DAG_SWITCHOVER)
	builder.
		AddTask(newSwitchoverPreCheckTask(), false).
		AddTask(newSwitchoverSetLogRestoreSrcTask(), false).
		AddTask(newSwitchoverPrimaryToStandbyTask(), false).
		AddTask(newSwitchoverStandbyToPrimaryTask(), false)
	if err := observer.AppendRestartTasks(builder, ctx); err != nil {
		return nil, err
	}
	builder.
		AddTask(newSwitchoverFinalizeOldPrimaryTask(), false).
		AddTask(newSwitchoverPostCheckTask(), false).
		SetMaintenance(task.GlobalMaintenance())

	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

// SwitchoverPreCheckTask verifies:
//  1. Local role is PRIMARY.
//  2. The target peer role is STANDBY.
//  3. Replication lag (local.sync_scn - peer.sync_scn) is within threshold.
type SwitchoverPreCheckTask struct {
	task.Task
}

func newSwitchoverPreCheckTask() *SwitchoverPreCheckTask {
	t := &SwitchoverPreCheckTask{
		Task: *task.NewSubTask(constant.TASK_SWITCHOVER_PRECHECK),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SwitchoverPreCheckTask) Execute() error {
	var peerHost string
	var peerObshellPort int
	var delayThreshold int

	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_HOST, &peerHost); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, &peerObshellPort); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_SWITCHOVER_DELAY_THRESHOLD, &delayThreshold); err != nil {
		delayThreshold = constant.DefaultSwitchoverDelayThresholdSeconds
	}

	t.ExecuteLogf("Checking switchover pre-conditions (role, peer %s:%d, lag threshold %ds)",
		peerHost, peerObshellPort, delayThreshold)
	if err := CheckSwitchoverPreConditions(peerHost, peerObshellPort, delayThreshold); err != nil {
		return err
	}
	t.ExecuteLog("Switchover pre-check passed")
	return nil
}

// SwitchoverSetLogRestoreSrcTask sets log_restore_source on the soon-to-be primary
// (current primary's address) so that after SWITCHOVER TO STANDBY the new standby
// can immediately begin replaying from the new primary.
type SwitchoverSetLogRestoreSrcTask struct {
	task.Task
}

func newSwitchoverSetLogRestoreSrcTask() *SwitchoverSetLogRestoreSrcTask {
	t := &SwitchoverSetLogRestoreSrcTask{
		Task: *task.NewSubTask(constant.TASK_SWITCHOVER_SET_LOG_RESTORE_SRC),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SwitchoverSetLogRestoreSrcTask) Execute() error {
	var peerHost string
	var peerObshellPort int
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_HOST, &peerHost); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, &peerObshellPort); err != nil {
		return err
	}

	peer, err := standbyService.GetPeerByAddr(peerHost, peerObshellPort)
	if err != nil {
		return err
	}

	t.ExecuteLogf("Setting log_restore_source to %s:%d", peerHost, peer.PeerRpcPort)
	return standbyService.SetLogRestoreSource(peerHost, peer.PeerRpcPort)
}

// SwitchoverPrimaryToStandbyTask prepares the current primary for a restart as
// STANDBY. New SeekDB keeps the active role PRIMARY, fences writes, and persists
// pending_role=STANDBY until the process restarts.
type SwitchoverPrimaryToStandbyTask struct {
	task.Task
}

func newSwitchoverPrimaryToStandbyTask() *SwitchoverPrimaryToStandbyTask {
	t := &SwitchoverPrimaryToStandbyTask{
		Task: *task.NewSubTask(constant.TASK_SWITCHOVER_PRIMARY_TO_STANDBY),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SwitchoverPrimaryToStandbyTask) Execute() error {
	var peerHost string
	var peerObshellPort int
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_HOST, &peerHost); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, &peerObshellPort); err != nil {
		return err
	}

	t.ExecuteLog("Executing ALTER SYSTEM SWITCHOVER TO STANDBY")
	if err := standbyService.SwitchoverToStandby(); err != nil {
		return err
	}

	status, err := standbyService.GetLocalStatus()
	if err != nil {
		return fmt.Errorf("failed to query prepared old primary: %w", err)
	}
	if status.Role != "PRIMARY" || status.PendingRole != "STANDBY" || status.SwitchoverStatus != "PREPARING" {
		return fmt.Errorf("old primary was not prepared for restart: role=%s pending_role=%s switchover_status=%s",
			status.Role, status.PendingRole, status.SwitchoverStatus)
	}
	t.ExecuteLog("Old primary is fenced with pending_role=STANDBY and ready for restart")
	return nil
}

// SwitchoverStandbyToPrimaryTask calls the internal RPC on the peer standby to
// promote it to PRIMARY, then flips the peer direction metadata to DOWNSTREAM
// (we become the new standby that replicates from the new primary).
type SwitchoverStandbyToPrimaryTask struct {
	task.Task
}

func newSwitchoverStandbyToPrimaryTask() *SwitchoverStandbyToPrimaryTask {
	t := &SwitchoverStandbyToPrimaryTask{
		Task: *task.NewSubTask(constant.TASK_SWITCHOVER_STANDBY_TO_PRIMARY),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SwitchoverStandbyToPrimaryTask) Execute() error {
	var peerHost string
	var peerObshellPort int
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_HOST, &peerHost); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, &peerObshellPort); err != nil {
		return err
	}

	t.ExecuteLogf("Calling peer %s:%d to switchover to primary", peerHost, peerObshellPort)
	rpcParam := param.RpcSwitchoverToPrimaryParam{
		CallerHost:        meta.OCS_AGENT.GetIp(),
		CallerObshellPort: meta.OCS_AGENT.GetPort(),
	}
	if err := callPeerRpcSwitchoverToPrimary(peerHost, peerObshellPort, rpcParam); err != nil {
		return err
	}
	return nil
}

// SwitchoverPostCheckTask verifies that roles and directions have been swapped
// successfully after a Switchover:
//  1. Local role should now be STANDBY (was PRIMARY before switchover).
//  2. Peer role should now be PRIMARY (was STANDBY before switchover).
//  3. Local peer direction should now be UPSTREAM (was DOWNSTREAM before switchover).
type SwitchoverPostCheckTask struct {
	task.Task
}

func newSwitchoverPostCheckTask() *SwitchoverPostCheckTask {
	t := &SwitchoverPostCheckTask{
		Task: *task.NewSubTask(constant.TASK_SWITCHOVER_POSTCHECK),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SwitchoverPostCheckTask) Execute() error {
	var peerHost string
	var peerObshellPort int
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_HOST, &peerHost); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, &peerObshellPort); err != nil {
		return err
	}

	// Check local role is now STANDBY — hard fail: a wrong local role means
	// the OB-level switchover did not complete, and the DAG must not succeed.
	t.ExecuteLog("Verifying local role is now STANDBY")
	localStatus, err := standbyService.GetLocalStatus()
	if err != nil {
		return fmt.Errorf("postcheck: failed to query local status: %w", err)
	}
	if localStatus.Role != "STANDBY" {
		return fmt.Errorf("postcheck: local role is %s, expected STANDBY after switchover; OB-level switchover may have failed", localStatus.Role)
	}
	if localStatus.PendingRole != "INVALID" || localStatus.SwitchoverStatus != "NORMAL" {
		return fmt.Errorf("postcheck: local role transition is not stable: pending_role=%s switchover_status=%s",
			localStatus.PendingRole, localStatus.SwitchoverStatus)
	}

	// Check peer role is now PRIMARY.
	t.ExecuteLogf("Verifying peer %s:%d role is now PRIMARY", peerHost, peerObshellPort)
	peerStatus, err := callPeerGetStatus(peerHost, peerObshellPort)
	if err != nil {
		return fmt.Errorf("postcheck: failed to query peer status: %w", err)
	}
	if peerStatus.Local.Role != "PRIMARY" {
		return fmt.Errorf("postcheck: peer role is %s, expected PRIMARY after switchover", peerStatus.Local.Role)
	}
	if peerStatus.Local.PendingRole != "INVALID" || peerStatus.Local.SwitchoverStatus != "NORMAL" {
		return fmt.Errorf("postcheck: peer role transition is not stable: pending_role=%s switchover_status=%s",
			peerStatus.Local.PendingRole, peerStatus.Local.SwitchoverStatus)
	}

	// Check local peer direction has been flipped to UPSTREAM.
	peer, err := standbyService.GetPeerByAddr(peerHost, peerObshellPort)
	if err != nil {
		return fmt.Errorf("postcheck: failed to query local peer record: %w", err)
	}
	if peer.Direction != constant.STANDBY_DIRECTION_UPSTREAM {
		return fmt.Errorf("postcheck: local peer direction is %s, expected UPSTREAM after switchover", peer.Direction)
	}

	// Check peer's direction for our address has been flipped to DOWNSTREAM.
	// The RPC status response includes the peer's local SQLite records.
	peerDirectionOk := false
	selfHost := meta.OCS_AGENT.GetIp()
	selfPort := meta.OCS_AGENT.GetPort()
	for _, remotePeer := range peerStatus.Peers {
		if remotePeer.PeerHost == selfHost && remotePeer.PeerObshellPort == selfPort {
			if remotePeer.Direction == constant.STANDBY_DIRECTION_DOWNSTREAM {
				peerDirectionOk = true
			} else {
				return fmt.Errorf("postcheck: peer's direction for us is %s, expected DOWNSTREAM after switchover", remotePeer.Direction)
			}
			break
		}
	}
	if !peerDirectionOk {
		return fmt.Errorf("postcheck: peer has no DOWNSTREAM record pointing to us")
	}

	t.ExecuteLog(fmt.Sprintf("PostCheck passed: local=%s, peer=%s, localDirection=%s",
		localStatus.Role, peerStatus.Local.Role, peer.Direction))
	return nil
}

// ExecuteSwitchoverToPrimary is the server-side handler for the internal
// POST /rpc/v1/seekdb/standby/switchover-to-primary RPC.
// It executes ALTER SYSTEM SWITCHOVER TO PRIMARY on the local standby and
// flips the caller peer's direction record from UPSTREAM → DOWNSTREAM.
func ExecuteSwitchoverToPrimary(p param.RpcSwitchoverToPrimaryParam) error {
	if err := standbyService.SwitchoverToPrimary(); err != nil {
		return err
	}
	status, err := standbyService.GetLocalStatus()
	if err != nil {
		return fmt.Errorf("failed to query promoted standby: %w", err)
	}
	if status.Role != "PRIMARY" || status.PendingRole != "INVALID" || status.SwitchoverStatus != "NORMAL" {
		return fmt.Errorf("standby promotion did not reach a stable primary state: role=%s pending_role=%s switchover_status=%s",
			status.Role, status.PendingRole, status.SwitchoverStatus)
	}

	// Flip the caller (original primary, now standby) to DOWNSTREAM.
	return standbyService.FlipDirection(
		p.CallerHost, p.CallerObshellPort,
		constant.STANDBY_DIRECTION_DOWNSTREAM,
	)
}

// SwitchoverFinalizeOldPrimaryTask runs after the seekdb-only restart. The
// persisted pending role must have been consumed by SeekDB startup before the
// local obshell metadata is changed to UPSTREAM.
type SwitchoverFinalizeOldPrimaryTask struct {
	task.Task
}

func newSwitchoverFinalizeOldPrimaryTask() *SwitchoverFinalizeOldPrimaryTask {
	t := &SwitchoverFinalizeOldPrimaryTask{
		Task: *task.NewSubTask(constant.TASK_SWITCHOVER_FINALIZE_OLD_PRIMARY),
	}
	t.SetCanRetry().SetCanContinue().SetCanPass().SetCanCancel()
	return t
}

func (t *SwitchoverFinalizeOldPrimaryTask) Execute() error {
	var peerHost string
	var peerObshellPort int
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_HOST, &peerHost); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(constant.PARAM_STANDBY_PEER_OBSHELL_PORT, &peerObshellPort); err != nil {
		return err
	}

	status, err := standbyService.GetLocalStatus()
	if err != nil {
		return fmt.Errorf("failed to query restarted old primary: %w", err)
	}
	if status.Role != "STANDBY" || status.PendingRole != "INVALID" || status.SwitchoverStatus != "NORMAL" {
		return fmt.Errorf("old primary restart did not activate standby role: role=%s pending_role=%s switchover_status=%s",
			status.Role, status.PendingRole, status.SwitchoverStatus)
	}

	t.ExecuteLogf("Flipping peer %s:%d direction to UPSTREAM", peerHost, peerObshellPort)
	return standbyService.FlipDirection(peerHost, peerObshellPort, constant.STANDBY_DIRECTION_UPSTREAM)
}
