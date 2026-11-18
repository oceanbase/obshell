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

package oceanbase

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
)

func TestFillConfigPortWithLegacyAndLocalStore(t *testing.T) {
	previousArgs, previousAgent, previousPort := os.Args, meta.OCS_AGENT, meta.GetMysqlPort()
	os.Args = []string{previousArgs[0], "agent", "start", "--seekdb"}
	meta.OCS_AGENT = &meta.AgentInstance{AgentInfo: meta.AgentInfo{Ip: "127.0.0.1"}}
	t.Setenv(path.ENV_INTERNAL_BASE_DIR, t.TempDir())
	t.Cleanup(func() {
		os.Args, meta.OCS_AGENT = previousArgs, previousAgent
		meta.SetMysqlPort(previousPort)
	})
	meta.SetMysqlPort(0)
	if err := fillConfigPort(config.NewObMysqlDataSourceConfig()); err == nil {
		t.Fatal("legacy TCP configuration accepted an unknown port")
	}
	explicit := config.NewObMysqlDataSourceConfig().SetPort(13306)
	if err := fillConfigPort(explicit); err != nil || explicit.GetPort() != 13306 {
		t.Fatalf("explicit TCP port changed: port=%d error=%v", explicit.GetPort(), err)
	}
	meta.SetMysqlPort(13307)
	cached := config.NewObMysqlDataSourceConfig()
	if err := fillConfigPort(cached); err != nil || cached.GetPort() != 13307 {
		t.Fatalf("cached TCP port not used: port=%d error=%v", cached.GetPort(), err)
	}
	if err := os.MkdirAll(filepath.Dir(path.LocalStoreDataPath()), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.LocalStoreDataPath(), []byte("persisted KV data"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path.LocalStoreDataPath()) })
	meta.SetMysqlPort(0)
	socket := config.NewObMysqlDataSourceConfig()
	if socket.GetSocket() == "" {
		t.Fatal("new local store did not select a SQL socket")
	}
	if err := fillConfigPort(socket); err != nil || socket.GetPort() != 0 {
		t.Fatalf("SQL socket required a TCP port: port=%d error=%v", socket.GetPort(), err)
	}
}

func TestFillConfigPortConcurrentUpdates(t *testing.T) {
	previousPort := meta.GetMysqlPort()
	t.Cleanup(func() { meta.SetMysqlPort(previousPort) })
	meta.SetMysqlPort(4347)
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 10000; i++ {
			meta.SetMysqlPort(0)
			meta.SetMysqlPort(4347)
		}
	}()
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for j := 0; j < 10000; j++ {
				dsConfig := &config.ObDataSourceConfig{}
				if err := fillConfigPort(dsConfig); err == nil && dsConfig.GetPort() != 4347 {
					t.Errorf("TCP configuration accepted an unknown port: %d", dsConfig.GetPort())
					return
				}
			}
		}()
	}
	close(start)
	workers.Wait()
}
