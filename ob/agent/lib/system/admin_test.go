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

package system

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/oceanbase/obshell/ob/agent/global"
	obpath "github.com/oceanbase/obshell/ob/agent/lib/path"
)

func TestGetOBAdminCtxByURIDoesNotExecuteShell(t *testing.T) {
	tests := []struct {
		name     string
		buildURI func(marker string) string
	}{
		{
			name: "file path",
			buildURI: func(marker string) string {
				return fmt.Sprintf("file:///tmp/backup'; touch %s; #", marker)
			},
		},
		{
			name: "encoded OSS parameter",
			buildURI: func(marker string) string {
				return buildObjectStorageInjectionURI("oss", marker)
			},
		},
		{
			name: "encoded S3 parameter",
			buildURI: func(marker string) string {
				return buildObjectStorageInjectionURI("s3", marker)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marker := filepath.Join(t.TempDir(), "shell-injection-marker")
			uri := tt.buildURI(marker)

			output, commandErr := getOBAdminCtxByURI(uri)
			if commandErr != nil {
				t.Logf("ob_admin returned an expected error in the test environment: %v (output: %q)", commandErr, output)
			}

			assertFileDoesNotExist(t, marker)
		})
	}
}

func TestGetRestoreSourceTenantInfoDoesNotExecuteShell(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "shell-injection-marker")
	uri := fmt.Sprintf("file:///tmp/backup'; touch %s; #", marker)

	result, commandErr := GetRestoreSourceTenantInfo(uri, uri)
	if commandErr != nil {
		t.Logf("ob_admin returned an expected error in the test environment: %v (result: %#v)", commandErr, result)
	}

	assertFileDoesNotExist(t, marker)
}

func TestNewOBAdminDumpBackupCommandPreservesArgumentsAndEnvironment(t *testing.T) {
	originalHomePath := global.HomePath
	global.HomePath = "/opt/oceanbase"
	t.Cleanup(func() {
		global.HomePath = originalHomePath
	})

	storageURI := "file:///backup path/with'quote;$(command)"
	storageParams := "host=http://example.invalid&access_id=test&access_key=aA+/'\";$()"
	cmd := newOBAdminDumpBackupCommand(storageURI, storageParams)
	wantArgs := []string{obpath.OBAdmin(), "dump_backup", "-q", "-d", storageURI, "-s", storageParams}
	if !reflect.DeepEqual(cmd.Args, wantArgs) {
		t.Fatalf("unexpected ob_admin arguments:\nwant: %#v\n got: %#v", wantArgs, cmd.Args)
	}

	wantLibraryPath := "LD_LIBRARY_PATH=/opt/oceanbase/lib"
	if len(cmd.Env) == 0 || cmd.Env[len(cmd.Env)-1] != wantLibraryPath {
		t.Fatalf("LD_LIBRARY_PATH must be set without a shell, want final environment entry %q", wantLibraryPath)
	}
}

func buildObjectStorageInjectionURI(scheme, marker string) string {
	host := fmt.Sprintf("http://example.invalid'; touch %s; #", marker)
	encodedHost := strings.ReplaceAll(url.QueryEscape(host), "+", "%20")
	return fmt.Sprintf("%s://bucket/backup?host=%s&access_id=test-id&access_key=test-key", scheme, encodedHost)
}

func assertFileDoesNotExist(t *testing.T, filename string) {
	t.Helper()
	if _, err := os.Stat(filename); err == nil {
		t.Fatalf("storage URI was interpreted as shell syntax and created %s", filename)
	} else if !os.IsNotExist(err) {
		t.Fatalf("check shell injection marker: %v", err)
	}
}
