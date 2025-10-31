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
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

const TIMEOUT_KEY = "timeout"

// State
const (
	PENDING = iota + 1
	READY
	RUNNING
	FAILED
	SUCCEED

	PENDING_STR = "PENDING"
	READY_STR   = "READY"
	RUNNING_STR = "RUNNING"
	FAILED_STR  = "FAILED"
	SUCCEED_STR = "SUCCEED"
)

var STATE_MAP = map[int]string{
	PENDING: PENDING_STR,
	READY:   READY_STR,
	RUNNING: RUNNING_STR,
	FAILED:  FAILED_STR,
	SUCCEED: SUCCEED_STR,
}

// Operator
const (
	RUN = iota + 1
	RETRY
	ROLLBACK
	CANCEL
	PASS

	RUN_STR      = "RUN"
	RETRY_STR    = "RETRY"
	ROLLBACK_STR = "ROLLBACK"
	CANCEL_STR   = "CANCEL"
	PASS_STR     = "PASS"
)

var OPERATOR_MAP = map[int]string{
	RUN:      RUN_STR,
	RETRY:    RETRY_STR,
	ROLLBACK: ROLLBACK_STR,
	CANCEL:   CANCEL_STR,
	PASS:     PASS_STR,
}

const DEFAULT_TIMEOUT = 3600 * time.Second

var (
	ERR_WAIT_OPERATOR = errors.New("wait operator to advance")
)

type TaskStatusInterface interface {
	IsSuccess() bool
	IsFail() bool
	IsRunning() bool
	IsPending() bool
	IsReady() bool
	IsFinished() bool
	IsRollback() bool
	SetOperator(operator int)
	GetOperator() int
	GetStartTime() time.Time
	GetEndTime() time.Time
	SetStartTime(startTime time.Time)
	SetEndTime(endTime time.Time)
}

type TaskInfoInterface interface {
	GetID() int64
	GetName() string
	SetState(state int)
	GetState() int
	IsLocalTask() bool
	TaskStatusInterface
}

type TaskLogInterface interface {
	ExecuteLog(text string)
	ExecuteInfoLog(text string)
	ExecuteWarnLog(err error)
	ExecuteErrorLog(err error)
	ExecuteLogf(format string, args ...interface{})
	ExecuteInfoLogf(format string, args ...interface{})
	ExecuteWarnLogf(format string, args ...interface{})
	ExecuteErrorLogf(format string, args ...interface{})
}

type Executable interface {
	IsRun() bool
	Execute() error
	SetLogChannel(logChan chan<- TaskExecuteLogDTO)
	GetTimeout() time.Duration
	TimeoutCheck()
	CanPass() bool
	GetResult() TaskResult
	GetContext() *TaskContext
	SetContext(context *TaskContext)
	GetLocalData(key string) interface{}
	GetLocalDataWithValue(key string, value interface{}) error
	SetLocalData(key string, data interface{})
	GetExecuteAgent() meta.AgentInfo
	GetExecuteTimes() int
	AddExecuteTimes()
	Finish(err error)
}

type CancelableTask interface {
	IsCancel() bool
	CanCancel() bool
	Cancel()
	SetCancelFunc(cancel context.CancelFunc)
}
type ContinuableTask interface {
	CanContinue() bool
	IsContinue() bool
	SetIsContinue()
}

type Retryable interface {
	IsRetry() bool
	CanRetry() bool
}

type RollableTask interface {
	IsRollback() bool
	CanRollback() bool
	Rollback() error
}

type AdditionalData interface {
	GetAdditionalData() map[string]interface{}
}

type ExecutableTask interface {
	TaskInfoInterface
	Executable
	TaskLogInterface
	Retryable
	CancelableTask
	ContinuableTask
	RollableTask
	AdditionalData
	SetExecuteAgent(agent meta.AgentInfo)
	GetExecuteAgent() meta.AgentInfo
}

type TaskResult struct {
	Finished    bool
	Ok          bool
	LogContents []string
}

type TaskInfo struct {
	id          int64
	name        string
	state       int
	operator    int
	canCancel   bool
	canContinue bool
	canPass     bool
	canRetry    bool
	canRollback bool
	isLocalTask bool
	startTime   time.Time
	endTime     time.Time
}

type Task struct {
	TaskInfo
	taskContext   *TaskContext
	logChan       chan<- TaskExecuteLogDTO
	logContents   []string
	cancel        context.CancelFunc
	executeTimes  int
	isContinue    bool
	executerAgent meta.AgentInfo
	localAgentKey string
}

func (task *TaskInfo) GetID() int64 {
	return task.id
}

func (task *TaskInfo) GetName() string {
	return task.name
}

func (task *TaskInfo) SetState(state int) {
	task.state = state
}

func (task *TaskInfo) GetState() int {
	return task.state
}

func (task *Task) SetExecuteAgent(agent meta.AgentInfo) {
	task.executerAgent = agent
}

