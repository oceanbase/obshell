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
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

var ZERO_TIME = time.Unix(0, 0)

func (s *taskService) GetDagInstance(dagId int64) (*task.Dag, error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	dest := s.getDagModel()
	if err = db.Model(dest).Where("id=?", dagId).First(dest).Error; err != nil {
		return nil, err
	}
	return s.convertDagInstance(s.convertDagInstanceBO(dest))
}

func (s *taskService) GetDagDetail(dagId int64) (dagDetailDTO *task.DagDetailDTO, err error) {
	dag, err := s.GetDagInstance(dagId)
	if err != nil {
		return
	}
	nodes, err := s.GetNodes(dag)
	if err != nil {
		return
	}
	dagDetailDTO = task.NewDagDetailDTO(dag)
	for i := 0; i < len(nodes); i++ {
		if _, err = s.GetSubTasks(nodes[i]); err != nil {
			return nil, err
		}

		nodeDetailDTO, err := getNodeDetail(s, nodes[i])
		if err != nil {
			return nil, err
		}
		dagDetailDTO.Nodes = append(dagDetailDTO.Nodes, nodeDetailDTO)
	}
	return dagDetailDTO, nil
}

func getNodeDetail(service TaskServiceInterface, node *task.Node) (nodeDetailDTO *task.NodeDetailDTO, err error) {
	nodeDetailDTO = task.NewNodeDetailDTO(node)
	subTasks := node.GetSubTasks()
	n := len(subTasks)
	for i := 0; i < n; i++ {
		taskDetailDTO, err := getSubTaskDetail(service, subTasks[i])
		if err != nil {
			return nil, err
		}
		nodeDetailDTO.SubTasks = append(nodeDetailDTO.SubTasks, taskDetailDTO)
	}
	return
}

func getSubTaskDetail(service TaskServiceInterface, subTask task.ExecutableTask) (taskDetailDTO *task.TaskDetailDTO, err error) {
	taskDetailDTO = task.NewTaskDetailDTO(subTask)
	if subTask.IsRunning() || subTask.IsFinished() {
		taskDetailDTO.TaskLogs, err = service.GetSubTaskLogsByTaskID(subTask.GetID())
	}
	return
}

func (s *taskService) GetUnfinishedDagInstance() (*task.Dag, error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	dest := s.getDagModel()
	if err = db.Model(dest).Where("is_finished = false").First(dest).Error; err != nil {
		return nil, err
	}
	return s.convertDagInstance(s.convertDagInstanceBO(dest))
}

func (s *taskService) GetAllUnfinishedDagInstance() ([]*task.Dag, error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	dest := s.getDagModelSlice()
	if err = db.Model(s.getDagModel()).Where("is_finished = false").Find(dest).Error; err != nil {
		return nil, err
	}
	dagInstancesBO := s.convertDagInstanceBOSlice(dest)
	dags := make([]*task.Dag, 0, len(dagInstancesBO))
	for _, dagInstanceBO := range dagInstancesBO {
		dag, err := s.convertDagInstance(dagInstanceBO)
		if err != nil {
			return nil, err
		}
		dags = append(dags, dag)
	}
	return dags, nil
}

func (s *taskService) GetLastMaintenanceDag() (*task.Dag, error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	dest := s.getDagModel()
	if err = db.Model(dest).Where("is_maintenance = true").Order("id desc").First(dest).Error; err != nil {
		return nil, err
	}
	return s.convertDagInstance(s.convertDagInstanceBO(dest))
}

func (s *taskService) FindLastMaintenanceDag() (*task.Dag, error) {
	dag, err := s.GetLastMaintenanceDag()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return dag, err
}

func (s *taskService) GetDagIDBySubTaskId(taskID int64) (dagID int64, err error) {
	db, err := s.getDbInstance()
	if err != nil {
		return
	}
	var nodeID int64
	if err = db.Model(s.getSubTaskModel()).Select("node_id").Where("id=?", taskID).First(&nodeID).Error; err != nil {
		return
	}
	err = db.Model(s.getNodeModel()).Select("dag_id").Where("id=?", nodeID).First(&dagID).Error
	return
}

func (s *taskService) GetDagGenericIDBySubTaskId(taskID int64) (dagGenericID string, err error) {
	var dagID int64
	if dagID, err = s.GetDagIDBySubTaskId(taskID); err != nil {
		return
	}
	dagGenericID = task.ConvertIDToGenericID(dagID, s.isLocal)
	return
}

