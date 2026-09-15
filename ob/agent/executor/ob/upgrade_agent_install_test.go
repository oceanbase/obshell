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
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallVerifiedObshellBinary(t *testing.T) {
	const signedBinary = "binary extracted from the signed RPM"
	const tamperedBinary = "binary replaced in the upgrade directory"
	const currentBinary = "currently installed obshell"

	digest := sha256.Sum256([]byte(signedBinary))
	expectedSHA256 := hex.EncodeToString(digest[:])
	testDir := t.TempDir()
	sourcePath := filepath.Join(testDir, "staged-obshell")
	destPath := filepath.Join(testDir, "bin", "obshell")
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		t.Fatalf("create destination directory: %v", err)
	}

	t.Run("installs bytes matching the signed RPM digest", func(t *testing.T) {
		if err := os.WriteFile(sourcePath, []byte(signedBinary), 0600); err != nil {
			t.Fatalf("write signed staging binary: %v", err)
		}
		if err := os.WriteFile(destPath, []byte(currentBinary), 0755); err != nil {
			t.Fatalf("write current OBShell binary: %v", err)
		}
		if err := installVerifiedObshellBinary(sourcePath, destPath, expectedSHA256); err != nil {
			t.Fatalf("install verified OBShell binary: %v", err)
		}
		contents, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("read installed OBShell binary: %v", err)
		}
		if string(contents) != signedBinary {
			t.Fatalf("installed OBShell content is %q, want %q", contents, signedBinary)
		}
	})

	t.Run("rejects a staging binary replaced after RPM extraction", func(t *testing.T) {
		if err := os.WriteFile(sourcePath, []byte(tamperedBinary), 0600); err != nil {
			t.Fatalf("write tampered staging binary: %v", err)
		}
		if err := os.WriteFile(destPath, []byte(currentBinary), 0755); err != nil {
			t.Fatalf("restore current OBShell binary: %v", err)
		}
		if err := installVerifiedObshellBinary(sourcePath, destPath, expectedSHA256); err == nil {
			t.Fatal("expected replaced staging binary to be rejected")
		}
		contents, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("read current OBShell binary after rejection: %v", err)
		}
		if string(contents) != currentBinary {
			t.Fatalf("current OBShell changed after rejection: got %q, want %q", contents, currentBinary)
		}
	})
}
