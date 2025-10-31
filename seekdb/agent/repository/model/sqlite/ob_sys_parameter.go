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
	Name         string `gorm:"column:NAME;primaryKey;not null;type:varchar(128)"`
	DataType     string `gorm:"column:DATA_TYPE;not null;type:varchar(128)"`
	Value        string `gorm:"column:VALUE;not null;type:text"`
	Info         string `gorm:"column:INFO;not null;type:varchar(4096)"`
	Section      string `gorm:"column:SECTION;not null;type:varchar(128)"`
	EditLevel    string `gorm:"column:EDIT_LEVEL;not null;type:varchar(64)"`
	DefaultValue string `gorm:"column:DEFAULT_VALUE;not null;type:text"`
	IsDefault    bool   `gorm:"column:IS_DEFAULT;not null;default:false"`
}
