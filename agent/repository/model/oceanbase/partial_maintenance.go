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

type PartialMaintenance struct {
	Id        int64     `gorm:"primaryKey;autoIncrement;not null"`
	LockType  int       `gorm:"not null; index:idx_lock_type_name,unique"`
	LockName  string    `gorm:"not null; type:varchar(64);index:idx_lock_type_name,unique"`
	DagID     int64     `gorm:"not null"`
	Count     int       `gorm:"default:0"`
	GmtLocked time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	GmtCreate time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	GmtModify time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}
