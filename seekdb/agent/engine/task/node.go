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
	"reflect"
	"time"
)

const (
	NORMAL   = "normal"
	PARALLEL = "parallel"
)

type Node struct {
	subtasks   []ExecutableTask
	taskType   reflect.Type
	nodeType   string
	upStream   *Node
	downStream *Node
	dagId      int
	TaskInfo
	ctx *TaskContext
}

func (node *Node) GetTaskType() reflect.Type {
	return node.taskType
}

func (node *Node) GetNodeType() string {
	return node.nodeType
}

func (node *Node) GetDagId() int {
	return node.dagId
}

func (node *Node) GetSubTasks() []ExecutableTask {
	return node.subtasks
}

func (node *Node) GetContext() *TaskContext {
	return node.ctx
}

func (node *Node) MergeContext(ctx *TaskContext) {
	node.ctx.MergeContextWithoutKeyords(ctx)
}

func (node *Node) SetContext(ctx *TaskContext) {
	node.ctx = ctx
}

func (node *Node) IsParallel() bool {
	return node.nodeType == PARALLEL
}

func (node *Node) AddSubTask(task ExecutableTask) error {
	if node.taskType == nil {
		node.taskType = reflect.TypeOf(task).Elem()
	} else if reflect.TypeOf(task).Elem() != node.taskType {
		return errors.New("task type not match")
	}
	node.subtasks = append(node.subtasks, task)
	return nil
}

func (node *Node) GetUpstream() *Node {
	return node.upStream
}

func (node *Node) GetDownstream() *Node {
	return node.downStream
}

func (node *Node) AddUpstream(upstream *Node) {
	if node.upStream != nil {
		panic("node already has upstream")
	}
	node.upStream = upstream
}

func (node *Node) AddDownstream(downstream *Node) {
	if node.downStream != nil {
		panic("node already has downStream")
	}
	node.downStream = downstream
}

func (node *Node) CanCancel() bool {
	for _, task := range node.subtasks {
		if !task.CanCancel() {
			return false
		}
	}
	return true
}

func (node *Node) CanContinue() bool {
	for _, task := range node.subtasks {
		if !task.CanContinue() {
			return false
		}
	}
	return true
}

func (node *Node) CanRetry() bool {
	for _, task := range node.subtasks {
		if !task.CanRetry() {
			return false
		}
	}
	return true
}

func (node *Node) CanRollback() bool {
	for _, task := range node.subtasks {
		if !task.CanRollback() {
			return false
		}
	}
	return true
}

func (node *Node) CanPass() bool {
	for _, task := range node.subtasks {
		if !task.CanPass() {
			return false
		}
	}
	return true
}

func NewNode(task ExecutableTask, paralle bool) *Node {
	taskType := reflect.TypeOf(task).Elem()
	node := &Node{
		TaskInfo: TaskInfo{
			name: task.GetName(),
		},
		taskType: taskType,
		subtasks: make([]ExecutableTask, 0),
	}
	if paralle {
		node.nodeType = PARALLEL
	} else {
		node.nodeType = NORMAL
	}
	node.AddSubTask(task)
	return node
}

func NewNodeWithContext(task ExecutableTask, paralle bool, ctx *TaskContext) *Node {
	node := NewNode(task, paralle)
	node.ctx = ctx
	return node
}

func NewNodeWithId(id int64, name string, dagId int, nodeType string, state int, operator int, structName string, ctx *TaskContext, isLocalTask bool, startTime time.Time, endTime time.Time) *Node {
	node := &Node{
		taskType: TASK_TYPE[structName],
		subtasks: make([]ExecutableTask, 0),
		nodeType: nodeType,
		ctx:      ctx,
		dagId:    dagId,
		TaskInfo: TaskInfo{
			id:          id,
			name:        name,
			state:       state,
			operator:    operator,
			isLocalTask: isLocalTask,
			startTime:   startTime,
			endTime:     endTime,
		},
	}
	return node
}
