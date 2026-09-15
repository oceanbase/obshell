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

package signature

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/cavaliergopher/rpm"
	"github.com/oceanbase/obshell/rpm/trustedkeys"
	"golang.org/x/crypto/openpgp"
)

var (
	trustedKeyringOnce sync.Once
	trustedKeyring     openpgp.KeyRing
	trustedKeyringErr  error
)

// Verify checks that an RPM is signed by one of the public keys built into
// OBShell. It preserves the reader position so callers can add the check to an
// existing parsing flow without changing later header or payload processing.
func Verify(input io.ReadSeeker) (signer string, err error) {
	keyring, err := loadTrustedKeyring()
	if err != nil {
		return "", fmt.Errorf("load trusted RPM public keys: %w", err)
	}
	return verifyWithKeyring(input, keyring)
}

func verifyWithKeyring(input io.ReadSeeker, keyring openpgp.KeyRing) (signer string, err error) {
	originalOffset, err := input.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", fmt.Errorf("get reader position before RPM signature verification: %w", err)
	}
	if _, err = input.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek to start before RPM signature verification: %w", err)
	}

	signer, verifyErr := rpm.GPGCheck(input, keyring)
	if _, err = input.Seek(originalOffset, io.SeekStart); err != nil {
		if verifyErr != nil {
			return "", fmt.Errorf("RPM signature verification failed (%v), and restoring reader position failed: %w", verifyErr, err)
		}
		return "", fmt.Errorf("restore reader position after RPM signature verification: %w", err)
	}
	if verifyErr != nil {
		return "", fmt.Errorf("RPM is not signed by a trusted OceanBase release key: %w", verifyErr)
	}
	return signer, nil
}

func loadTrustedKeyring() (openpgp.KeyRing, error) {
	trustedKeyringOnce.Do(func() {
		keyNames := trustedkeys.Names()
		entities := make(openpgp.EntityList, 0, len(keyNames))
		for _, name := range keyNames {
			publicKey, err := trustedkeys.Files.ReadFile(name)
			if err != nil {
				trustedKeyringErr = fmt.Errorf("read embedded key %q: %w", name, err)
				return
			}
			keyring, err := rpm.ReadKeyRing(bytes.NewReader(publicKey))
			if err != nil {
				trustedKeyringErr = fmt.Errorf("parse embedded key %q: %w", name, err)
				return
			}
			keyEntities, ok := keyring.(openpgp.EntityList)
			if !ok {
				trustedKeyringErr = fmt.Errorf("embedded key %q has unexpected keyring type %T", name, keyring)
				return
			}
			entities = append(entities, keyEntities...)
		}
		trustedKeyring = entities
	})
	return trustedKeyring, trustedKeyringErr
}
