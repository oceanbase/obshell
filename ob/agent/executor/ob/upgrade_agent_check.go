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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/coordinator"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
	"github.com/oceanbase/obshell/ob/utils"
)

func AgentUpgradeCheck(param param.UpgradeCheckParam) (*task.DagDetailDTO, error) {
	agentErr := preCheckForAgentUpgrade(param)
	if agentErr != nil {
		return nil, agentErr
	}
	agents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, err
	}
	agentUpgradeCheckTemplate := buildAgentUpgradeCheckTemplate(param)
	agentUpgradeCheckTaskContext := buildAgentUpgradeCheckTaskContext(param, agents)
	agentUpgradeCheckDag, err := taskService.CreateDagInstanceByTemplate(agentUpgradeCheckTemplate, agentUpgradeCheckTaskContext)
	if err != nil {
		log.WithError(err).Error("create dag instance by template failed")
		return nil, err
	}
	return task.NewDagDetailDTO(agentUpgradeCheckDag), nil
}

func preCheckForAgentUpgrade(param param.UpgradeCheckParam) (err error) {
	log.Info("Starting obshell upgrade pre-check.")
	if !meta.OCS_AGENT.IsClusterAgent() {
		return errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}
	allAgents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return errors.Wrap(err, "failed to query all agents from ob")
	}
	agentInfo := coordinator.OCS_COORDINATOR.Maintainer
	agentsStatus := make(map[string]http.AgentStatus)
	resErr := secure.SendGetRequest(agentInfo, "/api/v1/agents/status", nil, &agentsStatus)
	if resErr != nil {
		return errors.Wrap(resErr, "failed to query all agents status")
	}
	unavailableAgents := make([]string, 0)
	for agent, agentStatus := range agentsStatus {
		if agentStatus.State != 2 {
			unavailableAgents = append(unavailableAgents, agent)
		}
	}
	for _, agent := range allAgents {
		if _, ok := agentsStatus[agent.String()]; !ok {
			unavailableAgents = append(unavailableAgents, agent.String())
		}
	}
	if len(unavailableAgents) > 0 {
		return errors.Occur(errors.ErrAgentUnavailable, strings.Join(unavailableAgents, ","))
	}
	if err := checkUpgradeDir(&param.UpgradeDir); err != nil {
		return err
	}
	if err := checkTargetVersionSupport(param.Version, param.Release); err != nil {
		return err
	}
	if err := findTargetPkg(param.Version, param.Release); err != nil {
		return err
	}
	return nil
}

func checkTargetVersionSupport(version, release string) error {
	buildNumber, _, err := pkg.SplitRelease(release)
	if err != nil {
		return err
	}

	targetVR := fmt.Sprintf("%s-%s", version, buildNumber)
	if pkg.CompareVersion(targetVR, constant.VERSION_RELEASE) <= 0 {
		return errors.Occur(errors.ErrAgentUpgradeToLowerVersion, targetVR, constant.VERSION_RELEASE)
	}

	return nil
}

func findTargetPkg(version, release string) error {
	archList, err := obclusterService.GetAllArchs()
	if err != nil {
		return err
	}
	var errs []error
	buildNumber, distribution, _ := pkg.SplitRelease(release)
	var notFoundPkgs []string
	for _, arch := range archList {
		_, err := obclusterService.GetUpgradePkgInfoByVersionAndRelease(constant.PKG_OBSHELL, version, buildNumber, distribution, arch)
		if err != nil {
			notFoundPkgs = append(notFoundPkgs, fmt.Sprintf("%s-%s-%s.%s.rpm", constant.PKG_OBSHELL, version, release, arch))
		}
	}
	if len(errs) > 0 {
		return errors.Occur(errors.ErrAgentPackageNotFound, strings.Join(notFoundPkgs, ","))
	}
	return nil
}

func buildAgentUpgradeCheckTaskContext(param param.UpgradeCheckParam, agents []meta.AgentInfo) *task.TaskContext {
	ctx := task.NewTaskContext()
	buildNumber, distribution, _ := pkg.SplitRelease(param.Release)
	taskTime := strconv.Itoa(int(time.Now().UnixMilli()))
	ctx.SetParam(PARAM_ALL_AGENTS, agents).
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_UPGRADE_DIR, param.UpgradeDir).
		SetParam(PARAM_TASK_TIME, taskTime).
		SetParam(PARAM_VERSION, param.Version).
		SetParam(PARAM_BUILD_NUMBER, buildNumber).
		SetParam(PARAM_DISTRIBUTION, distribution).
		SetParam(PARAM_RELEASE_DISTRIBUTION, param.Release).
		SetParam(PARAM_ONLY_FOR_AGENT, true)
	agentUpgradeRoute := []RouteNode{
		{
			Version:        param.Version,
			Release:        buildNumber,
			BuildVersion:   fmt.Sprintf("%s-%s", param.Version, buildNumber),
			DeprecatedInfo: []string{},
		},
	}
	ctx.SetParam(PARAM_AGENT_UPGRADE_ROUTE, agentUpgradeRoute)
	return ctx
}

func buildAgentUpgradeCheckTemplate(param param.UpgradeCheckParam) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_UPGRADE_CHECK_OBSHELL, param.Version, param.Release)
	agentUpgradeCheckTemplateBuilder := task.NewTemplateBuilder(name)
	agentUpgradeCheckTemplateBuilder.
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCreateUpgradeDirTask(), true).
		AddTask(newGetAllRequiredPkgsTask(), true).
		AddTask(newCheckAllRequiredPkgsTask(), true).
		AddTask(newInstallAllRequiredPkgsTask(), true).
		AddTask(newRemoveUpgradeCheckDirTask(), true)
	return agentUpgradeCheckTemplateBuilder.Build()
}

func checkUpgradeDir(path *string) (err error) {
	log.Infof("checking upgrade directory: '%s'", *path)
	str := *path

	*path = strings.TrimSpace(*path)
	if len(*path) == 0 {
		return nil
	}

	return utils.CheckPathValid(str)
}
