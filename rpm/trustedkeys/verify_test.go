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

package trustedkeys

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/cavaliergopher/rpm"
	"golang.org/x/crypto/openpgp"
)

func TestEmbeddedKeysMatchReviewedFingerprints(t *testing.T) {
	expectedKeys := map[string]struct {
		digest      string
		fingerprint string
	}{
		"RPM-GPG-KEY-OceanBase": {
			digest:      "c0a9f872be9a99a5f28ea9fa3332df0bb345e9abfcec53b0ff054fa57f6bb8d3",
			fingerprint: "a109e086b8f1d712a2555cd4371265da09e1cebf",
		},
		"RPM-GPG-KEY-OceanBase-el7": {
			digest:      "a8f7ec5059e8ef476d41cc7a1cb4f40ec1d35afa50fff67250104b6d576b2586",
			fingerprint: "a74fe069a852a639a2412347283bca0963476b06",
		},
		"RPM-GPG-KEY-OceanBase-old": {
			digest:      "dff33e8dcee747a23c285eaaea1d7dea14175000b726e9452bd90c431c21aafe",
			fingerprint: "ef7de8e36987b60cacf99a532ff845a6e9b4a7aa",
		},
	}
	keyNames := Names()
	if len(keyNames) != len(expectedKeys) {
		t.Fatalf("expected %d trusted keys, got %d", len(expectedKeys), len(keyNames))
	}

	for _, name := range keyNames {
		expected, ok := expectedKeys[name]
		if !ok {
			t.Fatalf("trusted key %q has not been reviewed", name)
		}
		publicKey, err := Files.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded key %q: %v", name, err)
		}

		digest := sha256.Sum256(publicKey)
		if actual := hex.EncodeToString(digest[:]); actual != expected.digest {
			t.Fatalf("key %q digest changed: got %s", name, actual)
		}

		keyring, err := rpm.ReadKeyRing(bytes.NewReader(publicKey))
		if err != nil {
			t.Fatalf("parse embedded key %q: %v", name, err)
		}
		entities := keyring.(openpgp.EntityList)
		if len(entities) != 1 {
			t.Fatalf("key %q contains %d entities", name, len(entities))
		}
		actualFingerprint := hex.EncodeToString(entities[0].PrimaryKey.Fingerprint[:])
		if actualFingerprint != expected.fingerprint {
			t.Fatalf("key %q fingerprint changed: got %s", name, actualFingerprint)
		}
	}
}
