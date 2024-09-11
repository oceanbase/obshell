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
)

type DbaObTenant struct {
	TenantID     int       `gorm:"column:TENANT_ID"`
	TenantName   string    `gorm:"column:TENANT_NAME"`
	Mode         string    `gorm:"column:COMPATIBILITY_MODE"`
	Status       string    `gorm:"column:STATUS"`
	Locked       string    `gorm:"column:LOCKED"`
	PrimaryZone  string    `gorm:"column:PRIMARY_ZONE"`
	Locality     string    `gorm:"column:LOCALITY"`
	InRecyclebin string    `gorm:"column:IN_RECYCLEBIN"`
	CreatedTime  time.Time `gorm:"column:CREATE_TIME"`
}

type DbaOBResourcePool struct {
	Name         string `gorm:"column:NAME"`
	Id           int    `gorm:"column:RESOURCE_POOL_ID"`
	ZoneList     string `gorm:"column:ZONE_LIST"`
	UnitNum      int    `gorm:"column:UNIT_COUNT"`
	UnitConfigId int    `gorm:"column:UNIT_CONFIG_ID"`
	TenantId     int    `gorm:"column:TENANT_ID"`
	ReplicaType  string `gorm:"column:REPLICA_TYPE"`
}

type CdbObSysVariable struct {
	Name  string `gorm:"column:NAME"`
	Value string `gorm:"column:VALUE"`
	Info  string `gorm:"column:INFO"`
}

type GvObParameter struct {
	Name      string `gorm:"column:NAME"`
	Value     string `gorm:"column:VALUE"`
	DataType  string `gorm:"column:DATA_TYPE"`
	Info      string `gorm:"column:INFO"`
	EditLevel string `gorm:"column:EDIT_LEVEL"`
}

type DbaRecyclebin struct {
	Name         string `gorm:"column:OBJECT_NAME"`
	OriginalName string `gorm:"column:ORIGINAL_NAME"`
	CanUndrop    string `gorm:"column:CAN_UNDROP"`
	CanPurge     string `gorm:"column:CAN_PURGE"`
}

// select * from information_schema.collations
type Collations struct {
	Charset   string `gorm:"column:CHARACTER_SET_NAME"`
	Collation string `gorm:"column:COLLATION_NAME"`
}

type ObServerCapacity struct {
	Zone            string  `gorm:"column:ZONE"`
	SvrIp           string  `gorm:"column:SVR_IP"`
	SvrPort         int     `gorm:"column:SVR_PORT"`
	SqlPort         int     `gorm:"column:SQL_PORT"`
	CpuCapacity     float64 `gorm:"column:CPU_CAPACITY"`
	CpuCapacityMax  float64 `gorm:"column:CPU_CAPACITY_MAX"`
	MemCapacity     int     `gorm:"column:MEM_CAPACITY"`
	LogDiskCapacity int     `gorm:"column:LOG_DISK_CAPACITY"`
}

func (ObServerCapacity) TableName() string {
	return "oceanbase.GV$OB_SERVERS"
}

type DbaObUnit struct {
	UnitId         int     `gorm:"column:UNIT_ID"`
	TenantId       int     `gorm:"column:TENANT_ID"`
	Status         string  `gorm:"column:STATUS"`
	ResourcePoolId int     `gorm:"column:RESOURCE_POOL_ID"`
	Zone           string  `gorm:"column:ZONE"`
	SvrIp          string  `gorm:"column:SVR_IP"`
	SvrPort        int     `gorm:"column:SVR_PORT"`
	UnitConfigId   int     `gorm:"column:UNIT_CONFIG_ID"`
	MaxCpu         float64 `gorm:"column:MAX_CPU"`
	MinCpu         float64 `gorm:"column:MIN_CPU"`
	MemorySize     int     `gorm:"column:MEMORY_SIZE"`
	LogDiskSize    int     `gorm:"column:LOG_DISK_SIZE"`
	MaxIops        int     `gorm:"column:MAX_IOPS"`
	MinIops        int     `gorm:"column:MIN_IOPS"`
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
