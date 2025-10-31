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

package host

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/utils"
)

var once sync.Once
var hostStaticInfo *bo.HostInfo

func getOSInfo() *bo.OSInfo {
	osInfo := &bo.OSInfo{}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_OS_NAME).CombinedOutput(); err == nil {
		osInfo.OS = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_OS_RELEASE).CombinedOutput(); err == nil {
		osInfo.Version = strings.Trim(strings.TrimSpace(string(output)), "\"")
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	return osInfo
}

func getHostBasicInfo() *bo.HostBasicInfo {
	basicInfo := &bo.HostBasicInfo{}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_HOST_TYPE).CombinedOutput(); err == nil {
		basicInfo.HostType = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	basicInfo.HostHash = utils.Sha1(meta.OCS_AGENT.GetIp())
	return basicInfo
}

func getMemoryStaticInfo() *bo.MemoryInfo {
	memoryInfo := &bo.MemoryInfo{}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_MEMORY_TOTAL).CombinedOutput(); err == nil {
		memoryInfo.Total = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	return memoryInfo
}

func getMemoryDynamicInfo() *bo.MemoryInfo {
	memoryInfo := &bo.MemoryInfo{}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_MEMORY_FREE).CombinedOutput(); err == nil {
		memoryInfo.Free = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_MEMORY_AVAILABLE).CombinedOutput(); err == nil {
		memoryInfo.Available = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	return memoryInfo
}

func getUlimitInfo() *bo.UlimitInfo {
	ulimitInfo := &bo.UlimitInfo{}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_ULIMIT_NOFILE).CombinedOutput(); err == nil {
		ulimitInfo.NofileHard = strings.TrimSpace(string(output))
		ulimitInfo.NofileSoft = ulimitInfo.NofileHard
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_ULIMIT_MAX_USER_PROCESSES).CombinedOutput(); err == nil {
		ulimitInfo.MaxUserProcessesHard = strings.TrimSpace(string(output))
		ulimitInfo.MaxUserProcessesSoft = ulimitInfo.MaxUserProcessesHard
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	return ulimitInfo
}

func getCpuInfo() *bo.CpuInfo {
	cpuInfo := &bo.CpuInfo{}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_CPU_FREQUENCY).CombinedOutput(); err == nil {
		cpuInfo.Frequency = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_CPU_MODEL).CombinedOutput(); err == nil {
		cpuInfo.ModelName = strings.TrimSpace(string(output))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_CPU_LOGIC_CORES).CombinedOutput(); err == nil {
		cpuInfo.LogicalCores, _ = strconv.Atoi(strings.TrimSpace(string(output)))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	if output, err := exec.Command("bash", "-c", constant.COMMAND_CPU_PHYSICAL_CORES).CombinedOutput(); err == nil {
		cpuInfo.PhysicalCores, _ = strconv.Atoi(strings.TrimSpace(string(output)))
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	return cpuInfo
}

func getDiskInfo() []bo.DiskInfo {
	diskList := make([]bo.DiskInfo, 0)
	log.Info("get disk info")
	if output, err := exec.Command("bash", "-c", constant.COMMAND_DF).CombinedOutput(); err == nil {
		diskInfoList := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, diskInfo := range diskInfoList {
			fields := strings.Fields(diskInfo)
			if len(fields) >= 6 {
				diskList = append(diskList, bo.DiskInfo{
					Used:       fields[2],
					Total:      fields[1],
					MountHash:  utils.Sha1(fields[5]),
					DeviceName: fields[0],
				})
			}
		}
	} else {
		log.Errorf("Got error when executing command: %v", err)
	}
	return diskList
}

func getStaticInfo() *bo.HostInfo {
	hostInfo := &bo.HostInfo{
		OS:     getOSInfo(),
		Cpu:    getCpuInfo(),
		Basic:  getHostBasicInfo(),
		Ulimit: getUlimitInfo(),
		Memory: getMemoryStaticInfo(),
	}
	return hostInfo
}

func GetInfo() *bo.HostInfo {
	// Get static info only once
	once.Do(func() {
		log.Info("Get static info")
		hostStaticInfo = getStaticInfo()
	})
	hostInfo := &bo.HostInfo{
		OS:     hostStaticInfo.OS,
		Cpu:    hostStaticInfo.Cpu,
		Basic:  hostStaticInfo.Basic,
		Ulimit: hostStaticInfo.Ulimit,
		Memory: hostStaticInfo.Memory,
	}

	// Get dynamic info each time
	memoryDynamicInfo := getMemoryDynamicInfo()
	hostInfo.Memory.Available = memoryDynamicInfo.Available
	hostInfo.Memory.Free = memoryDynamicInfo.Free
	hostInfo.Disks = getDiskInfo()
	return hostInfo
}
