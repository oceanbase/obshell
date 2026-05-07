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

package api

import (
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/seekdb/agent/api/common"
	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	standbyexec "github.com/oceanbase/obshell/seekdb/agent/executor/standby"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
	standbyservice "github.com/oceanbase/obshell/seekdb/agent/service/standby"
	"github.com/oceanbase/obshell/seekdb/param"
)

var (
	standbyService = standbyservice.StandbyService{}
)

// ─── User-facing handlers ────────────────────────────────────────────────────

// @ID putStandbyPair
// @Summary Create or update a standby peer relationship
// @Tags seekdb
// @Accept application/json
// @Produce application/json
// @Param body body param.PairParam true "pair params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/standby/pair [put]
func standbyPairUpsertHandler(c *gin.Context) {
	var p param.PairParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := standbyexec.WritePair(p)
	common.SendResponse(c, nil, err)
}

// @ID deleteStandbyPair
// @Summary Remove a standby peer relationship
// @Tags seekdb
// @Accept application/json
// @Produce application/json
// @Param body body param.PairDeleteParam true "pair delete params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/standby/pair [delete]
func standbyPairDeleteHandler(c *gin.Context) {
	var p param.PairDeleteParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if err := standbyexec.DeletePair(p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, gin.H{"deleted": true}, nil)
}

// @ID getStandbyStatus
// @Summary Get local standby role and peer statuses
// @Tags seekdb
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=param.StandbyStatusResp}
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/standby/status [get]
func standbyStatusHandler(c *gin.Context) {
	resp, err := standbyexec.GetFullStatus()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, resp, nil)
}

// @ID postStandbySwitchover
// @Summary Trigger a Switchover DAG (primary → standby)
// @Tags seekdb
// @Accept application/json
// @Produce application/json
// @Param body body param.SwitchoverParam true "switchover params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/standby/switchover [post]
func standbySwitchoverHandler(c *gin.Context) {
	var p param.SwitchoverParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	data, err := standbyexec.Switchover(p)
	common.SendResponse(c, data, err)
}

// @ID postStandbyActivate
// @Summary Trigger an Activate DAG (unilateral promotion to primary)
// @Tags seekdb
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/standby/activate [post]
func standbyActivateHandler(c *gin.Context) {
	data, err := standbyexec.Activate()
	common.SendResponse(c, data, err)
}

// @ID postStandbyToken
// @Summary Generate or retrieve the local standby token
// @Tags seekdb
// @Accept application/json
// @Produce application/json
// @Param body body param.TokenParam true "token params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=param.TokenResp}
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/seekdb/standby/token [post]
func standbyTokenHandler(c *gin.Context) {
	var p param.TokenParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	token, err := secure.GenerateStandbyToken(p.Force)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	// Local-route (Unix socket) callers are already protected by UID checking;
	// return the token in plaintext so the deploy/admin scripts can use plain curl.
	if config.IsEncryptionDisabled() || c.GetBool(constant.LOCAL_ROUTE_KEY) {
		common.SendResponse(c, param.TokenResp{Token: token}, nil)
		return
	}
	// Encrypt the token with the caller's AES keys (carried in X-OCS-Header.Keys).
	// This mirrors OB's Login behaviour: the raw credential never travels in plaintext.
	obHeaderByte, _ := c.Get(constant.OCS_HEADER)
	header, ok := obHeaderByte.(secure.HttpHeader)
	if !ok || len(header.Keys) == 0 {
		common.SendResponse(c, nil, errors.Occur(errors.ErrRequestAesKeyNotFound))
		return
	}
	encrypted, err := secure.EncryptWithAesKeys([]byte(token), string(header.Keys))
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, encrypted, nil)
}

// ─── Internal RPC handlers (Token auth, peer-to-peer only) ──────────────────

// @ID rpcSwitchoverToPrimary
// @Summary Internal RPC: execute SWITCHOVER TO PRIMARY on this standby node
// @Tags seekdb
// @Accept application/json
// @Produce application/json
// @Param body body param.RpcSwitchoverToPrimaryParam true "rpc params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /rpc/v1/seekdb/standby/switchover-to-primary [post]
func rpcSwitchoverToPrimaryHandler(c *gin.Context) {
	var p param.RpcSwitchoverToPrimaryParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	err := standbyexec.ExecuteSwitchoverToPrimary(p)
	common.SendResponse(c, nil, err)
}

// @ID rpcStandbyStatus
// @Summary Internal RPC: get local standby status and SQLite peer rows (no remote peer queries)
// @Description Returns local role/status and peer records from SQLite only; does not call remote obshell (avoids RPC recursion).
// @Tags seekdb
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=param.StandbyStatusResp}
// @Failure 500 object http.OcsAgentResponse
// @Router /rpc/v1/seekdb/standby/status [get]
func rpcStandbyStatusHandler(c *gin.Context) {
	local, err := standbyService.GetLocalStatus()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	resp := param.StandbyStatusResp{Local: local}
	// Include local peer records (SQLite-only, no remote calls) so callers
	// like PostCheck can verify direction flips without a second RPC.
	if peers, peerErr := standbyService.GetPeers(); peerErr == nil {
		for _, p := range peers {
			resp.Peers = append(resp.Peers, param.PeerStandbyStatus{
				PeerInfo: param.PeerInfo{
					PeerHost:        p.PeerHost,
					PeerObshellPort: p.PeerObshellPort,
					PeerRpcPort:     p.PeerRpcPort,
					Direction:       p.Direction,
				},
			})
		}
	}
	common.SendResponse(c, resp, nil)
}

// @ID rpcPairDelete
// @Summary Internal RPC: delete local standby pair record (peer cleanup notification)
// @Description Deletes the local SQLite pair row by peer identity; bypasses role checks and Activate DAG.
// @Tags seekdb
// @Accept application/json
// @Produce application/json
// @Param body body param.RpcPairDeleteParam true "rpc pair delete params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /rpc/v1/seekdb/standby/pair [delete]
func rpcPairDeleteHandler(c *gin.Context) {
	var p param.RpcPairDeleteParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	_, err := standbyService.DeletePairRecord(param.PairDeleteParam{
		PeerHost:        p.PeerHost,
		PeerObshellPort: p.PeerObshellPort,
	})
	common.SendResponse(c, gin.H{"deleted": true}, err)
}
