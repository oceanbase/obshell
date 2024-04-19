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

//go:build sm4
// +build sm4

package crypto

import (
	"encoding/base64"

	gmssl "github.com/GmSSL/GmSSL-Go"
)

var sm4_key_size = 16

func SetSm4KeySize(size int) {
	sm4_key_size = size
}

func GetSm4KeySize() int {
	return sm4_key_size
}

func Sm4Encrypt(plaintext []byte, key []byte, iv []byte) (string, error) {
	gmssl.GetGmSSLLibraryVersion()
	sm4_cbc, err := gmssl.NewSm4Cbc(key, iv, true)
	if err != nil {
		return "", err
	}
	ciphertext, _ := sm4_cbc.Update(plaintext)
	ciphertext_last, _ := sm4_cbc.Finish()
	ciphertext = append(ciphertext, ciphertext_last...)
	return base64.StdEncoding.EncodeToString(ciphertext), err
}

func Sm4DecryptAndReturnBytes(ciphertext string, key []byte, iv []byte) ([]byte, error) {
	sm4_cbc, err := gmssl.NewSm4Cbc(key, iv, false)
	if err != nil {
		return nil, err
	}
	decCiphertext, _ := base64.StdEncoding.DecodeString(ciphertext)
	decrypted, _ := sm4_cbc.Update(decCiphertext)
	decrypted_last, _ := sm4_cbc.Finish()
	decrypted = append(decrypted, decrypted_last...)
	return decrypted, nil
}
