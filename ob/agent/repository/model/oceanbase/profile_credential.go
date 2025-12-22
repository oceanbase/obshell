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

type ProfileCredential struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id;type:bigint(20);not null"`
	AccessTarget string    `gorm:"column:access_target;type:varchar(64);not null"`
	Name         string    `gorm:"column:name;type:varchar(64);not null"`
	Secret       string    `gorm:"column:secret;type:varchar(65536);not null"`
	Deleted      bool      `gorm:"column:deleted;type:tinyint(1);not null;default:0"`
	Description  string    `gorm:"column:description;type:varchar(256)"`
	CreateTime   time.Time `gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP"`
	UpdateTime   time.Time `gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

func (ProfileCredential) TableName() string {
	return "profile_credential"
}
