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

package standby

import (
	"strings"
	"testing"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	modelsqlite "github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestValidatePeerDirectionUpdateRejectsMissingPeer(t *testing.T) {
	if err := validatePeerDirectionUpdate(0, "10.0.0.1", 2886); err == nil {
		t.Fatal("expected zero-row peer direction update to fail")
	}
}

func TestValidatePeerDirectionUpdateAcceptsOnePeer(t *testing.T) {
	if err := validatePeerDirectionUpdate(1, "10.0.0.1", 2886); err != nil {
		t.Fatalf("expected one-row peer direction update to succeed: %v", err)
	}
}

func TestValidatePeerDirectionUpdateRejectsDuplicatePeers(t *testing.T) {
	err := validatePeerDirectionUpdate(2, "10.0.0.1", 2886)
	if err == nil {
		t.Fatal("expected multi-row peer direction update to fail")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "updated 2 records") {
		t.Fatalf("expected data consistency error, got %v", err)
	}
}

func TestPeerDirectionUpdateRowsAffectedWithSQLite(t *testing.T) {
	db, err := gorm.Open(gormsqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite database: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close sqlite database: %v", err)
		}
	})
	if err := db.AutoMigrate(&modelsqlite.SeekdbStandbyPeer{}); err != nil {
		t.Fatalf("migrate peer table: %v", err)
	}

	peer := modelsqlite.SeekdbStandbyPeer{
		PeerHost:        "10.0.0.1",
		PeerObshellPort: 2886,
		PeerRpcPort:     2882,
		Direction:       constant.STANDBY_DIRECTION_DOWNSTREAM,
	}
	if err := db.Create(&peer).Error; err != nil {
		t.Fatalf("create peer: %v", err)
	}

	result := db.Model(&modelsqlite.SeekdbStandbyPeer{}).
		Where("peer_host = ? AND peer_obshell_port = ?", peer.PeerHost, peer.PeerObshellPort).
		Update("direction", constant.STANDBY_DIRECTION_DOWNSTREAM)
	if result.Error != nil {
		t.Fatalf("update peer to existing direction: %v", result.Error)
	}
	if err := validatePeerDirectionUpdate(result.RowsAffected, peer.PeerHost, peer.PeerObshellPort); err != nil {
		t.Fatalf("same-value SQLite update should remain retryable: rows=%d err=%v", result.RowsAffected, err)
	}

	duplicate := peer
	duplicate.ID = 0
	if err := db.Create(&duplicate).Error; err != nil {
		t.Fatalf("create duplicate peer: %v", err)
	}
	result = db.Model(&modelsqlite.SeekdbStandbyPeer{}).
		Where("peer_host = ? AND peer_obshell_port = ?", peer.PeerHost, peer.PeerObshellPort).
		Update("direction", constant.STANDBY_DIRECTION_UPSTREAM)
	if result.Error != nil {
		t.Fatalf("update duplicate peers: %v", result.Error)
	}
	if err := validatePeerDirectionUpdate(result.RowsAffected, peer.PeerHost, peer.PeerObshellPort); err == nil || !strings.Contains(err.Error(), "updated 2 records") {
		t.Fatalf("expected duplicate-row consistency error, rows=%d err=%v", result.RowsAffected, err)
	}
}
