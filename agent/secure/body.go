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
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/lib/crypto"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/meta"
)

const (
	EncryptMethodAes = "aes"
	EncryptMethodRsa = "rsa"
	EncryptMethodSm4 = "sm4"
)

var encryptMethod = EncryptMethodAes

func BodyDecrypt(body []byte, keys ...string) ([]byte, error) {
	if encryptMethod == EncryptMethodAes {
		if len(keys) == 0 {
			return nil, errors.New("no key for aes")
		}
		return bodyDecryptWithAes(string(body), keys[0])
	} else if encryptMethod == EncryptMethodRsa {
		return bodyDecryptWithRsa(string(body))
	} else if encryptMethod == EncryptMethodSm4 {
		if len(keys) == 0 {
			return nil, errors.New("no key for sm4")
		}
		return bodyDecryptWithSm4(string(body), keys[0])
	}
	return body, nil
}

func EncryptBodyWithSm4(body interface{}) (encryptedBody interface{}, key []byte, iv []byte, err error) {
	if body == nil {
		return
	}
	mBody, err := json.Marshal(body)
	if err != nil {
		log.WithError(err).Error("json marshal failed")
		return
	}
	key = make([]byte, 16)
	iv = make([]byte, 16)
	_, err = rand.Read(key)
	if err != nil {
		return
	}
	_, err = rand.Read(iv)
	if err != nil {
		return
	}
	encryptedBody, err = crypto.Sm4Encrypt(mBody, key, iv)
	return
}

func EncryptBodyWithRsa(agentInfo meta.AgentInfoInterface, body interface{}) (encryptedBody interface{}, err error) {
	if body == nil {
		return nil, nil
	}
	mBody, err := json.Marshal(body)
	if err != nil {
		log.WithError(err).Error("json marshal failed")
		return
	}
	pk := GetAgentPublicKey(agentInfo)
	if pk == "" {
		log.Warnf("no key for agent '%s'", agentInfo.String())
		return
	}
	encryptedBody, err = crypto.RSAEncrypt(mBody, pk)
	return
}

func EncryptBodyWithAes(body interface{}) (encryptedBody interface{}, key []byte, iv []byte, err error) {
	if body == nil {
		return
	}
	mBody, err := json.Marshal(body)
	if err != nil {
		log.WithError(err).Error("json marshal failed")
		return
	}
	key = make([]byte, crypto.GetAesKeySize())
	iv = make([]byte, 16) // Equal to block_size，16 bytes.
	_, err = rand.Read(key)
	if err != nil {
		return
	}
	_, err = rand.Read(iv)
	if err != nil {
		return
	}
	encryptedBody, err = crypto.AESEncrypt(mBody, key, iv)
	return
}

func bodyDecryptWithRsa(ciphertext string) ([]byte, error) {
	plaintext, err := Crypter.DecryptAndReturnBytes(ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func bodyDecryptWithAes(ciphertext string, keys string) ([]byte, error) {
	key, iv, err := transferKeys(keys)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.AesDecryptAndReturnBytes(ciphertext, key, iv)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func bodyDecryptWithSm4(ciphertext string, keys string) ([]byte, error) {
	key, iv, err := transferKeys(keys)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.Sm4DecryptAndReturnBytes(ciphertext, key, iv)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func transferKeys(Keys string) (aesKey []byte, aesIv []byte, err error) {
	keys := []byte(Keys)
	key_size := 16
	if encryptMethod == EncryptMethodAes {
		key_size = crypto.GetAesKeySize()
	}
	if len(keys) < key_size {
		return nil, nil, errors.New("aes key and iv size error")
	}
	aesKey = keys[:key_size]
	aesIv = keys[key_size:]
	return
}
