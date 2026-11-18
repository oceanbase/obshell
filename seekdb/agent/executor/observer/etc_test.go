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

package observer

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
)

func TestSeekdbPersistentConfiguration(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "base with & percent%")
	if err := os.Mkdir(baseDir, 0755); err != nil {
		t.Fatal(err)
	}
	previousArgs, previousPort := os.Args, meta.GetMysqlPort()
	os.Args = []string{previousArgs[0], "agent", "start", "--seekdb"}
	t.Setenv(path.ENV_INTERNAL_BASE_DIR, baseDir)
	t.Cleanup(func() {
		os.Args = previousArgs
		meta.SetMysqlPort(previousPort)
	})
	if err := os.MkdirAll(filepath.Dir(path.MetaDbPath()), 0755); err != nil {
		t.Fatal(err)
	}

	assertStarted := func(want bool) {
		t.Helper()
		got, err := HasStarted()
		if err != nil || got != want {
			t.Fatalf("HasStarted() = %v, %v, want %v, nil", got, err, want)
		}
	}
	assertStarted(false)
	if value, ok := getParamFromMetaDb(ETC_KEY_MYSQL_PORT); ok || value != "" {
		t.Fatalf("missing metadata returned %q, %v", value, ok)
	}
	if _, err := os.Stat(path.MetaDbPath()); !os.IsNotExist(err) {
		t.Fatalf("reading absent metadata created meta.db: %v", err)
	}
	meta.SetMysqlPort(13306)
	if err := LoadOBConfigFromConfigFile(); err != nil {
		t.Fatal(err)
	}
	if meta.GetMysqlPort() != 13306 {
		t.Fatalf("cached port changed to %d", meta.GetMysqlPort())
	}

	if err := os.MkdirAll(filepath.Dir(path.LocalStoreDataPath()), 0755); err != nil {
		t.Fatal(err)
	}
	assertStarted(false)
	if err := os.WriteFile(path.LocalStoreDataPath(), nil, 0644); err != nil {
		t.Fatal(err)
	}
	assertStarted(false)
	if err := os.WriteFile(path.LocalStoreDataPath(), []byte("persisted KV data"), 0644); err != nil {
		t.Fatal(err)
	}
	assertStarted(true)
	if err := LoadOBConfigFromConfigFile(); err != nil {
		t.Fatal(err)
	}
	if meta.GetMysqlPort() != 13306 {
		t.Fatalf("local store reset cached port to %d", meta.GetMysqlPort())
	}
	if _, err := os.Stat(path.MetaDbPath()); !os.IsNotExist(err) {
		t.Fatalf("local store access created legacy meta.db: %v", err)
	}
	if err := os.Remove(path.LocalStoreDataPath()); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", path.MetaDbPath())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE __all_sys_parameter (name TEXT PRIMARY KEY, value TEXT)")
	if err == nil {
		_, err = db.Exec("INSERT INTO __all_sys_parameter (name, value) VALUES (?, ?), (?, ?)", ETC_KEY_MYSQL_PORT, "13307", ETC_KEY_USE_IPV6, "true")
	}
	closeErr := db.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if err := os.Chmod(path.MetaDbPath(), 0444); err != nil {
		t.Fatal(err)
	}
	assertStarted(true)
	ip, port := GetConfFromObMeta()
	if ip != "::1" || port != 13307 {
		t.Fatalf("legacy metadata = %q:%d, want [::1]:13307", ip, port)
	}
	if value, ok := getParamFromMetaDb("missing"); ok || value != "" {
		t.Fatalf("unknown parameter returned %q, %v", value, ok)
	}
}
