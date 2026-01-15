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
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/tenant"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/secure"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
)

// only for mysql
func GetObclusterSummary() (*bo.ClusterInfo, error) {
	var info bo.ClusterInfo

	// Get all tenant infos
	tenantOverviews, err := tenantService.GetTenantsOverView()
	if err != nil {
		return nil, err
	}
	for _, tenantOverview := range tenantOverviews {
		tenantInfo, err := tenant.GetTenantInfo(tenantOverview.TenantName)
		if err != nil {
			return nil, err
		}
		info.Tenants = append(info.Tenants, *tenantInfo)
	}

	// Set basic cluster info
	if err := observerService.GetOBParatemerByName("cluster", &info.ClusterName); err != nil {
		return nil, err
	}
	if err := observerService.GetOBParatemerByName("cluster_id", &info.ClusterId); err != nil {
		return nil, err
	}
	var lclOpInterval string
	if err := observerService.GetOBParatemerByName("_lcl_op_interval", &lclOpInterval); err != nil {
		return nil, err
	}
	info.DeadLockDetectionEnabled = (lclOpInterval != "0ms" && lclOpInterval != "0")
	obType, err := obclusterService.GetOBType()
	if err != nil {
		return nil, err
	}
	info.IsCommunityEdition = obType == modelob.OBTypeCommunity
	info.IsStandalone = obType == modelob.OBTypeStandalone
	info.Status = oceanbase.OBStateShortMap[oceanbase.GetState()]
	version, err := obclusterService.GetObVersion()
	if err != nil {
		return nil, err
	}
	info.ObVersion = version

	rootServers, err := obclusterService.GetAllZoneRootServers()
	if err != nil {
		return nil, err
	}

	zones, err := obclusterService.GetAllZone()
	if err != nil {
		return nil, err
	}

	serverResourceMap, err := obclusterService.GetAllObserverResourceMap()
	if err != nil {
		return nil, err
	}

	allAgents, _ := agentService.GetAllAgentsDOFromOB()
	archMap := make(map[string]string)
	for _, agent := range allAgents {
		archMap[meta.NewAgentInfo(agent.Ip, agent.RpcPort).String()] = agent.Architecture
	}

	for _, zone := range zones {
		var zoneInfo bo.Zone
		zoneInfo.Name = zone.Zone
		zoneInfo.IdcName = zone.Idc
		zoneInfo.RegionName = zone.Region
		zoneInfo.InnerStatus = zone.Status
		if zoneInfo.InnerStatus == "ACTIVE" {
			zoneInfo.Status = "RUNNING"
		} else if zoneInfo.InnerStatus == "INACTIVE" {
			zoneInfo.Status = "SERVICE_STOPPED"
		}
		if rootServer, ok := rootServers[zone.Zone]; ok {
			zoneInfo.RootServer = rootServer.ToBO()
		}
		observers, err := obclusterService.GetOBServersByZone(zone.Zone)
		if err != nil {
			return nil, err
		}

		for _, server := range observers {
			observerBo := server.ToBo()
			observerBo.Architecture = archMap[meta.NewAgentInfo(server.SvrIp, server.SvrPort).String()]
			if baseResourceStat, ok := serverResourceMap[meta.ObserverSvrInfo{
				Ip:   server.SvrIp,
				Port: server.SvrPort,
			}]; !ok {
				zoneInfo.Servers = append(zoneInfo.Servers, observerBo)
				continue
			} else {
				observerBo.Stats.BaseResourceStats = baseResourceStat.ToBO()
				observerBo.Stats.FillExtendDiskStats()
				observerBo.Stats.Zone = zone.Zone
				observerBo.Stats.Ip = server.SvrIp
				observerBo.Stats.Port = server.SvrPort
			}

			if server.SvrIp == meta.OCS_AGENT.GetIp() && server.SvrPort == meta.RPC_PORT {
				observerBo.DataDir = getLocalObserverDataDir()
				observerBo.RedoDir = getLocalObserverRedoDir()
			} else {
				for _, agent := range allAgents {
					if server.SvrIp == agent.Ip && server.SvrPort == agent.RpcPort {
						agentInfo := meta.NewAgentInfo(agent.Ip, agent.Port)
						var observerInfo bo.Observer
						err := secure.SendGetRequest(agentInfo, constant.URI_OBSERVER_API_PREFIX+constant.URI_INFO, nil, &observerInfo)
						if err != nil {
							log.Warnf("Failed to get observer info from agent %s: %v", agentInfo.String(), err)
						} else {
							observerBo.DataDir = observerInfo.DataDir
							observerBo.RedoDir = observerInfo.RedoDir
						}
						break
					}
				}
			}

			zoneInfo.Servers = append(zoneInfo.Servers, observerBo)
		}
		info.Zones = append(info.Zones, zoneInfo)
	}

	// Set cluster stats
	for _, zone := range info.Zones {
		for _, server := range zone.Servers {
			info.Stats.Add(&server.Stats.BaseResourceStats)
		}
	}

	// Set tenant stats
	for _, tenant := range info.Tenants {
		tenantSysStatsMap, err := obclusterService.GetTenantMutilSysStat(tenant.Id, []int{SYS_STAT_CPU_USAGE_STAT_ID, SYS_STAT_MEMORY_USAGE_STAT_ID, SYS_STAT_MAX_CPU_STAT_ID, SYS_STAT_MEMORY_SIZE_STAT_ID})
		if err != nil {
			return nil, err
		}

		var tenantResourceStat bo.TenantResourceStat
		tenantResourceStat.TenantId = tenant.Id
		tenantResourceStat.TenantName = tenant.Name
		var cpuUsage, maxCpu, memoryUsage, memorySize float64
		if sysStat, ok := tenantSysStatsMap[SYS_STAT_CPU_USAGE_STAT_ID]; ok {
			cpuUsage = float64(sysStat.Value)
		}
		if sysStat, ok := tenantSysStatsMap[SYS_STAT_MAX_CPU_STAT_ID]; ok {
			maxCpu = float64(sysStat.Value)
			tenantResourceStat.CpuCoreTotal = (maxCpu / 100)
		}
		if sysStat, ok := tenantSysStatsMap[SYS_STAT_MEMORY_USAGE_STAT_ID]; ok {
			memoryUsage = float64(sysStat.Value)
		}
		if sysStat, ok := tenantSysStatsMap[SYS_STAT_MEMORY_SIZE_STAT_ID]; ok {
			memorySize = float64(sysStat.Value)
			tenantResourceStat.MemoryInBytesTotal = sysStat.Value
		}
		tenantResourceStat.CpuUsedPercent = cpuUsage / maxCpu * 100
		tenantResourceStat.MemoryUsedPercent = memoryUsage / memorySize * 100
		tenantDataDiskUsageMap, err := tenantService.GetTenantDataDiskUsageMap()
		if err != nil {
			return nil, err
		}
		tenantResourceStat.DataDiskUsage = tenantDataDiskUsageMap[tenant.Id]
		info.TenantStats = append(info.TenantStats, tenantResourceStat)
	}

	if info.IsStandalone {
		oblicense, err := obclusterService.GetObLicense()
		if err != nil {
			return nil, err
		}
		if oblicense != nil {
			info.License = oblicense.ToBO()
		}
	}

	return &info, nil
}

