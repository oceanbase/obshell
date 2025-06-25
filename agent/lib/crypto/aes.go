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

package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/oceanbase/obshell/agent/errors"
)

var aes_key_size = 16

func SetAesKeySize(size int) {
	aes_key_size = size
}

func GetAesKeySize() int {
	return aes_key_size
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AESEncrypt(raw []byte, key []byte, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	raw = pkcs5Padding(raw, block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(raw))
	mode.CryptBlocks(ciphertext, raw)

	return base64.StdEncoding.EncodeToString(ciphertext), err
}

func AesDecryptAndReturnBytes(raw string, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decRaw, _ := base64.StdEncoding.DecodeString(raw)
	decrypted := make([]byte, len(decRaw))
	mode := cipher.NewCBCDecrypter(block, iv)
	if len(decRaw)%aes.BlockSize != 0 {
		return nil, errors.Occur(errors.ErrRequestBodyDecryptAesContentLengthInvalid)
	}
	mode.CryptBlocks(decrypted, []byte(decRaw))
	decrypted = pkcs5UnPadding(decrypted)
	return decrypted, nil
}
