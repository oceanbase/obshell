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
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	bo "github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

// convertDagInstance converts DagInstance to task.Dag.
func (s *taskService) convertDagInstance(bo *bo.DagInstance) (*task.Dag, error) {
	var ctx task.TaskContext
	if err := json.Unmarshal(bo.Context, &ctx); err != nil {
		return nil, err
	}

	maintenance := task.NewMaintenance(bo.MaintenanceType, bo.MaintenanceKey)
	return task.NewDag(bo.Id, bo.Name, bo.Type, bo.State, bo.Stage, bo.MaxStage, bo.Operator, maintenance, &ctx, s.isLocal, bo.StartTime, bo.EndTime), nil
}

// convertNodeInstance converts NodeInstance to task.TaskNode.
func (s *taskService) convertNodeInstance(bo *bo.NodeInstance) (*task.Node, error) {
	var ctx task.TaskContext
	if err := json.Unmarshal(bo.Context, &ctx); err != nil {
		return nil, err
	}
	return task.NewNodeWithId(bo.Id, bo.Name, bo.State, bo.Operator, bo.StructName, &ctx, s.isLocal, bo.StartTime, bo.EndTime), nil
}

// convertSubTaskInstance convert SubTaskInstance to task.ExecutableTask.
func (s *taskService) convertSubTaskInstance(bo *bo.SubTaskInstance) (task.ExecutableTask, error) {
	var ctx task.TaskContext
	if err := json.Unmarshal(bo.Context, &ctx); err != nil {
		return nil, err
	}

	agentInfo := meta.NewAgentInfo(bo.ExecuterAgentIp, bo.ExecuterAgentPort)
	isLocal := s.isLocal && bo.NodeId != 0
	return task.CreateSubTaskInstance(
		bo.StructName, bo.Id, bo.Name, &ctx, bo.State, bo.Operator, bo.CanCancel, bo.CanContinue, bo.CanPass,
		bo.CanRetry, bo.CanRollback, bo.ExecuteTimes, *agentInfo, isLocal, bo.StartTime, bo.EndTime)
}

// convertDagInstanceBOToDO converts DagInstance to sqlite.DagInstance or oceanbase.DagInstance.
func (s *taskService) convertDagInstanceBOToDO(dagInstanceBO *bo.DagInstance) interface{} {
	if s.isLocal {
		return sqlite.ConvertDagInstanceBOToDO(dagInstanceBO)
	}
	return oceanbase.ConvertDagInstanceBOToDO(dagInstanceBO)
}

// convertNodeInstanceBOToDO convert NodeInstance to sqlite.NodeInstance or oceanbase.NodeInstance.
func (s *taskService) convertNodeInstanceBOToDO(nodeInstanceBO *bo.NodeInstance) interface{} {
	if s.isLocal {
		return sqlite.ConvertNodeInstanceBOToDO(nodeInstanceBO)
	}
	return oceanbase.ConvertNodeInstanceBOToDO(nodeInstanceBO)
}

// convertSubTaskInstanceBOToDO convert SubTaskInstance to sqlite.SubtaskInstance or oceanbase.SubtaskInstance.
func (s *taskService) convertSubTaskInstanceBOToDO(subTaskBO *bo.SubTaskInstance) interface{} {
	if s.isLocal {
		return sqlite.ConvertSubTaskInstanceBOToDO(subTaskBO)
	}
	return oceanbase.ConvertSubTaskInstanceBOToDO(subTaskBO)
}

func (s *taskService) convertSubTaskLogBOToDO(subTaskLogBO *bo.SubTaskLog) interface{} {
	if s.isLocal {
		return sqlite.ConvertSubTaskLogBOToDO(subTaskLogBO)
	}
	return oceanbase.ConvertSubTaskLogBOToDO(subTaskLogBO)
}

// convertDagInstanceBO convert sqlite.DagInstance or oceanbase.DagInstance to DagInstance.
func (s *taskService) convertDagInstanceBO(model interface{}) *bo.DagInstance {
	if s.isLocal {
		dagInstance := model.(*sqlite.DagInstance)
		return dagInstance.ToBO()
	}
	dagInstance := model.(*oceanbase.DagInstance)
	return dagInstance.ToBO()
}

// convertNodeInstanceBO convert sqlite.NodeInstance or oceanbase.NodeInstance to NodeInstance.
func (s *taskService) convertNodeInstanceBO(model interface{}) *bo.NodeInstance {
	if s.isLocal {
		nodeInstance := model.(*sqlite.NodeInstance)
		return nodeInstance.ToBO()
	}
	nodeInstance := model.(*oceanbase.NodeInstance)
	return nodeInstance.ToBO()
}

