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
	"net"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	agenthttp "github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
	"github.com/oceanbase/obshell/seekdb/param"
)

// isValidHost returns true when host contains only characters safe for use in
// ALTER SYSTEM SET log_restore_source (alphanumeric, dots, hyphens), preventing
// injection into that DDL statement.
func isValidHost(host string) bool {
	if net.ParseIP(host) != nil {
		return true
	}
	if len(host) == 0 {
		return false
	}
	for _, c := range host {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '-') {
			return false
		}
	}
	return true
}

// WritePair validates the request and delegates the DB upsert to the service.
func WritePair(p param.PairParam) error {
	if p.Direction != constant.STANDBY_DIRECTION_UPSTREAM && p.Direction != constant.STANDBY_DIRECTION_DOWNSTREAM {
		return errors.Occur(errors.ErrStandbyInvalidDirection, p.Direction)
	}
	if !isValidHost(p.PeerHost) {
		return errors.Occur(errors.ErrStandbyInvalidPeerHost, p.PeerHost)
	}
	if p.PeerObshellPort < 1 || p.PeerObshellPort > 65535 {
		return errors.Occur(errors.ErrStandbyInvalidPeerPort, p.PeerObshellPort)
	}
	if p.PeerRpcPort < 1 || p.PeerRpcPort > 65535 {
		return errors.Occur(errors.ErrStandbyInvalidPeerPort, p.PeerRpcPort)
	}
	if p.Direction == constant.STANDBY_DIRECTION_UPSTREAM {
		localStatus, err := standbyService.GetLocalStatus()
		if err != nil {
			return fmt.Errorf("failed to query local seekdb status: %w", err)
		}
		if localStatus.Role != "STANDBY" {
			return errors.Occur(errors.ErrStandbyLocalRoleNotStandby, localStatus.Role)
		}
	}
	return standbyService.UpsertPeer(p)
}

// DeletePair checks the local role, activates the standby if needed, then
// removes the peer record and asynchronously notifies the remote peer.
// Only STANDBY nodes are allowed to call this; PRIMARY nodes are rejected.
func DeletePair(p param.PairDeleteParam) error {
	localStatus, err := standbyService.GetLocalStatus()
	if err != nil {
		return fmt.Errorf("failed to query local seekdb status: %w", err)
	}
	if localStatus.Role == "PRIMARY" {
		return errors.Occur(errors.ErrStandbyDecoupleFromPrimaryNotAllowed)
	}
	if localStatus.Role == "STANDBY" {
		if err := standbyService.ActivateStandby(); err != nil {
			return fmt.Errorf("failed to activate standby before decouple: %w", err)
		}
	}

	peer, err := standbyService.DeletePairRecord(p)
	if err != nil {
		return err
	}

	if p.NotifyPeer && peer != nil {
		go notifyPeerDeletePair(*peer)
	}
	return nil
}

// notifyPeerDeletePair sends a best-effort DELETE /pair notification to the
// peer via the RPC channel. Token and public key are taken from the peer struct
// rather than re-queried from SQLite: by the time this runs the local record
// may already be deleted.
func notifyPeerDeletePair(peer sqlite.SeekdbStandbyPeer) bool {
	uri := constant.URI_SEEKDB_STANDBY_RPC_PREFIX + constant.URI_PAIR
	rpcParam := param.RpcPairDeleteParam{
		PeerHost:        meta.OCS_AGENT.GetIp(),
		PeerObshellPort: meta.OCS_AGENT.GetPort(),
	}
	encBody, headers, err := secure.BuildStandbyBodyAndHeader(uri, peer.PeerToken, peer.PeerPublicKey, rpcParam)
	if err != nil {
		return false
	}
	err = agenthttp.SendRequestAndBuildReturn(
		&meta.AgentInfo{Ip: peer.PeerHost, Port: peer.PeerObshellPort},
		uri, agenthttp.DELETE, encBody, nil, headers)
	return err == nil
}
