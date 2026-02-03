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
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/tenant"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	modeloceanbase "github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	modelsqlite "github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/ob/agent/secure"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

// agentsToBoFromOB converts oceanbase agents to []bo.AllAgent.
func agentsToBoFromOB(agents []modeloceanbase.AllAgent) []bo.AllAgent {
	out := make([]bo.AllAgent, 0, len(agents))
	for i := range agents {
		out = append(out, *agents[i].ToBo())
	}
	return out
}

// agentsToBoFromSQLite converts sqlite agents to []bo.AllAgent.
func agentsToBoFromSQLite(agents []modelsqlite.AllAgent) []bo.AllAgent {
	out := make([]bo.AllAgent, 0, len(agents))
	for i := range agents {
		out = append(out, *agents[i].ToBo())
	}
	return out
}

// isClusterUnavailableError checks if the error indicates cluster is unavailable
func isClusterUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	// Check for observer not exist error
	if errors.Is(err, oceanbase.ERR_OBSERVER_NOT_EXIST) {
		return true
	}
	// Check for OcsAgentError with specific error codes
	if ocsAgentErr, ok := err.(errors.OcsAgentErrorInterface); ok {
		errorCode := ocsAgentErr.ErrorCode()
		return errorCode.Code == errors.ErrAgentOceanbaseNotHold.Code ||
			errorCode.Code == errors.ErrAgentOceanbaseUesless.Code
	}
	return false
}

func GetObclusterSummary() (*bo.ClusterInfo, error) {
	// Quick health check with 1-second timeout to detect unresponsive cluster early
	if err := oceanbase.QuickHealthCheck(); err != nil {
		log.Warnf("Quick health check failed, returning degraded cluster info: %v", err)
		return buildDegradedClusterInfo()
	}

	var err error
	// buildBasicClusterInfo now supports partial failure and preserves local info
	var fromLocal bool
	info, err := buildBasicClusterInfo()
	if err != nil {
		if !isClusterUnavailableError(err) {
			log.Warnf("Failed to build basic cluster info: %v", err)
			return &info, err
		}
		fromLocal = true
		log.Warnf("Cluster is unavailable, skipping basic cluster info: %v", err)
	}

	taskIdMap := buildTaskIdMap(&info)
	mainDagTaskInfo := getMainDagTaskInfoWithName()

	zones, rootServers, serverResourceMap, allAgentsBo, err := getZonesAndAgents(fromLocal)
	if err != nil {
		if !fromLocal {
			// Return partial info (basic cluster info) so caller can still use it; Zones will be empty
			return &info, err
		}
		// fromLocal: keep going with nil zones (same as ignoring local collect error before)
	}
	buildZonesIntoInfo(&info, zones, allAgentsBo, rootServers, serverResourceMap, taskIdMap, mainDagTaskInfo, fromLocal)

	if !fromLocal {
		_ = calculateClusterStats(&info)
		_ = calculateTenantStats(&info)
		_ = buildLicenseInfo(&info)
	}

	return &info, nil
}

// buildDegradedClusterInfo builds cluster info from local data when OB cluster is unavailable.
// It returns basic cluster topology from SQLite without querying OB database or remote agents.
func buildDegradedClusterInfo() (*bo.ClusterInfo, error) {
	info := bo.ClusterInfo{}

	// Get cluster name from local config (SQLite only, no OB query)
	clusterName, _ := getClusterNameFromLocal()
	info.ClusterName = clusterName

	// Get cluster ID from local config (SQLite only, no OB query)
	clusterID, _ := getClusterIDFromLocal()
	info.ClusterId = clusterID

	// Set status based on current OB state (use quick version to avoid blocking)
	info.Status = oceanbase.OBStateShortMap[oceanbase.GetStateQuick()]

	taskIdMap := buildTaskIdMap(&info)
	mainDagTaskInfo := getMainDagTaskInfoWithName()
	zones, rootServers, serverResourceMap, allAgentsBo, _ := getZonesAndAgents(true)
	buildZonesIntoInfo(&info, zones, allAgentsBo, rootServers, serverResourceMap, taskIdMap, mainDagTaskInfo, true)

	// Tenants and TenantStats will be empty (require OB queries)
	// License will be nil (requires OB query)

	return &info, nil
}