func (s *taskService) CreateDagInstanceByTemplate(template *task.Template, ctx *task.TaskContext) (*task.Dag, error) {
	inited, err := s.IsInited()
	if err != nil {
		return nil, err
	}
	if !inited {
		return nil, errors.New("Status Maintainer is not inited")
	}
	if template.IsEmpty() {
		return nil, errors.New("empty template")
	}
	dagInstanceBO, err := s.newDagInstanceBO(template, ctx)
	if err != nil {
		return nil, err
	}

	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	var dag *task.Dag
	err = db.Transaction(func(tx *gorm.DB) error {
		if template.IsMaintenance() {
			if err := s.StartMaintenance(tx, template); err != nil {
				return err
			}
		}

		dagInstanceBO, err = s.insertNewDag(tx, dagInstanceBO)
		if err != nil {
			return err
		}

		dag, err = s.convertDagInstance(dagInstanceBO)
		if err != nil {
			return err
		}

		if err := s.UpdateMaintenanceTask(tx, dag); err != nil {
			return err
		}

		nodeInstancesBO, err := s.newNodes(template, ctx)
		if err != nil {
			return err
		}
		nodes := template.GetNodes()
		for idx, nodeInstanceBO := range nodeInstancesBO {
			node := nodes[idx]
			nodeInstanceBO, err = s.insertNewNode(tx, node, nodeInstanceBO, dagInstanceBO.Id)
			if err != nil {
				return err
			}
			if err := s.insertNewSubTasks(tx, nodeInstanceBO, node); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dag, nil
}

// insertNewDag creates a new dag based on BO in the transaction.
func (s *taskService) insertNewDag(tx *gorm.DB, dagInstanceBO *bo.DagInstance) (*bo.DagInstance, error) {
	dagInstance := s.convertDagInstanceBOToDO(dagInstanceBO)
	if resp := tx.Create(dagInstance); resp.Error != nil {
		return nil, resp.Error
	}
	return s.convertDagInstanceBO(dagInstance), nil
}

// newDagInstanceBO creates a new DagInstance based on template and ctx.
func (s *taskService) newDagInstanceBO(template *task.Template, ctx *task.TaskContext) (*bo.DagInstance, error) {
	ctxJsonStr, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	return &bo.DagInstance{
		Name:            template.Name,
		Stage:           1,
		MaxStage:        len(template.GetNodes()),
		State:           task.READY,
		Operator:        task.RUN,
		IsMaintenance:   template.IsMaintenance(),
		MaintenanceType: template.GetMaintenanceType(),
		MaintenanceKey:  template.GetMaintenanceKey(),
		Context:         ctxJsonStr,
	}, nil
}

func (s *taskService) SetDagRollback(dag *task.Dag) error {
	rollbackNodes, err := s.getNodesCanRollback(dag)
	if err != nil {
		return err
	}
	return s.txForRollbackDag(dag, rollbackNodes)
}

func (s *taskService) SetDagRetryAndReady(dag *task.Dag) error {
	node, err := s.getNodeCanRetry(dag)
	if err != nil {
		return err
	}
	return s.txForRetryAndReadyDag(dag, node)
}

func (s *taskService) CancelDag(dag *task.Dag) error {
	node, err := s.getNodeCanCancel(dag)
	if err != nil {
		return err
	}
	return s.txForCancelDag(dag, node)
}

func (s *taskService) PassDag(dag *task.Dag) error {
	nodes, err := s.getNodesCanPass(dag)
	if err != nil {
		return err
	}
	return s.txForPassDag(dag, nodes)
}

func (s *taskService) getNodesCanRollback(dag *task.Dag) ([]*task.Node, error) {
	if dag.GetState() != task.FAILED {
		return nil, errors.New("failed to set dag rollback: dag state is not failed")
	}

	nodes, err := s.GetNodes(dag)
	if err != nil {
		return nil, err
	}

	var _idx int
	for idx, node := range nodes {
		_idx = idx
		if _, err := s.GetSubTasks(node); err != nil {
			return nil, err
		}
		if !node.IsRollback() && node.IsPending() {
			break
		}
		if !node.CanRollback() {
			return nil, fmt.Errorf("failed to set dag rollback: node %d. %s can not rollback", node.GetID(), node.GetName())
		}
		if node.IsFail() {
			_idx += 1
			break
		}
	}

	if _idx == 0 {
		return nil, errors.New("failed to set dag rollback: no node failed")
	}

	return nodes[:_idx], nil
}

func (s *taskService) getNodeCanCancel(dag *task.Dag) (*task.Node, error) {
	if dag.IsFinished() {
		return nil, errors.New("failed to cancel dag: dag is finished")
	}

	nodes, err := s.GetNodes(dag)
	if err != nil {
		return nil, err
	}
	var currentNode *task.Node
	for _, node := range nodes {
		currentNode = node
		if _, err := s.GetSubTasks(node); err != nil {
			return nil, err
		}
		if !node.CanCancel() {
			return nil, fmt.Errorf("failed to cancel dag: %s can not cancel", node.GetName())
		}
		if !node.IsFinished() {
			break
		}
	}
	if currentNode == nil {
		return nil, errors.New("failed to cancel dag: no node found")
	}
	return currentNode, nil
}

func (s *taskService) getNodesCanPass(dag *task.Dag) ([]*task.Node, error) {
	if !dag.IsFail() {
		return nil, errors.New("failed to pass dag: dag is not failed")
	}
	nodes, err := s.GetNodes(dag)
	if err != nil {
		return nil, err
	}
	var idx int
	for _idx, node := range nodes {
		idx = _idx
		if _, err := s.GetSubTasks(node); err != nil {
			return nil, err
		}
		if !node.CanPass() {
			return nil, fmt.Errorf("failed to pass dag: %s can not pass", node.GetName())
		}
		if node.IsFail() {
			break
		}
	}
	return nodes[idx:], nil
}

func (s *taskService) getNodeCanRetry(dag *task.Dag) (*task.Node, error) {
	if !dag.IsFail() {
		return nil, errors.New("failed to set dag retry: dag state is not failed")
	}
	node, err := s.GetNodeByStage(dag.GetID(), dag.GetStage())
	if err != nil {
		return nil, err
	}
	if _, err = s.GetSubTasks(node); err != nil {
		return nil, err
	}
	if !node.CanRetry() {
		return nil, fmt.Errorf("failed to set dag retry: %s can not retry", node.GetName())
	}
	return node, nil
}

func (s *taskService) txForRollbackDag(dag *task.Dag, rollbackNodes []*task.Node) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if dag.IsMaintenance() && dag.GetContext().GetParam(task.FAILURE_EXIT_MAINTENANCE) != nil {
			if err := s.StartMaintenance(tx, dag); err != nil {
				return err
			}
			if err := s.UpdateMaintenanceTask(tx, dag); err != nil {
				return err
			}
		}

		// Update dag state & operator
		if err := s.updateDagOperator(tx, dag, task.ROLLBACK); err != nil {
			return errors.Wrap(err, "failed to set dag rollback")
		}

		for _, node := range rollbackNodes {
			if !node.IsFinished() {
				continue
			}
			// Update node state & operator
			if err := s.updateNodeOperator(tx, node, task.ROLLBACK); err != nil {
				return errors.Wrap(err, "failed to set node rollback")
			}
		}
		return nil
	})
}

