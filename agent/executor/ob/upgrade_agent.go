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

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
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
