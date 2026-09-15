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

package pkg

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ulikunitz/xz"
)

func TestNewXzSystemReaderUsesOpenedFile(t *testing.T) {
	if _, err := exec.LookPath("xzcat"); err != nil {
		if _, err := exec.LookPath("unxz"); err != nil {
			t.Skip("xzcat/unxz not available")
		}
	}

	const prefix = "signed-rpm-prefix"
	const verifiedPayload = "payload from the verified file descriptor"
	const replacementPayload = "payload from the replaced path"

	rpmPath := filepath.Join(t.TempDir(), "obshell.rpm")
	writeXzFixture(t, rpmPath, prefix, verifiedPayload)

	verifiedFile, err := os.Open(rpmPath)
	if err != nil {
		t.Fatalf("open verified RPM fixture: %v", err)
	}
	defer func() {
		if err := verifiedFile.Close(); err != nil {
			t.Errorf("close verified RPM fixture: %v", err)
		}
	}()

	verifiedPath := rpmPath + ".verified"
	if err := os.Rename(rpmPath, verifiedPath); err != nil {
		t.Fatalf("move verified RPM fixture: %v", err)
	}
	writeXzFixture(t, rpmPath, prefix, replacementPayload)

	reader, cleanup, err := NewXzSystemReader(verifiedFile, int64(len(prefix)))
	if err != nil {
		t.Fatalf("create system xz reader: %v", err)
	}
	defer cleanup()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read decompressed payload: %v", err)
	}
	if string(content) != verifiedPayload {
		t.Fatalf("system xz read payload from replaced path: got %q, want %q", content, verifiedPayload)
	}
}

func writeXzFixture(t *testing.T, path, prefix, payload string) {
	t.Helper()

	var compressed bytes.Buffer
	writer, err := xz.NewWriter(&compressed)
	if err != nil {
		t.Fatalf("create xz writer: %v", err)
	}
	if _, err := writer.Write([]byte(payload)); err != nil {
		t.Fatalf("write xz payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close xz writer: %v", err)
	}

	content := append([]byte(prefix), compressed.Bytes()...)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("write xz fixture: %v", err)
	}
}