// buildZonesInfoLocal builds zones info for degraded mode when cluster is unavailable.
// Local observer uses quick methods; for remote observers it still requests info from agents (fallback to basic info on failure).
func buildZonesInfoLocal(info *bo.ClusterInfo, zones []bo.Zone, allAgents []bo.AllAgent, taskIdMap map[string]string, mainDagTaskInfo *MainDagTaskInfo) {
	for _, zone := range zones {
		zone.Status = zoneStatusMap[fixZoneStatus(&zone, mainDagTaskInfo)]
		for i := range zone.Servers {
			server := &zone.Servers[i]
			buildObserverInfoLocal(server, zone.Name, allAgents, taskIdMap, mainDagTaskInfo)
		}
		if localTaskId, ok := taskIdMap[zone.Name]; ok {
			zone.LocalTaskId = localTaskId
		}
		info.Zones = append(info.Zones, zone)
	}
}

// buildObserverInfoLocal builds observer info for degraded mode.
// For local observer it uses quick methods (no OB/remote). For remote observers it still requests info from the agent and falls back to basic info on failure.
func buildObserverInfoLocal(server *bo.Observer, zoneName string, allAgents []bo.AllAgent, taskIdMap map[string]string, mainDagTaskInfo *MainDagTaskInfo) {
	obState := oceanbase.STATE_PROCESS_NOT_RUNNING
	var agentPort int
	if server.Ip == meta.OCS_AGENT.GetIp() && server.SvrPort == meta.RPC_PORT {
		// Local observer: use quick methods
		obState = oceanbase.GetStateQuick()
		server.DataDir = getLocalObserverDataDirFast()
		server.RedoDir = getLocalObserverRedoDir()
		agentPort = meta.OCS_AGENT.GetPort()
	} else {
		// Remote observer: request info from agent, fall back to basic info on failure
		agentPort = server.ObshellPort
		agentInfo := meta.NewAgentInfo(server.Ip, agentPort)
		var observerInfo bo.Observer
		err := secure.SendGetRequest(agentInfo, constant.URI_OBSERVER_API_PREFIX+constant.URI_INFO, nil, &observerInfo)
		if err != nil {
			log.Warnf("Failed to get observer info from agent %s: %v", agentInfo.String(), err)
		} else {
			server.DataDir = observerInfo.DataDir
			server.RedoDir = observerInfo.RedoDir
			obState = observerInfo.ObStatus
		}
	}
	server.ObStatus = obState
	server.Status = observerStatusMap[fixObserverStatus(server, zoneName, obState, mainDagTaskInfo, allAgents)]
	if agentPort > 0 {
		if localTaskId, ok := taskIdMap[meta.NewAgentInfo(server.Ip, agentPort).String()]; ok {
			server.LocalTaskId = localTaskId
		}
	}
	server.ObshellPort = agentPort
}

func GetObclusterTopology() (*bo.ClusterTopology, error) {
	// Quick health check to detect unresponsive cluster early
	if err := oceanbase.QuickHealthCheck(); err != nil {
		log.Warnf("Quick health check failed, returning degraded topology: %v", err)
		return buildDegradedTopology()
	}

	info := bo.ClusterInfo{}
	taskIdMap := buildTaskIdMap(&info)
	mainDagTaskInfo := getMainDagTaskInfoWithName()

	zones, rootServers, serverResourceMap, allAgentsBo, err := getZonesAndAgents(false)
	if err != nil {
		if isClusterUnavailableError(err) {
			log.Warnf("Failed to collect zone and server data: %v", err)
			return buildDegradedTopology()
		}
		return &bo.ClusterTopology{Zones: nil}, nil
	}
	buildZonesIntoInfo(&info, zones, allAgentsBo, rootServers, serverResourceMap, taskIdMap, mainDagTaskInfo, false)
	return &bo.ClusterTopology{Zones: info.Zones}, nil
}

