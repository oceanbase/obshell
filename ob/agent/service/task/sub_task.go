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

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/json"
	"github.com/oceanbase/obshell/ob/agent/meta"
	sqlitedb "github.com/oceanbase/obshell/ob/agent/repository/db/sqlite"
	bo "github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
)

func (s *taskService) newSubTasks(nodeInstanceBO *bo.NodeInstance, ctx *task.TaskContext) []*bo.SubTaskInstance {
	subtasks := make([]*bo.SubTaskInstance, 0, nodeInstanceBO.MaxStage)
	agents := s.GetExecuteAgents(ctx)
	if len(agents) > 0 {
		for _, agent := range agents {
			subtasks = append(subtasks, s.newSubTask(nodeInstanceBO, agent.Ip, agent.Port))
		}
	} else {
		subtasks = append(subtasks, s.newSubTask(nodeInstanceBO, meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()))
	}
	return subtasks
}

func (s *taskService) newSubTask(nodeInstanceBO *bo.NodeInstance, agentIP string, agentPort int) *bo.SubTaskInstance {
	return &bo.SubTaskInstance{
		NodeId:            nodeInstanceBO.Id,
		Name:              nodeInstanceBO.Name,
		State:             task.PENDING,
		Context:           nodeInstanceBO.Context,
		Operator:          task.RUN,
		StructName:        nodeInstanceBO.StructName,
		ExecuterAgentIp:   agentIP,
		ExecuterAgentPort: agentPort,
	}
}

func (s *taskService) insertNewSubTasks(tx *gorm.DB, nodeInstanceBO *bo.NodeInstance, node *task.Node) error {
	subTasksBO := s.newSubTasks(nodeInstanceBO, node.GetContext())
	for _, subTaskBO := range subTasksBO {
		subTaskBO.NodeId = nodeInstanceBO.Id
		subTaskBO.CanCancel = node.CanCancel()
		subTaskBO.CanContinue = node.CanContinue()
		subTaskBO.CanPass = node.CanPass()
		subTaskBO.CanRetry = node.CanRetry()
		subTaskBO.CanRollback = node.CanRollback()
		subTask := s.convertSubTaskInstanceBOToDO(subTaskBO)
		if resp := tx.Create(subTask); resp.Error != nil {
			return resp.Error
		}
	}
	return nil
}

func (s *taskService) GetSubTasks(node *task.Node) ([]task.ExecutableTask, error) {
	subtaskInstances := s.getSubTaskModelSlice()
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	if err := db.Model(s.getSubTaskModel()).Where("node_id=?", node.GetID()).Find(subtaskInstances).Error; err != nil {
		return nil, err
	}
	subTaskInstancesBO := s.convertSubTaskInstanceBOSlice(subtaskInstances)
	for _, subtaskInstanceBO := range subTaskInstancesBO {
		subtask, err := s.convertSubTaskInstance(subtaskInstanceBO)
		if err != nil {
			return nil, err
		}
		node.AddSubTask(subtask)
	}
	return node.GetSubTasks(), nil
}

func (s *taskService) GetNodeBySubTask(taskID int64) (*task.Node, error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	var nodeID int64
	if err = db.Model(s.getSubTaskModel()).Select("node_id").Where("id=?", taskID).First(&nodeID).Error; err != nil {
		return nil, err
	}
	return s.GetNodeByNodeId(nodeID)
}

func (s *taskService) GetSubTaskByTaskID(taskID int64) (task.ExecutableTask, error) {
	subTaskInstance := s.getSubTaskModel()
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	if err := db.Model(s.getSubTaskModel()).Where("id=?", taskID).First(subTaskInstance).Error; err != nil {
		return nil, err
	}
	subTaskInstanceBO := s.convertSubTaskInstanceBO(subTaskInstance)
	return s.convertSubTaskInstance(subTaskInstanceBO)
}

