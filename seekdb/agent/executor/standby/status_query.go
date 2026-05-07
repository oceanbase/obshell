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
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/param"
)

// computeSyncStatus derives the sync_status enum for a STANDBY node from the
// upstream's authoritative sync_scn and the standby's own sync_scn. The
// semantics are aligned with OCP's standby_tenant_delay_seconds metric
// (ocp-service CheckAllStandbyTenantTask): delay is primary - standby SCN
// diff in seconds, and SYNC_DELAYED fires at >= SyncDelayThresholdSeconds.
//
// When upstreamAvailable is false (e.g. the upstream obshell management-plane
// is unreachable), we return UNKNOWN rather than asserting a network error,
// because seekdb data-plane replication runs directly between the seekdb
// processes' rpc ports and is independent of obshell availability.
func computeSyncStatus(upstreamAvailable bool, upstreamSyncScn, standbySyncScn uint64, hasRestoreSource bool) string {
	if !hasRestoreSource {
		return constant.SyncStatusSyncPaused
	}
	if !upstreamAvailable {
		return constant.SyncStatusUnknown
	}
	var fetchDeltaNs uint64
	if upstreamSyncScn > standbySyncScn {
		fetchDeltaNs = upstreamSyncScn - standbySyncScn
	}
	if fetchDeltaNs/1_000_000_000 >= constant.SyncDelayThresholdSeconds {
		return constant.SyncStatusDelayed
	}
	return constant.SyncStatusNormal
}

// QueryPeerStatus queries a specific peer obshell's local status via the RPC
// channel (carries X-Standby-Token).
func QueryPeerStatus(host string, port int) (param.LocalStandbyStatus, error) {
	var resp param.StandbyStatusResp
	if err := callPeerRpcStandbyStatus(host, port, &resp); err != nil {
		return param.LocalStandbyStatus{Role: "unknown"}, err
	}
	return resp.Local, nil
}

// GetFullStatus assembles local status and remote peer statuses. SyncStatus
// is populated here because it depends on cross-referencing local and peer
// sync_scn values.
func GetFullStatus() (param.StandbyStatusResp, error) {
	local, localErr := standbyService.GetLocalStatus()
	if localErr != nil {
		local.Role = "unknown"
	}

	peers, err := standbyService.GetPeers()
	if err != nil {
		return param.StandbyStatusResp{Local: local}, err
	}

	// Query each peer, remembering the authoritative upstream sync_scn (if
	// we are a STANDBY) so we can derive local.SyncStatus after the loop.
	var upstreamReachable bool
	var upstreamSyncScn uint64

	peerStatuses := make([]param.PeerStandbyStatus, 0, len(peers))
	for _, p := range peers {
		ps := param.PeerStandbyStatus{
			PeerInfo: param.PeerInfo{
				PeerHost:        p.PeerHost,
				PeerObshellPort: p.PeerObshellPort,
				PeerRpcPort:     p.PeerRpcPort,
				Direction:       p.Direction,
			},
		}

		var remoteResp param.StandbyStatusResp
		if callErr := callPeerRpcStandbyStatus(p.PeerHost, p.PeerObshellPort, &remoteResp); callErr != nil {
			// The peer's obshell management-plane is unreachable.
			// DOWNSTREAM: surface UNKNOWN sync_status.
			// UPSTREAM: lag_seconds cannot be computed without peer's sync_scn; leave nil.
			ps.Error = callErr.Error()
			if p.Direction == constant.STANDBY_DIRECTION_DOWNSTREAM {
				ps.SyncStatus = constant.SyncStatusUnknown
			}
		} else {
			ps.Role = remoteResp.Local.Role
			ps.InstanceName = remoteResp.Local.InstanceName
			ps.SyncScn = remoteResp.Local.SyncScn
			ps.ReadableScn = remoteResp.Local.ReadableScn
			if p.Direction == constant.STANDBY_DIRECTION_DOWNSTREAM {
				if remoteResp.Local.Role == "PRIMARY" {
					// Peer promoted itself (e.g. after ACTIVATE STANDBY); our
					// record is stale. Clean it up and exclude from response.
					_, _ = standbyService.DeletePairRecord(param.PairDeleteParam{
						PeerHost:        p.PeerHost,
						PeerObshellPort: p.PeerObshellPort,
					})
					continue
				}
				if localErr != nil {
					// Local seekdb is unreachable; local.SyncScn is the zero value
					// and cannot serve as the authoritative upstream SCN.
					ps.SyncStatus = constant.SyncStatusUnknown
				} else {
					// Local is PRIMARY relative to this peer; we hold the
					// authoritative sync_scn.
					ps.SyncStatus = computeSyncStatus(true, local.SyncScn, remoteResp.Local.SyncScn, remoteResp.Local.LogRestoreSource != "")
					// Compute lag for all known-SCN states: NORMAL, DELAYED, and
					// SYNC_PAUSED. SYNC_PAUSED means the standby paused its restore
					// source but the SCNs are still valid, so lag is meaningful.
					// Skip only UNKNOWN (upstream SCN unavailable).
					if ps.SyncStatus != constant.SyncStatusUnknown {
						var lag int64
						if local.SyncScn >= remoteResp.Local.SyncScn {
							lag = int64(local.SyncScn-remoteResp.Local.SyncScn) / 1_000_000_000
						}
						ps.LagSeconds = &lag
					}
				}
			} else if p.Direction == constant.STANDBY_DIRECTION_UPSTREAM {
				// Peer is our upstream PRIMARY; record its sync_scn for
				// computing our local.SyncStatus below.
				// lag_seconds here reflects how far the local standby is
				// behind the upstream primary: (primary.sync_scn - local.sync_scn) / 1e9.
				// Skip when local seekdb is unreachable: local.SyncScn would be
				// the zero value and produce a meaningless inflated lag reading.
				upstreamReachable = true
				upstreamSyncScn = remoteResp.Local.SyncScn
				if localErr == nil {
					var lag int64
					if remoteResp.Local.SyncScn > local.SyncScn {
						lag = int64(remoteResp.Local.SyncScn-local.SyncScn) / 1_000_000_000
					}
					ps.LagSeconds = &lag
				}
			}
		}
		peerStatuses = append(peerStatuses, ps)
	}

	// Populate local.SyncStatus now that the upstream has been queried.
	if local.Role == "STANDBY" {
		local.SyncStatus = computeSyncStatus(upstreamReachable, upstreamSyncScn, local.SyncScn, local.LogRestoreSource != "")
	}

	return param.StandbyStatusResp{
		Local: local,
		Peers: peerStatuses,
	}, nil
}