// buildDegradedTopology builds topology from local data without OB queries or remote requests
func buildDegradedTopology() (*bo.ClusterTopology, error) {
	info := bo.ClusterInfo{}
	taskIdMap := buildTaskIdMap(&info)
	mainDagTaskInfo := getMainDagTaskInfoWithName()
	zones, rootServers, serverResourceMap, allAgentsBo, _ := getZonesAndAgents(true)
	log.Infof("[buildDegradedTopology] collect zone and server data from local: %v", zones)
	buildZonesIntoInfo(&info, zones, allAgentsBo, rootServers, serverResourceMap, taskIdMap, mainDagTaskInfo, true)
	return &bo.ClusterTopology{Zones: info.Zones}, nil
}

// buildBasicClusterInfo builds the basic cluster information
func buildBasicClusterInfo() (info bo.ClusterInfo, err error) {
	// Get all tenant infos
	tenantOverviews, err := tenantService.GetTenantsOverView()
	if err != nil {
		return info, err
	}
	for _, tenantOverview := range tenantOverviews {
		tenantInfo, err := tenant.GetTenantInfo(tenantOverview.TenantName)
		if err != nil {
			return info, err
		}
		info.Tenants = append(info.Tenants, *tenantInfo)
	}

	// Set basic cluster info
	// Optimization: Batch get all OB parameters in one query
	paramNames := []string{"cluster", "cluster_id", "_lcl_op_interval"}
	paramsMap, err := observerService.GetOBParatemersByNames(paramNames)
	if err != nil {
		return info, err
	}
	info.ClusterName = paramsMap["cluster"]
	if clusterIdStr := paramsMap["cluster_id"]; clusterIdStr != "" {
		if clusterId, err := strconv.Atoi(clusterIdStr); err == nil {
			info.ClusterId = clusterId
		}
	}
	lclOpInterval := paramsMap["_lcl_op_interval"]
	info.DeadLockDetectionEnabled = (lclOpInterval != "0ms" && lclOpInterval != "0")
	obType, err := obclusterService.GetOBType()
	if err != nil {
		return info, err
	}
	info.IsCommunityEdition = obType == modelob.OBTypeCommunity
	info.IsStandalone = obType == modelob.OBTypeStandalone
	info.Status = oceanbase.OBStateShortMap[oceanbase.GetState()]
	version, err := obclusterService.GetObVersion()
	if err != nil {
		return info, err
	}
	info.ObVersion = version
	return info, nil
}

// collectZoneAndServerData collects zones, root servers, resource maps, and agent architecture data
func collectZoneAndServerData() (zones []bo.Zone, rootServers map[string]modeloceanbase.RootServer, serverResourceMap map[meta.ObserverSvrInfo]modeloceanbase.ObServerCapacity, err error) {
	rootServers, err = obclusterService.GetAllZoneRootServers()
	if err != nil {
		return nil, nil, nil, err
	}

	allAgents, _ := agentService.GetAllAgentsDOFromOB()
	var archMap = make(map[string]string)
	var obshellPortMap = make(map[string]int)
	for _, agent := range allAgents {
		archMap[meta.NewAgentInfo(agent.Ip, agent.RpcPort).String()] = agent.Architecture
		obshellPortMap[meta.NewAgentInfo(agent.Ip, agent.RpcPort).String()] = agent.Port
	}

	servers, err := obclusterService.GetAllOBServers()
	if err != nil {
		return nil, nil, nil, err
	}
	var zoneToServersMap = make(map[string][]bo.Observer)
	for _, server := range servers {
		if _, ok := zoneToServersMap[server.Zone]; !ok {
			zoneToServersMap[server.Zone] = make([]bo.Observer, 0)
		}
		serverBo := server.ToBo()
		serverBo.Architecture = archMap[meta.NewAgentInfo(server.SvrIp, server.SvrPort).String()]
		serverBo.ObshellPort = obshellPortMap[meta.NewAgentInfo(server.SvrIp, server.SvrPort).String()]
		zoneToServersMap[server.Zone] = append(zoneToServersMap[server.Zone], serverBo)
	}

	zonesDo, err := obclusterService.GetAllZone()
	if err != nil {
		return nil, nil, nil, err
	}
	zones = make([]bo.Zone, 0, len(zoneToServersMap))
	for _, zoneDo := range zonesDo {
		zone := bo.Zone{
			Name:        zoneDo.Zone,
			IdcName:     zoneDo.Idc,
			RegionName:  zoneDo.Region,
			Status:      zoneDo.Status,
			InnerStatus: zoneDo.Status,
			Servers:     zoneToServersMap[zoneDo.Zone],
		}
		zones = append(zones, zone)
	}

	serverResourceMap, err = obclusterService.GetAllObserverResourceMap()
	if err != nil {
		return nil, nil, nil, err
	}

	return zones, rootServers, serverResourceMap, nil
}

