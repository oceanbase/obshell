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

type InspectionReport struct {
	Id            int       `gorm:"primaryKey;autoIncrement;not null"`
	StartTime     time.Time `gorm:"not null;type:time"`
	FinishTime    time.Time `gorm:"not null;type:time"`
	CriticalCount int       `gorm:"type:int;default:0"`
	FailCount     int       `gorm:"type:int;default:0"`
	WarningCount  int       `gorm:"type:int;default:0"`
	PassCount     int       `gorm:"type:int;default:0"`
	Scenario      string    `gorm:"type:varchar(128);not null"`
	Report        string    `gorm:"type:text;not null"`
	LocalTaskId   string    `gorm:"type:varchar(128);default:''"`
	Status        string    `gorm:"type:varchar(32);default:''"`
	ErrorMessage  string    `gorm:"type:text"`
}
