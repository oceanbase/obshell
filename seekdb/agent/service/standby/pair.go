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
	"fmt"

	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/seekdb/param"
)

// UpsertPeer inserts or updates a SeekdbStandbyPeer record. When direction is
// UPSTREAM it also issues ALTER SYSTEM SET log_restore_source so the local
// standby starts replicating from the peer's RPC endpoint.
// All input validation and role checks are the caller's responsibility.
func (s *StandbyService) UpsertPeer(p param.PairParam) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	var existing sqlite.SeekdbStandbyPeer
	result := db.Where("peer_host = ? AND peer_obshell_port = ?", p.PeerHost, p.PeerObshellPort).First(&existing)

	peer := sqlite.SeekdbStandbyPeer{
		PeerHost:        p.PeerHost,
		PeerObshellPort: p.PeerObshellPort,
		PeerRpcPort:     p.PeerRpcPort,
		Direction:       p.Direction,
		PeerToken:       p.Token,
	}

	isUpdate := result.Error == nil
	if isUpdate {
		// Record exists — update it.
		peer.ID = existing.ID
		if err := db.Save(&peer).Error; err != nil {
			return err
		}
	} else {
		// New record: atomically check-and-insert to prevent concurrent UPSTREAM
		// duplicates (count + create must be a single SQLite transaction).
		if err := db.Transaction(func(tx *gorm.DB) error {
			if p.Direction == constant.STANDBY_DIRECTION_UPSTREAM {
				var count int64
				tx.Model(&sqlite.SeekdbStandbyPeer{}).
					Where("direction = ? AND NOT (peer_host = ? AND peer_obshell_port = ?)",
						constant.STANDBY_DIRECTION_UPSTREAM, p.PeerHost, p.PeerObshellPort).
					Count(&count)
				if count > 0 {
					return errors.Occur(errors.ErrStandbyUpstreamPeerAlreadyExists)
				}
			}
			return tx.Create(&peer).Error
		}); err != nil {
			return err
		}
	}

	if p.Direction == constant.STANDBY_DIRECTION_UPSTREAM {
		if err := s.SetLogRestoreSource(p.PeerHost, p.PeerRpcPort); err != nil {
			// Rollback SQLite write to keep metadata consistent with seekdb state.
			if isUpdate {
				db.Save(&existing)
			} else {
				db.Delete(&peer)
			}
			return err
		}
	}
	return nil
}

// DeletePairRecord removes the peer record from SQLite and clears
// log_restore_source if the record is UPSTREAM. It does not perform role
// checks or activate — used by RPC handlers and Activate cleanup.
func (s *StandbyService) DeletePairRecord(p param.PairDeleteParam) (*sqlite.SeekdbStandbyPeer, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}

	var peer sqlite.SeekdbStandbyPeer
	if err = db.Where("peer_host = ? AND peer_obshell_port = ?",
		p.PeerHost, p.PeerObshellPort).First(&peer).Error; err != nil {
		return nil, errors.Occur(errors.ErrStandbyPeerNotFound, p.PeerHost, p.PeerObshellPort)
	}

	if peer.Direction == constant.STANDBY_DIRECTION_UPSTREAM {
		if clearErr := s.ClearLogRestoreSource(); clearErr != nil {
			return nil, clearErr
		}
	}

	if err = db.Delete(&peer).Error; err != nil {
		return nil, err
	}

	return &peer, nil
}

// GetPeers returns all recorded peer relationships.
func (s *StandbyService) GetPeers() ([]sqlite.SeekdbStandbyPeer, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	var peers []sqlite.SeekdbStandbyPeer
	err = db.Find(&peers).Error
	return peers, err
}

// GetPeerByAddr returns the peer record that matches host+port.
func (s *StandbyService) GetPeerByAddr(host string, port int) (*sqlite.SeekdbStandbyPeer, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	var peer sqlite.SeekdbStandbyPeer
	err = db.Where("peer_host = ? AND peer_obshell_port = ?", host, port).First(&peer).Error
	return &peer, err
}

// GetUpstreamPeer returns the UPSTREAM peer, or nil if none.
func (s *StandbyService) GetUpstreamPeer() (*sqlite.SeekdbStandbyPeer, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	var peer sqlite.SeekdbStandbyPeer
	if err = db.Where("direction = ?",
		constant.STANDBY_DIRECTION_UPSTREAM).First(&peer).Error; err != nil {
		return nil, err
	}
	return &peer, nil
}

// FlipDirection swaps a peer's direction to newDirection.
func (s *StandbyService) FlipDirection(host string, port int, newDirection string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Model(&sqlite.SeekdbStandbyPeer{}).
		Where("peer_host = ? AND peer_obshell_port = ?", host, port).
		Update("direction", newDirection).Error
}

// SetLogRestoreSource issues ALTER SYSTEM SET log_restore_source.
func (s *StandbyService) SetLogRestoreSource(host string, rpcPort int) error {
	sql := fmt.Sprintf(
		"ALTER SYSTEM SET log_restore_source = '%s:%d'",
		host, rpcPort)
	return s.execOBSql(sql)
}

// ClearLogRestoreSource removes the replication source configuration.
func (s *StandbyService) ClearLogRestoreSource() error {
	return s.execOBSql("ALTER SYSTEM SET log_restore_source = ''")
}

func (s *StandbyService) execOBSql(sql string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(sql).Error
}

// SwitchoverToStandby executes ALTER SYSTEM SWITCHOVER TO STANDBY.
func (s *StandbyService) SwitchoverToStandby() error {
	return s.execOBSql("ALTER SYSTEM SWITCHOVER TO STANDBY")
}

// SwitchoverToPrimary executes ALTER SYSTEM SWITCHOVER TO PRIMARY.
func (s *StandbyService) SwitchoverToPrimary() error {
	return s.execOBSql("ALTER SYSTEM SWITCHOVER TO PRIMARY")
}

// ActivateStandby executes ALTER SYSTEM ACTIVATE STANDBY.
func (s *StandbyService) ActivateStandby() error {
	return s.execOBSql("ALTER SYSTEM ACTIVATE STANDBY")
}

// UpdatePeerPublicKey caches the peer's RSA public key in SQLite.
func (s *StandbyService) UpdatePeerPublicKey(host string, port int, publicKey string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Model(&sqlite.SeekdbStandbyPeer{}).
		Where("peer_host = ? AND peer_obshell_port = ?", host, port).
		Update("peer_public_key", publicKey).Error
}