func (task *Task) GetExecuteAgent() meta.AgentInfo {
	return task.executerAgent
}

func (task *Task) GetContext() *TaskContext {
	return task.taskContext
}

func (task *Task) SetContext(taskContext *TaskContext) {
	task.taskContext = taskContext
}

func (task *Task) GetLocalData(key string) interface{} {
	if task.taskContext != nil {
		return task.taskContext.GetAgentDataByAgentKey(task.localAgentKey, key)
	}
	return nil
}

func (task *Task) GetLocalDataWithValue(key string, value interface{}) error {
	if task.taskContext != nil {
		return task.taskContext.GetAgentDataByAgentKeyWithValue(task.localAgentKey, key, value)
	}
	panic(fmt.Sprintf("task %d context is nil", task.id))
}

func (task *Task) SetLocalData(key string, data interface{}) {
	if task.taskContext != nil {
		task.taskContext.SetAgentDataByAgentKey(task.localAgentKey, key, data)
		return
	}
	panic(fmt.Sprintf("task %d context is nil", task.id))
}

func (task *TaskInfo) GetStartTime() time.Time {
	return task.startTime
}

func (task *TaskInfo) GetEndTime() time.Time {
	return task.endTime
}

func (task *TaskInfo) SetStartTime(startTime time.Time) {
	task.startTime = startTime
}

func (task *TaskInfo) SetEndTime(endTime time.Time) {
	task.endTime = endTime
}

func (task *Task) SetLogChannel(logChan chan<- TaskExecuteLogDTO) {
	task.logChan = logChan
}

func (task *Task) TimeoutCheck() {
	if task.IsCancel() {
		log.Warningf("Check: task %d cancel", task.id)
		runtime.Goexit()
	} else if task.logChan == nil {
		log.Warningf("Check: task %d Interrupted", task.id)
		runtime.Goexit()
	}
}

func (task *Task) executeLog(level log.Level, text string) {
	logContext := fmt.Sprintf("task %d %s execute log: %s", task.id, task.name, text)
	switch level {
	case log.WarnLevel:
		log.Warn(logContext)
	case log.ErrorLevel:
		log.Error(logContext)
	default:
		log.Info(logContext)
	}

	task.TimeoutCheck()

	task.logChan <- TaskExecuteLogDTO{
		TaskId:       task.id,
		ExecuteTimes: task.executeTimes,
		LogContent:   text,
		IsSync:       task.isLocalTask,
	}
	task.logContents = append(task.logContents, text)
}

func (task *Task) ExecuteLog(text string) {
	task.ExecuteInfoLog(text)
}

func (task *Task) ExecuteInfoLog(text string) {
	task.executeLog(log.InfoLevel, text)
}

func (task *Task) ExecuteWarnLog(err error) {
	task.executeLog(log.WarnLevel, fmt.Sprintf("WARN: %s", err.Error()))
}

func (task *Task) ExecuteErrorLog(err error) {
	if isNotPrintErr(err) {
		task.ExecuteInfoLog(err.Error())
		return
	}

	var OcsAgentError errors.OcsAgentErrorInterface
	if tmp, ok := err.(errors.OcsAgentErrorInterface); ok {
		OcsAgentError = tmp
	} else if errors.IsMysqlError(err) {
		OcsAgentError = errors.Occur(errors.ErrMysqlError, err.Error())
	} else {
		OcsAgentError = errors.Occur(errors.ErrCommonUnexpected, err.Error())
	}
	task.executeLog(log.ErrorLevel, fmt.Sprintf("ERROR: %s", OcsAgentError.ErrorMessage()))
}

func isNotPrintErr(err error) bool {
	return errors.Is(err, ERR_WAIT_OPERATOR)
}

func (task *Task) ExecuteLogf(format string, args ...interface{}) {
	task.ExecuteInfoLog(fmt.Sprintf(format, args...))
}

func (task *Task) ExecuteInfoLogf(format string, args ...interface{}) {
	task.ExecuteInfoLog(fmt.Sprintf(format, args...))
}

func (task *Task) ExecuteWarnLogf(format string, args ...interface{}) {
	task.ExecuteWarnLog(fmt.Errorf(format, args...))
}

func (task *Task) ExecuteErrorLogf(format string, args ...interface{}) {
	task.ExecuteErrorLog(fmt.Errorf(format, args...))
}

func (task *Task) GetResult() TaskResult {
	return TaskResult{
		Finished:    task.IsFinished(),
		Ok:          task.IsSuccess(),
		LogContents: task.logContents,
	}
}

func (task *Task) GetTimeout() time.Duration {
	if task.taskContext != nil {
		if timeout, ok := task.taskContext.GetParam(TIMEOUT_KEY).(time.Duration); ok {
			return timeout
		}

		var timeout int
		if err := task.taskContext.GetParamWithValue(TIMEOUT_KEY, &timeout); err == nil {
			return time.Duration(timeout) * time.Second
		}
	}
	return DEFAULT_TIMEOUT
}

