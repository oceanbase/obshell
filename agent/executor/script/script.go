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

package script

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/service/tenant"
)

type ImportScriptForTenantTask struct {
	task.Task
	tenantName      string // tenant name
	alwaysSuccess   bool   // if true, the task will always return success, even if the import process fails
	parallelExecute bool
}

const (
	TASK_NAME_IMPORT_SCRIPT     = "Import script for tenant"
	MYSQL_CONNECTOR             = "mysql.connector"
	PARAM_TENANT_NAME           = "tenantName"
	PARAM_IMPORT_ALWAYS_SUCCESS = "alwaysSuccess"
	PARAM_PARALLEL_EXECUTE      = "parallelExecute"
)

var modules = []string{MYSQL_CONNECTOR}
var tenantService = tenant.TenantService{}

// alwaysSuccess: if true, the task will always return success, even if the import process fails
// alwaysSuccess should be true only in take over and upgrade.
func NewParallelImportScriptForTenantNode(executeAgents []meta.AgentInfo, alwaysSuccess bool) *task.Node {
	context := task.NewTaskContext().
		SetParam(PARAM_IMPORT_ALWAYS_SUCCESS, alwaysSuccess).
		SetParam(task.EXECUTE_AGENTS, executeAgents).
		SetParam(PARAM_PARALLEL_EXECUTE, true)
	return task.NewNodeWithContext(newImportScriptForTenantTask(), true, context)

}
func NewImportScriptForTenantNode(alwaysSuccess bool) *task.Node {
	context := task.NewTaskContext().
		SetParam(PARAM_IMPORT_ALWAYS_SUCCESS, alwaysSuccess).
		SetParam(PARAM_PARALLEL_EXECUTE, false)
	return task.NewNodeWithContext(newImportScriptForTenantTask(), false, context)

}

func newImportScriptForTenantTask() *ImportScriptForTenantTask {
	newTask := &ImportScriptForTenantTask{
		Task: *task.NewSubTask(TASK_NAME_IMPORT_SCRIPT),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel().SetCanPass()
	return newTask
}

// If the python environment is not installed, the task will do nothing
func (t *ImportScriptForTenantTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_IMPORT_ALWAYS_SUCCESS, &t.alwaysSuccess); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_PARALLEL_EXECUTE, &t.parallelExecute); err != nil {
		return err
	}

	server, err := tenantService.GetTenantActiveServer(t.tenantName)
	if err != nil {
		return errors.Wrap(err, "Get tenant active server failed")
	}
	if t.parallelExecute && (server.SvrIp != meta.OCS_AGENT.GetIp() || server.SqlPort != meta.MYSQL_PORT) {
		return nil
	}

	/* check env */
	t.ExecuteLog("Checking if python is installed.")
	cmd := exec.Command("python", "-c", "import sys; print(sys.version_info.major)")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return errors.New("Python is not installed, please install it first.")
	}
	output := strings.TrimSpace(out.String())
	t.ExecuteLogf("Python major version %s", output)

	for _, module := range modules {
		t.ExecuteLogf("Checking if python module '%s' is installed.", module)
		cmd = exec.Command("python", "-c", "import "+module)
		if err := cmd.Run(); err != nil {
			return errors.New("Python module not installed, please install it first.")
		}
	}

	/* import timezone info */
	t.ExecuteLog("Importing timezone info.")
	if err != nil {
		return errors.Wrap(err, "Check timezone table failed")
	}
	pwd := strings.ReplaceAll(meta.GetOceanbasePwd(), "'", "'\"'\"'")
	str := fmt.Sprintf("%s -h%s -P%d -t%s -f%s", path.ImportTimeZoneInfoScriptPath(), constant.LOCAL_IP, server.SqlPort, t.tenantName, path.ImportTimeZoneInfoFilePath())
	if meta.GetOceanbasePwd() != "" {
		str = fmt.Sprintf("%s -p'%s'", str, pwd)
	}
	cmd = exec.Command("/bin/bash", "-c", str)
	if res, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("Import timezone info failed: %s", string(res))
	}

	/* import srs data */
	if t.tenantName == constant.SYS_TENANT {
		return nil
	}
	t.ExecuteLog("Importing srs data.")
	str = fmt.Sprintf("%s -h%s -P%d -t%s -f%s", path.ImportSrsDataScriptPath(), constant.LOCAL_IP, server.SqlPort, t.tenantName, path.ImportSrsDataFilePath())
	if meta.GetOceanbasePwd() != "" {
		str = fmt.Sprintf("%s -p'%s'", str, pwd)
	}
	cmd = exec.Command("/bin/bash", "-c", str)
	if res, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("Import srs data failed: %s", string(res))
	}
	return nil
}