func GetObclusterLicense() (license *bo.ObLicense, err error) {
	obType, err := obclusterService.GetOBType()
	if err != nil {
		return nil, err
	}
	if obType != modelob.OBTypeStandalone {
		return nil, errors.Occur(errors.ErrCommonUnexpected, "Not a standalone cluster")
	}
	oblicense, err := obclusterService.GetObLicense()
	if err != nil {
		return nil, err
	}
	if oblicense != nil {
		return oblicense.ToBO(), nil
	}
	return
}

func getLocalObserverDataDir() string {
	dataDir, err := observerService.GetOBStringParatemerByName(constant.CONFIG_DATA_DIR)
	if err != nil {
		log.Warnf("Failed to get data dir: %v", err)
		return ""
	}
	return dataDir
}

func getLocalObserverRedoDir() string {
	clogPath := filepath.Join(global.HomePath, constant.OB_DIR_STORE, constant.OB_DIR_CLOG)
	realPath, err := filepath.EvalSymlinks(clogPath)
	if err != nil {
		log.Warnf("Failed to resolve clog symlink: %v", err)
		return ""
	}
	return realPath
}

func GetLocalObserverInfo() (*bo.Observer, error) {
	svrInfo := meta.ObserverSvrInfo{
		Ip:   meta.OCS_AGENT.GetIp(),
		Port: meta.RPC_PORT,
	}
	server, err := obclusterService.GetOBServer(svrInfo)
	if err != nil {
		return nil, err
	}

	observerBo := server.ToBo()
	observerBo.DataDir = getLocalObserverDataDir()
	observerBo.RedoDir = getLocalObserverRedoDir()

	return &observerBo, nil
}
