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

package session

import (
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

func Logout(c *gin.Context, sessionID string) error {
	sessionIDInHeader, exist := c.Get(constant.HTTP_HEADER_SESSION_ID)
	if exist || sessionIDInHeader != nil {
		sessionIDInHeader, ok := sessionIDInHeader.(string)
		if !ok {
			return errors.Occur(errors.ErrSecurityAuthenticationSessionInvalid)
		}
		if sessionID == sessionIDInHeader {
			// When the request is verify by session id
			// only if the session id is the same as the session id in the header, will invalidate the session id in the header
			secure.InvalidateSession(sessionIDInHeader)
		}
	} else { // else means is verify by auth instead of session id
		// invalidate the session id in the body
		secure.InvalidateSession(sessionID)
	}
	return nil
}
