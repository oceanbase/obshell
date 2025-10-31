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

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/crypto"
	"github.com/oceanbase/obshell/seekdb/agent/lib/json"
)

var encryptMethod = "aes"

func BodyDecrypt(body []byte, keys ...string) ([]byte, error) {
	if encryptMethod == "aes" {
		if len(keys) == 0 {
			return nil, errors.Occur(errors.ErrRequestBodyDecryptAesNoKey)
		}
		return bodyDecryptWithAes(string(body), keys[0])
	} else if encryptMethod == "rsa" {
		return bodyDecryptWithRsa(string(body))
	} else if encryptMethod == "sm4" {
		if len(keys) == 0 {
			return nil, errors.Occur(errors.ErrRequestBodyDecryptSm4NoKey)
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
	if encryptMethod == "aes" {
		key_size = crypto.GetAesKeySize()
	}
	if len(keys) < key_size {
		return nil, nil, errors.Occur(errors.ErrRequestBodyDecryptAesKeyAndIvInvalid)
	}
	aesKey = keys[:key_size]
	aesIv = keys[key_size:]
	return
}
