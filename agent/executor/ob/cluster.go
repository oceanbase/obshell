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
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/tenant"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

// only for mysql
func GetObclusterSummary() (*bo.ClusterInfo, *errors.OcsAgentError) {
	var info bo.ClusterInfo

	// Get all tenant infos
	tenantOverviews, err := tenantService.GetTenantsOverViewByMode(constant.MYSQL_MODE)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
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
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	if err := observerService.GetOBParatemerByName("cluster_id", &info.ClusterId); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	isCommunityEdition, err := obclusterService.IsCommunityEdition()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	info.IsCommunityEdition = isCommunityEdition
	info.Status = oceanbase.OBStateShortMap[oceanbase.GetState()]
	version, err := obclusterService.GetObVersion()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	info.ObVersion = version

	rootServers, err := obclusterService.GetAllZoneRootServers()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}

	zones, err := obclusterService.GetAllZone()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}

	serverResourceMap, err := obclusterService.GetAllObserverResourceMap()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
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
			return nil, errors.Occur(errors.ErrUnexpected, err.Error())
		}

		for _, server := range observers {
			observerBo := server.ToBo()
			observerBo.Architecture = archMap[meta.NewAgentInfo(server.SvrIp, server.SvrPort).String()]
			if baseResourceStat, ok := serverResourceMap[meta.ObserverSvrInfo{
				Ip:   server.SvrIp,
				Port: server.SvrPort,
			}]; !ok {
				continue
			} else {
				observerBo.Stats.BaseResourceStats = baseResourceStat.ToBO()
				observerBo.Stats.FillExtendDiskStats()
				observerBo.Stats.Zone = zone.Zone
				observerBo.Stats.Ip = server.SvrIp
				observerBo.Stats.Port = server.SvrPort
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
			return nil, errors.Occur(errors.ErrUnexpected, err.Error())
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
		}
		if sysStat, ok := tenantSysStatsMap[SYS_STAT_MEMORY_USAGE_STAT_ID]; ok {
			memoryUsage = float64(sysStat.Value)
		}
		if sysStat, ok := tenantSysStatsMap[SYS_STAT_MEMORY_SIZE_STAT_ID]; ok {
			memorySize = float64(sysStat.Value)
		}
		tenantResourceStat.CpuUsedPercent = cpuUsage / maxCpu * 100
		tenantResourceStat.MemoryUsedPercent = memoryUsage / memorySize * 100
		tenantDataDiskUsageMap, err := tenantService.GetTenantDataDiskUsageMap()
		if err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err.Error())
		}
		tenantResourceStat.DataDiskUsage = tenantDataDiskUsageMap[tenant.Id]
		info.TenantStats = append(info.TenantStats, tenantResourceStat)
	}
	return &info, nil
}
