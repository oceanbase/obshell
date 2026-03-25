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

package oceanbase

import (
	"github.com/oceanbase/obshell/ob/agent/lib/parse"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type ObServerResource struct {
	SvrIp             string  `gorm:"column:SVR_IP" json:"svr_ip"`
	SvrPort           int     `gorm:"column:SVR_PORT" json:"svr_port"`
	Zone              string  `gorm:"column:ZONE" json:"zone"`
	CpuCapacity       float64 `gorm:"column:CPU_CAPACITY" json:"cpu_capacity"`               // in cores
	CpuAssigned       float64 `gorm:"column:CPU_ASSIGNED" json:"cpu_assigned"`               // in cores
	MemCapacity       float64 `gorm:"column:MEM_CAPACITY" json:"mem_capacity"`               // in GB
	MemAssigned       float64 `gorm:"column:MEM_ASSIGNED" json:"mem_assigned"`               // in GB
	LogDiskCapacity   float64 `gorm:"column:LOG_DISK_CAPACITY" json:"log_disk_capacity"`     // in GB
	LogDiskAssigned   float64 `gorm:"column:LOG_DISK_ASSIGNED" json:"log_disk_assigned"`     // in GB
	LogDiskInUse      float64 `gorm:"column:LOG_DISK_IN_USE" json:"log_disk_in_use"`         // in GB
	DataDiskCapacity  float64 `gorm:"column:DATA_DISK_CAPACITY" json:"data_disk_capacity"`   // in GB - tenant cache disk capacity
	DataDiskAssigned  float64 `gorm:"column:DATA_DISK_ASSIGNED" json:"data_disk_assigned"`   // in GB - tenant cache disk assigned
	DataDiskInUse     float64 `gorm:"column:DATA_DISK_IN_USE" json:"data_disk_in_use"`       // in GB - tenant cache disk in use
	MemoryLimit       float64 `gorm:"column:MEMORY_LIMIT" json:"memory_limit"`               // in GB
	DataDiskAllocated float64 `gorm:"column:DATA_DISK_ALLOCATED" json:"data_disk_allocated"` // in GB
}

func (ObServerResource) TableName() string {
	return "oceanbase.GV$OB_SERVERS"
}

func (o *ObServerResource) ToBO() bo.BaseResourceStats {
	var cpuPercent, memPercent, diskPercent float64
	if o.CpuCapacity > 0 {
		cpuPercent = o.CpuAssigned / o.CpuCapacity * 100
	}
	if o.MemCapacity > 0 {
		memPercent = o.MemAssigned / o.MemCapacity * 100
	}
	if o.DataDiskCapacity > 0 {
		diskPercent = o.DataDiskAssigned / o.DataDiskCapacity * 100
	}
	// Convert GB to bytes for FormatCapacity usage
	memCapacityBytes := int64(o.MemCapacity * 1024 * 1024 * 1024)
	memAssignedBytes := int64(o.MemAssigned * 1024 * 1024 * 1024)
	dataDiskCapacityBytes := int64(o.DataDiskCapacity * 1024 * 1024 * 1024)
	dataDiskAssignedBytes := int64(o.DataDiskAssigned * 1024 * 1024 * 1024)
	return bo.BaseResourceStats{
		CpuCoreTotal:           o.CpuCapacity,
		CpuCoreAssigned:        o.CpuAssigned,
		CpuCoreAssignedPercent: cpuPercent,
		MemoryTotal:            parse.FormatCapacity(memCapacityBytes),
		MemoryAssigned:         parse.FormatCapacity(memAssignedBytes),
		MemoryInBytesTotal:     memCapacityBytes,
		MemoryInBytesAssigned:  memAssignedBytes,
		MemoryAssignedPercent:  memPercent,
		DiskTotal:              parse.FormatCapacity(dataDiskCapacityBytes),
		DiskAssigned:           parse.FormatCapacity(dataDiskAssignedBytes),
		DiskInBytesTotal:       dataDiskCapacityBytes,
		DiskInBytesAssigned:    dataDiskAssignedBytes,
		DiskAssignedPercent:    diskPercent,
	}
}

// GvObUnitsAggregated aggregated row shape for GV$OB_UNITS (shared storage mode).
type GvObUnitsAggregated struct {
	SvrIp              string  `gorm:"column:svr_ip" json:"svr_ip"`
	SvrPort            int     `gorm:"column:svr_port" json:"svr_port"`
	UnitId             int     `gorm:"column:unit_id" json:"unit_id"`
	TenantId           int     `gorm:"column:tenant_id" json:"tenant_id"`
	Zone               string  `gorm:"column:zone" json:"zone"`
	IopsWeight         int64   `gorm:"column:iops_weight" json:"iops_weight"`
	NetBandwidthWeight int64   `gorm:"column:net_bandwidth_weight" json:"net_bandwidth_weight"`
	MemorySize         float64 `gorm:"column:memory_size" json:"memory_size"`           // in GB
	LogDiskSize        float64 `gorm:"column:log_disk_size" json:"log_disk_size"`       // in GB
	LogDiskInUse       float64 `gorm:"column:log_disk_in_use" json:"log_disk_in_use"`   // in GB
	DataDiskInUse      float64 `gorm:"column:data_disk_in_use" json:"data_disk_in_use"` // in GB - data disk in use
	DataDiskSize       float64 `gorm:"column:data_disk_size" json:"data_disk_size"`     // in GB - locally assigned cache disk size
}

func (GvObUnitsAggregated) TableName() string {
	return "oceanbase.GV$OB_UNITS"
}
