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

package path

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasLocalStore(t *testing.T) {
	previousPath := mypath
	mypath = t.TempDir()
	t.Cleanup(func() { mypath = previousPath })
	dataPath := filepath.Join(mypath, "store", "local", "data")
	if got := LocalStoreDataPath(); got != dataPath {
		t.Fatalf("local store path = %q, want %q", got, dataPath)
	}
	if got := ObserverSqlSocketPath(); got != filepath.Join(mypath, "run", "sql.sock") {
		t.Fatalf("SQL socket path = %q", got)
	}
	assertStore := func(want bool) {
		t.Helper()
		got, err := HasLocalStore()
		if err != nil || got != want {
			t.Fatalf("HasLocalStore() = %v, %v, want %v, nil", got, err, want)
		}
	}
	assertStore(false)
	if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
		t.Fatal(err)
	}
	assertStore(false)
	if err := os.WriteFile(dataPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	assertStore(false)
	if err := os.WriteFile(dataPath, []byte("persisted KV data"), 0644); err != nil {
		t.Fatal(err)
	}
	assertStore(true)
	if err := os.Remove(dataPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dataPath, 0755); err != nil {
		t.Fatal(err)
	}
	assertStore(false)
}