func collectZoneAndServerDataFromLocal() (zones []bo.Zone, rootServers map[string]modeloceanbase.RootServer, serverResourceMap map[meta.ObserverSvrInfo]modeloceanbase.ObServerCapacity, err error) {
	allAgentsSqlite, err := agentService.GetAllAgentsDO()
	if err != nil {
		return nil, nil, nil, err
	}

	var zoneMap = make(map[string]bo.Zone)
	var zoneToServersMap = make(map[string][]bo.Observer)
	for _, agent := range allAgentsSqlite {
		if _, ok := zoneMap[agent.Zone]; !ok {
			zoneMap[agent.Zone] = bo.Zone{
				Name: agent.Zone,
			}
		}
		zoneToServersMap[agent.Zone] = append(zoneToServersMap[agent.Zone], bo.Observer{
			Ip:           agent.Ip,
			SvrPort:      agent.RpcPort,
			SqlPort:      agent.MysqlPort,
			ObshellPort:  agent.Port,
			Architecture: agent.Architecture,
		})
	}
	zones = make([]bo.Zone, 0, len(zoneMap))
	for zone, servers := range zoneToServersMap {
		zoneBo := zoneMap[zone]
		zoneBo.Servers = servers
		zones = append(zones, zoneBo)
	}
	return zones, nil, nil, nil
}

// getZonesAndAgents returns zones, rootServers, serverResourceMap and allAgentsBo.
// When fromLocal is true, data comes from SQLite (rootServers and serverResourceMap are nil); otherwise from OB.
func getZonesAndAgents(fromLocal bool) (zones []bo.Zone, rootServers map[string]modeloceanbase.RootServer, serverResourceMap map[meta.ObserverSvrInfo]modeloceanbase.ObServerCapacity, allAgentsBo []bo.AllAgent, err error) {
	if fromLocal {
		zones, rootServers, serverResourceMap, err = collectZoneAndServerDataFromLocal()
		if err != nil {
			return nil, nil, nil, nil, err
		}
		allAgents, _ := agentService.GetAllAgentsDO()
		allAgentsBo = agentsToBoFromSQLite(allAgents)
		return zones, rootServers, serverResourceMap, allAgentsBo, nil
	}
	zones, rootServers, serverResourceMap, err = collectZoneAndServerData()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	allAgents, _ := agentService.GetAllAgentsDOFromOB()
	allAgentsBo = agentsToBoFromOB(allAgents)
	return zones, rootServers, serverResourceMap, allAgentsBo, nil
}

// buildZonesIntoInfo fills info.Zones from zones and agents; fromLocal selects buildZonesInfoLocal vs buildZonesInfo.
func buildZonesIntoInfo(info *bo.ClusterInfo, zones []bo.Zone, allAgentsBo []bo.AllAgent, rootServers map[string]modeloceanbase.RootServer, serverResourceMap map[meta.ObserverSvrInfo]modeloceanbase.ObServerCapacity, taskIdMap map[string]string, mainDagTaskInfo *MainDagTaskInfo, fromLocal bool) {
	if fromLocal {
		buildZonesInfoLocal(info, zones, allAgentsBo, taskIdMap, mainDagTaskInfo)
		return
	}
	_ = buildZonesInfo(info, zones, allAgentsBo, rootServers, serverResourceMap, taskIdMap, mainDagTaskInfo)
}

