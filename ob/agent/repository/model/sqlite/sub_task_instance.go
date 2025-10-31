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

	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type SubtaskInstance struct {
	Id                int64     `gorm:"primaryKey;autoIncrement;not null"`
	NodeId            int64     `gorm:"not null; index:idx_node_id"` // The remote node node_id is fixed to 0.
	Name              string    `gorm:"type:varchar(64);not null"`
	StructName        string    `gorm:"type:varchar(128)"`
	ExecuterAgentIp   string    `gorm:"type:varchar(64);not null"`
	ExecuterAgentPort int       `gorm:"type:int;not null"`
	ExecuteTimes      int       `gorm:"type:int;default:0"`
	CanCancel         bool      `gorm:"type:bool;default:false"`
	CanContinue       bool      `gorm:"type:bool;default:false"`
	CanPass           bool      `gorm:"type:bool;default:false"`
	CanRetry          bool      `gorm:"type:bool;default:false"`
	CanRollback       bool      `gorm:"type:bool;default:false"`
	Context           []byte    `gorm:"type:text"`
	State             int       `gorm:"not null"`
	Operator          int       `gorm:"not null"`
	StartTime         time.Time `gorm:"autoCreateTime"`
	EndTime           time.Time `gorm:"autoCreateTime"`
	GmtCreate         time.Time `gorm:"autoCreateTime"`
	GmtModify         time.Time `gorm:"autoUpdateTime"`
}

func (s *SubtaskInstance) ToBO() *bo.SubTaskInstance {
	return &bo.SubTaskInstance{
		Id:                s.Id,
		NodeId:            s.NodeId,
		Name:              s.Name,
		StructName:        s.StructName,
		ExecuterAgentIp:   s.ExecuterAgentIp,
		ExecuterAgentPort: s.ExecuterAgentPort,
		ExecuteTimes:      s.ExecuteTimes,
		CanCancel:         s.CanCancel,
		CanContinue:       s.CanContinue,
		CanPass:           s.CanPass,
		CanRetry:          s.CanRetry,
		CanRollback:       s.CanRollback,
		Context:           s.Context,
		State:             s.State,
		Operator:          s.Operator,
		StartTime:         s.StartTime,
		EndTime:           s.EndTime,
		GmtCreate:         s.GmtCreate,
		GmtModify:         s.GmtModify,
	}
}

func ConvertSubTaskInstanceBOToDO(s *bo.SubTaskInstance) *SubtaskInstance {
	return &SubtaskInstance{
		Id:                s.Id,
		NodeId:            s.NodeId,
		Name:              s.Name,
		StructName:        s.StructName,
		ExecuterAgentIp:   s.ExecuterAgentIp,
		ExecuterAgentPort: s.ExecuterAgentPort,
		ExecuteTimes:      s.ExecuteTimes,
		CanCancel:         s.CanCancel,
		CanContinue:       s.CanContinue,
		CanPass:           s.CanPass,
		CanRetry:          s.CanRetry,
		CanRollback:       s.CanRollback,
		Context:           s.Context,
		State:             s.State,
		Operator:          s.Operator,
		StartTime:         s.StartTime,
		EndTime:           s.EndTime,
		GmtCreate:         s.GmtCreate,
		GmtModify:         s.GmtModify,
	}
}
