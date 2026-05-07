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

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	agenthttp "github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
	"github.com/oceanbase/obshell/seekdb/param"
)

// getPeerPublicKey returns the RSA public key for the given peer.
// Checks the SQLite cache first; if empty, fetches from the peer's
// /api/v1/secret endpoint and updates the cache for subsequent calls.
func getPeerPublicKey(host string, port int) string {
	peer, err := standbyService.GetPeerByAddr(host, port)
	if err != nil {
		return ""
	}
	if peer.PeerPublicKey != "" {
		return peer.PeerPublicKey
	}
	// Not cached — fetch from peer.
	var secret meta.AgentSecret
	if err := agenthttp.SendGetRequest(
		&meta.AgentInfo{Ip: host, Port: port},
		constant.URI_API_V1+"/"+constant.URI_SECRET,
		nil, &secret); err != nil {
		log.WithError(err).Warnf("failed to fetch public key from peer %s:%d", host, port)
		return ""
	}
	if secret.PublicKey == "" {
		return ""
	}
	if err := standbyService.UpdatePeerPublicKey(host, port, secret.PublicKey); err != nil {
		log.WithError(err).Warnf("failed to cache public key for peer %s:%d", host, port)
	}
	return secret.PublicKey
}

// BuildTokenHeader looks up the peer_token for the given peer, lazily fetches
// the peer's RSA public key if not yet cached, and returns an encrypted
// X-OCS-Header map for the specified URI.
func BuildTokenHeader(host string, port int, uri string) map[string]string {
	peer, err := standbyService.GetPeerByAddr(host, port)
	if err != nil || peer.PeerToken == "" {
		return nil
	}
	pk := getPeerPublicKey(host, port)
	encrypted, err := secure.BuildStandbyHeader(uri, peer.PeerToken, pk)
	if err != nil {
		return nil
	}
	return map[string]string{constant.OCS_HEADER: encrypted}
}

// BuildBodyAndTokenHeader encrypts body with AES, embeds the key in the header,
// and RSA-encrypts the header with the peer's public key.
// Returns the encrypted body and X-OCS-Header map for use in RPC calls.
func BuildBodyAndTokenHeader(uri string, host string, port int, body interface{}) (encryptedBody interface{}, headers map[string]string, err error) {
	peer, err := standbyService.GetPeerByAddr(host, port)
	if err != nil || peer.PeerToken == "" {
		return nil, nil, fmt.Errorf("peer token not found for %s:%d", host, port)
	}
	pk := getPeerPublicKey(host, port)
	if pk == "" {
		return nil, nil, fmt.Errorf("peer public key not found for %s:%d", host, port)
	}
	return secure.BuildStandbyBodyAndHeader(uri, peer.PeerToken, pk, body)
}

// callPeerRpcStandbyStatus queries the remote obshell's local standby status
// via the RPC channel (returns only local status, no peer aggregation).
func callPeerRpcStandbyStatus(host string, port int, resp *param.StandbyStatusResp) error {
	uri := constant.URI_SEEKDB_STANDBY_RPC_PREFIX + constant.URI_STANDBY_STATUS
	return agenthttp.SendRequestAndBuildReturn(
		&meta.AgentInfo{Ip: host, Port: port},
		uri,
		agenthttp.GET, nil, resp,
		BuildTokenHeader(host, port, uri))
}

// callPeerGetStatus fetches the standby status from a remote obshell via the
// RPC channel (carries encrypted X-OCS-Header with standby token).
func callPeerGetStatus(host string, port int) (*param.StandbyStatusResp, error) {
	uri := constant.URI_SEEKDB_STANDBY_RPC_PREFIX + constant.URI_STANDBY_STATUS
	var resp param.StandbyStatusResp
	err := agenthttp.SendRequestAndBuildReturn(
		&meta.AgentInfo{Ip: host, Port: port},
		uri,
		agenthttp.GET, nil, &resp,
		BuildTokenHeader(host, port, uri))
	return &resp, err
}

// callPeerRpcSwitchoverToPrimary calls the internal switchover-to-primary RPC
// on the peer obshell (the standby that will become primary).
func callPeerRpcSwitchoverToPrimary(host string, port int, p param.RpcSwitchoverToPrimaryParam) error {
	uri := constant.URI_SEEKDB_STANDBY_RPC_PREFIX + constant.URI_SWITCHOVER_TO_PRIMARY
	encBody, headers, err := BuildBodyAndTokenHeader(uri, host, port, p)
	if err != nil {
		return err
	}
	return agenthttp.SendRequestAndBuildReturn(
		&meta.AgentInfo{Ip: host, Port: port},
		uri, agenthttp.POST, encBody, nil, headers)
}