func (s *taskService) convertSubTaskInstanceBO(model interface{}) *bo.SubTaskInstance {
	if s.isLocal {
		subTaskInstance := model.(*sqlite.SubtaskInstance)
		return subTaskInstance.ToBO()
	}
	subTaskInstance := model.(*oceanbase.SubtaskInstance)
	return subTaskInstance.ToBO()
}

func (s *taskService) convertDagInstanceBOSlice(model interface{}) []*bo.DagInstance {
	dagInstancesBO := make([]*bo.DagInstance, 0)
	if s.isLocal {
		dagInstances := model.(*[]sqlite.DagInstance)
		for _, dagInstance := range *dagInstances {
			dagInstancesBO = append(dagInstancesBO, dagInstance.ToBO())
		}
	} else {
		dagInstances := model.(*[]oceanbase.DagInstance)
		for _, dagInstance := range *dagInstances {
			dagInstancesBO = append(dagInstancesBO, dagInstance.ToBO())
		}
	}
	return dagInstancesBO
}

func (s *taskService) convertNodeInstanceBOSlice(model interface{}) []*bo.NodeInstance {
	nodeInstancesBO := make([]*bo.NodeInstance, 0)
	if s.isLocal {
		nodeInstances := model.(*[]sqlite.NodeInstance)
		for _, nodeInstance := range *nodeInstances {
			nodeInstancesBO = append(nodeInstancesBO, nodeInstance.ToBO())
		}
	} else {
		nodeInstances := model.(*[]oceanbase.NodeInstance)
		for _, nodeInstance := range *nodeInstances {
			nodeInstancesBO = append(nodeInstancesBO, nodeInstance.ToBO())
		}
	}
	return nodeInstancesBO
}

func (s *taskService) convertSubTaskInstanceBOSlice(model interface{}) []*bo.SubTaskInstance {
	subTaskInstancesBO := make([]*bo.SubTaskInstance, 0)
	if s.isLocal {
		subTaskInstances := model.(*[]sqlite.SubtaskInstance)
		for _, subTaskInstance := range *subTaskInstances {
			subTaskInstancesBO = append(subTaskInstancesBO, subTaskInstance.ToBO())
		}
	} else {
		subTaskInstances := model.(*[]oceanbase.SubtaskInstance)
		for _, subTaskInstance := range *subTaskInstances {
			subTaskInstancesBO = append(subTaskInstancesBO, subTaskInstance.ToBO())
		}
	}
	return subTaskInstancesBO
}

func (s *taskService) getDagModel() interface{} {
	if s.isLocal {
		return &sqlite.DagInstance{}
	}
	return &oceanbase.DagInstance{}
}

func (s *taskService) getNodeModel() interface{} {
	if s.isLocal {
		return &sqlite.NodeInstance{}
	}
	return &oceanbase.NodeInstance{}
}

func (s *taskService) getSubTaskModel() interface{} {
	if s.isLocal {
		return &sqlite.SubtaskInstance{}
	}
	return &oceanbase.SubtaskInstance{}
}

func (s *taskService) getSubTaskLogModel() interface{} {
	if s.isLocal {
		return &sqlite.SubTaskLog{}
	}
	return &oceanbase.SubTaskLog{}
}

func (s *taskService) getDagModelSlice() interface{} {
	if s.isLocal {
		return &[]sqlite.DagInstance{}
	}
	return &[]oceanbase.DagInstance{}
}

func (s *taskService) getNodeModelSlice() interface{} {
	if s.isLocal {
		return &[]sqlite.NodeInstance{}
	}
	return &[]oceanbase.NodeInstance{}
}

func (s *taskService) getSubTaskModelSlice() interface{} {
	if s.isLocal {
		return &[]sqlite.SubtaskInstance{}
	}
	return &[]oceanbase.SubtaskInstance{}
}

func (s *taskService) getDbInstance() (*gorm.DB, error) {
	if s.isLocal {
		return sqlitedb.GetSqliteInstance()
	}
	return oceanbasedb.GetOcsInstance()
}

func (s *taskService) getCurrentTime(db *gorm.DB) (t time.Time) {
	if s.isLocal {
		return time.Now()
	}
	db.Raw("SELECT NOW(6)").Scan(&t)
	return
}
