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
	"time"

	"github.com/oceanbase/obshell/agent/lib/parse"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

type TenantOverview struct {
	DbaObTenant       `json:",inline"`
	ConnectionStrings []bo.ObproxyAndConnectionString `json:"connection_strings"`
}

type DbaObTenant struct {
	TenantID     int       `gorm:"column:TENANT_ID" json:"tenant_id"`
	TenantName   string    `gorm:"column:TENANT_NAME" json:"tenant_name"`
	Mode         string    `gorm:"column:COMPATIBILITY_MODE" json:"mode"`
	Status       string    `gorm:"column:STATUS" json:"status"`
	Locked       string    `gorm:"column:LOCKED" json:"locked"`
	PrimaryZone  string    `gorm:"column:PRIMARY_ZONE" json:"primary_zone"`
	Locality     string    `gorm:"column:LOCALITY" json:"locality"`
	InRecyclebin string    `gorm:"column:IN_RECYCLEBIN" json:"in_recyclebin"`
	CreatedTime  time.Time `gorm:"column:CREATE_TIME" json:"created_time"`
	ReadOnly     bool      `json:"read_only"`
}

type DbaObResourcePool struct {
	ResourcePoolID int       `gorm:"column:RESOURCE_POOL_ID" json:"id"`
	Name           string    `gorm:"column:NAME" json:"name"`
	ZoneList       string    `gorm:"column:ZONE_LIST" json:"zone_list"`
	UnitNum        int       `gorm:"column:UNIT_COUNT" json:"unit_num"`
	UnitCount      int       `gorm:"column:UNIT_COUNT"`
	UnitConfigId   int       `gorm:"column:UNIT_CONFIG_ID" json:"unit_config_id"`
	TenantId       int       `gorm:"column:TENANT_ID" json:"tenant_id"`
	ReplicaType    string    `gorm:"column:REPLICA_TYPE" json:"replica_type"`
	CreateTime     time.Time `gorm:"column:CREATE_TIME"`
	ModifyTime     time.Time `gorm:"column:MODIFY_TIME"`
}

type CdbObSysVariable struct {
	Name  string `gorm:"column:NAME" json:"name"`
	Value string `gorm:"column:VALUE" json:"value"`
	Info  string `gorm:"column:INFO" json:"info"`
}

type ObSysVariableWithValue struct {
	Name  string `gorm:"column:Variable_name" json:"name"`
	Value string `gorm:"column:Value" json:"value"`
}

type GvObParameter struct {
	Name      string `gorm:"column:NAME" json:"name"`
	Value     string `gorm:"column:VALUE" json:"value"`
	DataType  string `gorm:"column:DATA_TYPE" json:"data_type"`
	Info      string `gorm:"column:INFO" json:"info"`
	EditLevel string `gorm:"column:EDIT_LEVEL" json:"edit_level"`
}

type DbaRecyclebin struct {
	Name         string `gorm:"column:OBJECT_NAME" json:"object_name"`
	OriginalName string `gorm:"column:ORIGINAL_NAME" json:"original_tenant_name"`
	CanUndrop    string `gorm:"column:CAN_UNDROP" json:"can_undrop"`
	CanPurge     string `gorm:"column:CAN_PURGE" json:"can_purge"`
}

type ObServerCapacity struct {
	Zone             string  `gorm:"column:ZONE"`
	SvrIp            string  `gorm:"column:SVR_IP"`
	SvrPort          int     `gorm:"column:SVR_PORT"`
	SqlPort          int     `gorm:"column:SQL_PORT"`
	CpuCapacity      float64 `gorm:"column:CPU_CAPACITY"`
	CpuCapacityMax   float64 `gorm:"column:CPU_CAPACITY_MAX"`
	CpuAssigned      float64 `gorm:"column:CPU_ASSIGNED"`
	CpuAssignedMax   float64 `gorm:"column:CPU_ASSIGNED_MAX"`
	MemCapacity      int64   `gorm:"column:MEM_CAPACITY"`
	MemAssigned      int64   `gorm:"column:MEM_ASSIGNED"`
	LogDiskCapacity  int64   `gorm:"column:LOG_DISK_CAPACITY"`
	LogDiskAssigned  int64   `gorm:"column:LOG_DISK_ASSIGNED"`
	LogDiskInUse     int64   `gorm:"column:LOG_DISK_IN_USE"`
	DataDiskCapacity int64   `gorm:"column:DATA_DISK_CAPACITY"`
	DataDiskAssigned int64   `gorm:"column:DATA_DISK_IN_USE"`
}

func (ObServerCapacity) TableName() string {
	return "oceanbase.GV$OB_SERVERS"
}

