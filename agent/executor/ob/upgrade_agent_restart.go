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
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
)

type RestartAgentTask struct {
	task.Task

	realExecAgent      meta.AgentInfo
	targetBuildVersion string
	backupDir          string
	prevBuildVersion   string
}

func newRestartAgentTask() *RestartAgentTask {
	newTask := &RestartAgentTask{
		Task: *task.NewSubTask(TASK_RESTART_OBSHELL),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanCancel()
	return newTask
}

func (t *RestartAgentTask) Execute() (err error) {
	if t.IsContinue() {
		t.ExecuteLog("restart agent task continue")
		if err = t.checkVersion(); err != nil {
			return
		}
		t.ExecuteLog("Connecting to the database.")
		if err = getOcsInstance(); err != nil {
			return err
		}
		t.ExecuteLog("Updating version.")
		if err = agentService.UpdateAgentVersion(); err != nil {
			return
		}
		return nil
	}
	t.ExecuteLog("restart agent")
	if err = t.restartAgent(); err != nil {
		return
	}
	t.ExecuteLog("restart agent finished")
	return nil
}

func getOcsInstance() (err error) {
	for i := 1; i <= constant.MAX_GET_INSTANCE_RETRIES; i++ {
		if _, err = oceanbase.GetOcsInstance(); err == nil {
			return nil
		}
		log.Infof("get ocs instance failed: %v , retry [%d/%d]", err, i, constant.MAX_GET_INSTANCE_RETRIES)
		time.Sleep(time.Second * constant.GET_INSTANCE_RETRY_INTERVAL)
	}
	return errors.New("get ocs instance timeout")
}

func (t *RestartAgentTask) getParams() (err error) {
	_, t.realExecAgent, err = isRealExecuteAgent(t)
	if err != nil {
		return err
	}
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(t.realExecAgent.String(), DATA_BACKUP_DIR, &t.backupDir); err != nil {
		return err
	}
	return nil
}

func (t *RestartAgentTask) getPrevAgentVersion() (err error) {
	if err := t.getParams(); err != nil {
		return err
	}

	backAgentPath := filepath.Join(t.backupDir, constant.PROC_OBSHELL)
	t.prevBuildVersion, err = system.GetBinaryVersion(backAgentPath)
	if err != nil {
		return errors.Wrapf(err, "get binary version failed %s", backAgentPath)
	}
	t.ExecuteLogf("previous version is %s", t.prevBuildVersion)
	return nil
}

func (t *RestartAgentTask) checkVersion() (err error) {
	t.ExecuteLogf("current obshell version is %s", constant.VERSION_RELEASE)

	// If the target version is not set, then get the previous version which was backed up in the backup dir.
	if t.GetContext().GetParam(PARAM_TARGET_AGENT_BUILD_VERSION) == nil {
		t.ExecuteLog("upgrade from version which not set target version, check previous version.")
		if err := t.getPrevAgentVersion(); err != nil {
			return err
		}

		// If the previous version is lower than the current version, then return error.
		if pkg.CompareVersion(constant.VERSION_RELEASE, t.prevBuildVersion) < 0 {
			err = fmt.Errorf("current version %s is lower than previous version %s", constant.VERSION_RELEASE, t.prevBuildVersion)
		}

	} else {
		// If the target version is set, then get the target version and compare with the current version.
		if err := t.GetContext().GetParamWithValue(PARAM_TARGET_AGENT_BUILD_VERSION, &t.targetBuildVersion); err != nil {
			return err
		}
		t.ExecuteLogf("target obshell version is %s", t.targetBuildVersion)

		// If the current version is not the target version, then return error.
		if pkg.CompareVersion(constant.VERSION_RELEASE, t.targetBuildVersion) < 0 {
			err = fmt.Errorf("current version %s is lower than target version %s", constant.VERSION_RELEASE, t.targetBuildVersion)
		}
	}
	return
}

func (t *RestartAgentTask) restartAgent() (err error) {
	cmd := fmt.Sprintf("%s/bin/%s admin restart --ip %s --port %d --pid %d --takeover 0", global.HomePath, constant.PROC_OBSHELL,
		meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort(), global.Pid)
	t.ExecuteLogf("cmd is %s", cmd)
	return execStartCmd(cmd)

}
