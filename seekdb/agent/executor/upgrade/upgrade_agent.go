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

package upgrade

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/param"
)

func AgentUpgrade(param param.UpgradeCheckParam) (*task.DagDetailDTO, error) {
	err := preCheckForAgentUpgrade(param)
	if err != nil {
		log.WithError(err).Error("pre check for agent upgrade failed")
		return nil, err
	}
	agentUpgradeTemplate := buildAgentUpgradeTemplate(param)
	agentUpgradeTaskContext := buildAgentUpgradeCheckTaskContext(param)
	agentUpgradeDag, err := taskService.CreateDagInstanceByTemplate(agentUpgradeTemplate, agentUpgradeTaskContext)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(agentUpgradeDag), nil
}

func buildAgentCheckAndUpgradeTemplate() *task.Template {
	return task.NewTemplateBuilder(DAG_CHECK_AND_UPGRADE_OBSHELL).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newCreateUpgradeDirTask(), false).
		AddTask(newGetAllRequiredPkgsTask(), false).
		AddTask(newCheckAllRequiredPkgsTask(), false).
		AddTask(newInstallAllRequiredPkgsTask(), false).
		AddTask(newBackupAgentForUpgradeTask(), false).
		AddTask(newInstallNewAgentTask(), false).
		AddTask(newRestartAgentTask(), false).
		AddTask(newUpgradePostTableMaintainTask(), false).
		Build()
}

func buildAgentUpgradeTemplate(param param.UpgradeCheckParam) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_UPGRADE_OBSHELL, param.Version, param.Release)
	return task.NewTemplateBuilder(name).
		AddTemplate(buildAgentCheckAndUpgradeTemplate()).
		AddTask(newRemoveUpgradeCheckDirTask(), false).
		Build()
}