func (task *TaskInfo) GetOperator() int {
	return task.operator
}

func (task *TaskInfo) SetOperator(operator int) {
	task.operator = operator
}

func (task *TaskInfo) IsFail() bool {
	return task.GetState() == FAILED
}

func (task *TaskInfo) IsSuccess() bool {
	return task.GetState() == SUCCEED
}

func (task *TaskInfo) IsRunning() bool {
	return task.GetState() == RUNNING
}

func (task *TaskInfo) IsPending() bool {
	return task.GetState() == PENDING
}

func (task *TaskInfo) IsReady() bool {
	return task.GetState() == READY
}

func (task *TaskInfo) IsFinished() bool {
	return task.IsFail() || task.IsSuccess()
}

func (task *TaskInfo) IsRollback() bool {
	return task.operator == ROLLBACK
}

func (task *TaskInfo) IsCancel() bool {
	return task.operator == CANCEL
}

func (task *TaskInfo) IsRetry() bool {
	return task.operator == RETRY
}

func (task *TaskInfo) IsRun() bool {
	return task.operator == RUN
}

func (task *Task) GetExecuteTimes() int {
	return task.executeTimes
}

func (task *Task) AddExecuteTimes() {
	task.executeTimes++
}

func (task *TaskInfo) CanCancel() bool {
	return task.canCancel
}

func (task *TaskInfo) CanContinue() bool {
	return task.canContinue
}

func (task *TaskInfo) CanRetry() bool {
	return task.canRetry
}

func (task *TaskInfo) CanRollback() bool {
	return task.canRollback
}

func (task *TaskInfo) CanPass() bool {
	return task.canPass
}

func (task *Task) SetCanCancel() *Task {
	task.canCancel = true
	return task
}

func (task *Task) SetCanContinue() *Task {
	task.canContinue = true
	return task
}

// SetCanRetry set task can retry, and Retryable task must be rollbackable.
func (task *Task) SetCanRetry() *Task {
	task.canRetry = true
	return task
}

func (task *Task) SetCanRollback() *Task {
	task.canRollback = true
	return task
}

func (task *Task) SetCanPass() *Task {
	task.canPass = true
	return task
}

func (task *Task) SetIsContinue() {
	task.isContinue = true
}

func (task *Task) IsContinue() bool {
	return task.isContinue
}

func (task *TaskInfo) IsLocalTask() bool {
	return task.isLocalTask
}

func (task *Task) Finish(err error) {
	if err == nil {
		task.SetState(SUCCEED)
	} else {
		task.SetState(FAILED) // Set state before add log to avoid panic.
		task.ExecuteErrorLog(err)
	}
	task.cancel = nil
}

func (task *Task) SetCancelFunc(cancel context.CancelFunc) {
	task.cancel = cancel
}

func (task *Task) Cancel() {
}

func (task *Task) Rollback() error {
	if !task.IsRollback() {
		return fmt.Errorf("task %d can not rollback", task.id)
	}
	return nil
}

func (task *Task) GetAdditionalData() map[string]interface{} {
	return nil
}

func NewSubTask(name string) *Task {
	return &Task{
		TaskInfo: TaskInfo{
			name: name,
		},
	}
}

var TASK_TYPE = make(map[string]reflect.Type)

func RegisterTaskType(typedNil interface{}) {
	t := reflect.TypeOf(typedNil)
	taskType := t.Name()
	log.Info("Register Task:", taskType)
	TASK_TYPE[taskType] = t
}

func CreateSubTaskInstance(
	taskType string, id int64, taskName string, ctx *TaskContext, state int, operator int,
	canCancel bool, canContinue bool, canPass bool, canRetry bool, canRollback bool, executeTimes int,
	executerAgent meta.AgentInfo, isLocalTask bool, startTime time.Time, endTime time.Time) (ExecutableTask, error) {

	if TASK_TYPE[taskType] == nil {
		return nil, fmt.Errorf("task type %s not register", taskType)
	}
	task := Task{
		taskContext: ctx,
		TaskInfo: TaskInfo{
			id:          id,
			name:        taskName,
			state:       state,
			operator:    operator,
			canCancel:   canCancel,
			canContinue: canContinue,
			canPass:     canPass,
			canRetry:    canRetry,
			canRollback: canRollback,
			isLocalTask: isLocalTask,
			startTime:   startTime,
			endTime:     endTime,
		},
		executeTimes:  executeTimes,
		executerAgent: executerAgent,
		localAgentKey: executerAgent.String(),
	}

	taskInstance := reflect.New(TASK_TYPE[taskType]).Elem()
	v := taskInstance.FieldByName("Task")
	v.Set(reflect.ValueOf(task))

	return taskInstance.Addr().Interface().(ExecutableTask), nil
}