// buildTaskIdMap builds the task ID map based on maintenance scope
func buildTaskIdMap(info *bo.ClusterInfo) map[string]string {
	taskIdMap := make(map[string]string)
	if taskId, scopeType, targets := getMainDagTaskInfo(); taskId != "" && scopeType != "" && targets != nil {
		if scopeType == SCOPE_GLOBAL {
			info.LocalTaskId = taskId
		} else if scopeType == SCOPE_ZONE {
			for _, target := range targets {
				taskIdMap[target] = taskId
			}
		} else if scopeType == SCOPE_SERVER {
			for _, target := range targets {
				taskIdMap[target] = taskId
			}
		}
	}
	return taskIdMap
}

// buildZonesInfo builds the zones and observers information
// When zones is nil or empty, it will build observer info from agents (cluster unavailable scenario)
func buildZonesInfo(info *bo.ClusterInfo, zones []bo.Zone, allAgents []bo.AllAgent, rootServers map[string]modeloceanbase.RootServer, serverResourceMap map[meta.ObserverSvrInfo]modeloceanbase.ObServerCapacity, taskIdMap map[string]string, mainDagTaskInfo *MainDagTaskInfo) error {
	for _, zone := range zones {
		zone.Status = zoneStatusMap[fixZoneStatus(&zone, mainDagTaskInfo)]
		if rootServers != nil {
			if rootServer, ok := rootServers[zone.Name]; ok {
				zone.RootServer = rootServer.ToBO()
			}
		}

		for i := range zone.Servers {
			server := &zone.Servers[i]
			buildObserverInfo(server, zone.Name, allAgents, serverResourceMap, taskIdMap, mainDagTaskInfo)
		}
		if localTaskId, ok := taskIdMap[zone.Name]; ok {
			zone.LocalTaskId = localTaskId
		}
		info.Zones = append(info.Zones, zone)
	}
	return nil
}

// buildObserverInfo builds the observer information for a single server
func buildObserverInfo(server *bo.Observer, zoneName string, allAgents []bo.AllAgent, serverResourceMap map[meta.ObserverSvrInfo]modeloceanbase.ObServerCapacity, taskIdMap map[string]string, mainDagTaskInfo *MainDagTaskInfo) {
	obState := oceanbase.STATE_PROCESS_NOT_RUNNING
	var agentPort int
	if server.Ip == meta.OCS_AGENT.GetIp() && server.SvrPort == meta.RPC_PORT {
		// Use GetStateQuick to avoid blocking when cluster is unresponsive
		obState = oceanbase.GetStateQuick()
		server.DataDir = getLocalObserverDataDirFast()
		server.RedoDir = getLocalObserverRedoDir()
		agentPort = meta.OCS_AGENT.GetPort()
	} else {
		agentPort = server.ObshellPort
		agentInfo := meta.NewAgentInfo(server.Ip, agentPort)
		if server.InnerStatus == "ACTIVE" {
			obState = oceanbase.STATE_CONNECTION_AVAILABLE
		}
		var observerInfo bo.Observer
		err := secure.SendGetRequest(agentInfo, constant.URI_OBSERVER_API_PREFIX+constant.URI_INFO, nil, &observerInfo)
		if err != nil {
			log.Warnf("Failed to get observer info from agent %s: %v", agentInfo.String(), err)
		} else {
			server.DataDir = observerInfo.DataDir
			server.RedoDir = observerInfo.RedoDir
			obState = observerInfo.ObStatus
		}
	}
	server.ObStatus = obState
	server.Status = observerStatusMap[fixObserverStatus(server, zoneName, obState, mainDagTaskInfo, allAgents)]
	if agentPort > 0 {
		if localTaskId, ok := taskIdMap[meta.NewAgentInfo(server.Ip, agentPort).String()]; ok {
			server.LocalTaskId = localTaskId
		}
	}
	server.ObshellPort = agentPort
	if baseResourceStat, ok := serverResourceMap[meta.ObserverSvrInfo{
		Ip:   server.Ip,
		Port: server.SvrPort,
	}]; ok {
		server.Stats.BaseResourceStats = baseResourceStat.ToBO()
		server.Stats.FillExtendDiskStats()
		server.Stats.Zone = zoneName
		server.Stats.Ip = server.Ip
		server.Stats.Port = server.SvrPort
	}
}

