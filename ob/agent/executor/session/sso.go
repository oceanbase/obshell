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

// CreateSSOToken creates a one-time SSO token. Caller must have passed Verify(ROUTE_LOGIN).
func CreateSSOToken(c *gin.Context) (token string, err error) {
	key, exist := c.Get(constant.HTTP_HEADER_AES_KEY)
	if !exist || key == nil {
		return "", errors.Occur(errors.ErrRequestAESKeyNotFound)
	}
	keyBytes, ok := key.([]byte)
	if !ok {
		return "", errors.Occur(errors.ErrRequestAESKeyNotFound)
	}
	// only support aes currently, other algorithms will be supported in the future
	token, err = secure.CreateSSOToken()
	if err != nil {
		return "", err
	}
	encryptedToken, err := secure.EncryptWithAesKeys([]byte(token), string(keyBytes))
	if err != nil {
		return "", err
	}

	return encryptedToken, nil
}

// ExchangeSSOToken consumes the token from context, creates a session, encrypts session_id with request Keys and returns it.
func ExchangeSSOToken(c *gin.Context) (encryptedSessionID string, err error) {
	tokenVal, _ := c.Get(constant.HTTP_HEADER_SSO_TOKEN)
	ssoToken, _ := tokenVal.(string)
	if ssoToken == "" {
		return "", errors.Occur(errors.ErrSecuritySSOTokenMissing)
	}
	keyVal, _ := c.Get(constant.HTTP_HEADER_AES_KEY)
	keyBytes, ok := keyVal.([]byte)
	if !ok || len(keyBytes) < 32 {
		return "", errors.Occur(errors.ErrRequestAESKeyNotFound)
	}
	if err = secure.ExchangeSSOToken(ssoToken); err != nil {
		return "", err
	}
	sessionID, err := secure.CreateSession()
	if err != nil {
		return "", err
	}
	encryptedSessionID, err = secure.EncryptWithAesKeys([]byte(sessionID), string(keyBytes))
	if err != nil {
		return "", err
	}
	return encryptedSessionID, nil
}