func (s *taskService) SetSubTaskReady(subtask task.ExecutableTask, operator int) error {
	ctx, err := json.Marshal(subtask.GetContext())
	if err != nil {
		return err
	}
	subTaskInstanceBO := &bo.SubTaskInstance{
		Id:                subtask.GetID(),
		State:             task.READY,
		Operator:          operator,
		ExecuteTimes:      subtask.GetExecuteTimes() + 1,
		ExecuterAgentIp:   subtask.GetExecuteAgent().Ip,
		ExecuterAgentPort: subtask.GetExecuteAgent().Port,
		Context:           ctx,
	}

	// Update based on ID and ExecuteTimes.
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	subTaskInstance := s.convertSubTaskInstanceBOToDO(subTaskInstanceBO)
	resp := db.Model(s.getSubTaskModel()).Where("id=? and execute_times=? and state!=?", subtask.GetID(), subtask.GetExecuteTimes(), task.READY).Updates(subTaskInstance)
	err = resp.Error
	if err != nil {
		return err
	}
	if resp.RowsAffected == 0 {
		return errors.Occur(errors.ErrGormNoRowAffected, "failed to set task ready")
	}
	subtask.SetState(subTaskInstanceBO.State)
	subtask.SetOperator(subTaskInstanceBO.Operator)
	subtask.AddExecuteTimes()
	return nil
}

func (s *taskService) StartSubTask(subtask task.ExecutableTask) error {
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}

	taskInstanceBO := &bo.SubTaskInstance{
		Id:        subtask.GetID(),
		State:     task.RUNNING,
		StartTime: s.getCurrentTime(db),
	}

	taskInstance := s.convertSubTaskInstanceBOToDO(taskInstanceBO)
	resp := db.Model(s.getSubTaskModel()).Where("id=? and execute_times=? and state=?", subtask.GetID(), subtask.GetExecuteTimes(), task.READY).Updates(taskInstance)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		if err := db.Model(s.getSubTaskModel()).Where("id=?", subtask.GetID()).First(&taskInstance).Error; err != nil {
			return err
		}

		taskInstanceBO = s.convertSubTaskInstanceBO(taskInstance)
		if taskInstanceBO.State != task.RUNNING {
			return errors.Occurf(errors.ErrCommonUnexpected, "failed to start task: sub task %d state is %d now", subtask.GetID(), taskInstanceBO.State)
		} else if taskInstanceBO.ExecuteTimes != subtask.GetExecuteTimes() {
			return errors.Occurf(errors.ErrCommonUnexpected, "failed to start task: sub task %d execute times is %d now", subtask.GetID(), taskInstanceBO.ExecuteTimes)
		} else if taskInstanceBO.ExecuterAgentIp != subtask.GetExecuteAgent().Ip || taskInstanceBO.ExecuterAgentPort != subtask.GetExecuteAgent().Port {
			return errors.Occurf(errors.ErrCommonUnexpected, "failed to start task: sub task %d execute agent is %s now", subtask.GetID(), meta.NewAgentInfo(taskInstanceBO.ExecuterAgentIp, taskInstanceBO.ExecuterAgentPort).String())
		}
	}
	subtask.SetState(taskInstanceBO.State)
	subtask.SetStartTime(taskInstanceBO.StartTime)
	return nil
}

func (s *taskService) FinishSubTask(subtask task.ExecutableTask, state int) error {
	if state != task.SUCCEED && state != task.FAILED {
		return errors.Occur(errors.ErrCommonUnexpected, "invalid state")
	}
	ctx, err := json.Marshal(subtask.GetContext())
	if err != nil {
		return err
	}

	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	taskInstanceBO := &bo.SubTaskInstance{
		State:   state,
		Context: ctx,
		EndTime: s.getCurrentTime(db),
	}
	taskInstance := s.convertSubTaskInstanceBOToDO(taskInstanceBO)

	return db.Transaction(func(tx *gorm.DB) error {
		resp := tx.Model(s.getSubTaskModel()).Where("id=? and execute_times=? and state=?", subtask.GetID(), subtask.GetExecuteTimes(), task.RUNNING).Updates(taskInstance)
		if resp.Error != nil {
			return resp.Error
		}
		if resp.RowsAffected == 0 {
			return errors.Occur(errors.ErrGormNoRowAffected, "failed to finish sub task")
		}
		if s.isLocal && !subtask.IsLocalTask() {
			// After executing the remote task locally, synchronization is required.
			if err := tx.Model(&sqlite.TaskMapping{}).Where("local_task_id = ?", subtask.GetID()).Update("is_sync", false).Error; err != nil {
				return err
			}
		}
		subtask.SetState(state)
		subtask.SetEndTime(taskInstanceBO.EndTime)
		return nil
	})
}

