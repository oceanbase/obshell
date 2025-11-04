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
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	"github.com/oceanbase/obshell/seekdb/agent/lib/parse"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	obmodel "github.com/oceanbase/obshell/seekdb/model/observer"
)

const (
	STATUS_STOPPING    = "STOPPING"    // 停止中
	STATUS_STOPPED     = "STOPPED"     // 已停止
	STATUS_STARTING    = "STARTING"    // 启动中
	STATUS_RESTARTING  = "RESTARTING"  // 重启中
	STATUS_AVAILABLE   = "AVAILABLE"   // 可用
	STATUS_UNAVAILABLE = "UNAVAILABLE" // 不可用
	STATUS_UNKNOWN     = "UNKNOWN"
)

func GetObserverInfo() (info obmodel.ObserverInfo) {
	var err error
	info.LogDir = filepath.Join(path.AgentDir(), "log")
	info.BaseDir = path.AgentDir()
	info.ObshellPort = meta.OCS_AGENT.GetPort()
	info.Version, err = obclusterService.GetObVersion()
	if err != nil {
		log.Warnf("Failed to get observer version: %v", err)
		// get from sqlite
		err = agentService.GetObConfig(constant.CONFIG_OB_VERSION, &info.Version)
		if err != nil {
			log.Warnf("Failed to get observer version from sqlite: %v", err)
		}
	}
	observer, err := obclusterService.GetOBServer()
	if err != nil {
		log.Warnf("Failed to get observer: %v", err)
	}
	info.ClusterName, err = observerService.GetOBStringParatemerByName(constant.OB_PARAM_CLUSTER_NAME)
	if err != nil {
		log.Warnf("Failed to get cluster name: %v", err)
	}

	info.DataDir, err = observerService.GetOBStringParatemerByName(constant.CONFIG_DATA_DIR)
	if err != nil {
		log.Warnf("Failed to get data dir: %v", err)
		// get from sqlite
		err = agentService.GetObConfig(constant.CONFIG_DATA_DIR, &info.DataDir)
		if err != nil {
			log.Warnf("Failed to get data dir from sqlite: %v", err)
		}
	}
	if !strings.HasPrefix(info.DataDir, "/") {
		info.DataDir = filepath.Join(path.AgentDir(), info.DataDir)
	}

	info.RedoDir, err = observerService.GetOBStringParatemerByName(constant.CONFIG_REDO_DIR)
	if err != nil {
		log.Warnf("Failed to get redo dir: %v", err)
		// get from sqlite
		err = agentService.GetObConfig(constant.CONFIG_REDO_DIR, &info.RedoDir)
		if err != nil {
			log.Warnf("Failed to get redo dir from sqlite: %v", err)
		}
	}
	if !strings.HasPrefix(info.RedoDir, "/") {
		info.RedoDir = filepath.Join(path.AgentDir(), info.RedoDir)
	}

	// Get observer pid
	info.BinPath, err = process.GetObserverBinPath()
	if err != nil {
		log.Warnf("Failed to get observer pid: %v", err)
	}
	err = agentService.GetObConfig(constant.CONFIG_USER, &info.User)
	if err != nil {
		log.Warnf("Failed to get user: %v", err)
	}

	whitelist, err := tenantService.GetTenantVariable(constant.VARIABLE_OB_TCP_INVITED_NODES)
	if err != nil {
		log.Warnf("Failed to get whitelist: %v", err)
	}
	if whitelist == nil {
		info.Whitelist = ""
	} else {
		info.Whitelist = whitelist.Value
	}

	if observer != nil {
		info.Port = observer.SqlPort
		info.InnerStatus = observer.Status
		info.StartTime = &observer.StartServiceTime
		info.CreatedTime = &observer.CreateTime
		info.LifeTime = time.Since(observer.StartServiceTime).String()
	} else {
		err = agentService.GetObConfig(constant.CONFIG_MYSQL_PORT, &info.Port)
		if err != nil {
			log.Warnf("Failed to get mysql port from sqlite: %v", err)
		}
		var createTimeStr string
		err = agentService.GetObConfig(constant.CONFIG_CREATED_TIME, &createTimeStr)
		if err != nil {
			log.Warnf("Failed to get created time from sqlite: %v", err)
		}
		createTime, err := time.Parse(time.RFC3339, createTimeStr)
		if err != nil {
			log.Warnf("Failed to parse created time: %v", err)
		}
		info.CreatedTime = &createTime
	}
	info.Status = getObserverStatus()
	info.DatabaseCount, err = tenantService.GetDatabaseCount()
	if err != nil {
		log.Warnf("Failed to get database count: %v", err)
	}
	info.UserCount, err = userService.GetUserCount()
	if err != nil {
		log.Warnf("Failed to get user count: %v", err)
	}

	resource, err := obclusterService.GetObserverResource()
	if err != nil {
		log.Warnf("Failed to get observer resource: %v", err)
	}
	if resource != nil {
		info.ObserverResourceInfo = obmodel.ObserverResourceInfo{
			CpuCount:     resource.CpuCapacity,
			MemorySize:   parse.FormatCapacity(resource.MemCapacity),
			LogDiskSize:  parse.FormatCapacity(resource.LogDiskCapacity),
			DataDiskSize: parse.FormatCapacity(resource.DataDiskCapacity),
		}
	}
	info.Architecture = global.Architecture
	info.ConnectionString = fmt.Sprintf("obclient -h%s -P%d -uroot -p", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT)
	return
}

func getObserverStatus() (status string) {
	lastMaintainDag, err := localTaskService.FindLastMaintenanceDag()
	if err != nil {
		log.Warnf("Failed to find last maintenance dag: %v", err)
		return
	}
	if lastMaintainDag != nil && !lastMaintainDag.IsFinished() {
		switch lastMaintainDag.GetName() {
		case DAG_START_OBSERVER:
			return STATUS_STARTING
		case DAG_STOP_OBSERVER:
			return STATUS_STOPPING
		case DAG_RESTART_OBSERVER:
			return STATUS_RESTARTING
		default:
		}
	}
	obStatus := oceanbase.GetState()
	switch obStatus {
	case oceanbase.STATE_PROCESS_NOT_RUNNING:
		return STATUS_STOPPED
	case oceanbase.STATE_PROCESS_RUNNING:
		return STATUS_UNAVAILABLE
	case oceanbase.STATE_CONNECTION_RESTRICTED:
		return STATUS_UNAVAILABLE
	case oceanbase.STATE_CONNECTION_AVAILABLE:
		return STATUS_AVAILABLE
	default:
		return STATUS_UNKNOWN
	}
}
