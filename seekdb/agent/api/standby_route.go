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
	"github.com/oceanbase/obshell/seekdb/agent/constant"
)

// InitStandbyRoutes registers all standby API routes.
//   - seekdbGroup: the already-created /api/v1/seekdb group (Verify middleware applied).
//   - r: root gin.Engine for the internal /rpc/v1 group (Token auth).
func InitStandbyRoutes(r *gin.Engine, seekdbGroup *gin.RouterGroup, isLocalRoute bool) {
	// ── User-facing API (/api/v1/seekdb/standby/*) — Verify middleware ──
	standby := seekdbGroup.Group(constant.URI_STANDBY_GROUP)
	if !isLocalRoute {
		standby.Use(common.Verify())
	}

	standby.PUT(constant.URI_PAIR, standbyPairUpsertHandler)
	standby.DELETE(constant.URI_PAIR, standbyPairDeleteHandler)
	standby.GET(constant.URI_STANDBY_STATUS, standbyStatusHandler)
	standby.POST(constant.URI_SWITCHOVER, standbySwitchoverHandler)
	standby.POST(constant.URI_ACTIVATE, standbyActivateHandler)
	standby.POST(constant.URI_TOKEN, standbyTokenHandler)

	// ── Internal RPC (/rpc/v1/*) — Token auth, TCP only ──
	if !isLocalRoute {
		// /rpc/v1/seekdb/standby/*
		rpcStandby := r.Group(constant.URI_RPC_V1).Group(constant.URI_SEEKDB_GROUP + constant.URI_STANDBY_GROUP)
		rpcStandby.Use(common.PaddingBody())
		rpcStandby.Use(common.VerifyStandbyToken())
		rpcStandby.GET(constant.URI_STANDBY_STATUS, rpcStandbyStatusHandler)
		rpcStandby.POST(constant.URI_SWITCHOVER_TO_PRIMARY, rpcSwitchoverToPrimaryHandler)
		rpcStandby.DELETE(constant.URI_PAIR, rpcPairDeleteHandler)
	}
}
