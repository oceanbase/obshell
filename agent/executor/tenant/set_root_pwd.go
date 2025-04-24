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
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/service/tenant"
	"github.com/oceanbase/obshell/param"
)

type SetRootPwdTask struct {
	task.Task
	tenantName  string
	newPassword string
}

func GetExecuteAgentForTenant(tenantName string) (meta.AgentInfoInterface, error) {
	isTenantOn, err := tenantService.IsTenantActiveAgent(tenantName, meta.OCS_AGENT.GetIp(), meta.RPC_PORT)
	if err != nil {
		return nil, err
	}
	if isTenantOn {
		return meta.OCS_AGENT, nil
	}
	executeAgent, err := tenantService.GetTenantActiveAgent(tenantName)
	if err != nil {
		return nil, err
	}
	if executeAgent == nil {
		return executeAgent, errors.New("tenant is not active")
	}
	return executeAgent, err
}

func PersistTenantRootPassword(c *gin.Context, tenantName, rootPassword string) {
	tenant.GetPasswordMap().Set(tenantName, rootPassword)
}

func ModifyTenantRootPassword(c *gin.Context, tenantName string, pwdParam param.ModifyTenantRootPasswordParam) (*errors.OcsAgentError, bool) {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return err, false
	}
	if tenantName == constant.TENANT_SYS {
		return errors.Occur(errors.ErrIllegalArgument, "Can not modify root password for sys tenant."), false
	}
	executeAgent, err := GetExecuteAgentForTenant(tenantName)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "get execute agent failed: %s", err.Error()), false
	}

	if meta.OCS_AGENT.Equal(executeAgent) {
		if err := tenantService.ModifyTenantRootPassword(tenantName, pwdParam.OldPwd, *pwdParam.NewPwd); err != nil {
			return err, false
		}
	} else {
		common.ForwardRequest(c, executeAgent, pwdParam)
		return nil, true
	}
	return nil, false
}

func newSetRootPwdNode(newPwd string) (*task.Node, error) {
	agents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, errors.Wrap(err, "create set root password task failed")
	}
	ctx := task.NewTaskContext().
		SetParam(PARAM_TENANT_NEW_PASSWORD, newPwd).
		SetParam(task.EXECUTE_AGENTS, agents)
	return task.NewNodeWithContext(newSetRootPwdTask(), true, ctx), nil
}

func newSetRootPwdTask() *SetRootPwdTask {
	newTask := &SetRootPwdTask{
		Task: *task.NewSubTask(TASK_NAME_SET_ROOT_PWD),
	}

	newTask.SetCanContinue().SetCanRollback().SetCanRetry().SetCanCancel()
	return newTask
}

func (t *SetRootPwdTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return errors.Wrap(err, "Get tenant name failed")
	}
	t.ExecuteLogf("Set root password for tenant '%s'", t.tenantName)

	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NEW_PASSWORD, &t.newPassword); err != nil {
		return errors.Wrap(err, "Get tenant new password failed")
	}

	executeAgent, err := tenantService.GetTenantActiveAgent(t.tenantName)
	if err != nil {
		return err
	}
	if executeAgent == nil {
		return errors.New("tenant is not active")
	}

	if meta.OCS_AGENT.Equal(executeAgent) {
		if err := tenantService.ModifyTenantRootPassword(t.tenantName, "", t.newPassword); err != nil {
			return errors.Occurf(errors.ErrUnexpected, "modify tenant root password failed: %s", err.Error())
		}
	}
	tenant.GetPasswordMap().Set(t.tenantName, t.newPassword)
	return nil
}
