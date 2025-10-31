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

// select * from information_schema.collations
type ObCollation struct {
	Charset   string `gorm:"column:CHARACTER_SET_NAME"`
	Collation string `gorm:"column:COLLATION_NAME"`
	Id        int    `gorm:"column:ID"`
	IsDefault string `gorm:"column:IS_DEFAULT"`
}

type ObCharset struct {
	Charset          string `gorm:"column:Charset"`
	Description      string `gorm:"column:Description"`
	DefaultCollation string `gorm:"column:Default collation"`
	MaxLen           int64  `gorm:"column:Maxlen"`
}
