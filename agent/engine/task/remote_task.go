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
	"time"

	"github.com/oceanbase/obshell/agent/meta"
)

type RemoteTask struct {
	TaskID        int64          `json:"task_id" binding:"required"`
	Name          string         `json:"name"`
	StructName    string         `json:"struct_name"`
	State         int            `json:"state" binding:"required"`
	Operator      int            `json:"operator" binding:"required"`
	CanCancel     bool           `json:"can_cancel"`
	CanContinue   bool           `json:"can_continue"`
	CanPass       bool           `json:"can_pass"`
	CanRetry      bool           `json:"can_retry"`
	CanRollback   bool           `json:"can_rollback"`
	Context       TaskContext    `json:"context" binding:"required"`
	ExecuteTimes  int            `json:"execute_times" binding:"required"`
	ExecuterAgent meta.AgentInfo `json:"executer_agent" binding:"required"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
}

func (t *RemoteTask) GetStructName() string {
	return t.StructName
}

func (t *RemoteTask) Execute() error {
	return errors.New("remote task struct can not execute")
}

func NewRemoteTask(
	structName string, taskID int64, taskName string, ctx *TaskContext, state int, operator int,
	canCancel bool, canContinue bool, canPass bool, canRetry bool, canRollback bool, executeTimes int,
	executerAgent meta.AgentInfo, startTime time.Time, endTime time.Time) *RemoteTask {

	return &RemoteTask{
		TaskID:        taskID,
		Name:          taskName,
		StructName:    structName,
		Context:       *ctx,
		State:         state,
		Operator:      operator,
		CanCancel:     canCancel,
		CanContinue:   canContinue,
		CanPass:       canPass,
		CanRetry:      canRetry,
		CanRollback:   canRollback,
		StartTime:     startTime,
		EndTime:       endTime,
		ExecuteTimes:  executeTimes,
		ExecuterAgent: executerAgent,
	}
}
