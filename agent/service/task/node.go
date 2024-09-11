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
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

func (s *taskService) newNodes(template *task.Template, ctx *task.TaskContext) ([]*bo.NodeInstance, error) {
	nodes := template.GetNodes()
	nodeInstancesBO := make([]*bo.NodeInstance, 0, len(nodes))
	for idx, node := range nodes {
		maxStage := 1
		s.mergeNodeContext(node, ctx)
		agents := s.GetExecuteAgents(node.GetContext())
		if node.IsParallel() {
			if len(agents) == 0 {
				return nil, fmt.Errorf("parallel node %s has no execute agents", node.GetName())
			}
			maxStage = len(agents)
		} else {
			if len(agents) > 1 {
				return nil, fmt.Errorf("serial node %s has more than one execute agents", node.GetName())
			}
		}
		nodeInstancesBO = append(nodeInstancesBO, &bo.NodeInstance{
			DagStage:   idx + 1,
			Name:       node.GetName(),
			Type:       node.GetNodeType(),
			Operator:   task.RUN,
			State:      task.PENDING,
			MaxStage:   maxStage,
			StructName: node.GetTaskType().Name(),
		})
	}
	return nodeInstancesBO, nil
}

func (s *taskService) insertNewNode(tx *gorm.DB, node *task.Node, nodeInstanceBO *bo.NodeInstance, dagId int64) (*bo.NodeInstance, error) {
	nodeCtxJsonStr, err := json.Marshal(node.GetContext())
	if err != nil {
		return nil, err
	}
	nodeInstanceBO.DagId = dagId
	nodeInstanceBO.Context = nodeCtxJsonStr
	nodeInstance := s.convertNodeInstanceBOToDO(nodeInstanceBO)
	if resp := tx.Create(nodeInstance); resp.Error != nil {
		return nil, resp.Error
	}
	return s.convertNodeInstanceBO(nodeInstance), nil
}

func (s *taskService) GetExecuteAgents(ctx *task.TaskContext) []meta.AgentInfo {
	agents, _ := ctx.GetParam(task.EXECUTE_AGENTS).([]meta.AgentInfo)
	return agents
}

func (s *taskService) mergeNodeContext(node *task.Node, ctx *task.TaskContext) {
	nodeCtx := node.GetContext()
	if nodeCtx == nil {
		nodeCtx = task.NewTaskContext()
	}
	if node.IsParallel() {
		nodeCtx.MergeContext(ctx)
	} else {
		nodeCtx.MergeContextWithoutExecAgents(ctx)
	}
	node.SetContext(nodeCtx)
}

func (s *taskService) GetNodes(dag *task.Dag) ([]*task.Node, error) {
	nodeInstances := s.getNodeModelSlice()
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	if err = db.Model(s.getNodeModel()).Where("dag_id=?", dag.GetID()).Order("dag_stage asc").Find(nodeInstances).Error; err != nil {
		return nil, err
	}
	nodeInstancesBO := s.convertNodeInstanceBOSlice(nodeInstances)
	nodes := make([]*task.Node, 0, len(nodeInstancesBO))
	for idx, nodeInstance := range nodeInstancesBO {
		node, err := s.convertNodeInstance(nodeInstance)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
		if idx > 0 {
			nodes[idx-1].AddDownstream(node)
			node.AddUpstream(nodes[idx-1])
		}
	}
	return nodes, nil
}

func (s *taskService) GetNodeByNodeId(nodeID int64) (*task.Node, error) {
	nodeinstance := s.getNodeModel()
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	if err := db.Model(s.getNodeModel()).Where("id=?", nodeID).First(nodeinstance).Error; err != nil {
		return nil, err
	}
	nodeInstanceBO := s.convertNodeInstanceBO(nodeinstance)
	return s.convertNodeInstance(nodeInstanceBO)
}

func (s *taskService) GetNodeByStage(dagID int64, stage int) (*task.Node, error) {
	nodeinstance := s.getNodeModel()
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	if err := db.Model(s.getNodeModel()).Where("dag_id=? and dag_stage=?", dagID, stage).First(nodeinstance).Error; err != nil {
		return nil, err
	}
	nodeInstanceBO := s.convertNodeInstanceBO(nodeinstance)
	return s.convertNodeInstance(nodeInstanceBO)
}

func (s *taskService) StartNode(node *task.Node) error {
	ctx, err := json.Marshal(node.GetContext())
	if err != nil {
		return err
	}
	nodeInstanceBO := &bo.NodeInstance{
		Id:      node.GetID(),
		State:   node.GetState(),
		Context: ctx,
	}
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	nodeInstanceBO.StartTime = s.getCurrentTime(db)
	nodeInstanceBO.EndTime = nodeInstanceBO.StartTime
	nodeInstance := s.convertNodeInstanceBOToDO(nodeInstanceBO)
	resp := db.Model(s.getNodeModel()).Where("id=?", node.GetID()).Updates(nodeInstance)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.New("failed to start node: no row affected")
	}
	node.SetState(nodeInstanceBO.State)
	node.SetStartTime(nodeInstanceBO.StartTime)
	return nil
}

func (s *taskService) FinishNode(node *task.Node) error {
	nodeInstanceBO := &bo.NodeInstance{
		Id:    node.GetID(),
		State: node.GetState(),
	}
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	nodeInstanceBO.EndTime = s.getCurrentTime(db)
	// Update based on ID and StartTime
	nodeInstance := s.convertNodeInstanceBOToDO(nodeInstanceBO)
	resp := db.Model(s.getNodeModel()).Where("id=? and state=? and start_time=? ", node.GetID(), task.RUNNING, node.GetStartTime()).Updates(nodeInstance)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.New("failed to finish node: no row affected")
	}
	node.SetState(nodeInstanceBO.State)
	node.SetEndTime(nodeInstanceBO.EndTime)
	return nil
}
