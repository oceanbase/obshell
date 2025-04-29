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

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/utils"
)

func GetStatisticsInfo() *bo.ObclusterStatisticInfo {
	info := &bo.ObclusterStatisticInfo{
		Reporter:         "obshell-dashboard",
		ObshellVersion:   fmt.Sprintf("OBShell %s (for OceanBase_CE)", constant.VERSION),
		ObshellRevision:  fmt.Sprintf("%s-%s", constant.RELEASE, config.GitCommitId),
		TelemetryVersion: 1,
		ReportTime:       time.Now().Unix(),
	}
	// fetch all host info
	hostInfoList := make([]bo.HostInfo, 0)
	allAgents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		log.Errorf("Failed to get all agents: %v", err)
	} else {
		for _, agent := range allAgents {
			hostInfo := bo.HostInfo{}
			resErr := secure.SendGetRequest(&agent, "/api/v1/agent/host-info", nil, &hostInfo)
			if resErr == nil {
				hostInfoList = append(hostInfoList, hostInfo)
			}
		}
	}
	info.Hosts = hostInfoList

	// fetch all instance info
	obclusterSummary, summaryErr := GetObclusterSummary()
	observerInfoList := make([]bo.ObserverInfo, 0)
	if summaryErr != nil {
		log.Errorf("Failed to get obcluster config: %v", summaryErr)
	} else {
		info.ClusterId = fmt.Sprintf("%d", obclusterSummary.ClusterBasicInfo.ClusterId)
		typeSuffix := ""
		if obclusterSummary.ClusterBasicInfo.IsCommunityEdition {
			typeSuffix = "-ce"
		}
		parameters, parameterErr := GetAllParameters()
		parameterMap := make(map[string]bo.ClusterParameter)
		if parameterErr != nil {
			log.Errorf("Failed to get obcluster parameters: %v", parameterErr)
		} else {
			for _, parameter := range parameters {
				parameterMap[parameter.Name] = parameter
			}
		}
		for _, zone := range obclusterSummary.Zones {
			for _, server := range zone.Servers {
				cpuCount := 0
				memoryLimit := ""
				dataFileSize := ""
				logDiskSize := ""
				version := obclusterSummary.ObVersion
				revision := ""
				if cpuCountParameter, ok := parameterMap["cpu_count"]; ok {
					for _, serverValue := range cpuCountParameter.ServerValue {
						if serverValue.SvrIp == server.Ip && serverValue.SvrPort == server.SvrPort {
							cpuCount, _ = strconv.Atoi(serverValue.Value)
						}
					}
				}
				if memoryLimitParameter, ok := parameterMap["memory_limit"]; ok {
					for _, serverValue := range memoryLimitParameter.ServerValue {
						if serverValue.SvrIp == server.Ip && serverValue.SvrPort == server.SvrPort {
							memoryLimit = serverValue.Value
						}
					}
				}
				if dataFileSizeParameter, ok := parameterMap["datafile_size"]; ok {
					for _, serverValue := range dataFileSizeParameter.ServerValue {
						if serverValue.SvrIp == server.Ip && serverValue.SvrPort == server.SvrPort {
							dataFileSize = serverValue.Value
						}
					}
				}
				if logDiskSizeParameter, ok := parameterMap["log_disk_size"]; ok {
					for _, serverValue := range logDiskSizeParameter.ServerValue {
						if serverValue.SvrIp == server.Ip && serverValue.SvrPort == server.SvrPort {
							logDiskSize = serverValue.Value
						}
					}
				}
				fields := strings.Split(server.Version, "-")
				parts := strings.Split(fields[0], "_")
				if len(parts) >= 2 {
					version = parts[0]
					revision = parts[1]
				}
				observerInfo := bo.ObserverInfo{
					Type:         fmt.Sprintf("oceanbase%s", typeSuffix),
					Version:      version,
					Revision:     revision,
					HostHash:     utils.Sha1(server.Ip),
					CpuCount:     cpuCount,
					MemoryLimit:  memoryLimit,
					DataFileSize: dataFileSize,
					LogDiskSize:  logDiskSize,
				}
				observerInfoList = append(observerInfoList, observerInfo)
			}
		}

	}
	info.Instances = observerInfoList
	return info
}