func (s *taskService) txForRetryAndReadyDag(dag *task.Dag, node *task.Node) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.updateDagOperator(tx, dag, task.RUN); err != nil {
			return errors.Wrap(err, "failed to rerun dag")
		}
		if err := s.updateNodeOperator(tx, node, task.RETRY); err != nil {
			return errors.Wrap(err, "failed to retry node")
		}

		return nil
	})

}

func (s *taskService) txForCancelDag(dag *task.Dag, node *task.Node) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.updateDagOperator(tx, dag, task.CANCEL); err != nil {
			return errors.Wrap(err, "failed to cancel dag")
		}
		if err := s.updateNodeOperator(tx, node, task.CANCEL); err != nil {
			return errors.Wrap(err, "failed to cancel node")
		}
		return nil
	})
}

func (s *taskService) txForPassDag(dag *task.Dag, nodes []*task.Node) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		dag.SetEndTime(s.getCurrentTime(tx))
		if err := s.updateDagOperator(tx, dag, task.PASS); err != nil {
			return errors.Wrap(err, "failed to pass dag")
		}
		for _, node := range nodes {
			node.SetEndTime(dag.GetEndTime())
			if err := s.updateNodeOperator(tx, node, task.PASS); err != nil {
				return errors.Wrap(err, "failed to pass node")
			}

			subTaskInstanceBO := &bo.SubTaskInstance{
				State:    task.SUCCEED,
				EndTime:  dag.GetEndTime(),
				Operator: task.PASS,
			}
			subTaskInstance := s.convertSubTaskInstanceBOToDO(subTaskInstanceBO)
			if err := tx.Model(subTaskInstance).Where("node_id=? and state!=?", node.GetID(), task.SUCCEED).Updates(subTaskInstance).Error; err != nil {
				return err
			}
		}
		if dag.IsMaintenance() {
			if err := s.StopMaintenance(tx, dag); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *taskService) StartDag(dag *task.Dag) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	if err := s.updateDagState(db, dag, task.RUNNING); err != nil {
		return err
	}
	return nil
}