func (s *taskService) SetSubTaskFailed(subtask task.ExecutableTask, logContent string) error {
	taskInstanceBO := &bo.SubTaskInstance{
		Id:    subtask.GetID(),
		State: task.FAILED,
	}
	db, err := s.getDbInstance()
	if err != nil {
		return err
	}
	taskInstanceBO.EndTime = s.getCurrentTime(db)
	taskInstance := s.convertSubTaskInstanceBOToDO(taskInstanceBO)
	return db.Transaction(func(tx *gorm.DB) error {
		subTaskLogBO := &bo.SubTaskLog{
			SubTaskId:    subtask.GetID(),
			ExecuteTimes: subtask.GetExecuteTimes(),
			LogContent:   logContent,
			IsSync:       subtask.IsLocalTask(),
		}
		subTaskLog := s.convertSubTaskLogBOToDO(subTaskLogBO)
		if err := tx.Create(subTaskLog).Error; err != nil {
			return err
		}
		resp := tx.Model(s.getSubTaskModel()).Where("id=? and execute_times=?", subtask.GetID(), subtask.GetExecuteTimes()).Updates(taskInstance)
		if resp.Error != nil {
			return err
		}
		if resp.RowsAffected == 0 {
			return errors.Occur(errors.ErrGormNoRowAffected, "failed to set task failed")
		}
		subtask.SetState(taskInstanceBO.State)
		subtask.SetEndTime(taskInstanceBO.EndTime)
		return nil
	})
}

func (s *taskService) GetAllUnfinishedSubTasks() ([]task.ExecutableTask, error) {
	db, err := s.getDbInstance()
	if err != nil {
		return nil, err
	}
	subTaskInstances := s.getSubTaskModelSlice()
	if err := db.Model(s.getSubTaskModel()).Where("state in (?)", []int{task.READY, task.RUNNING}).Find(subTaskInstances).Error; err != nil {
		return nil, err
	}
	subTaskInstancesBO := s.convertSubTaskInstanceBOSlice(subTaskInstances)
	var subTasks []task.ExecutableTask
	for _, localTask := range subTaskInstancesBO {
		subTask, err := s.convertSubTaskInstance(localTask)
		if err != nil {
			return nil, err
		}
		subTasks = append(subTasks, subTask)
	}
	return subTasks, nil
}

func (s *taskService) CreateLocalTaskInstanceByRemoteTask(remoteTask *task.RemoteTask) (int64, error) {
	taskMapping := &sqlite.TaskMapping{
		RemoteTaskId: remoteTask.TaskID,
		ExecuteTimes: remoteTask.ExecuteTimes,
		IsSync:       true, // synchronization is not required yet when the local task is created
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return 0, err
	}
	if err := sqliteDb.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&taskMapping).Error
		if err != nil {
			return err
		}
		localTask, err := s.convertLocalTaskInstance(remoteTask)
		if err != nil {
			return err
		}
		if err = tx.Create(localTask).Error; err != nil {
			return err
		}
		taskMapping.LocalTaskId = localTask.Id
		err = tx.Save(&taskMapping).Error
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return taskMapping.LocalTaskId, nil
}

func (s *taskService) UpdateLocalTaskInstanceByRemoteTask(remoteTask *task.RemoteTask) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		var taskMapping sqlite.TaskMapping
		if err := tx.Model(&taskMapping).Where("remote_task_id = ?", remoteTask.TaskID).First(&taskMapping).Error; err != nil {
			return err
		}

		var localTask *sqlite.SubtaskInstance
		if err := tx.Model(&localTask).Where("id = ?", taskMapping.LocalTaskId).First(&localTask).Error; err != nil {
			return err
		}

		if remoteTask.ExecuteTimes > localTask.ExecuteTimes {
			localTask.ExecuteTimes = remoteTask.ExecuteTimes
			localTask.State = remoteTask.State
			localTask.Operator = remoteTask.Operator
			localTask.StartTime = remoteTask.StartTime
			localTask.EndTime = remoteTask.EndTime
			ctxJsonStr, err := json.Marshal(remoteTask.Context)
			if err != nil {
				return err
			}
			localTask.Context = ctxJsonStr
			if err := tx.Save(localTask).Error; err != nil {
				return err
			}
			taskMapping.IsSync = true // no need to synchronize again
			taskMapping.ExecuteTimes = remoteTask.ExecuteTimes
			if err := tx.Save(&taskMapping).Error; err != nil {
				return err
			}
		} else {
			return errors.Occur(errors.ErrCommonUnexpected, "remote task execute times is less than local task execute times")
		}

		return nil
	})
}

