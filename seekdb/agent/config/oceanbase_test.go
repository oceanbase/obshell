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

package config

import (
	"os"
	"path/filepath"
	"testing"

	mysql "github.com/oceanbase/go-oceanbase-driver"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
)

func TestMysqlDataSourceWithLegacyAndLocalStore(t *testing.T) {
	baseDir := t.TempDir()
	previousArgs, previousAgent := os.Args, meta.OCS_AGENT
	os.Args = []string{previousArgs[0], "agent", "start", "--seekdb"}
	meta.OCS_AGENT = &meta.AgentInstance{AgentInfo: meta.AgentInfo{Ip: "127.0.0.1"}}
	t.Setenv(path.ENV_INTERNAL_BASE_DIR, baseDir)
	t.Cleanup(func() { os.Args, meta.OCS_AGENT = previousArgs, previousAgent })

	const username, password = "root", "pass:word@with/slash?mark"
	assertDSN := func(network, address string) {
		t.Helper()
		ds := NewObMysqlDataSourceConfig().SetUsername(username).SetPassword(password).SetPort(13306)
		parsed, err := mysql.ParseDSN(ds.GetDSN())
		if err != nil {
			t.Fatal(err)
		}
		if parsed.Net != network || parsed.Addr != address {
			t.Fatalf("endpoint = %s(%s), want %s(%s)", parsed.Net, parsed.Addr, network, address)
		}
		if parsed.User != username || parsed.Passwd != password || parsed.DBName != "oceanbase" {
			t.Fatalf("DSN lost credentials or database: user=%q database=%q", parsed.User, parsed.DBName)
		}
		if !parsed.ParseTime || parsed.Params["ob_query_timeout"] != "60000000" {
			t.Fatalf("DSN lost connection parameters: %#v", parsed.Params)
		}
	}
	assertDSN("tcp", "127.0.0.1:13306")
	if err := os.MkdirAll(filepath.Dir(path.LocalStoreDataPath()), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.LocalStoreDataPath(), []byte("persisted KV data"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.ObserverSqlSocketPath()); !os.IsNotExist(err) {
		t.Fatalf("test requires absent SQL socket, stat error = %v", err)
	}
	assertDSN("unix", filepath.Join(baseDir, "run", "sql.sock"))
}
