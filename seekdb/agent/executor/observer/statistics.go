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

package observer

import (
	"fmt"
	"strings"
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/executor/host"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/binary"
	"github.com/oceanbase/obshell/seekdb/agent/lib/parse"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
	"github.com/oceanbase/obshell/seekdb/utils"
	log "github.com/sirupsen/logrus"
)

func GetStatisticsInfo() *bo.ObclusterStatisticInfo {
	telemetryEnabled := global.EnableTelemetry
	_, isCommunityEdition, _ := binary.GetMyOBVersion() // ignore the error
	if !isCommunityEdition {
		telemetryEnabled = false
	}
	info := &bo.ObclusterStatisticInfo{
		TelemetryEnabled: telemetryEnabled,
		Reporter:         "seekdb",
		ObshellVersion:   fmt.Sprintf("obshell %s", constant.VERSION),
		ObshellRevision:  fmt.Sprintf("%s-%s", constant.RELEASE, config.GitCommitId),
		TelemetryVersion: 1,
		ReportTime:       time.Now().Unix(),
	}

	// read cluster id from observer file
	clusterId, err := binary.GetClusterId()
	if err != nil {
		log.Warnf("Failed to get cluster id: %v", err)
	}
	info.ClusterId = clusterId

	hostInfo := host.GetInfo()
	if hostInfo != nil {
		info.Hosts = *hostInfo
	}

	observer, err := obclusterService.GetOBServer()
	if err != nil {
		log.Warnf("Failed to get observer info: %v", err)
	}

	info.Instances = bo.ObserverInfo{
		HostHash: utils.Sha1(observer.SvrIp),
	}

	observerResource, err := obclusterService.GetObserverResource()
	if err != nil {
		log.Warnf("Failed to get observer resource: %v", err)
	}

	if observerResource != nil {
		info.Instances.CpuCount = observerResource.CpuCapacity
		info.Instances.MemoryLimit = parse.FormatCapacity(observerResource.MemCapacity)
		info.Instances.DataFileSize = parse.FormatCapacity(observerResource.DataDiskCapacity)
		info.Instances.LogDiskSize = parse.FormatCapacity(observerResource.LogDiskCapacity)
	}

	item := strings.Split(observer.BuildVersion, "-")
	if len(item) > 1 {
		info.Instances.Version = item[0]
		info.Instances.Revision = item[1]
	}
	return info
}
