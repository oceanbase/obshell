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

//go:build !sm4
// +build !sm4

package crypto

func SetSm4KeySize(size int) {
}

func GetSm4KeySize() int {
	return 0
}

func Sm4Encrypt(plaintext []byte, key []byte, iv []byte) (string, error) {
	return "", nil
}

func Sm4DecryptAndReturnBytes(ciphertext string, key []byte, iv []byte) ([]byte, error) {
	return nil, nil
}
