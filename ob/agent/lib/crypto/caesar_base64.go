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
	"encoding/base64"
	"strings"
)

const (
	base64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
)

// CaesarBase64Encode encodes a string using Caesar cipher on Base64 characters with shift value
func CaesarBase64Encode(data string, shift int) string {
	// First encode to base64
	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	// Apply Caesar cipher to each character
	var result strings.Builder
	for _, char := range encoded {
		pos := strings.IndexRune(base64Alphabet, char)
		if pos == -1 {
			// If character not in alphabet, keep as is
			result.WriteRune(char)
			continue
		}
		// Apply shift with wrap-around
		newPos := (pos + shift) % len(base64Alphabet)
		result.WriteByte(base64Alphabet[newPos])
	}
	return result.String()
}

// CaesarBase64Decode decodes a Caesar-encoded Base64 string
func CaesarBase64Decode(encoded string, shift int) (string, error) {
	// Reverse Caesar cipher
	var decoded strings.Builder
	for _, char := range encoded {
		pos := strings.IndexRune(base64Alphabet, char)
		if pos == -1 {
			// If character not in alphabet, keep as is
			decoded.WriteRune(char)
			continue
		}
		// Reverse shift with wrap-around
		newPos := (pos - shift + len(base64Alphabet)) % len(base64Alphabet)
		decoded.WriteByte(base64Alphabet[newPos])
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(decoded.String())
	if err != nil {
		return "", err
	}
	return string(data), nil
}
