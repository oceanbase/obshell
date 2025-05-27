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

type UpgradePkgInfo struct {
	PkgId               int       `gorm:"primaryKey;autoIncrement;not null"`
	Name                string    `gorm:"type:varchar(128);not null"`
	Version             string    `gorm:"type:varchar(128);not null"`
	ReleaseDistribution string    `gorm:"type:varchar(128);not null"`
	Distribution        string    `gorm:"type:varchar(128);not null"`
	Release             string    `gorm:"type:varchar(128);not null"`
	Architecture        string    `gorm:"type:varchar(128);not null"`
	Size                uint64    `gorm:"not null"`
	PayloadSize         uint64    `gorm:"not null"`
	ChunkCount          int       `gorm:"not null"`
	Md5                 string    `gorm:"type:varchar(128);not null"`
	GmtModify           time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
}
