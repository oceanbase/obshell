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

type ObSysParameter struct {
	Zone         string `gorm:"primaryKey;not null;type:varchar(128)"`
	SvrType      string `gorm:"primaryKey;not null;type:varchar(16)"`
	SvrIp        string `gorm:"primaryKey;not null;type:varchar(46)"`
	SvrPort      int64  `gorm:"primaryKey;not null;"`
	Name         string `gorm:"primaryKey;not null;type:varchar(128)"`
	DataType     string `gorm:"type:varchar(128)"`
	Value        string `gorm:"not null;type:text"`
	Info         string `gorm:"not null;type:varchar(4096)"`
	NeedReboot   int64  `gorm:"not null;"`
	Section      string `gorm:"not null;type:varchar(128)"`
	VisibleLevel string `gorm:"not null;type:varchar(64)"`
	Scope        string `gorm:"not null;type:varchar(64)"`
	Source       string `gorm:"not null;type:varchar(64)"`
	EditLevel    string `gorm:"not null;type:varchar(128)"`
}
