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
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

const (
	OCEANBASE_HOMEPATH = "/home/admin/oceanbase"
	OCEANBASE_HOME     = "/home"

	MYSQL_CONNECTOR = "mysql.connector"
)

var (
	confficient = 1.1
	modules     = []string{MYSQL_CONNECTOR}
)

func ObUpgradeCheck(param param.UpgradeCheckParam) (*task.DagDetailDTO, error) {
	log.Info("ob upgrade check")
	obType, err := obclusterService.GetOBType()
	if err != nil {
		return nil, err
	}

	upgradeRoute, err := preCheckForObUpgradeCheck(param, obType)
	if err != nil {
		return nil, err
	}
	agents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, err
	}
	obUpgradeCheckTemplate := buildObUpgradeCheckTemplate(param)
	obUpgradeCheckTaskContext := buildObUpgradeCheckTaskContext(param, upgradeRoute, agents, obType)
	obUpgradeCheckDag, err := taskService.CreateDagInstanceByTemplate(obUpgradeCheckTemplate, obUpgradeCheckTaskContext)
	if err != nil {
		log.WithError(err).Error("create dag instance by template failed")
		return nil, err
	}
	return task.NewDagDetailDTO(obUpgradeCheckDag), nil
}

func buildObUpgradeCheckTaskContext(param param.UpgradeCheckParam, upgradeRoute []RouteNode, agents []meta.AgentInfo, obType modelob.OBType) *task.TaskContext {
	ctx := task.NewTaskContext()
	buildNumer, distribution, _ := pkg.SplitRelease(param.Release)
	taskTime := strconv.Itoa(int(time.Now().UnixMilli()))
	ctx.SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_ALL_AGENTS, agents).
		SetParam(PARAM_UPGRADE_DIR, param.UpgradeDir).
		SetParam(PARAM_VERSION, param.Version).
		SetParam(PARAM_BUILD_NUMBER, buildNumer).
		SetParam(PARAM_DISTRIBUTION, distribution).
		SetParam(PARAM_TASK_TIME, taskTime).
		SetParam(PARAM_UPGRADE_ROUTE, upgradeRoute).
		SetParam(PARAM_OB_TYPE, obType)
	return ctx
}

func buildObUpgradeCheckTemplate(param param.UpgradeCheckParam) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_UPGRADE_CHECK_OB, param.Version, param.Release)
	obUpgradeCheckTemplateBuilder := task.NewTemplateBuilder(name)
	obUpgradeCheckTemplateBuilder.
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCheckEnvTask(), true).
		AddTask(newCreateUpgradeDirTask(), true).
		AddTask(newGetAllRequiredPkgsTask(), true).
		AddTask(newCheckAllRequiredPkgsTask(), true).
		AddTask(newInstallAllRequiredPkgsTask(), true).
		AddTask(newRemoveUpgradeCheckDirTask(), true)
	return obUpgradeCheckTemplateBuilder.Build()
}

func getKeyForPkgInfoMap(ctx *task.TaskContext) (keys []string, err error) {
	upgradeRoute, err := getUpgradeRouteForTask(ctx)
	if err != nil {
		return nil, err
	}
	for _, node := range upgradeRoute {
		if ctx.GetParam(PARAM_ONLY_FOR_AGENT) == nil {
			obTypeParam := ctx.GetParam(PARAM_OB_TYPE)
			if obTypeParam == nil {
				keys = append(keys, GenerateLibsBuildVersion(node.BuildVersion))
			} else if obType, ok := ctx.GetParam(PARAM_OB_TYPE).(string); ok && modelob.OBType(obType) == modelob.OBTypeCommunity {
				keys = append(keys, GenerateLibsBuildVersion(node.BuildVersion))
			}
		}
		keys = append(keys, node.BuildVersion)
	}
	return keys, nil
}

func getUpgradeRouteForTask(taskContext *task.TaskContext) (upgradeRoute []RouteNode, err error) {
	if taskContext.GetParam(PARAM_ONLY_FOR_AGENT) != nil {
		err = taskContext.GetParamWithValue(PARAM_AGENT_UPGRADE_ROUTE, &upgradeRoute)
	} else {
		err = taskContext.GetParamWithValue(PARAM_UPGRADE_ROUTE, &upgradeRoute)
	}
	return
}

func preCheckForObUpgradeCheck(param param.UpgradeCheckParam, obType modelob.OBType) (upgradeRoute []RouteNode, err error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}
	if err = checkUpgradeDir(&param.UpgradeDir); err != nil {
		return nil, err
	}
	upgradeRoute, err = checkForAllRequiredPkgs(param.Version, param.Release, obType)
	if err != nil {
		return
	}

	return upgradeRoute, nil
}

func getTargetObUpgradeDepYaml(targetVersion string, targetRelease string, obType modelob.OBType) ([]RouteNode, error) {
	// Param 'targetRelease' is like '***.**.el7'.
	targetBuildNumber, _, err := pkg.SplitRelease(targetRelease)
	if err != nil {
		return nil, err
	}

	log.Info("get target pkg info")
	var pkgName string
	switch obType {
	case modelob.OBTypeCommunity:
		pkgName = constant.PKG_OCEANBASE_CE
	case modelob.OBTypeBusiness:
		pkgName = constant.PKG_OCEANBASE
	case modelob.OBTypeStandalone:
		pkgName = constant.PKG_OCEANBASE_STANDALONE
	}
	pkgInfo, err := obclusterService.GetUpgradePkgInfoByVersionAndReleaseDist(pkgName, targetVersion, targetRelease, global.Architecture)
	if err != nil {
		return nil, errors.Wrapf(err, "%s-%s-%s.%s.rpm", pkgName, targetVersion, targetRelease, global.Architecture)
	}

	upgradeRoute, err := generateUpgradeRouteList(targetVersion, targetBuildNumber, pkgInfo.UpgradeDepYaml)
	if err != nil {
		log.WithError(err).Error("generate upgrade route failed")
		return nil, err
	}
	upgradeRoute[len(upgradeRoute)-1].Release = targetBuildNumber
	upgradeRoute[len(upgradeRoute)-1].BuildVersion = fmt.Sprintf("%s-%s", targetVersion, targetBuildNumber)
	log.Infof("upgrade route: %v", upgradeRoute)
	return upgradeRoute, nil
}