func (o *ObServerCapacity) ToBO() bo.BaseResourceStats {
	return bo.BaseResourceStats{
		CpuCoreTotal:           o.CpuCapacity,
		CpuCoreAssigned:        o.CpuAssigned,
		CpuCoreAssignedPercent: o.CpuAssigned / o.CpuCapacity * 100,
		MemoryTotal:            parse.FormatCapacity(o.MemCapacity),
		MemoryAssigned:         parse.FormatCapacity(o.MemAssigned),
		MemoryInBytesTotal:     o.MemCapacity,
		MemoryInBytesAssigned:  o.MemAssigned,
		MemoryAssignedPercent:  float64(o.MemAssigned) / float64(o.MemCapacity) * 100,
		DiskTotal:              parse.FormatCapacity(o.DataDiskCapacity),
		DiskAssigned:           parse.FormatCapacity(o.DataDiskAssigned),
		DiskInBytesTotal:       o.DataDiskCapacity,
		DiskInBytesAssigned:    o.DataDiskAssigned,
		DiskAssignedPercent:    float64(o.DataDiskAssigned) / float64(o.DataDiskCapacity) * 100,
	}
}

type DbaObUnit struct {
	UnitId         int     `gorm:"column:UNIT_ID" json:"unit_id"`
	TenantId       int     `gorm:"column:TENANT_ID" json:"tenant_id"`
	Status         string  `gorm:"column:STATUS" json:"status"`
	ResourcePoolId int     `gorm:"column:RESOURCE_POOL_ID" json:"resource_pool_id"`
	Zone           string  `gorm:"column:ZONE" json:"zone"`
	SvrIp          string  `gorm:"column:SVR_IP" json:"svr_ip"`
	SvrPort        int     `gorm:"column:SVR_PORT" json:"svr_port"`
	UnitConfigId   int     `gorm:"column:UNIT_CONFIG_ID" json:"unit_config_id"`
	MaxCpu         float64 `gorm:"column:MAX_CPU" json:"max_cpu"`
	MinCpu         float64 `gorm:"column:MIN_CPU" json:"min_cpu"`
	MemorySize     int64   `gorm:"column:MEMORY_SIZE" json:"memory_size"`
	LogDiskSize    int64   `gorm:"column:LOG_DISK_SIZE" json:"log_disk_size"`
	MaxIops        uint    `gorm:"column:MAX_IOPS" json:"max_iops"`
	MinIops        uint    `gorm:"column:MIN_IOPS" json:"min_iops"`
}

type DbaObTenantJob struct {
	JobId      int       `gorm:"column:JOB_ID"`
	JobType    string    `gorm:"column:JOB_TYPE"`
	JobStatus  string    `gorm:"column:JOB_STATUS"`
	ResultCode int       `gorm:"column:RESULT_CODE"`
	Progress   int       `gorm:"column:PROGRESS"`
	TenantId   int       `gorm:"column:TENANT_ID"`
	SqlText    string    `gorm:"column:SQL_TEXT"`
	ExtraInfo  string    `gorm:"column:EXTRA_INFO"`
	StartTime  time.Time `gorm:"column:START_TIME"`
	ModifyTime time.Time `gorm:"column:MODIFY_TIME"`
}

type CdbObMajorCompaction struct {
	TenantId           int       `gorm:"column:TENANT_ID"`
	FrozenScn          int64     `gorm:"column:FROZEN_SCN"`
	FrozenTime         time.Time `gorm:"column:FROZEN_TIME"`
	GlobalBroadcastScn int64     `gorm:"column:GLOBAL_BROADCAST_SCN"`
	LastScn            int64     `gorm:"column:LAST_SCN"`
	LastFinishTime     time.Time `gorm:"column:LAST_FINISH_TIME"`
	StartTime          time.Time `gorm:"column:START_TIME"`
	Status             string    `gorm:"column:STATUS"`
	IsError            string    `gorm:"column:IS_ERROR"`
	IsSuspended        string    `gorm:"column:IS_SUSPENDED"`
	Info               string    `gorm:"column:INFO"`
}

func (CdbObMajorCompaction) TableName() string {
	return "oceanbase.CDB_OB_MAJOR_COMPACTION"
}

func (c *CdbObMajorCompaction) ToBO() *bo.TenantCompaction {
	return &bo.TenantCompaction{
		TenantId:           c.TenantId,
		FrozenScn:          c.FrozenScn,
		FrozenTime:         c.FrozenTime,
		GlobalBroadcastScn: c.GlobalBroadcastScn,
		LastScn:            c.LastScn,
		LastFinishTime:     c.LastFinishTime,
		StartTime:          c.StartTime,
		Status:             c.Status,
		IsError:            c.IsError,
		IsSuspended:        c.IsSuspended,
	}
}
