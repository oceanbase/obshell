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

package agent

import (
	"strconv"
	"sync"
	"testing"

	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

func mysqlPortTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(gormsqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&sqlite.ObConfig{}, &sqlite.AllAgent{}); err != nil {
		t.Fatal(err)
	}
	previousAgent, previousPort := meta.OCS_AGENT, meta.GetMysqlPort()
	meta.OCS_AGENT = &meta.AgentInstance{
		AgentInfo: meta.AgentInfo{Ip: "127.0.0.1", Port: 2886},
	}
	meta.SetMysqlPort(2881)
	t.Cleanup(func() {
		meta.OCS_AGENT = previousAgent
		meta.SetMysqlPort(previousPort)
	})
	if err := db.Create(&sqlite.ObConfig{Name: constant.CONFIG_MYSQL_PORT, Value: "2881"}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSyncMysqlPortPersistsBoundPort(t *testing.T) {
	for _, port := range []int{4347, 0} {
		t.Run(strconv.Itoa(port), func(t *testing.T) {
			db := mysqlPortTestDB(t)
			agents := []sqlite.AllAgent{
				{Ip: "127.0.0.1", Port: 2886, MysqlPort: 2881},
				{Ip: "127.0.0.2", Port: 2886, MysqlPort: 2881},
			}
			if err := db.Create(&agents).Error; err != nil {
				t.Fatal(err)
			}
			service := AgentService{}
			if err := service.syncMysqlPort(db, port); err != nil {
				t.Fatal(err)
			}
			var config sqlite.ObConfig
			if err := db.Where("name = ?", constant.CONFIG_MYSQL_PORT).First(&config).Error; err != nil {
				t.Fatal(err)
			}
			if config.Value != strconv.Itoa(port) || meta.GetMysqlPort() != port {
				t.Fatalf("port not persisted: config=%q, meta=%d, want %d", config.Value, meta.GetMysqlPort(), port)
			}
			var ownAgent, otherAgent sqlite.AllAgent
			if err := db.Where("ip = ?", "127.0.0.1").First(&ownAgent).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Where("ip = ?", "127.0.0.2").First(&otherAgent).Error; err != nil {
				t.Fatal(err)
			}
			if ownAgent.MysqlPort != port || otherAgent.MysqlPort != 2881 {
				t.Fatalf("unexpected agent ports: own=%d, other=%d", ownAgent.MysqlPort, otherAgent.MysqlPort)
			}
		})
	}
}

func TestSyncMysqlPortBeforeTakeOver(t *testing.T) {
	db := mysqlPortTestDB(t)
	service := AgentService{}
	if err := service.syncMysqlPort(db, 4347); err != nil {
		t.Fatalf("sync before agent row exists: %v", err)
	}
	if meta.GetMysqlPort() != 4347 {
		t.Fatalf("port=%d, want 4347", meta.GetMysqlPort())
	}
}

func TestSyncMysqlPortFailureKeepsPreviousPort(t *testing.T) {
	db := mysqlPortTestDB(t)
	if err := db.Create(&sqlite.AllAgent{Ip: "127.0.0.1", Port: 2886, MysqlPort: 2881}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER reject_port BEFORE UPDATE OF mysql_port ON all_agent
		BEGIN SELECT RAISE(ABORT, 'reject port update'); END`).Error; err != nil {
		t.Fatal(err)
	}
	service := AgentService{}
	if err := service.syncMysqlPort(db, 4347); err == nil {
		t.Fatal("expected failed agent update")
	}
	var config sqlite.ObConfig
	if err := db.Where("name = ?", constant.CONFIG_MYSQL_PORT).First(&config).Error; err != nil {
		t.Fatal(err)
	}
	if config.Value != "2881" || meta.GetMysqlPort() != 2881 {
		t.Fatalf("failed transaction changed port: config=%q, meta=%d", config.Value, meta.GetMysqlPort())
	}
}

func TestSyncMysqlPortConcurrentReaders(t *testing.T) {
	db := mysqlPortTestDB(t)
	if err := db.Create(&sqlite.AllAgent{Ip: "127.0.0.1", Port: 2886, MysqlPort: 2881}).Error; err != nil {
		t.Fatal(err)
	}
	stop := make(chan struct{})
	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					port := meta.GetMysqlPort()
					if port != 2881 && port != 4347 && port != 4348 && port != 0 {
						t.Errorf("unexpected cached port: %d", port)
						return
					}
				}
			}
		}()
	}
	var writers sync.WaitGroup
	for _, port := range []int{4347, 4348, 0} {
		writers.Add(1)
		go func(port int) {
			defer writers.Done()
			service := AgentService{}
			for i := 0; i < 20; i++ {
				if err := service.syncMysqlPort(db, port); err != nil {
					t.Errorf("sync port %d: %v", port, err)
					return
				}
			}
		}(port)
	}
	writers.Wait()
	close(stop)
	readers.Wait()
	var config sqlite.ObConfig
	if err := db.Where("name = ?", constant.CONFIG_MYSQL_PORT).First(&config).Error; err != nil {
		t.Fatal(err)
	}
	var ownAgent sqlite.AllAgent
	if err := db.Where("ip = ?", "127.0.0.1").First(&ownAgent).Error; err != nil {
		t.Fatal(err)
	}
	if port := meta.GetMysqlPort(); config.Value != strconv.Itoa(port) || ownAgent.MysqlPort != port {
		t.Fatalf("concurrent writes left inconsistent ports: config=%q, agent=%d, cache=%d", config.Value, ownAgent.MysqlPort, port)
	}
}