// calculateClusterStats calculates and sets the cluster resource statistics
func calculateClusterStats(info *bo.ClusterInfo) error {
	for _, zone := range info.Zones {
		for _, server := range zone.Servers {
			info.Stats.Add(&server.Stats.BaseResourceStats)
		}
	}
	return nil
}

// calculateTenantStats calculates and sets the tenant resource statistics
func calculateTenantStats(info *bo.ClusterInfo) error {
	tenantDataDiskUsageMap, err := tenantService.GetTenantDataDiskUsageMap()
	if err != nil {
		return err
	}

	// Optimization: Batch get all tenant sys stats in one query instead of multiple separate queries
	if len(info.Tenants) > 0 {
		tenantIds := make([]int, 0, len(info.Tenants))
		for _, tenant := range info.Tenants {
			tenantIds = append(tenantIds, tenant.Id)
		}

		statIds := []int{SYS_STAT_CPU_USAGE_STAT_ID, SYS_STAT_MEMORY_USAGE_STAT_ID, SYS_STAT_MAX_CPU_STAT_ID, SYS_STAT_MEMORY_SIZE_STAT_ID}
		tenantStatsMap, err := obclusterService.GetTenantsMutilSysStatBatch(tenantIds, statIds)
		if err != nil {
			return err
		}

		for _, tenant := range info.Tenants {
			tenantSysStatsMap := tenantStatsMap[tenant.Id]
			if tenantSysStatsMap == nil {
				tenantSysStatsMap = make(map[int64]modeloceanbase.SysStat)
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
			if maxCpu > 0 {
				tenantResourceStat.CpuUsedPercent = cpuUsage / maxCpu * 100
			}
			if memorySize > 0 {
				tenantResourceStat.MemoryUsedPercent = memoryUsage / memorySize * 100
			}
			tenantResourceStat.DataDiskUsage = tenantDataDiskUsageMap[tenant.Id]
			info.TenantStats = append(info.TenantStats, tenantResourceStat)
		}
	}
	return nil
}

// buildLicenseInfo builds the license information for standalone clusters
func buildLicenseInfo(info *bo.ClusterInfo) error {
	if info.IsStandalone {
		oblicense, err := obclusterService.GetObLicense()
		if err != nil {
			return err
		}
		if oblicense != nil {
			info.License = oblicense.ToBO()
		}
	}
	return nil
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

// getLocalObserverDataDirFast returns the data directory without querying OB.
// This is used when we need fast response and OB might be unresponsive.
func getLocalObserverDataDirFast() string {
	dataDirPath := filepath.Join(global.HomePath, constant.OB_DIR_STORE)
	realPath, err := filepath.EvalSymlinks(dataDirPath)
	if err != nil {
		log.Warnf("Failed to resolve data dir symlink: %v", err)
		return dataDirPath
	}
	return realPath
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

	var observerBo bo.Observer

	// Use quick health check to determine if we should query OB
	if oceanbase.QuickHealthCheck() == nil {
		// OB is responsive, try to get server info from OB
		server, err := obclusterService.GetOBServer(svrInfo)
		if err == nil {
			observerBo = server.ToBo()
		}
	}

	// Fill in basic info if not already set
	if observerBo.Ip == "" {
		observerBo.Ip = meta.OCS_AGENT.GetIp()
		observerBo.SvrPort = meta.RPC_PORT
		observerBo.SqlPort = meta.MYSQL_PORT
	}

	// Use fast local methods that don't query OB
	observerBo.DataDir = getLocalObserverDataDirFast()
	observerBo.RedoDir = getLocalObserverRedoDir()
	observerBo.ObStatus = oceanbase.GetStateQuick()

	return &observerBo, nil
}

// isInMaintenanceScope checks if the observer or zone is in the maintenance scope of the main dag
func isInMaintenanceScope(server *bo.Observer, zoneName string, taskInfo *MainDagTaskInfo, allAgents []bo.AllAgent) bool {
	if taskInfo == nil || taskInfo.TaskId == "" {
		return false
	}

	switch taskInfo.ScopeType {
	case SCOPE_GLOBAL:
		return true
	case SCOPE_ZONE:
		for _, target := range taskInfo.Targets {
			if target == zoneName {
				return true
			}
		}
	case SCOPE_SERVER:
		// Find the corresponding obshell port
		var agentPort int
		for _, agent := range allAgents {
			if server.Ip == agent.Ip && server.SvrPort == agent.RpcPort {
				agentPort = agent.Port
				break
			}
		}
		if agentPort > 0 {
			agentInfo := meta.NewAgentInfo(server.Ip, agentPort)
			for _, target := range taskInfo.Targets {
				if target == agentInfo.String() {
					return true
				}
			}
		}
	}
	return false
}

// MainDagTaskInfo contains information about the main dag task
type MainDagTaskInfo struct {
	TaskId              string   // Main dag generic ID
	ScopeType           string   // GLOBAL/ZONE/SERVER
	Targets             []string // Scope targets
	DagName             string   // Main dag name (e.g., DAG_START_OB)
	StopObserverProcess bool     // Only for stop dag, if true, will terminate the observer process
}

// getMainDagTaskInfo returns the main dag task ID, scope type, targets and dag name
// It checks the last maintenance sub dag, and gets task ID and scope from the main dag
func getMainDagTaskInfo() (taskId string, scopeType string, targets []string) {
	mainDagTaskInfo := getMainDagTaskInfoWithName()
	if mainDagTaskInfo == nil {
		return "", "", nil
	}
	return mainDagTaskInfo.TaskId, mainDagTaskInfo.ScopeType, mainDagTaskInfo.Targets
}

// getMainDagTaskInfoWithName returns the main dag task information including dag name
func getMainDagTaskInfoWithName() *MainDagTaskInfo {
	// Get the last maintenance dag (sub dag like DAG_STOP_OBSERVER)
	lastMaintainDag, err := localTaskService.FindLastMaintenanceDag()
	if err != nil {
		log.Warnf("Failed to find last maintenance dag: %v", err)
		return nil
	}
	if lastMaintainDag == nil {
		log.Infof("No maintenance dag found")
		return nil
	}
	if lastMaintainDag.IsSuccess() {
		log.Infof("Maintenance dag %d is succeed, name: %s", lastMaintainDag.GetID(), lastMaintainDag.GetName())
		return nil
	}

	subDagCtx := lastMaintainDag.GetContext()
	if subDagCtx == nil {
		log.Warnf("Maintenance dag context is nil for dag id=%d, name=%s",
			lastMaintainDag.GetID(), lastMaintainDag.GetName())
		return nil
	}

	// Get main dag ID from sub dag context
	var mainDagID string
	if err := subDagCtx.GetParamWithValue(PARAM_MAIN_DAG_ID, &mainDagID); err != nil {
		log.Infof("Failed to get main dag ID from sub dag context: %v", err)
		return nil
	}

	log.Infof("Sub dag has main dag ID: %s", mainDagID)

	var scope param.Scope
	if err := subDagCtx.GetParamWithValue(PARAM_SCOPE, &scope); err != nil {
		log.Warnf("Failed to get scope from main dag context: %v", err)
		return nil
	}
	var mainDagName string
	if err := subDagCtx.GetParamWithValue(PARAM_MAIN_DAG_NAME, &mainDagName); err != nil {
		log.Warnf("Failed to get main dag name from main dag context: %v", err)
		return nil
	}

	var stopObserverProcess bool
	if err := subDagCtx.GetParamWithValue(PARAM_STOP_OBSERVER_PROCESS, &stopObserverProcess); err != nil {
		log.Warnf("Failed to get stop observer process from main dag context: %v", err)
	} // ignore the error, it's not critical

	log.Infof("Got scope from main dag context: type=%s, target=%v", scope.Type, scope.Target)

	return &MainDagTaskInfo{
		TaskId:              mainDagID,
		ScopeType:           scope.Type,
		Targets:             scope.Target,
		DagName:             mainDagName,
		StopObserverProcess: stopObserverProcess,
	}
}
