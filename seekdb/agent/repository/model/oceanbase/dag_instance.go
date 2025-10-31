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

import (
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
)

type DagInstance struct {
	Id                int64     `gorm:"primaryKey;autoIncrement;not null"`
	Name              string    `gorm:"type:varchar(128);not null"`
	Type              string    `gorm:"type:varchar(128);not null"`
	Stage             int       `gorm:"not null"`
	MaxStage          int       `gorm:"not null"`
	State             int       `gorm:"not null"`
	ExecuterAgentIp   string    `gorm:"type:varchar(64);not null"`
	ExecuterAgentPort int       `gorm:"type:int;not null"`
	IsMaintenance     bool      `gorm:"not null"`
	MaintenanceType   int       `gorm:"not null;default:1"`
	MaintenanceKey    string    `gorm:"type:varchar(128);default:''"`
	IsFinished        bool      `gorm:"not null"`
	Context           []byte    `gorm:"type:text"`
	Operator          int       `gorm:"not null"`
	StartTime         time.Time `gorm:"type:TIMESTAMP(6);default:CURRENT_TIMESTAMP(6)"`
	EndTime           time.Time `gorm:"type:TIMESTAMP(6);default:CURRENT_TIMESTAMP(6)"`
	GmtCreate         time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	GmtModify         time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (d *DagInstance) ToBO() *bo.DagInstance {
	MaintenanceType := d.MaintenanceType
	if d.IsMaintenance && MaintenanceType == task.NOT_UNDER_MAINTENANCE {
		MaintenanceType = task.GLOBAL_MAINTENANCE
	}
	return &bo.DagInstance{
		Id:                d.Id,
		Name:              d.Name,
		Type:              d.Type,
		Stage:             d.Stage,
		MaxStage:          d.MaxStage,
		State:             d.State,
		ExecuterAgentIp:   d.ExecuterAgentIp,
		ExecuterAgentPort: d.ExecuterAgentPort,
		IsMaintenance:     d.IsMaintenance,
		MaintenanceType:   MaintenanceType,
		MaintenanceKey:    d.MaintenanceKey,
		IsFinished:        d.IsFinished,
		Context:           d.Context,
		Operator:          d.Operator,
		StartTime:         d.StartTime,
		EndTime:           d.EndTime,
		GmtCreate:         d.GmtCreate,
		GmtModify:         d.GmtModify,
	}
}

func ConvertDagInstanceBOToDO(d *bo.DagInstance) *DagInstance {
	return &DagInstance{
		Id:                d.Id,
		Name:              d.Name,
		Type:              d.Type,
		Stage:             d.Stage,
		MaxStage:          d.MaxStage,
		State:             d.State,
		ExecuterAgentIp:   d.ExecuterAgentIp,
		ExecuterAgentPort: d.ExecuterAgentPort,
		IsMaintenance:     d.IsMaintenance,
		MaintenanceType:   d.MaintenanceType,
		MaintenanceKey:    d.MaintenanceKey,
		IsFinished:        d.IsFinished,
		Context:           d.Context,
		Operator:          d.Operator,
		StartTime:         d.StartTime,
		EndTime:           d.EndTime,
		GmtCreate:         d.GmtCreate,
		GmtModify:         d.GmtModify,
	}
}
