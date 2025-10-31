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

	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
)

func GetAllUpgradePkgInfos() ([]bo.UpgradePkgInfo, error) {
	upgradePkgInfos, err := obclusterService.GetAllUpgradePkgInfos()
	if err != nil {
		return nil, err
	}
	log.Infof("Get all upgrade pkg infos: %v", upgradePkgInfos)
	infos := make([]bo.UpgradePkgInfo, 0)
	onlyMap := make(map[string]bool)
	for i := len(upgradePkgInfos) - 1; i >= 0; i-- {
		onlyFlag := fmt.Sprintf("%s-%s-%s.%s", upgradePkgInfos[i].Name, upgradePkgInfos[i].Version, upgradePkgInfos[i].ReleaseDistribution, upgradePkgInfos[i].Architecture)
		if _, ok := onlyMap[onlyFlag]; !ok {
			infos = append(infos, upgradePkgInfos[i].ToBO())
			onlyMap[onlyFlag] = true
		}
	}

	return infos, nil
}
