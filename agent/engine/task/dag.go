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

package task

import (
	"time"
)

type Dag struct {
	dagType  string
	stage    int
	maxStage int
	TaskInfo
	maintenance Maintainer
	ctx         *TaskContext
}

func NewDag(dagId int64, dagName string, dagType string, state int, stage int, maxStage int, operator int, maintenance Maintainer, ctx *TaskContext, isLocalTask bool, startTime time.Time, endTime time.Time) *Dag {
	return &Dag{
		dagType:     dagType,
		stage:       stage,
		maxStage:    maxStage,
		maintenance: maintenance,
		ctx:         ctx,
		TaskInfo: TaskInfo{
			id:          dagId,
			name:        dagName,
			state:       state,
			operator:    operator,
			isLocalTask: isLocalTask,
			startTime:   startTime,
			endTime:     endTime,
		},
	}
}

func (dag *Dag) GetDagType() string {
	return dag.dagType
}

func (dag *Dag) GetStage() int {
	return dag.stage
}

func (dag *Dag) GetContext() *TaskContext {
	return dag.ctx
}

func (dag *Dag) MergeContext(ctx *TaskContext) {
	dag.ctx.MergeContext(ctx)
}

func (dag *Dag) GetMaxStage() int {
	return dag.maxStage
}

func (dag *Dag) SetStage(stage int) {
	dag.stage = stage
}

func (dag *Dag) IsMaintenance() bool {
	return dag.maintenance.IsMaintenance()
}

func (dag *Dag) GetMaintenanceType() int {
	return dag.maintenance.GetMaintenanceType()
}

func (dag *Dag) GetMaintenanceKey() string {
	return dag.maintenance.GetMaintenanceKey()
}
