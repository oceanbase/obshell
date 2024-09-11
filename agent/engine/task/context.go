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

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/meta"
)

const (
	EXECUTE_AGENTS           = "execute_agents"
	FAILURE_EXIT_MAINTENANCE = "failure_exit_maintenance"
)

type TaskContext struct {
	Params               map[string]interface{} // params can not be rewritten when merge context
	Data                 map[string]interface{} // global data will be rewritten when merge context
	AgentData            map[string]map[string]interface{}
	AgentDataUpdateCount map[string]int
}

func convertInterface(src interface{}, dest interface{}) error {
	jstr, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jstr, dest)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *TaskContext) GetParam(key string) interface{} {
	return ctx.Params[key]
}

func (ctx *TaskContext) GetData(key string) interface{} {
	return ctx.Data[key]
}

func (ctx *TaskContext) GetAgentData(agent meta.AgentInfoInterface, key string) interface{} {
	return ctx.GetAgentDataByAgentKey(fmt.Sprintf("%s:%d", agent.GetIp(), agent.GetPort()), key)
}

func (ctx *TaskContext) GetAgentDataByAgentKey(agentKey string, key string) interface{} {
	if ctx.AgentData[agentKey] == nil {
		return nil
	}
	return ctx.AgentData[agentKey][key]
}

func (ctx *TaskContext) GetParamWithValue(key string, value interface{}) error {
	v, ok := ctx.Params[key]
	if !ok {
		return fmt.Errorf("param `%s` not set", key)
	}
	if err := convertInterface(v, value); err != nil {
		return errors.Wrapf(err, "convert `%s` failed", key)
	}
	return nil
}

func (ctx *TaskContext) GetDataWithValue(key string, value interface{}) error {
	v, ok := ctx.Data[key]
	if !ok {
		return fmt.Errorf("data `%s` not set", key)
	}
	if err := convertInterface(v, value); err != nil {
		return errors.Wrapf(err, "convert `%s` failed", key)
	}
	return nil
}

func (ctx *TaskContext) GetAgentDataWithValue(agent meta.AgentInfoInterface, key string, value interface{}) error {
	return ctx.GetAgentDataByAgentKeyWithValue(agent.String(), key, value)
}

func (ctx *TaskContext) GetAgentDataByAgentKeyWithValue(agentKey string, key string, value interface{}) error {
	if ctx.AgentData[agentKey] == nil {
		return fmt.Errorf("agent %s data %s not set", agentKey, key)
	}
	v, ok := ctx.AgentData[agentKey][key]
	if !ok {
		return fmt.Errorf("agent %s data `%s` not set", agentKey, key)
	}
	if err := convertInterface(v, value); err != nil {
		return errors.Wrapf(err, "convert `%s` failed", key)
	}
	return nil
}

func (ctx *TaskContext) SetParam(key string, value interface{}) *TaskContext {
	ctx.Params[key] = value
	return ctx
}

func (ctx *TaskContext) SetData(key string, value interface{}) *TaskContext {
	ctx.Data[key] = value
	return ctx
}

func (ctx *TaskContext) SetAgentData(agent meta.AgentInfoInterface, key string, value interface{}) *TaskContext {
	return ctx.SetAgentDataByAgentKey(agent.String(), key, value)
}

func (ctx *TaskContext) SetAgentDataByAgentKey(agentKey string, key string, value interface{}) *TaskContext {
	if ctx.AgentData[agentKey] == nil {
		ctx.AgentData[agentKey] = make(map[string]interface{})
		ctx.AgentDataUpdateCount[agentKey] = 0
	}
	ctx.AgentData[agentKey][key] = value
	ctx.AgentDataUpdateCount[agentKey]++
	return ctx
}

func (ctx *TaskContext) MergeContext(other *TaskContext) {
	for k, v := range other.Params {
		if ctx.Params[k] == nil {
			ctx.Params[k] = v
		}
	}
	mergeMap(ctx.Data, other.Data)
	for k, v := range other.AgentDataUpdateCount {
		if v > ctx.AgentDataUpdateCount[k] {
			ctx.AgentData[k] = other.AgentData[k]
			ctx.AgentDataUpdateCount[k] = v
		}
	}
}

func (ctx *TaskContext) MergeContextWithoutExecAgents(other *TaskContext) {
	for k, v := range other.Params {
		if k == EXECUTE_AGENTS {
			continue
		}
		if ctx.Params[k] == nil {
			ctx.Params[k] = v
		}
	}
	mergeMap(ctx.Data, other.Data)
	for k, v := range other.AgentDataUpdateCount {
		if v > ctx.AgentDataUpdateCount[k] {
			ctx.AgentData[k] = other.AgentData[k]
			ctx.AgentDataUpdateCount[k] = v
		}
	}
}

func mergeMap(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	for k, v2 := range m2 {
		v1 := m1[k]
		if v1 == nil {
			m1[k] = v2
		} else {
			switch v1.(type) {
			case map[string]interface{}:
				m1[k] = mergeMap(v1.(map[string]interface{}), v2.(map[string]interface{}))
			default:
				m1[k] = v2
			}
		}
	}
	return m1
}

func NewTaskContext() *TaskContext {
	return &TaskContext{
		Params:               make(map[string]interface{}),
		Data:                 make(map[string]interface{}),
		AgentData:            make(map[string]map[string]interface{}),
		AgentDataUpdateCount: make(map[string]int),
	}
}
