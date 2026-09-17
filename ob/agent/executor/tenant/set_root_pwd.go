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
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
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
	allowForward := true
	if c != nil {
		if value, exists := c.Get(constant.OCS_HEADER); exists {
			if header, ok := value.(secure.HttpHeader); ok {
				allowForward = header.ForwardType == secure.NotForward
			}
		}
	}
	return newPasswordPersister().persist(passwordRequestContext(c), tenantName, rootPassword, allowForward)
}

func persistTenantRootPasswordOnMaintainer(c *gin.Context, tenantName, rootPassword string) error {
	return newPasswordPersister().persist(passwordRequestContext(c), tenantName, rootPassword, true)
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
		if err != nil {
			return err, false
		}
		defer CloseDbConnection(db)

		// Reject known maintainer outages before changing database credentials.
		if _, err := getPasswordMaintainer(passwordRequestContext(c)); err != nil {
			return err, false
		}
		userService := user.GetUserService(db)
		if err := userService.ModifyTenantRootPassword(*pwdParam.NewPwd); err != nil {
			return err, false
		}
		if err := newPasswordPersister().persist(passwordCompletionContext(c), tenantName, *pwdParam.NewPwd, true); err != nil {
			uri := constant.URI_TENANT_API_PREFIX + "/" + tenantName + constant.URI_ROOTPASSWORD + constant.URI_PERSIST
			return errors.WrapOverride(errors.ErrObTenantPasswordCacheSyncFailed, err, uri), false
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
	return t.executePasswordChange(func() error {
		return persistTenantRootPasswordOnMaintainer(nil, t.tenantName, t.newPassword)
	}, t.modifyPassword)
}

// Per-call operations make task recovery testable without replacing shared
// package functions. Only a repeated/continued task can use desired-password
// recovery; a public PUT never calls this method.
func (t *SetRootPwdTask) executePasswordChange(persist, modify func() error) error {
	// Only this persisted create-tenant task may recover an interrupted change
	// by checking its desired password. The public PUT API always checks old_password.
	if t.GetExecuteTimes() > 1 || t.IsContinue() {
		if err := persist(); err == nil {
			return nil
		} else if !isTenantPasswordIncorrect(err) {
			return errors.Wrap(err, "resume tenant password cache synchronization failed")
		}
	}
	return modify()
}

func (t *SetRootPwdTask) modifyPassword() error {
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
