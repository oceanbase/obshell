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
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
)

func AgentUpgrade(param param.UpgradeCheckParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	agentErr := preCheckForAgentUpgrade(param)
	if agentErr != nil {
		log.WithError(agentErr).Error("pre check for agent upgrade failed")
		return nil, agentErr
	}
	agents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	agentUpgradeTemplate := buildAgentUpgradeTemplate(param)
	agentUpgradeTaskContext := buildAgentUpgradeCheckTaskContext(param, agents)
	agentUpgradeDag, err := taskService.CreateDagInstanceByTemplate(agentUpgradeTemplate, agentUpgradeTaskContext)
	if err != nil {
		log.WithError(err).Error("create dag instance by template failed")
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(agentUpgradeDag), nil
}

func buildAgentCheckAndUpgradeTemplate() *task.Template {
	return task.NewTemplateBuilder(DAG_CHECK_AND_UPGRADE_OBSHELL).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newCreateUpgradeDirTask(), true).
		AddTask(newGetAllRequiredPkgsTask(), true).
		AddTask(newCheckAllRequiredPkgsTask(), true).
		AddTask(newInstallAllRequiredPkgsTask(), true).
		AddTask(newBackupAgentForUpgradeTask(), true).
		AddTask(newInstallNewAgentTask(), true).
		AddTask(newRestartAgentTask(), true).
		AddTask(newUpgradePostTableMaintainTask(), true).
		Build()
}

func buildAgentUpgradeTemplate(param param.UpgradeCheckParam) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_UPGRADE_OBSHELL, param.Version, param.Release)
	return task.NewTemplateBuilder(name).
		AddTemplate(buildAgentCheckAndUpgradeTemplate()).
		AddTask(newRemoveUpgradeCheckDirTask(), true).
		Build()
}

func TakeOverUpdateAgentVersion() (*task.DagDetailDTO, error) {
	targetVersion, err := needUpdateAgentBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "get update agent binary failed")
	} else if targetVersion == "" {
		return nil, nil
	}

	template := buildTakeOverUpdateAgentVersion()
	ctx := task.NewTaskContext().
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true).
		SetParam(PARAM_TARGET_AGENT_VERSION, targetVersion)

	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		if dag, err1 := localTaskService.FindLastMaintenanceDag(); err1 != nil {
			return nil, errors.Wrapf(err1, "get last maintenance dag failed")
		} else if dag == nil || dag.GetName() != DAG_TAKE_OVER_UPDATE_AGENT_VERSION {
			return nil, errors.Wrapf(err, "create dag instance by template failed")
		} else {
			return task.NewDagDetailDTO(dag), nil
		}
	}
	return task.NewDagDetailDTO(dag), nil
}

type UpgradeToClusterAgentVersionTask struct {
	task.Task
	targetVersion string
	RestartAgentTask
}

func newInstallClusterAgentBinaryTask() *UpgradeToClusterAgentVersionTask {
	newTask := &UpgradeToClusterAgentVersionTask{
		Task: *task.NewSubTask(TASK_INSTALL_CLUSTER_AGENT_VERSION),
	}
	newTask.
		SetCanRetry().
		SetCanRollback().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *UpgradeToClusterAgentVersionTask) Execute() (err error) {
	if err := t.GetContext().GetParamWithValue(PARAM_TARGET_AGENT_VERSION, &t.targetVersion); err != nil {
		return errors.Wrap(err, "get target version failed")
	}

	if t.targetVersion == "" {
		t.ExecuteLog("current agent version is the same as the cluster agent version, no need to upgrade")
		return nil
	}

	if t.targetVersion == constant.VERSION {
		return HandleOBMeta()
	}

	defer func() {
		if err != nil {
			t.ExecuteLogf("install cluster agent binary failed, rollback")
			system.CopyFile(path.ObshellBinBackupPath(), path.ObshellBinPath())
		}
	}()

	// Download agent binary.
	t.ExecuteLogf("Install cluster agent binary, target version: %s", t.targetVersion)
	if err := installTargetAgent(t.targetVersion); err != nil {
		t.ExecuteErrorLogf("install target agent failed: %s", err.Error())
		return errors.Wrap(err, "install target agent failed")
	}
	t.ExecuteLogf("install agent %s success", t.targetVersion)

	t.RestartAgentTask.Task = t.Task
	return t.restartAgent()
}

func installTargetAgent(targetVersion string) error {
	if err := os.RemoveAll(path.ObshellBinPath()); err != nil {
		return errors.New("remove old agent binary failed")
	}
	if err := agentService.DownloadBinary(path.ObshellBinPath(), targetVersion); err != nil {
		return errors.Wrap(err, "download binary failed")
	}
	if err := os.Chmod(path.ObshellBinPath(), 0755); err != nil {
		return errors.Wrapf(err, "chmod %s failed to %s", path.ObshellBinPath(), "0755")
	}
	return nil
}

func buildTakeOverUpdateAgentVersion() *task.Template {
	return task.NewTemplateBuilder(DAG_TAKE_OVER_UPDATE_AGENT_VERSION).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newInstallClusterAgentBinaryTask(), false).
		Build()
}

func getClusterAgentVersion() (targetVersion string, err error) {
	agents, err := agentService.GetAllAgentsFromOB()
	if err != nil {
		return "", err
	}

	var clusterVersion string
	for _, agent := range agents {
		if agent.IsClusterAgent() {
			if clusterVersion == "" {
				clusterVersion = agent.Version
				if pkg.CompareVersion(clusterVersion, meta.OCS_AGENT.GetVersion()) < 0 {
					return "", fmt.Errorf("take over a higher version agent(%s) into cluster agent(%s) is not allowed", meta.OCS_AGENT.GetVersion(), clusterVersion)
				}
			} else if clusterVersion != agent.Version {
				return "", errors.New("unexpect error: cluster agent version is not consistent")
			}
		} else if agent.IsTakeover() {
			if agent.Version != meta.OCS_AGENT.GetVersion() {
				return "", errors.New("agent version is not consistent")
			}
		}
		// Other identifies is not considered.
	}
	return strings.Split(clusterVersion, "-")[0], nil
}

// needUpdateAgentBinary will check if need to update agent binary,
// and if need, return the target version which must not be empty.
func needUpdateAgentBinary() (targetVersion string, err error) {
	log.Info("check target version")
	if targetVersion, err = getClusterAgentVersion(); err != nil {
		return "", errors.Wrap(err, "get cluster agent version failed")
	}

	if targetVersion != "" && targetVersion != constant.VERSION {
		err = canGetTargetBinary(targetVersion)
		err = errors.Wrap(err, "can get target binary failed")
		return
	}
	return "", nil
}

func canGetTargetBinary(targetVersion string) error {
	log.Infof("the agent need to be upgraded to: %s", targetVersion)
	if exist, err := agentService.TargetVersionAgentExists(targetVersion); err != nil {
		// TODO: 查询 OB 是否有可用的 rpm 包。
		return err
	} else if !exist {
		return fmt.Errorf("There is no aviailable agent(version: %s, architecture: %s) in OB", targetVersion, global.Architecture)
	}
	return nil
}
