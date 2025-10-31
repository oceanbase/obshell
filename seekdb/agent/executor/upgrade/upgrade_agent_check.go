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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/pkg"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/param"
	"github.com/oceanbase/obshell/seekdb/utils"
)

func AgentUpgradeCheck(param param.UpgradeCheckParam) (*task.DagDetailDTO, error) {
	agentErr := preCheckForAgentUpgrade(param)
	if agentErr != nil {
		return nil, agentErr
	}
	agentUpgradeCheckTemplate := buildAgentUpgradeCheckTemplate(param)
	agentUpgradeCheckTaskContext := buildAgentUpgradeCheckTaskContext(param)
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
	buildNumber, distribution, _ := pkg.SplitRelease(release)
	_, err := obclusterService.GetUpgradePkgInfoByVersionAndRelease(constant.PKG_OBSHELL, version, buildNumber, distribution, global.Architecture)
	if err != nil {
		return errors.Occur(errors.ErrAgentPackageNotFound, fmt.Sprintf("%s-%s-%s.%s.rpm", constant.PKG_OBSHELL, version, release, global.Architecture))
	}

	return nil
}

func buildAgentUpgradeCheckTaskContext(param param.UpgradeCheckParam) *task.TaskContext {
	ctx := task.NewTaskContext()
	buildNumber, distribution, _ := pkg.SplitRelease(param.Release)
	taskTime := strconv.Itoa(int(time.Now().UnixMilli()))
	ctx.SetParam(PARAM_UPGRADE_DIR, param.UpgradeDir).
		SetParam(PARAM_TASK_TIME, taskTime).
		SetParam(PARAM_VERSION, param.Version).
		SetParam(PARAM_BUILD_NUMBER, buildNumber).
		SetParam(PARAM_DISTRIBUTION, distribution).
		SetParam(PARAM_RELEASE_DISTRIBUTION, param.Release)
	return ctx
}

func buildAgentUpgradeCheckTemplate(param param.UpgradeCheckParam) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_UPGRADE_CHECK_OBSHELL, param.Version, param.Release)
	agentUpgradeCheckTemplateBuilder := task.NewTemplateBuilder(name)
	agentUpgradeCheckTemplateBuilder.
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCreateUpgradeDirTask(), false).
		AddTask(newGetAllRequiredPkgsTask(), false).
		AddTask(newCheckAllRequiredPkgsTask(), false).
		AddTask(newInstallAllRequiredPkgsTask(), false).
		AddTask(newRemoveUpgradeCheckDirTask(), false)
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
