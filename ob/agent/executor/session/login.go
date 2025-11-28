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
	log "github.com/sirupsen/logrus"
)

// Login returns encrypted session id by aes algorithm
func Login(c *gin.Context) (string, error) {
	key, exist := c.Get(constant.HTTP_HEADER_AES_KEY)
	if !exist || key == nil {
		return "", errors.Occur(errors.ErrRequestAESKeyNotFound)
	}
	keyBytes, ok := key.([]byte)
	if !ok {
		return "", errors.Occur(errors.ErrRequestAESKeyNotFound)
	}
	// only support aes currently, other algorithms will be supported in the future
	sessionID, err := secure.CreateSession()
	if err != nil {
		return "", err
	}
	log.Infof("create session %s", sessionID)
	encryptedBody, err := secure.EncryptWithAesKeys([]byte(sessionID), string(keyBytes))
	if err != nil {
		return "", err
	}
	return encryptedBody, nil
}
