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

package tenant

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
)

func isUnkonwnTimeZoneErr(err error) bool {
	return err != nil && err.Error() == "unknown time zone"
}

func GetTenantVariables(tenantName string, filter string) ([]oceanbase.CdbObSysVariable, error) {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return nil, err
	}
	if filter == "" {
		filter = "%"
	}
	variables, err := tenantService.GetTenantVariables(tenantName, filter)
	if err != nil {
		return nil, err
	}
	return variables, nil
}

func GetTenantVariable(tenantName string, variableName string) (*oceanbase.CdbObSysVariable, error) {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return nil, err
	}
	variable, err := tenantService.GetTenantVariable(tenantName, variableName)
	if err != nil {
		return nil, err
	}
	if variable == nil {
		return nil, errors.Occur(errors.ErrObTenantVariableNotExist, variableName)
	}
	return variable, nil
}

func SetTenantVariables(c *gin.Context, tenantName string, param param.SetTenantVariablesParam) error {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return err
	}
	for k, v := range param.Variables {
		if k == "" || v == nil {
			return errors.Occur(errors.ErrObTenantEmptyVariable)
		}
	}
	transferNumber(param.Variables)

	needConnectTenant := false
	for k := range param.Variables {
		if utils.ContainsString(constant.VARIAbLES_COLLATION_OR_CHARACTER, k) {
			needConnectTenant = true
			break
		}
	}

	if !needConnectTenant {
		if err := tenantService.SetTenantVariables(tenantName, param.Variables); err != nil {
			if errors.IsUnkonwnTimeZoneErr(err) {
				if value, exist := param.Variables[constant.VARIABLE_TIME_ZONE]; exist {
					return timeZoneErrorReporter(value, err)
				}
			}
			return errors.Wrap(err, "set tenant variables failed")
		}
	} else {
		executeAgent, err := GetExecuteAgentForTenant(tenantName)
		if err != nil {
			return errors.Wrap(err, "get execute agent failed")
		}

		if meta.OCS_AGENT.Equal(executeAgent) {
			if err := tenantService.SetTenantVariablesWithTenant(tenantName, param.TenantPassword, param.Variables); err != nil {
				return err
			}
		} else {
			common.ForwardRequest(c, executeAgent, param)
			return nil
		}
	}

	return nil
}

func timeZoneErrorReporter(timeZone interface{}, err error) error {
	if v, ok := timeZone.(string); ok {
		pattern := `^[A-Za-z]+/[A-Za-z]+$`
		re := regexp.MustCompile(pattern)
		if re.MatchString(v) {
			if empty, _ := tenantService.IsTimeZoneTableEmpty(); empty {
				return errors.Occur(errors.ErrCommonUnexpected, errors.Wrapf(err, "Please check whether the sys tenant has been import time zone info").Error())
			}
		}
	}
	return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "time_zone", err.Error())
}

type SetTenantVariableTask struct {
	task.Task
	variables  map[string]interface{}
	tenantName string
}

func newSetTenantVariableNode(variables map[string]interface{}) (*task.Node, error) {
	agents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, err
	}
	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_TENANT_VARIABLES, variables)
	return task.NewNodeWithContext(newSetTenantVariableTask(), true, ctx), nil
}

func newSetTenantVariableTask() *SetTenantVariableTask {
	newTask := &SetTenantVariableTask{
		Task: *task.NewSubTask(TASK_NAME_SET_TENANT_VARIABLE),
	}

	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel()
	return newTask
}

func (t *SetTenantVariableTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_VARIABLES, &t.variables); err != nil {
		return err
	}

	executeAgent, err := tenantService.GetTenantActiveAgent(t.tenantName)
	if err != nil {
		return err
	}
	if executeAgent == nil {
		return errors.Occur(errors.ErrObTenantNoActiveServer, t.tenantName)
	}

	if meta.OCS_AGENT.Equal(executeAgent) {
		transferNumber(t.variables)
		if err := tenantService.SetTenantVariablesWithTenant(t.tenantName, "", t.variables); err != nil {
			return errors.Wrap(err, "set tenant variables failed")
		}
	}
	return nil
}