func (s *taskService) convertLocalTaskInstance(remoteTask *task.RemoteTask) (*sqlite.SubtaskInstance, error) {
	ctxJsonStr, err := json.Marshal(remoteTask.Context)
	if err != nil {
		return nil, err
	}

	localTask := &sqlite.SubtaskInstance{
		NodeId:            0,
		Name:              remoteTask.Name,
		StructName:        remoteTask.StructName,
		Context:           ctxJsonStr,
		State:             remoteTask.State,
		Operator:          remoteTask.Operator,
		CanCancel:         remoteTask.CanCancel,
		CanContinue:       remoteTask.CanContinue,
		CanPass:           remoteTask.CanPass,
		CanRetry:          remoteTask.CanRetry,
		CanRollback:       remoteTask.CanRollback,
		ExecuteTimes:      remoteTask.ExecuteTimes,
		ExecuterAgentIp:   remoteTask.ExecuterAgent.Ip,
		ExecuterAgentPort: remoteTask.ExecuterAgent.Port,
		StartTime:         remoteTask.StartTime,
		EndTime:           remoteTask.EndTime,
	}
	return localTask, nil
}

func (s *taskService) GetLocalTaskInstanceByRemoteTaskId(remoteTaskId int64) (*sqlite.SubtaskInstance, error) {
	var taskMapping sqlite.TaskMapping
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	if err := sqliteDb.Model(&taskMapping).Where("remote_task_id = ?", remoteTaskId).Scan(&taskMapping).Error; err != nil {
		return nil, err
	}
	if taskMapping.LocalTaskId == 0 {
		return nil, nil
	}
	var localTask sqlite.SubtaskInstance
	if err := sqliteDb.Model(&localTask).Where("id = ?", taskMapping.LocalTaskId).First(&localTask).Error; err != nil {
		return nil, err
	}
	return &localTask, nil
}

func (s *taskService) GetTaskMappingByRemoteTaskId(remoteTaskId int64) (*sqlite.TaskMapping, error) {
	var taskMapping sqlite.TaskMapping
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	if err := sqliteDb.Model(&taskMapping).Where("remote_task_id = ?", remoteTaskId).Scan(&taskMapping).Error; err != nil {
		return nil, err
	}
	if taskMapping.LocalTaskId == 0 {
		return nil, nil
	}
	return &taskMapping, nil
}

func (s *taskService) GetRemoteTaskIdByLocalTaskId(localTaskId int64) (int64, error) {
	var taskMapping sqlite.TaskMapping
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return 0, err
	}
	if err := sqliteDb.Model(&taskMapping).Where("local_task_id = ?", localTaskId).First(&taskMapping).Error; err != nil {
		return 0, err
	}
	return taskMapping.RemoteTaskId, nil
}

func (s *taskService) SetTaskMappingSync(remoteTaskId int64, executeTimes int) error {
	taskMapping := sqlite.TaskMapping{
		RemoteTaskId: remoteTaskId,
		ExecuteTimes: executeTimes,
		IsSync:       true,
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	resp := sqliteDb.Model(&sqlite.TaskMapping{}).Where("remote_task_id=? and execute_times=?", taskMapping.RemoteTaskId, taskMapping.ExecuteTimes).Updates(&taskMapping)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.Occur(errors.ErrGormNoRowAffected, "failed to set taskMapping sync")
	}
	return nil
}

func (s *taskService) GetUnSyncTaskMappingByTime(lastTime time.Time, limit int) (taskMappings []sqlite.TaskMapping, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.TaskMapping{}).Where("gmt_modify > ? and is_sync = false", lastTime).Order("gmt_modify asc").Limit(limit).Find(&taskMappings).Error
	return
}

func (s *taskService) IsRetryTask(localTaskId int64) (isRetry bool, err error) {
	taskId := localTaskId
	if !s.isLocal {
		taskId, err = s.GetRemoteTaskIdByLocalTaskId(localTaskId)
		if err != nil {
			return false, errors.Wrap(err, "get remote task id failed")
		}
	}

	node, err := s.GetNodeBySubTask(taskId)
	if err != nil {
		return false, err
	}
	return node.IsRetry(), nil
}
