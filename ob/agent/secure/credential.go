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
	"encoding/base64"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/crypto"
	"github.com/oceanbase/obshell/ob/agent/service/config"
)

const credentialAESKeySize = 32

// GetCredentialAESKey retrieves the AES key from ocs_config table
// Returns the raw key bytes after decoding Caesar Base64
func GetCredentialAESKey() ([]byte, error) {
	cfg, err := config.GetOcsConfig(constant.CREDENTIAL_AES_KEY_CONFIG)
	if err != nil {
		return nil, errors.Wrap(err, "get credential AES key from ocs_config failed")
	}
	if cfg == nil || cfg.Value == "" {
		// Key doesn't exist, generate and save it
		return EnsureCredentialAESKey()
	}

	// Decode Caesar Base64
	return decodeCredentialAESKey(cfg.Value)
}

// SaveCredentialAESKey generates a new AES key and saves it to ocs_config
// Key is encoded with Caesar Base64 before storage
func SaveCredentialAESKey() ([]byte, error) {
	// Generate 32-byte random key
	key := make([]byte, credentialAESKeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Wrap(err, "generate credential AES key failed")
	}

	// Encode with Caesar Base64
	encodedKey := crypto.CaesarBase64Encode(string(key), constant.CAESAR_SHIFT)

	// Save to ocs_config
	err = config.SaveOcsConfig(constant.CREDENTIAL_AES_KEY_CONFIG, encodedKey, "AES key for credential passphrase encryption")
	if err != nil {
		return nil, errors.Wrap(err, "save credential AES key to ocs_config failed")
	}

	log.Info("credential AES key generated and saved to ocs_config")
	return key, nil
}

// EnsureCredentialAESKey ensures the key exists, generating if needed
func EnsureCredentialAESKey() ([]byte, error) {
	cfg, err := config.GetOcsConfig(constant.CREDENTIAL_AES_KEY_CONFIG)
	if err != nil {
		return nil, errors.Wrap(err, "get credential AES key from ocs_config failed")
	}
	if cfg == nil || cfg.Value == "" {
		// Key doesn't exist, generate it
		return SaveCredentialAESKey()
	}

	// Decode existing key
	return decodeCredentialAESKey(cfg.Value)
}

// EncryptCredentialPassphrase encrypts a passphrase using the credential AES key
func EncryptCredentialPassphrase(passphrase string) (string, error) {
	key, err := EnsureCredentialAESKey()
	if err != nil {
		return "", errors.Wrap(err, "get credential AES key failed")
	}

	return EncryptCredentialPassphraseWithKey(passphrase, key)
}

// DecryptCredentialPassphrase decrypts an encrypted passphrase
func DecryptCredentialPassphrase(encryptedPassphrase string) (string, error) {
	key, err := GetCredentialAESKey()
	if err != nil {
		return "", errors.Wrap(err, "get credential AES key failed")
	}
	return DecryptCredentialPassphraseWithKey(encryptedPassphrase, key)
}

// EncryptCredentialPassphraseWithKey encrypts passphrase using provided AES key (raw bytes)
func EncryptCredentialPassphraseWithKey(passphrase string, key []byte) (string, error) {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return "", errors.Wrap(err, "generate iv failed")
	}

	encrypted, err := crypto.AESEncrypt([]byte(passphrase), key, iv)
	if err != nil {
		return "", errors.Wrap(err, "encrypt credential passphrase failed")
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", errors.Wrap(err, "decode cipher bytes failed")
	}

	combined := append(iv, cipherBytes...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptCredentialPassphraseWithKey decrypts passphrase using provided AES key (raw bytes)
func DecryptCredentialPassphraseWithKey(encryptedPassphrase string, key []byte) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encryptedPassphrase)
	if err != nil {
		return "", errors.Wrap(err, "decode encrypted passphrase failed")
	}
	if len(raw) < 16 {
		return "", errors.Occur(errors.ErrCredentialDecryptFailed, "encrypted passphrase too short")
	}

	iv := raw[:16]
	cipherBytes := raw[16:]
	cipherBase64 := base64.StdEncoding.EncodeToString(cipherBytes)

	decrypted, err := crypto.AesDecryptAndReturnBytes(cipherBase64, key, iv)
	if err != nil {
		return "", errors.Wrap(err, "decrypt credential passphrase failed")
	}
	return string(decrypted), nil
}

func decodeCredentialAESKey(encoded string) ([]byte, error) {
	keyStr, err := crypto.CaesarBase64Decode(encoded, constant.CAESAR_SHIFT)
	if err != nil {
		return nil, errors.Wrap(err, "decode credential AES key failed")
	}
	key := []byte(keyStr)
	if len(key) != credentialAESKeySize {
		return nil, errors.Occur(errors.ErrCredentialDecryptFailed, "invalid credential AES key length")
	}
	return key, nil
}
