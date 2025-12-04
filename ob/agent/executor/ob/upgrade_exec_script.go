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

package ob

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

type ExecScriptTask struct {
	task.Task
	localAgent    meta.AgentInfo
	realExecAgent meta.AgentInfo
	scriptFile    string
	zone          string
	rpmDir        string
	scriptPath    string
}

func (t *ExecScriptTask) getExecAgent() (err error) {
	_, t.realExecAgent, err = isRealExecuteAgent(t)
	if err != nil {
		return err
	}
	return nil
}

func (t *ExecScriptTask) getParams() (err error) {
	t.localAgent = t.GetExecuteAgent()
	if err = t.GetContext().GetParamWithValue(PARAM_SCRIPT_FILE, &t.scriptFile); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_ZONE, &t.zone); err != nil {
		return err
	}

	var upgradeRouteIndex float64
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_ROUTE_INDEX, &upgradeRouteIndex); err != nil {
		return err
	}
	upgradeRoute, err := getUpgradeRouteForTask(t.GetContext())
	if err != nil {
		return err
	}
	node := upgradeRoute[int(upgradeRouteIndex)]
	var rpmPkgInfo rpmPacakgeInstallInfo
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(t.realExecAgent.String(), node.BuildVersion, &rpmPkgInfo); err != nil {
		return err
	}
	t.rpmDir = rpmPkgInfo.RpmDir
	t.scriptPath = filepath.Join(rpmPkgInfo.RpmPkgHomepath, t.scriptFile)
	return nil
}

func (t *ExecScriptTask) Execute() (err error) {
	t.ExecuteLog("get real execute agent")
	if err = t.getExecAgent(); err != nil {
		return err
	}
	if err = t.getParams(); err != nil {
		return err
	}
	t.ExecuteLogf("execute script '%s' %s", t.scriptFile, t.zone)
	str := fmt.Sprintf("cd %s; python %s -h%s -P%d -uroot", t.rpmDir, t.scriptPath, t.localAgent.Ip, meta.MYSQL_PORT)
	if meta.GetOceanbasePwd() != "" {
		pwd := strings.ReplaceAll(meta.GetOceanbasePwd(), "'", "'\"'\"'")
		str = fmt.Sprintf("%s -p'%s'", str, pwd)
	}
	if t.zone != "" {
		str = fmt.Sprintf("%s -z'%s'", str, t.zone)
	}
	cmd := exec.Command("/bin/bash", "-c", str)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "execute script '%s' error, execute logs:\n%s", t.scriptFile, string(output))
	}

	return nil
}
