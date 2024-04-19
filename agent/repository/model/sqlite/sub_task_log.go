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

import (
	"time"

	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

type SubTaskLog struct {
	Id           int64     `gorm:"primaryKey;autoIncrement;not null"`
	SubTaskId    int64     `gorm:"not null"`
	ExecuteTimes int       `gorm:"not null"`
	LogContent   string    `gorm:"type:text"`
	IsSync       bool      `gorm:"type:bool;default:false"`
	CreateTime   time.Time `gorm:"autoCreateTime"`
	UpdateTime   time.Time `gorm:"autoUpdateTime"`
}

func (s *SubTaskLog) ToBO() *bo.SubTaskLog {
	return &bo.SubTaskLog{
		Id:           s.Id,
		SubTaskId:    s.SubTaskId,
		ExecuteTimes: s.ExecuteTimes,
		LogContent:   s.LogContent,
		IsSync:       s.IsSync,
		CreateTime:   s.CreateTime,
		UpdateTime:   s.UpdateTime,
	}
}

func ConvertSubTaskLogBOToDO(s *bo.SubTaskLog) *SubTaskLog {
	return &SubTaskLog{
		Id:           s.Id,
		SubTaskId:    s.SubTaskId,
		ExecuteTimes: s.ExecuteTimes,
		LogContent:   s.LogContent,
		IsSync:       s.IsSync,
		CreateTime:   s.CreateTime,
		UpdateTime:   s.UpdateTime,
	}
}
