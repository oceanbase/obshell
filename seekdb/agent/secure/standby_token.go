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

package secure

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/lib/crypto"
	"github.com/oceanbase/obshell/seekdb/agent/lib/json"
)

// GenerateStandbyToken returns the local node's standby token. When force is
// false the call is idempotent: if a token already exists it is returned as-is.
// When force is true a new token is generated and the old one is overwritten.
func GenerateStandbyToken(force bool) (string, error) {
	if !force {
		existing, err := GetStandbyToken()
		if err == nil && existing != "" {
			return existing, nil
		}
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)

	if err := updateOCSInfo(constant.OCS_INFO_STANDBY_TOKEN, token); err != nil {
		return "", err
	}
	log.Infof("standby token generated (force=%v)", force)
	return token, nil
}

// GetStandbyToken reads the local node's standby token from SQLite.
// Returns ("", nil) when no token has been generated yet.
func GetStandbyToken() (string, error) {
	var token string
	if err := getOCSInfo(constant.OCS_INFO_STANDBY_TOKEN, &token); err != nil {
		return "", nil // not found is not an error
	}
	return token, nil
}

// ValidateStandbyToken compares the provided token against the locally stored
// one using constant-time comparison. Returns false when no local token exists.
func ValidateStandbyToken(token string) bool {
	local, err := GetStandbyToken()
	if err != nil || local == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(local)) == 1
}

// BuildStandbyHeader constructs an encrypted X-OCS-Header value for standby
// RPC calls. uri is the full request URI; token is the local standby token;
// peerPublicKey is the receiver's RSA public key (stored in SQLite alongside
// the peer token during pairing). The returned string should be sent as the
// X-OCS-Header header value so the peer can decrypt it with its private key.
func BuildStandbyHeader(uri string, token string, peerPublicKey string) (string, error) {
	h := HttpHeader{
		StandbyToken: token,
		Uri:          uri,
		Ts:           fmt.Sprintf("%d", time.Now().Add(constant.DEFAULT_AUTH_EXPIRED_DURATION).Unix()),
	}
	data, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	peerCrypter, err := crypto.NewRSACryptoFromPublicKey(peerPublicKey)
	if err != nil {
		return "", err
	}
	return peerCrypter.Encrypt(string(data))
}

// BuildStandbyBodyAndHeader encrypts body with a random AES key, embeds the
// key+IV in HttpHeader.Keys, and RSA-encrypts the whole header with
// peerPublicKey. Returns the encrypted body (nil when body is nil) and the
// X-OCS-Header map so the peer's BodyDecrypt middleware can decrypt the body.
func BuildStandbyBodyAndHeader(uri, token, peerPublicKey string, body interface{}) (encryptedBody interface{}, headers map[string]string, err error) {
	h := HttpHeader{
		StandbyToken: token,
		Uri:          uri,
		Ts:           fmt.Sprintf("%d", time.Now().Add(constant.DEFAULT_AUTH_EXPIRED_DURATION).Unix()),
	}
	if body != nil {
		var key, iv []byte
		encryptedBody, key, iv, err = EncryptBodyWithAes(body)
		if err != nil {
			return nil, nil, err
		}
		h.Keys = append(key, iv...)
	}
	data, err := json.Marshal(h)
	if err != nil {
		return nil, nil, err
	}
	peerCrypter, err := crypto.NewRSACryptoFromPublicKey(peerPublicKey)
	if err != nil {
		return nil, nil, err
	}
	encrypted, err := peerCrypter.Encrypt(string(data))
	if err != nil {
		return nil, nil, err
	}
	return encryptedBody, map[string]string{constant.OCS_HEADER: encrypted}, nil
}
