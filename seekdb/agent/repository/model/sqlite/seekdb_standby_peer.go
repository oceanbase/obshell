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

package sqlite

import "time"

// SeekdbStandbyPeer records one directed relationship between two obshell instances.
// Direction=UPSTREAM means the remote peer is this node's replication source (primary).
// Direction=DOWNSTREAM means the remote peer is a replica that feeds from this node.
type SeekdbStandbyPeer struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	PeerHost        string    `gorm:"type:varchar(128);not null"`
	PeerObshellPort int       `gorm:"type:bigint;not null"`
	PeerRpcPort     int       `gorm:"type:bigint;not null"`
	Direction       string    `gorm:"type:varchar(16);not null"` // UPSTREAM / DOWNSTREAM
	PeerToken       string    `gorm:"type:text"`
	PeerPublicKey   string    `gorm:"type:text"` // RSA public key of the peer (base64-encoded PKCS1 DER)
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
