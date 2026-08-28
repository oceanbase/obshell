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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/meta"
	sqlitedb "github.com/oceanbase/obshell/ob/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
)

func TestUpdateStandaloneAgentIPPersistsOnlyStandaloneIdentity(t *testing.T) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		t.Fatalf("GetSqliteInstance returned error: %v", err)
	}
	if err := db.AutoMigrate(&sqlite.OcsInfo{}); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}
	var originalIPInfo sqlite.OcsInfo
	result := db.Where("name = ?", constant.OCS_INFO_IP).Find(&originalIPInfo)
	if result.Error != nil {
		t.Fatalf("read existing IP metadata: %v", result.Error)
	}
	hadOriginalIPInfo := result.RowsAffected > 0
	if err := db.Where("name = ?", constant.OCS_INFO_IP).Delete(&sqlite.OcsInfo{}).Error; err != nil {
		t.Fatalf("delete stale IP metadata: %v", err)
	}

	const managementIP = "10.0.0.8"
	if err := db.Create(&sqlite.OcsInfo{Name: constant.OCS_INFO_IP, Value: managementIP}).Error; err != nil {
		t.Fatalf("create initial IP metadata: %v", err)
	}

	agentInstance := meta.NewAgentInstance(managementIP, constant.DEFAULT_AGENT_PORT, "", meta.SINGLE, constant.VERSION_RELEASE)
	originalOcsAgent := ocsAgent
	originalMetaAgent := meta.OCS_AGENT
	ocsAgent = &Agent{AgentInstance: *agentInstance}
	meta.OCS_AGENT = ocsAgent
	t.Cleanup(func() {
		if err := db.Where("name = ?", constant.OCS_INFO_IP).Delete(&sqlite.OcsInfo{}).Error; err != nil {
			t.Errorf("delete test IP metadata: %v", err)
		}
		if hadOriginalIPInfo {
			if err := db.Create(&originalIPInfo).Error; err != nil {
				t.Errorf("restore existing IP metadata: %v", err)
			}
		}
		ocsAgent = originalOcsAgent
		meta.OCS_AGENT = originalMetaAgent
	})

	service := AgentService{}
	if err := service.UpdateAgentIP("10.0.0.9"); err != nil {
		t.Fatalf("UpdateAgentIP returned error: %v", err)
	}
	persistedIP, err := service.GetIP()
	if err != nil {
		t.Fatalf("GetIP after regular update returned error: %v", err)
	}
	if persistedIP != managementIP {
		t.Fatalf("regular update persisted IP %q, expected existing value %q", persistedIP, managementIP)
	}

	if err := service.UpdateStandaloneAgentIP(constant.LOCAL_IP); err != nil {
		t.Fatalf("UpdateStandaloneAgentIP returned error: %v", err)
	}
	persistedIP, err = service.GetIP()
	if err != nil {
		t.Fatalf("GetIP after standalone update returned error: %v", err)
	}
	if persistedIP != constant.LOCAL_IP {
		t.Fatalf("standalone update persisted IP %q, expected %q", persistedIP, constant.LOCAL_IP)
	}
}

func TestStandaloneModeIsPersistedForIndependentRestarts(t *testing.T) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		t.Fatalf("GetSqliteInstance returned error: %v", err)
	}
	if err := db.AutoMigrate(&sqlite.OcsInfo{}); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	var originalStandaloneInfo sqlite.OcsInfo
	result := db.Where("name = ?", constant.OCS_INFO_STANDALONE).Find(&originalStandaloneInfo)
	if result.Error != nil {
		t.Fatalf("read existing standalone metadata: %v", result.Error)
	}
	hadOriginalStandaloneInfo := result.RowsAffected > 0
	if err := db.Where("name = ?", constant.OCS_INFO_STANDALONE).Delete(&sqlite.OcsInfo{}).Error; err != nil {
		t.Fatalf("delete stale standalone metadata: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Where("name = ?", constant.OCS_INFO_STANDALONE).Delete(&sqlite.OcsInfo{}).Error; err != nil {
			t.Errorf("delete test standalone metadata: %v", err)
		}
		if hadOriginalStandaloneInfo {
			if err := db.Create(&originalStandaloneInfo).Error; err != nil {
				t.Errorf("restore existing standalone metadata: %v", err)
			}
		}
	})

	service := AgentService{}
	standalone, err := service.IsStandaloneMode()
	if err != nil {
		t.Fatalf("IsStandaloneMode without marker returned error: %v", err)
	}
	if standalone {
		t.Fatal("IsStandaloneMode returned true without a persisted marker")
	}

	if err := service.EnableStandaloneMode(); err != nil {
		t.Fatalf("EnableStandaloneMode returned error: %v", err)
	}
	standalone, err = service.IsStandaloneMode()
	if err != nil {
		t.Fatalf("IsStandaloneMode after persistence returned error: %v", err)
	}
	if !standalone {
		t.Fatal("IsStandaloneMode returned false after persistence")
	}
}
