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
	"testing"

	sqliteDriver "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	agenterrors "github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	modeloceanbase "github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

func TestUpdateAgentVersionRejectsMissingCurrentAgent(t *testing.T) {
	db, err := gorm.Open(sqliteDriver.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	if err := db.AutoMigrate(&modeloceanbase.AllAgent{}); err != nil {
		t.Fatalf("migrate all_agent: %v", err)
	}

	originalAgent := meta.OCS_AGENT
	meta.OCS_AGENT = meta.NewAgentInstance("10.0.0.1", 2886, "zone1", meta.CLUSTER_AGENT, "4.5.1.0-0")
	t.Cleanup(func() {
		meta.OCS_AGENT = originalAgent
	})

	service := AgentService{}
	if err := service.updateAgentVersion(db); err == nil {
		t.Fatal("updateAgentVersion returned nil after updating zero rows")
	} else if agentErr, ok := err.(agenterrors.OcsAgentErrorInterface); !ok || agentErr.ErrorCode().Code != agenterrors.ErrAgentNotExist.Code {
		t.Fatalf("updateAgentVersion error = %v, expected %s", err, agenterrors.ErrAgentNotExist.Code)
	}
}

func TestUpdateAgentVersionUpdatesCurrentAgent(t *testing.T) {
	db, err := gorm.Open(sqliteDriver.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	if err := db.AutoMigrate(&modeloceanbase.AllAgent{}); err != nil {
		t.Fatalf("migrate all_agent: %v", err)
	}

	const (
		agentIP        = "10.0.0.1"
		agentPort      = 2886
		currentVersion = "4.5.1.0-0"
	)
	if err := db.Create(&modeloceanbase.AllAgent{
		Ip:       agentIP,
		Port:     agentPort,
		Identity: string(meta.CLUSTER_AGENT),
		Version:  "4.5.0.0-0",
	}).Error; err != nil {
		t.Fatalf("create all_agent row: %v", err)
	}

	originalAgent := meta.OCS_AGENT
	meta.OCS_AGENT = meta.NewAgentInstance(agentIP, agentPort, "zone1", meta.CLUSTER_AGENT, currentVersion)
	t.Cleanup(func() {
		meta.OCS_AGENT = originalAgent
	})

	service := AgentService{}
	if err := service.updateAgentVersion(db); err != nil {
		t.Fatalf("updateAgentVersion returned error: %v", err)
	}

	var stored modeloceanbase.AllAgent
	if err := db.Where("ip = ? and port = ?", agentIP, agentPort).First(&stored).Error; err != nil {
		t.Fatalf("read all_agent row: %v", err)
	}
	if stored.Version != currentVersion {
		t.Fatalf("stored version = %q, expected %q", stored.Version, currentVersion)
	}
}

func TestUpdateAgentVersionIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqliteDriver.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	if err := db.AutoMigrate(&modeloceanbase.AllAgent{}); err != nil {
		t.Fatalf("migrate all_agent: %v", err)
	}

	const (
		agentIP      = "10.0.0.1"
		agentPort    = 2886
		agentVersion = "4.5.1.0-0"
	)
	if err := db.Create(&modeloceanbase.AllAgent{
		Ip:       agentIP,
		Port:     agentPort,
		Identity: string(meta.CLUSTER_AGENT),
		Version:  agentVersion,
	}).Error; err != nil {
		t.Fatalf("create all_agent row: %v", err)
	}

	originalAgent := meta.OCS_AGENT
	meta.OCS_AGENT = meta.NewAgentInstance(agentIP, agentPort, "zone1", meta.CLUSTER_AGENT, agentVersion)
	t.Cleanup(func() {
		meta.OCS_AGENT = originalAgent
	})

	service := AgentService{}
	const callbackName = "test:force-zero-agent-version-rows-affected"
	if err := db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
		tx.RowsAffected = 0
	}); err != nil {
		t.Fatalf("register zero rows affected callback: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Callback().Update().Remove(callbackName); err != nil {
			t.Errorf("remove zero rows affected callback: %v", err)
		}
	})

	if err := service.updateAgentVersion(db); err != nil {
		t.Fatalf("updateAgentVersion rejected an idempotent zero-row update: %v", err)
	}
}

func TestConfirmAgentVersionAcceptsConcurrentSameVersionUpdate(t *testing.T) {
	db, err := gorm.Open(sqliteDriver.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	if err := db.AutoMigrate(&modeloceanbase.AllAgent{}); err != nil {
		t.Fatalf("migrate all_agent: %v", err)
	}

	const (
		agentIP        = "10.0.0.1"
		agentPort      = 2886
		currentVersion = "4.5.1.0-0"
	)
	if err := db.Create(&modeloceanbase.AllAgent{
		Ip:       agentIP,
		Port:     agentPort,
		Identity: string(meta.CLUSTER_AGENT),
		Version:  currentVersion,
	}).Error; err != nil {
		t.Fatalf("create all_agent row: %v", err)
	}

	originalAgent := meta.OCS_AGENT
	meta.OCS_AGENT = meta.NewAgentInstance(agentIP, agentPort, "zone1", meta.CLUSTER_AGENT, currentVersion)
	t.Cleanup(func() {
		meta.OCS_AGENT = originalAgent
	})

	service := AgentService{}
	if err := service.confirmAgentVersion(db, currentVersion, meta.OCS_AGENT.String()); err != nil {
		t.Fatalf("confirmAgentVersion rejected a concurrently updated row: %v", err)
	}
}
