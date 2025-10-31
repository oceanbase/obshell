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

type TaskMapping struct {
	RemoteTaskId int64 `gorm:"primaryKey; not null;"`
	LocalTaskId  int64
	ExecuteTimes int       `gorm:"type:int;default:0"`
	IsSync       bool      `gorm:"default:false"`
	GmtModify    time.Time `gorm:"type:TIMESTAMP;autoUpdateTime;index:idx_gmt_modify"`
}
