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

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

type taskService struct {
	isLocal bool
	StatusMaintainerInterface
}

type TaskServiceInterface interface {
	DagServiceInterface
	NodeServiceInterface
	SubTaskServiceInterface
	SubTaskLogServiceInterface
	StatusMaintainerInterface

	// Get the agents that executes the task from TaskContext
	GetExecuteAgents(*task.TaskContext) []meta.AgentInfo
}

type StatusMaintainerInterface interface {
	StartMaintenance(*gorm.DB, task.Maintainer) error
	UpdateMaintenanceTask(*gorm.DB, *task.Dag) error
	StopMaintenance(*gorm.DB, task.Maintainer) error
	IsRunning() (bool, error)
	IsInited() (bool, error)
}

type DagServiceInterface interface {
	// Create dag, node, subTasks based on template and context
	CreateDagInstanceByTemplate(*task.Template, *task.TaskContext) (*task.Dag, error)

	GetDagInstance(int64) (*task.Dag, error)

	GetUnfinishedDagInstance() (*task.Dag, error)

	GetAllUnfinishedDagInstance() ([]*task.Dag, error)

	GetLastMaintenanceDag() (*task.Dag, error)

	// Advance dag from ready to running
	StartDag(*task.Dag) error

	// Advance dag to next stage
	UpdateDagStage(*task.Dag, int) error

	// Advance dag from running to failed
	FinishDagAsFailed(*task.Dag) error

	// Advance dag from running to succeed
	FinishDagAsSucceed(*task.Dag) error

	CancelDag(*task.Dag) error

	PassDag(*task.Dag) error

	SetDagRollback(*task.Dag) error

	SetDagRetryAndReady(*task.Dag) error
}

type NodeServiceInterface interface {
	GetNodes(*task.Dag) ([]*task.Node, error)

	GetNodeByNodeId(int64) (*task.Node, error)

	GetNodeByStage(int64, int) (*task.Node, error)

	// Advance node from pending to running
	StartNode(*task.Node) error

	// Advance node from running to succeed/failed
	FinishNode(*task.Node) error
}

type SubTaskServiceInterface interface {
	CreateLocalTaskInstanceByRemoteTask(*task.RemoteTask) (int64, error)

	GetLocalTaskInstanceByRemoteTaskId(int64) (*sqlite.SubtaskInstance, error)

	GetDagBySubTaskId(taskId int64) (*task.Dag, error)

	GetSubTasks(*task.Node) ([]task.ExecutableTask, error)

	GetSubTaskByTaskID(int64) (task.ExecutableTask, error)

	GetAllUnfinishedSubTasks() ([]task.ExecutableTask, error)

	GetTaskMappingByRemoteTaskId(int64) (*sqlite.TaskMapping, error)

	GetUnSyncTaskMappingByTime(time.Time, int) ([]sqlite.TaskMapping, error)

	GetRemoteTaskIdByLocalTaskId(int64) (int64, error)

	// Advance subTask from ready to running
	StartSubTask(task.ExecutableTask) error

	// Advance subTask from running to succeed/failed
	FinishSubTask(task.ExecutableTask, int) error

	// Advance subTask to ready
	SetSubTaskReady(task.ExecutableTask, int) error

	// Advance subTask from running to failed
	SetSubTaskFailed(task.ExecutableTask, string) error

	// Advance IsSync to true
	SetTaskMappingSync(int64, int) error

	UpdateLocalTaskInstanceByRemoteTask(remoteTask *task.RemoteTask) error
}

type SubTaskLogServiceInterface interface {
	GetSubTaskLogsByTaskID(int64) ([]string, error)
}
