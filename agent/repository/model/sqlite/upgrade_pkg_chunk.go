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

type UpgradePkgChunk struct {
	PkgId      int    `gorm:"primaryKey;column:pkg_id;not null"`
	ChunkId    int    `gorm:"primaryKey;column:chunk_id;not null"`
	ChunkCount int    `gorm:"not null"`
	Chunk      []byte `gorm:"type:MEDIUMBLOB;not null"`
}
