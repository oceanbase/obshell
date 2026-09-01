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
	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/coordinator"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/agent/service/tenant"
	"github.com/oceanbase/obshell/ob/agent/service/user"
	"github.com/oceanbase/obshell/ob/param"
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
		return executeAgent, errors.Occur(errors.ErrObTenantNoActiveServer, tenantName)
	}
	return executeAgent, err
}

func PersistTenantRootPassword(c *gin.Context, tenantName, rootPassword string) error {
	maintainer, err := coordinator.GetMaintainer()
	if err != nil {
		return errors.Wrap(err, "get maintainer failed")
	}
	if maintainer.GetPort() == 0 || !maintainer.IsActive() {
		return errors.Occur(errors.ErrAgentMaintainerNotActive)
	}
	return persistTenantRootPassword(c, tenantName, rootPassword, &maintainer)
}

func persistTenantRootPassword(c *gin.Context, tenantName, rootPassword string, maintainer meta.AgentInfoInterface) error {
	// check password connectable by calling precheck api with password
	body := &param.TenantRootPasswordParam{
		RootPassword: &rootPassword,
	}
	uri := constant.URI_API_V1 + constant.URI_TENANT + "/" + tenantName + constant.URI_PRECHECK
	result := &bo.ObTenantPreCheckResult{}
	err := secure.SendGetRequest(maintainer, uri, body, result)
	if err != nil {
		return errors.Wrap(err, "Failed to check tenant connectable using password.")
	}
	if !result.IsConnectable {
		return errors.Occur(errors.ErrObTenantRootPasswordIncorrect)
	}
	tenant.GetPasswordMap().Set(tenantName, rootPassword)
	return nil
}

func persistTenantRootPasswordOnMaintainer(c *gin.Context, tenantName, rootPassword string) error {
	maintainer, err := coordinator.GetMaintainer()
	if err != nil {
		return errors.Wrap(err, "get maintainer failed")
	}
	if maintainer.GetPort() == 0 || !maintainer.IsActive() {
		return errors.Occur(errors.ErrAgentMaintainerNotActive)
	}
	if meta.OCS_AGENT.Equal(&maintainer) {
		return persistTenantRootPassword(c, tenantName, rootPassword, &maintainer)
	}

	uri := constant.URI_API_V1 + constant.URI_TENANT + "/" + tenantName + constant.URI_ROOTPASSWORD + constant.URI_PERSIST
	if err := secure.SendPostRequest(&maintainer, uri, param.PersistTenantRootPasswordParam{
		Password: rootPassword,
	}, nil); err != nil {
		return errors.Wrap(err, "persist tenant root password on maintainer failed")
	}
	return nil
}

func ModifyTenantRootPassword(c *gin.Context, tenantName string, pwdParam param.ModifyTenantRootPasswordParam) (error, bool) {
	if tenantName == constant.TENANT_SYS {
		return errors.Occur(errors.ErrObTenantSysOperationNotAllowed), false
	}
	executeAgent, err := GetExecuteAgentForTenant(tenantName)
	if err != nil {
		return errors.Wrap(err, "get execute agent failed"), false
	}

	if meta.OCS_AGENT.Equal(executeAgent) {
		db, err := GetConnectionWithPassword(tenantName, &pwdParam.OldPwd)
		passwordAlreadyChanged := false
		if err != nil {
			oldPasswordErr := err
			// The previous request may have changed the database password but failed
			// while persisting it in OBShell. Accept the desired password on retry so
			// the operation can finish instead of retrying forever with the old one.
			db, err = GetConnectionWithPassword(tenantName, pwdParam.NewPwd)
			if err != nil {
				return oldPasswordErr, false
			}
			passwordAlreadyChanged = true
		}
		defer CloseDbConnection(db)

		if !passwordAlreadyChanged {
			userService := user.GetUserService(db)
			if err := userService.ModifyTenantRootPassword(*pwdParam.NewPwd); err != nil {
				return err, false
			}
		}
		if err := persistTenantRootPasswordOnMaintainer(c, tenantName, *pwdParam.NewPwd); err != nil {
			return errors.Wrap(err, "tenant root password was changed, but failed to save it in OBShell"), false
		}
	} else {
		common.ForwardRequest(c, executeAgent, pwdParam)
		return nil, true
	}
	return nil, false
}

func newSetRootPwdNode(newPwd string) (*task.Node, error) {
	ctx := task.NewTaskContext().
		SetParam(PARAM_TENANT_NEW_PASSWORD, newPwd)
	return task.NewNodeWithContext(newSetRootPwdTask(), false, ctx), nil
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
		return err
	}
	t.ExecuteLogf("Set root password for tenant '%s'", t.tenantName)

	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NEW_PASSWORD, &t.newPassword); err != nil {
		return err
	}

	executeAgent, err := tenantService.GetTenantActiveAgent(t.tenantName)
	if err != nil {
		return err
	}
	if executeAgent == nil {
		return errors.Occur(errors.ErrObTenantNoActiveServer, t.tenantName)
	}

	if err := secure.SendPutRequest(executeAgent, constant.URI_API_V1+constant.URI_TENANT+"/"+t.tenantName+constant.URI_ROOTPASSWORD, param.ModifyTenantRootPasswordParam{
		NewPwd: &t.newPassword,
	}, nil); err != nil {
		return errors.Wrap(err, "set root password failed")
	}

	return nil
}