func (s *taskService) UpdateDagStage(dag *task.Dag, nextSage int) error {
	dagInstanceBO := &bo.DagInstance{
		Id:    dag.GetID(),
		Stage: nextSage,
	}
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	dagInstance := s.convertDagInstanceBOToDO(dagInstanceBO)
	resp := db.Model(s.getDagModel()).Where("stage=? and start_time=?", dag.GetStage(), dag.GetStartTime()).Updates(dagInstance)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.New("failed to update dag: no row affected")
	}
	return nil
}

func (s *taskService) FinishDagAsFailed(dag *task.Dag) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.updateDagState(tx, dag, task.FAILED); err != nil {
			return err
		}

		if dag.IsMaintenance() && dag.GetContext().GetParam(task.FAILURE_EXIT_MAINTENANCE) != nil {
			return s.StopMaintenance(tx, dag)
		}
		return nil
	})
}

func (s *taskService) FinishDagAsSucceed(dag *task.Dag) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.updateDagState(tx, dag, task.SUCCEED); err != nil {
			return err
		}
		if dag.IsMaintenance() {
			return s.StopMaintenance(tx, dag)
		}
		return nil
	})
}

func (s *taskService) updateDagOperator(tx *gorm.DB, dag *task.Dag, operator int) error {
	data := map[string]interface{}{
		"state":      task.READY,
		"operator":   operator,
		"start_time": ZERO_TIME,
		"end_time":   ZERO_TIME,
	}
	currentState := task.FAILED

	switch operator {
	case task.CANCEL:
		currentState = dag.GetState()
	case task.ROLLBACK, task.RUN:
		data["is_finished"] = false
	case task.PASS:
		data["state"] = task.SUCCEED
		data["end_time"] = dag.GetEndTime()
		data["is_finished"] = true
	}

	resp := tx.Model(s.getDagModel()).Where("id=? and state=? and start_time=?", dag.GetID(), currentState, dag.GetStartTime()).Updates(data)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.New("no row affected")
	}
	return nil
}

func (s *taskService) updateDagState(tx *gorm.DB, dag *task.Dag, state int) error {
	bo := &bo.DagInstance{
		Id:    dag.GetID(),
		State: state,
		Stage: dag.GetStage(),
	}
	currentState := task.RUNNING

	switch state {
	case task.RUNNING:
		currentState = task.READY
		bo.StartTime = s.getCurrentTime(tx)
	case task.FAILED:
		if !dag.IsMaintenance() {
			bo.IsFinished = true
		}
		bo.EndTime = s.getCurrentTime(tx)
	case task.SUCCEED:
		bo.IsFinished = true
		bo.EndTime = s.getCurrentTime(tx)
	default:
		return fmt.Errorf("invalid dag state '%d'", state)
	}

	dagInstance := s.convertDagInstanceBOToDO(bo)
	resp := tx.Model(dagInstance).Where("state=? and start_time=?", currentState, dag.GetStartTime()).Updates(dagInstance)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.New("failed to update dag state: no row affected")
	}
	return nil
}

func (s *taskService) updateNodeOperator(tx *gorm.DB, node *task.Node, operator int) error {
	bo := &bo.NodeInstance{
		Id:       node.GetID(),
		State:    task.PENDING,
		Operator: operator,
		EndTime:  ZERO_TIME,
	}

	switch operator {
	case task.CANCEL, task.ROLLBACK, task.RETRY:
		bo.StartTime = ZERO_TIME
	case task.PASS:
		bo.State = task.SUCCEED
		bo.EndTime = node.GetEndTime()
	}

	nodeInstance := s.convertNodeInstanceBOToDO(bo)
	resp := tx.Model(nodeInstance).Where("state=? and start_time=?", node.GetState(), node.GetStartTime()).Updates(nodeInstance)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.New("failed to update instance in tx: no row affected")
	}
	return nil
}
