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

package bo

import "time"

type DagInstance struct {
	Id                int64
	Name              string
	Type              string
	Stage             int
	MaxStage          int
	State             int
	ExecuterAgentIp   string
	ExecuterAgentPort int
	IsMaintenance     bool
	MaintenanceType   int
	MaintenanceKey    string
	IsFinished        bool
	Context           []byte
	Operator          int
	StartTime         time.Time
	EndTime           time.Time
	GmtCreate         time.Time
	GmtModify         time.Time
}

type NodeInstance struct {
	Id                int64
	Name              string
	DagId             int64
	DagStage          int
	StructName        string
	Type              string
	State             int
	MaxStage          int
	ExecuterAgentIp   string
	ExecuterAgentPort int
	Context           []byte
	Operator          int
	StartTime         time.Time
	EndTime           time.Time
	GmtCreate         time.Time
	GmtModify         time.Time
}

type SubTaskInstance struct {
	Id                int64
	NodeId            int64
	Name              string
	StructName        string
	ExecuterAgentIp   string
	ExecuterAgentPort int
	ExecuteTimes      int
	CanCancel         bool
	CanContinue       bool
	CanPass           bool
	CanRetry          bool
	CanRollback       bool
	Context           []byte
	State             int
	Operator          int
	StartTime         time.Time
	EndTime           time.Time
	GmtCreate         time.Time
	GmtModify         time.Time
}

type SubTaskLog struct {
	Id           int64
	SubTaskId    int64
	ExecuteTimes int
	LogContent   string
	IsSync       bool
	CreateTime   time.Time
	UpdateTime   time.Time
}
