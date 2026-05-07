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

package param

// TokenParam is used by POST /api/v1/seekdb/standby/token.
type TokenParam struct {
	Force bool `json:"force"`
}

// TokenResp is returned by POST /api/v1/seekdb/standby/token.
type TokenResp struct {
	Token string `json:"token"`
}

// RpcPairDeleteParam is used by DELETE /rpc/v1/seekdb/standby/pair.
type RpcPairDeleteParam struct {
	PeerHost        string `json:"peer_host" binding:"required"`
	PeerObshellPort int    `json:"peer_obshell_port" binding:"required"`
}

// PairParam is used by PUT /api/v1/seekdb/standby/pair.
type PairParam struct {
	PeerHost        string `json:"peer_host" binding:"required"`
	PeerObshellPort int    `json:"peer_obshell_port" binding:"required"`
	PeerRpcPort     int    `json:"peer_rpc_port" binding:"required"`
	Direction       string `json:"direction" binding:"required"` // UPSTREAM / DOWNSTREAM
	Token           string `json:"token" binding:"required"`
}

// PairDeleteParam is used by DELETE /api/v1/seekdb/standby/pair.
type PairDeleteParam struct {
	PeerHost        string `json:"peer_host" binding:"required"`
	PeerObshellPort int    `json:"peer_obshell_port" binding:"required"`
	NotifyPeer      bool   `json:"notify_peer"`
}

// SwitchoverParam is used by POST /api/v1/seekdb/standby/switchover.
type SwitchoverParam struct {
	PeerHost              string `json:"peer_host" binding:"required"`
	PeerObshellPort       int    `json:"peer_obshell_port" binding:"required"`
	DelayThresholdSeconds int    `json:"delay_threshold_seconds"`
}

// ActivateParam is used by POST /api/v1/seekdb/standby/activate.
type ActivateParam struct{}

// RpcSwitchoverToPrimaryParam is used by the internal
// POST /rpc/v1/seekdb/standby/switchover-to-primary RPC.
type RpcSwitchoverToPrimaryParam struct {
	CallerHost        string `json:"caller_host" binding:"required"`
	CallerObshellPort int    `json:"caller_obshell_port" binding:"required"`
}

// PeerInfo is a compact address descriptor for a remote obshell peer.
type PeerInfo struct {
	PeerHost        string `json:"peer_host"`
	PeerObshellPort int    `json:"peer_obshell_port"`
	PeerRpcPort     int    `json:"peer_rpc_port"`
	Direction       string `json:"direction"`
}

// LocalStandbyStatus describes the local seekdb node state.
type LocalStandbyStatus struct {
	Role             string `json:"role"` // PRIMARY / STANDBY / unknown
	InstanceName     string `json:"instance_name,omitempty"`
	LogRestoreSource string `json:"log_restore_source"`
	SyncScn          uint64 `json:"sync_scn"`
	ReadableScn      uint64 `json:"readable_scn"`
	// SyncStatus is populated by StandbyService.GetFullStatus (which has the
	// upstream's sync_scn available) rather than by GetLocalStatus. PRIMARY
	// nodes leave this empty.
	SyncStatus string `json:"sync_status,omitempty"` // NORMAL_SYNC / SYNC_DELAYED / SYNC_PAUSED / UNKNOWN; empty for PRIMARY
}

// PeerStandbyStatus describes the state of a single remote peer as seen from
// the peer's own GET /standby/status response (nested under each peer entry).
type PeerStandbyStatus struct {
	PeerInfo
	Role         string `json:"role"`
	InstanceName string `json:"instance_name,omitempty"`
	SyncScn      uint64 `json:"sync_scn"`
	ReadableScn  uint64 `json:"readable_scn"`
	LagSeconds   *int64 `json:"lag_seconds,omitempty"`
	SyncStatus   string `json:"sync_status,omitempty"` // from peer's perspective; UNKNOWN when peer obshell unreachable
	Error        string `json:"error,omitempty"`
}

// StandbyStatusResp is returned by GET /api/v1/seekdb/standby/status.
type StandbyStatusResp struct {
	Local LocalStandbyStatus  `json:"local"`
	Peers []PeerStandbyStatus `json:"peers"`
}
