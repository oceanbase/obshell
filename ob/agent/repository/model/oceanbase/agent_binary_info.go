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

import "time"

type AgentBinaryInfo struct {
	BinId        int       `gorm:"primaryKey;autoIncrement;not null"`
	Version      string    `gorm:"type:varchar(128);not null; index:ver_arch_dest,unique"`
	Architecture string    `gorm:"type:varchar(128);not null; index:ver_arch_dest,unique"`
	Distribution string    `gorm:"type:varchar(128);not null; index:ver_arch_dest,unique"`
	ChunkCount   int       `gorm:"not null"` // if chunk_count = 0, means the binary is not uploaded yet
	GmtCreate    time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	GmtModify    time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}
