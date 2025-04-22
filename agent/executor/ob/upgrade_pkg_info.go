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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

func GetAllUpgradePkgInfos() ([]bo.UpgradePkgInfo, *errors.OcsAgentError) {
	upgradePkgInfos, err := obclusterService.GetAllUpgradePkgInfos()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	log.Infof("Get all upgrade pkg infos: %v", upgradePkgInfos)
	infos := make([]bo.UpgradePkgInfo, len(upgradePkgInfos))
	for i := range upgradePkgInfos {
		infos[i] = upgradePkgInfos[i].ToBO()
	}
	return infos, nil
}

func GetObPackageUpgradeDepYaml(version string, release string) ([]RouteNode, *errors.OcsAgentError) {
	upgradeRoute, err := getTargetObUpgradeDepYaml(version, release)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	if upgradeRoute == nil {
		return nil, errors.Occur(errors.ErrUnexpected, "No upgrade route found")
	}
	return upgradeRoute[1:], nil
}
