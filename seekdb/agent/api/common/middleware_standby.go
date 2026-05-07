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

package common

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
)

// verifyStandbyRouters is the RouteVerifier for /rpc/v1/ standby routes.
// It validates the replay-protection timestamp and the StandbyToken.
// The URI check is handled by the VerifyWith pipeline.
func verifyStandbyRouters(c *gin.Context, curTs int64, header *secure.HttpHeader) {
	if err := secure.VerifyTimeStamp(header.Ts, curTs); err != nil {
		log.WithContext(NewContextWithTraceId(c)).Warnf("standby token timestamp verification failed: %v", err)
		c.Abort()
		SendResponse(c, nil, err)
		return
	}
	if header.StandbyToken == "" {
		log.WithContext(NewContextWithTraceId(c)).Warn("standby token missing in RPC request")
		c.Abort()
		SendResponse(c, nil, errors.Occur(errors.ErrStandbyTokenMissing))
		return
	}
	if !secure.ValidateStandbyToken(header.StandbyToken) {
		log.WithContext(NewContextWithTraceId(c)).Warn("invalid standby token in RPC request")
		c.Abort()
		SendResponse(c, nil, errors.Occur(errors.ErrStandbyTokenInvalid))
		return
	}
}

// VerifyStandbyToken is the authentication middleware for /rpc/v1/ standby routes.
// It reuses the VerifyWith pipeline (header decode, URI check, timestamp) and
// validates the StandbyToken credential instead of the OB root-password auth.
func VerifyStandbyToken() func(*gin.Context) {
	return VerifyWith(verifyStandbyRouters)
}
