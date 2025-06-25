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
	"math/big"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
)

type AdditionalDataDTO struct {
	AdditionalData *map[string]any `json:"additional_data"`
	additionalData map[string]any
}

type TaskStatusDTO struct {
	State     string    `json:"state"`
	Operator  string    `json:"operator"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type TaskDetail struct {
	TaskID int64  `json:"task_id" uri:"task_id"`
	Name   string `json:"name"`
	TaskStatusDTO
	AdditionalDataDTO
	ExecuteTimes int            `json:"execute_times"`
	ExecuteAgent meta.AgentInfo `json:"execute_agent"`
	TaskLogs     []string       `json:"task_logs"`
}

type NodeDetail struct {
	NodeID int64  `json:"node_id" uri:"node_id"`
	Name   string `json:"name"`
	TaskStatusDTO
	AdditionalDataDTO
	SubTasks []*TaskDetailDTO `json:"sub_tasks"`
}

type TaskDetailDTO struct {
	*GenericDTO
	*TaskDetail
}

type NodeDetailDTO struct {
	*GenericDTO
	*NodeDetail
}

type DagDetailDTO struct {
	*GenericDTO
	*DagDetail
}

type DagDetail struct {
	DagID           int64  `json:"dag_id" uri:"dag_id"`
	Name            string `json:"name"`
	Stage           int    `json:"stage"`
	MaxStage        int    `json:"max_stage"`
	MaintenanceType int    `json:"maintenance_type"`
	MaintenanceKey  string `json:"maintenance_key"`
	TaskStatusDTO
	AdditionalDataDTO
	Nodes []*NodeDetailDTO `json:"nodes"`
}

type TaskExecuteLogDTO struct {
	TaskId       int64  `json:"task_id" binding:"required,min=1"`
	ExecuteTimes int    `json:"execute_times" binding:"required,min=1"`
	LogContent   string `json:"log_content" binding:"required"`
	IsSync       bool   `json:"is_sync"`
}

type NodeOperator struct {
	NodeDetailDTO
	Operator string `json:"operator" binding:"required"`
}

type DagOperator struct {
	DagDetailDTO
	Operator string `json:"operator" binding:"required"`
}

type GenericDTO struct {
	GenericID string `json:"id" uri:"id" binding:"required"`
}

func (a *AdditionalDataDTO) SetVisible(visible bool) {
	if visible {
		a.AdditionalData = &a.additionalData
	} else {
		a.AdditionalData = nil
	}
}

func (a *NodeDetailDTO) SetVisible(visible bool) {
	if !visible || a.SubTasks == nil || len(a.SubTasks) == 0 {
		a.AdditionalData = nil
		return
	}

	if a.additionalData == nil {
		data := make([]*AdditionalDataDTO, 0)
		for _, task := range a.SubTasks {
			task.SetVisible(false)
			data = append(data, &task.AdditionalDataDTO)
		}
		a.AdditionalDataDTO = mergeAdditionalData(data)
	}
	a.AdditionalData = &a.additionalData
}

func (a *DagDetailDTO) SetVisible(visible bool) {
	if !visible || a.Nodes == nil || len(a.Nodes) == 0 {
		a.AdditionalData = nil
		return
	}

	if a.additionalData == nil {
		data := make([]*AdditionalDataDTO, 0)
		for _, node := range a.Nodes {
			node.SetVisible(true)
			data = append(data, &node.AdditionalDataDTO)
			node.SetVisible(false)
		}
		a.AdditionalDataDTO = mergeAdditionalData(data)
	}
	a.AdditionalData = &a.additionalData
}

func mergeAdditionalData(a []*AdditionalDataDTO) AdditionalDataDTO {
	if len(a) == 0 {
		return AdditionalDataDTO{}
	}
	if len(a) == 1 {
		return *a[0]
	}
	result := AdditionalDataDTO{
		additionalData: make(map[string]any),
	}
	for _, item := range a {
		for k, v := range item.additionalData {
			r := result.additionalData[k]
			result.additionalData[k] = mergeValue(r, v)
		}
	}
	return result
}

func mergeValue(a any, b any) any {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)
	if typeA != typeB {
		panic(fmt.Sprintf("type of additional data is not same, a is %s, b is %s", typeA, typeB))
	}

	switch typeA.Kind() {
	case reflect.Slice:
		v1 := a.([]any)
		v2 := b.([]any)
		return append(v1, v2...)
	case reflect.Map:
		v1 := a.(map[string]any)
		v2 := b.(map[string]any)
		for k, v := range v2 {
			v1[k] = mergeValue(v1[k], v)
		}
	default:
		return b
	}
	return a
}

func NewTaskStatusDTO(task *TaskInfo) *TaskStatusDTO {
	return &TaskStatusDTO{
		State:     STATE_MAP[task.GetState()],
		Operator:  OPERATOR_MAP[task.GetOperator()],
		StartTime: task.GetStartTime(),
		EndTime:   task.GetEndTime(),
	}
}

func NewDagDetailDTO(dag *Dag) *DagDetailDTO {
	return &DagDetailDTO{
		GenericDTO: newGenericDTO(dag, dag.GetDagType()),
		DagDetail:  NewDagDetail(dag),
	}
}

func NewNodeDetailDTO(node *Node, dagType string) *NodeDetailDTO {
	return &NodeDetailDTO{
		GenericDTO: newGenericDTO(node, dagType),
		NodeDetail: NewNodeDetail(node),
	}
}

func NewTaskDetailDTO(task ExecutableTask, dagType string) *TaskDetailDTO {
	return &TaskDetailDTO{
		GenericDTO: newGenericDTO(task, dagType),
		TaskDetail: NewTaskDetail(task),
	}
}

func newGenericDTO(instance TaskInfoInterface, dagType string) *GenericDTO {
	return &GenericDTO{
		GenericID: ConvertToGenericID(instance, dagType),
	}
}

func NewDagDetail(dag *Dag) *DagDetail {
	return &DagDetail{
		DagID:           dag.GetID(),
		Name:            dag.GetName(),
		Stage:           dag.GetStage(),
		MaxStage:        dag.GetMaxStage(),
		MaintenanceType: dag.GetMaintenanceType(),
		MaintenanceKey:  dag.GetMaintenanceKey(),
		TaskStatusDTO:   *NewTaskStatusDTO(&dag.TaskInfo),
	}
}

func NewNodeDetail(node *Node) *NodeDetail {
	return &NodeDetail{
		NodeID:        node.GetID(),
		Name:          node.GetName(),
		TaskStatusDTO: *NewTaskStatusDTO(&node.TaskInfo),
	}
}

func NewTaskDetail(task ExecutableTask) *TaskDetail {
	taskDetailDTO := &TaskDetail{
		TaskID:       task.GetID(),
		Name:         task.GetName(),
		ExecuteTimes: task.GetExecuteTimes(),
		ExecuteAgent: task.GetExecuteAgent(),
		TaskStatusDTO: TaskStatusDTO{
			State:     STATE_MAP[task.GetState()],
			Operator:  OPERATOR_MAP[task.GetOperator()],
			StartTime: task.GetStartTime(),
			EndTime:   task.GetEndTime(),
		},
		AdditionalDataDTO: AdditionalDataDTO{
			AdditionalData: nil,
			additionalData: task.GetAdditionalData(),
		},
	}
	return taskDetailDTO
}

// ConvertToGenericID will convert task instance id to generic dto id.
func ConvertToGenericID(instance TaskInfoInterface, dagType string) string {
	if instance.IsLocalTask() {
		return ConvertLocalIDToGenericID(instance.GetID(), dagType)
	}
	return fmt.Sprintf("1%d", instance.GetID())
}

func ConvertIDToGenericID(dagID int64, isLocal bool, dagType string) string {
	if isLocal {
		return ConvertLocalIDToGenericID(dagID, dagType)
	} else {
		return fmt.Sprintf("1%d", dagID)
	}
}

func ConvertObproxyIDToGenericID(id int64) string {
	ipParsed := net.ParseIP(meta.OCS_AGENT.GetIp())
	if ipParsed.To4() != nil {
		bigInt := new(big.Int).SetBytes(ipParsed.To4())
		return fmt.Sprintf("4%010d%05d%d", bigInt, meta.OCS_AGENT.GetPort(), id)
	} else {
		bigInt := new(big.Int).SetBytes(ipParsed.To16())
		return fmt.Sprintf("5%039d%05d%d", bigInt, meta.OCS_AGENT.GetPort(), id)
	}
}

// ConvertLocalIDToGenericID will convert id of local task to generic id.
func ConvertLocalIDToGenericID(id int64, dagType string) string {
	if DAG_TYPE_MAP[DAG_OBPROXY] == dagType {
		return ConvertObproxyIDToGenericID(id)
	}
	ipParsed := net.ParseIP(meta.OCS_AGENT.GetIp())
	if ipParsed.To4() != nil {
		bigInt := new(big.Int).SetBytes(ipParsed.To4())
		return fmt.Sprintf("2%010d%05d%d", bigInt, meta.OCS_AGENT.GetPort(), id)
	} else {
		bigInt := new(big.Int).SetBytes(ipParsed.To16())
		return fmt.Sprintf("3%039d%05d%d", bigInt, meta.OCS_AGENT.GetPort(), id)
	}
}

func IsObproxyTask(genericID string) bool {
	return genericID[0] == constant.OBPROXY_TASK_IPV4_ID_PREFIX || genericID[0] == constant.OBPROXY_TASK_IPV6_ID_PREFIX
}

// ConvertGenericID will  onvert dto id to instance id.
func ConvertGenericID(genericID string) (id int64, agent meta.AgentInfoInterface, err error) {
	if genericID[0] == constant.CLUSTER_TASK_ID_PREFIX && len(genericID) <= 1 ||
		(genericID[0] == constant.LOCAL_TASK_IPV4_ID_PREFIX ||
			genericID[0] == constant.OBPROXY_TASK_IPV4_ID_PREFIX) && len(genericID) <= 16 ||
		(genericID[0] == constant.LOCAL_TASK_IPV6_ID_PREFIX ||
			genericID[0] == constant.OBPROXY_TASK_IPV6_ID_PREFIX) && len(genericID) <= 45 {
		err = errors.Occur(errors.ErrTaskGenericIDInvalid, genericID)
		return
	}
	var idIdx, ipIdx, portIdx int
	var isV6 bool

	switch genericID[0] {
	case constant.CLUSTER_TASK_ID_PREFIX:
		idIdx = 1
	case constant.LOCAL_TASK_IPV4_ID_PREFIX, constant.OBPROXY_TASK_IPV4_ID_PREFIX:
		// Ipv4 address.
		ipIdx, portIdx, idIdx = 11, 16, 16
	case constant.LOCAL_TASK_IPV6_ID_PREFIX, constant.OBPROXY_TASK_IPV6_ID_PREFIX:
		// Ipv6 address.
		ipIdx, portIdx, idIdx, isV6 = 40, 45, 45, true
	default:
		err = errors.OccurWithMessage("invalid generic id", errors.ErrTaskGenericIDInvalid, genericID)
		return
	}

	if ipIdx > 0 {
		ipInt, ok := new(big.Int).SetString(genericID[1:ipIdx], 10)
		if !ok {
			err = errors.OccurWithMessage("convert id to bigInt failed", errors.ErrTaskGenericIDInvalid, genericID)
			return
		}
		netIp := net.IP(ipInt.Bytes())
		if isV6 {
			netIp = netIp.To16()
		} else {
			netIp = netIp.To4()
		}
		if netIp == nil {
			err = errors.OccurWithMessage("convert id to ip failed", errors.ErrTaskGenericIDInvalid, genericID)
			return
		}
		port, perr := strconv.Atoi(genericID[ipIdx:portIdx])
		if perr != nil {
			err = errors.OccurWithMessage("parse port failed", errors.ErrTaskGenericIDInvalid, genericID)
			return
		}

		agent = meta.NewAgentInfo(netIp.String(), port)
	}

	id, perr := strconv.ParseInt(genericID[idIdx:], 10, 64)
	if perr != nil {
		err = errors.WrapRetain(errors.ErrTaskGenericIDInvalid, perr, genericID)
	}
	return
}

func (t *TaskStatusDTO) IsFailed() bool {
	return t.State == FAILED_STR
}

func (t *TaskStatusDTO) IsRunning() bool {
	return t.State == RUNNING_STR
}

func (t *TaskStatusDTO) IsSucceed() bool {
	return t.State == SUCCEED_STR
}

func (t *TaskStatusDTO) IsPending() bool {
	return t.State == PENDING_STR
}

func (t *TaskStatusDTO) IsReady() bool {
	return t.State == READY_STR
}

func (t *TaskStatusDTO) IsFinished() bool {
	return t.IsFailed() || t.IsSucceed()
}
