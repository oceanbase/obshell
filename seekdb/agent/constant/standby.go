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

package constant

// URI segments for standby API.
const (
	URI_RPC_V1 = "/rpc/v1"

	URI_STANDBY_GROUP         = "/standby"
	URI_PAIR                  = "/pair"
	URI_STANDBY_STATUS        = "/status"
	URI_SWITCHOVER            = "/switchover"
	URI_ACTIVATE              = "/activate"
	URI_SWITCHOVER_TO_PRIMARY = "/switchover-to-primary"

	URI_SEEKDB_STANDBY_API_PREFIX = URI_API_V1 + URI_SEEKDB_GROUP + URI_STANDBY_GROUP
	URI_SEEKDB_STANDBY_RPC_PREFIX = URI_RPC_V1 + URI_SEEKDB_GROUP + URI_STANDBY_GROUP
)

// Token authentication for inter-obshell RPC calls.
const (
	HEADER_STANDBY_TOKEN   = "X-Standby-Token"
	OCS_INFO_STANDBY_TOKEN = "standby_token"
)

// Peer relationship direction values stored in SeekdbStandbyPeer.Direction.
const (
	STANDBY_DIRECTION_UPSTREAM   = "UPSTREAM"
	STANDBY_DIRECTION_DOWNSTREAM = "DOWNSTREAM"
)

// DAG and task node names for standby operations.
const (
	DAG_SWITCHOVER = "Switchover"
	DAG_ACTIVATE   = "Activate"

	TASK_SWITCHOVER_PRECHECK             = "SwitchoverPreCheck"
	TASK_SWITCHOVER_SET_LOG_RESTORE_SRC  = "SwitchoverSetLogRestoreSource"
	TASK_SWITCHOVER_PRIMARY_TO_STANDBY   = "SwitchoverPrimaryToStandby"
	TASK_SWITCHOVER_STANDBY_TO_PRIMARY   = "SwitchoverStandbyToPrimary"
	TASK_SWITCHOVER_POSTCHECK            = "SwitchoverPostCheck"
	TASK_ACTIVATE_PRECHECK               = "ActivatePreCheck"
	TASK_ACTIVATE_NODE                   = "Activate"
	TASK_ACTIVATE_CLEANUP                = "ActivateCleanupMeta"
)

// Context parameter keys passed between DAG task nodes.
const (
	PARAM_STANDBY_PEER_HOST          = "peer_host"
	PARAM_STANDBY_PEER_OBSHELL_PORT  = "peer_obshell_port"
	PARAM_SWITCHOVER_DELAY_THRESHOLD = "delay_threshold_seconds"
)

// Default Switchover replication lag threshold (seconds).
const DefaultSwitchoverDelayThresholdSeconds = 10

// SyncStatus values reported for a STANDBY node in LocalStandbyStatus /
// PeerStandbyStatus. PRIMARY nodes return "". The enum is intentionally
// small: "completely caught up" is conveyed by lag_seconds == 0 on the
// numeric side rather than a dedicated status value.
const (
	SyncStatusNormal     = "NORMAL_SYNC"  // sync delay below threshold
	SyncStatusDelayed    = "SYNC_DELAYED" // fetch delay >= SyncDelayThresholdSeconds
	SyncStatusSyncPaused = "SYNC_PAUSED"  // log_restore_source is empty
	SyncStatusUnknown    = "UNKNOWN"      // cannot reach upstream management-plane to derive delay
)

// SyncDelayThresholdSeconds mirrors OCP's standby_tenant_sync_delay_too_long
// alarm threshold (ocp_alarm_template.yaml: 600s).
const SyncDelayThresholdSeconds uint64 = 600
