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

type NodeInstance struct {
	Id                int64     `gorm:"primaryKey;autoIncrement;not null"`
	Name              string    `gorm:"type:varchar(64);not null"`
	DagId             int64     `gorm:"not null; index:idx_dag_stage"`
	DagStage          int       `gorm:"not null; index:idx_dag_stage"`
	StructName        string    `gorm:"type:varchar(128)"`
	Type              string    `gorm:"type:varchar(128);not null"`
	State             int       `gorm:"not null"`
	MaxStage          int       `gorm:"not null"`
	ExecuterAgentIp   string    `gorm:"type:varchar(64);not null"`
	ExecuterAgentPort int       `gorm:"type:int;not null"`
	Context           []byte    `gorm:"type:text"`
	Operator          int       `gorm:"not null"`
	StartTime         time.Time `gorm:"autoCreateTime"`
	EndTime           time.Time `gorm:"autoCreateTime"`
	GmtCreate         time.Time `gorm:"autoCreateTime"`
	GmtModify         time.Time `gorm:"autoUpdateTime"`
}

func (n *NodeInstance) ToBO() *bo.NodeInstance {
	return &bo.NodeInstance{
		Id:                n.Id,
		Name:              n.Name,
		DagId:             n.DagId,
		DagStage:          n.DagStage,
		StructName:        n.StructName,
		Type:              n.Type,
		State:             n.State,
		MaxStage:          n.MaxStage,
		ExecuterAgentIp:   n.ExecuterAgentIp,
		ExecuterAgentPort: n.ExecuterAgentPort,
		Context:           n.Context,
		Operator:          n.Operator,
		StartTime:         n.StartTime,
		EndTime:           n.EndTime,
		GmtCreate:         n.GmtCreate,
		GmtModify:         n.GmtModify,
	}
}

func ConvertNodeInstanceBOToDO(n *bo.NodeInstance) *NodeInstance {
	return &NodeInstance{
		Id:                n.Id,
		Name:              n.Name,
		DagId:             n.DagId,
		DagStage:          n.DagStage,
		StructName:        n.StructName,
		Type:              n.Type,
		State:             n.State,
		MaxStage:          n.MaxStage,
		ExecuterAgentIp:   n.ExecuterAgentIp,
		ExecuterAgentPort: n.ExecuterAgentPort,
		Context:           n.Context,
		Operator:          n.Operator,
		StartTime:         n.StartTime,
		EndTime:           n.EndTime,
		GmtCreate:         n.GmtCreate,
		GmtModify:         n.GmtModify,
	}
}
