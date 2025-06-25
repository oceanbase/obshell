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
	"strings"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
)

const (
	defaultWhitelist = "127.0.0.1"
)

func ModifyTenantWhitelist(tenantName string, whitelist string) error {
	if _, err := checkTenantExistAndStatus(tenantName); err != nil {
		return err
	}
	if err := tenantService.ModifyTenantWhitelist(tenantName, mergeWhitelist(whitelist)); err != nil {
		return err
	}
	return nil
}

// mergeWhitelist merge s
func mergeWhitelist(specific string) string {
	if specific == "" {
		return defaultWhitelist
	}
	splits := strings.Split(specific, ",")
	splits = append(splits, defaultWhitelist)
	whitelistMap := make([]string, 0)
	// 去重
	unique := make(map[string]struct{})
	for _, item := range splits {
		if _, ok := unique[item]; !ok {
			unique[item] = struct{}{}
			whitelistMap = append(whitelistMap, item)
		}
	}
	return strings.Join(whitelistMap, ",")
}

type ModifyTenantWhitelistTask struct {
	task.Task
	tenantName string
	whitelist  string
}

func newModifyTenantWhitelistNode(whitelist string) *task.Node {
	ctx := task.NewTaskContext().SetParam(PARAM_TENANT_WHITELIST, whitelist)
	return task.NewNodeWithContext(newModifyTenantWhitelistTask(), false, ctx)
}

func newModifyTenantWhitelistTask() *ModifyTenantWhitelistTask {
	newTask := &ModifyTenantWhitelistTask{
		Task: *task.NewSubTask(Task_NAME_MODIFY_WHITELIST),
	}
	newTask.SetCanRollback().SetCanRetry().SetCanCancel() // could not pass, even if the task is passed, the tenant is still cannot use.
	return newTask
}

func (t *ModifyTenantWhitelistTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_WHITELIST, &t.whitelist); err != nil {
		return err
	}
	if err := tenantService.ModifyTenantWhitelist(t.tenantName, mergeWhitelist(t.whitelist)); err != nil {
		return errors.Wrapf(err, "modify tenant whitelist failed")
	}
	return nil
}
