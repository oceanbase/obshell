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
	"encoding/base64"
	"os"
	"testing"

	"golang.org/x/crypto/openpgp"
)

func TestLoadTrustedKeyring(t *testing.T) {
	keyring, err := loadTrustedKeyring()
	if err != nil {
		t.Fatalf("load trusted keyring: %v", err)
	}

	keys := keyring.KeysByIdUsage(0x283BCA0963476B06, 0)
	if len(keys) != 1 {
		t.Fatalf("expected one OceanBase EL7 key, got %d", len(keys))
	}

	if keys := keyring.KeysByIdUsage(0x371265DA09E1CEBF, 0); len(keys) != 1 {
		t.Fatalf("look up current OceanBase key: keys=%d", len(keys))
	}
	if keys := keyring.KeysByIdUsage(0x2FF845A6E9B4A7AA, 0); len(keys) != 1 {
		t.Fatalf("look up legacy OceanBase key: keys=%d", len(keys))
	}
}

func TestVerifyRejectsUntrustedSigner(t *testing.T) {
	fixture, err := os.ReadFile("../../ob/agent/lib/pkg/testdata/oceanbase-ce-libs-4.5.0.0.el7.x86_64.rpm.b64")
	if err != nil {
		t.Fatalf("read signed RPM fixture: %v", err)
	}
	signedRPM, err := base64.StdEncoding.DecodeString(string(fixture))
	if err != nil {
		t.Fatalf("decode signed RPM fixture: %v", err)
	}

	if _, err := verifyWithKeyring(bytes.NewReader(signedRPM), openpgp.EntityList{}); err == nil {
		t.Fatal("expected package signed by an untrusted key to be rejected")
	}
}
