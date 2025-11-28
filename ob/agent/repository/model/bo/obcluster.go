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

package bo

import (
	"time"

	"github.com/oceanbase/obshell/ob/agent/lib/parse"
)

type BaseResourceStats struct {
	CpuCoreTotal           float64 `json:"cpu_core_total"`
	CpuCoreAssigned        float64 `json:"cpu_core_assigned"`
	CpuCoreAssignedPercent float64 `json:"cpu_core_assigned_percent"`
	MemoryTotal            string  `json:"memory_total"`
	MemoryAssigned         string  `json:"memory_assigned"`
	MemoryInBytesTotal     int64   `json:"memory_in_bytes_total"`
	MemoryInBytesAssigned  int64   `json:"memory_in_bytes_assigned"`
	MemoryAssignedPercent  float64 `json:"memory_assigned_percent"`
	DiskTotal              string  `json:"disk_total"`
	DiskAssigned           string  `json:"disk_assigned"`
	DiskInBytesTotal       int64   `json:"disk_in_bytes_total"`
	DiskInBytesAssigned    int64   `json:"disk_in_bytes_assigned"`
	DiskAssignedPercent    float64 `json:"disk_assigned_percent"`
}

func (r *BaseResourceStats) Add(stats *BaseResourceStats) {
	r.CpuCoreTotal += stats.CpuCoreTotal
	r.CpuCoreAssigned += stats.CpuCoreAssigned
	r.MemoryInBytesTotal += stats.MemoryInBytesTotal
	r.MemoryInBytesAssigned += stats.MemoryInBytesAssigned
	r.DiskInBytesTotal += stats.DiskInBytesTotal
	r.DiskInBytesAssigned += stats.DiskInBytesAssigned

	r.CpuCoreAssignedPercent = r.CpuCoreAssigned / r.CpuCoreTotal * 100
	r.MemoryAssignedPercent = float64(r.MemoryInBytesAssigned) / float64(r.MemoryInBytesTotal) * 100
	r.DiskAssignedPercent = float64(r.DiskInBytesAssigned) / float64(r.DiskInBytesTotal) * 100
	r.MemoryTotal = parse.FormatCapacity(r.MemoryInBytesTotal)
	r.MemoryAssigned = parse.FormatCapacity(r.MemoryInBytesAssigned)
	r.DiskTotal = parse.FormatCapacity(r.DiskInBytesTotal)
	r.DiskAssigned = parse.FormatCapacity(r.DiskInBytesAssigned)
}

type ResourceStatsExtendDisk struct {
	BaseResourceStats
	DiskUsed        string  `json:"disk_used"`
	DiskFree        string  `json:"disk_free"`
	DiskInBytesUsed int64   `json:"disk_in_bytes_used"`
	DiskInBytesFree int64   `json:"disk_in_bytes_free"`
	DiskUsedPercent float64 `json:"disk_used_percent"`
}

func (r *ResourceStatsExtendDisk) FillExtendDiskStats() {
	r.DiskUsed = parse.FormatCapacity(r.DiskInBytesAssigned)
	r.DiskInBytesUsed = r.DiskInBytesAssigned
	r.DiskInBytesFree = r.DiskInBytesTotal - r.DiskInBytesAssigned
	r.DiskFree = parse.FormatCapacity(r.DiskInBytesFree)
	r.DiskUsedPercent = float64(r.DiskInBytesUsed) / float64(r.DiskInBytesTotal) * 100
}

type ServerResourceStats struct {
	ResourceStatsExtendDisk
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Zone string `json:"zone"`
}

type Observer struct {
	Id             int64               `json:"id"`
	Ip             string              `json:"ip"`
	SvrPort        int                 `json:"svr_port"`
	SqlPort        int                 `json:"sql_port"`
	Version        string              `json:"version"`
	WithRootserver bool                `json:"with_rootserver"`
	Status         string              `json:"status"`
	InnerStatus    string              `json:"inner_status"`
	StartTime      time.Time           `json:"start_time"`
	StopTime       time.Time           `json:"stop_time"`
	Stats          ServerResourceStats `json:"stats"`
	Architecture   string              `json:"architecture"`
}

type RootServer struct {
	Ip      string `json:"ip"`
	Role    string `json:"role"`
	Zone    string `json:"zone"`
	SvrPort int    `json:"svr_port"`
}

type Zone struct {
	Name        string     `json:"name"`
	IdcName     string     `json:"idc_name"`
	RegionName  string     `json:"region_name"`
	Status      string     `json:"status"`
	InnerStatus string     `json:"inner_status"`
	RootServer  RootServer `json:"root_server"`
	Servers     []Observer `json:"servers"`
}

type ClusterInfo struct {
	ClusterBasicInfo
	Stats       BaseResourceStats    `json:"stats"`
	Zones       []Zone               `json:"zones"`
	Tenants     []TenantInfo         `json:"tenants"`
	TenantStats []TenantResourceStat `json:"tenant_stats"`
}

type ClusterBasicInfo struct {
	ClusterName              string `json:"cluster_name"`
	ClusterId                int    `json:"cluster_id"`
	Status                   string `json:"status"`
	IsCommunityEdition       bool   `json:"is_community_edition"`
	IsStandalone             bool   `json:"is_standalone"`
	ObVersion                string `json:"ob_version"`
	DeadLockDetectionEnabled bool   `json:"dead_lock_detection_enabled"`
}

type TenantResourceStat struct {
	TenantId           int     `json:"tenant_id"`
	TenantName         string  `json:"tenant_name"`
	CpuUsedPercent     float64 `json:"cpu_used_percent"`
	MemoryUsedPercent  float64 `json:"memory_used_percent"`
	DataDiskUsage      int64   `json:"data_disk_usage"`
	CpuCoreTotal       float64 `json:"cpu_core_total"`
	MemoryInBytesTotal int64   `json:"memory_in_bytes_total"`
}

type ClusterParameter struct {
	Name          string                 `json:"name"`
	Scope         string                 `json:"scope"`
	EditLevel     string                 `json:"edit_level"`
	DefaultValue  string                 `json:"default_value"`
	Section       string                 `json:"section"`
	DataType      string                 `json:"data_type"`
	Info          string                 `json:"info"`
	ServerValue   []ObParameterValue     `json:"ob_parameters"`
	TenantValue   []TenantParameterValue `json:"tenant_value,omitempty"`
	Values        []string               `json:"values,omitempty"`
	IsSingleValue bool                   `json:"is_single_value"`
}

type ObParameterValue struct {
	SvrIp      string `json:"svr_ip"`
	SvrPort    int    `json:"svr_port"`
	Zone       string `json:"zone"`
	TenantId   int    `json:"tenant_id,omitempty"`
	TenantName string `json:"tenant_name,omitempty"`
	Value      string `json:"value"`
}

type TenantParameterValue struct {
	TenantId   int    `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Value      string `json:"value"`
}