func checkForAllRequiredPkgs(targetVersion, targetRelease string, obType modelob.OBType) ([]RouteNode, error) {
	// Param 'targetRelease' is like '***.**.el7'.
	targetBuildNumber, targetDistribution, err := pkg.SplitRelease(targetRelease)
	if err != nil {
		return nil, err
	}

	targetBV := fmt.Sprintf("%s-%s", targetVersion, targetBuildNumber)
	if err = checkTargetOBVersionSupport(targetBV); err != nil {
		return nil, err
	}

	upgradeRoute, err := getTargetObUpgradeDepYaml(targetVersion, targetRelease, obType)
	if err != nil {
		return nil, err
	}

	log.Info("check for all required pkgs")
	if err = checkForAllRequiredPkgsExist(upgradeRoute, targetDistribution, obType); err != nil {
		log.WithError(err).Error("check for all required pkgs failed")
		return nil, err
	}
	return upgradeRoute[1:], nil
}

func checkTargetOBVersionSupport(targetBV string) (err error) {
	if pkg.CompareVersion(targetBV, constant.SUPPORT_MIN_VERSION) < 0 {
		return errors.Occur(errors.ErrAgentOBVersionNotSupported, targetBV, constant.SUPPORT_MIN_VERSION)
	}
	currentBuildVersion, err := obclusterService.GetObBuildVersion()
	if err != nil {
		return errors.Wrap(err, "get current build version failed")
	}
	currentBV := strings.ReplaceAll(currentBuildVersion, "_", "-")
	log.Info("current build version is ", currentBV)
	if pkg.CompareVersion(targetBV, currentBV) <= 0 {
		return errors.Occur(errors.ErrObUpgradeToLowerVersion, targetBV, currentBV)
	}
	return nil
}

func checkForAllRequiredPkgsExist(upgradeRoute []RouteNode, distribution string, obType modelob.OBType) (err error) {
	archList, err := obclusterService.GetAllArchs()
	if err != nil {
		return err
	}
	var errs []error
	needPkgNameList := constant.REQUIRE_UPGRADE_PKG_NAMES_MAP[obType]
	var missingPkgs []string
	for _, node := range upgradeRoute[1:] {
		for _, arch := range archList {
			for _, pkgName := range needPkgNameList {
				var pkg oceanbase.UpgradePkgInfo
				var name string
				log.Infof("check pkg '%s' info '%v' arch '%s'", pkgName, node.BuildVersion, arch)
				if node.Release == RELEASE_NULL {
					name = fmt.Sprintf("%s-%s-${release}.%s.%s.rpm", pkgName, node.Version, distribution, arch)
					pkg, err = obclusterService.GetUpgradePkgInfoByVersion(pkgName, node.Version, distribution, arch, node.DeprecatedInfo)
				} else {
					name = fmt.Sprintf("%s-%s.%s.%s.rpm", pkgName, node.BuildVersion, distribution, arch)
					pkg, err = obclusterService.GetUpgradePkgInfoByVersionAndRelease(pkgName, node.Version, node.Release, distribution, arch)
				}
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						missingPkgs = append(missingPkgs, name)
						continue
					} else {
						return err
					}
				}

				if err = CheckPkgChunkCount(pkg.PkgId, pkg.ChunkCount, pkgName); err != nil {
					err = errors.Wrapf(err, "check pkg %s chunks count failed", name)
					log.Error(err)
					errs = append(errs, err)
				}
			}
		}
	}
	if len(missingPkgs) != 0 {
		return errors.Occur(errors.ErrObPackageNotExist, missingPkgs)
	}

	if len(errs) != 0 {
		return err
	}

	return nil
}

func CheckPkgChunkCount(pkgId, chunkCount int, pkgName string) error {
	actualChunkCount, err := obclusterService.GetUpgradePkgChunkCountByPkgId(pkgId)
	if err != nil {
		return err
	}
	if actualChunkCount != int64(chunkCount) {
		return errors.Occur(errors.ErrObPackageCorrupted, pkgName, fmt.Sprintf("actual chunk count %d is not equal to expected chunk count %d", actualChunkCount, chunkCount))
	}
	return nil
}

func generateUpgradeRouteList(targetVersion, targetRelease, upgradeDepYml string) ([]RouteNode, error) {
	currentBuildVersion, err := obclusterService.GetObBuildVersion()
	if err != nil {
		return nil, err
	}
	currentVersion := strings.Split(currentBuildVersion, "_")[0]
	currentRelease := strings.Split(currentBuildVersion, "_")[1]
	if currentVersion == targetVersion {
		return []RouteNode{
			{Version: currentVersion, Release: currentRelease},
			{Version: targetVersion, Release: targetRelease},
		}, nil
	}
	if upgradeDepYml == "" {
		return nil, errors.Occur(errors.ErrObUpgradeDepYamlMissing)
	}
	return GetOBUpgradeRoute(Repository{currentVersion, currentRelease}, Repository{targetVersion, targetRelease}, upgradeDepYml)
}

func GenerateLibsBuildVersion(buildVersion string) string {
	return fmt.Sprintf("libs-%s", buildVersion)
}
