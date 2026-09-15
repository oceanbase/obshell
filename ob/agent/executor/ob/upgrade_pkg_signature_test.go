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

package ob

import (
	"bytes"
	"os"
	"testing"

	"github.com/oceanbase/obshell/ob/agent/constant"
	agenterrors "github.com/oceanbase/obshell/ob/agent/errors"
)

func TestObshellUpgradeSignatureScope(t *testing.T) {
	t.Run("upload rejects unsigned obshell RPM", func(t *testing.T) {
		input := &upgradeMultipartFile{Reader: *bytes.NewReader([]byte("not a signed RPM"))}
		err := verifyObshellUpgradeRpm(input, constant.PKG_OBSHELL)
		if err == nil {
			t.Fatal("expected unsigned OBShell RPM to be rejected")
		}
		agentErr, ok := err.(agenterrors.OcsAgentErrorInterface)
		if !ok {
			t.Fatalf("unsigned OBShell RPM returned an untyped error: %T: %v", err, err)
		}
		if actual := agentErr.ErrorCode().Code; actual != agenterrors.ErrObshellPackageSignatureInvalid.Code {
			t.Fatalf("unsigned OBShell RPM error code is %q, want %q", actual, agenterrors.ErrObshellPackageSignatureInvalid.Code)
		}
	})

	t.Run("upload leaves OceanBase RPM unchanged", func(t *testing.T) {
		if err := verifyObshellUpgradeRpm(nil, constant.PKG_OCEANBASE_CE); err != nil {
			t.Fatalf("OceanBase RPM unexpectedly required a signature: %v", err)
		}
	})

}

func TestVerifyRealObshellUpgradeRpm(t *testing.T) {
	path := os.Getenv("OBSHELL_SIGNED_RPM_TEST_PATH")
	if path == "" {
		t.Skip("OBSHELL_SIGNED_RPM_TEST_PATH is not set")
	}

	input, err := os.Open(path)
	if err != nil {
		t.Fatalf("open OBShell release RPM: %v", err)
	}
	if err := verifyObshellUpgradeRpm(input, constant.PKG_OBSHELL); err != nil {
		if closeErr := input.Close(); closeErr != nil {
			t.Errorf("close OBShell release RPM after verification failure: %v", closeErr)
		}
		t.Fatalf("upload verification rejected signed OBShell RPM: %v", err)
	}
	if err := input.Close(); err != nil {
		t.Fatalf("close OBShell release RPM: %v", err)
	}

}

type upgradeMultipartFile struct {
	bytes.Reader
}

func (*upgradeMultipartFile) Close() error { return nil }
