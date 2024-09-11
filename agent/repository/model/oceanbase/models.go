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

type ObParameters struct {
	SvrIp      string `gorm:"column:SVR_IP"`
	SvrPort    int    `gorm:"column:SVR_PORT"`
	Zone       string `gorm:"column:ZONE"`
	Scope      string `gorm:"column:SCOPE"`
	TenantId   int    `gorm:"column:TENANT_ID"`
	Name       string `gorm:"column:NAME"`
	Value      string `gorm:"column:VALUE"`
	TenantName string
}

type DbaObZones struct {
	Zone   string `gorm:"column:ZONE"`
	Status string `gorm:"column:STATUS"`
	Region string `gorm:"column:REGION"`
}

type OBServer struct {
	Zone             string    `gorm:"column:ZONE"`
	SvrIp            string    `gorm:"column:SVR_IP"`
	SvrPort          int       `gorm:"column:SVR_PORT"`
	SqlPort          int       `gorm:"column:SQL_PORT"`
	StopTime         time.Time `gorm:"column:STOP_TIME"`
	StartServiceTime time.Time `gorm:"column:START_SERVICE_TIME"`
	WithRs           string    `gorm:"column:WITH_ROOTSERVER"`
	Status           string    `gorm:"column:STATUS"`
	BuildVersion     string    `gorm:"column:BUILD_VERSION"`
}

func (OBServer) TableName() string {
	return "oceanbase.DBA_OB_SERVERS"
}

type DbaObUnitConfigs struct {
	UnitConfigID int64     `gorm:"column:UNIT_CONFIG_ID"`
	Name         string    `gorm:"column:NAME"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME"`
	ModifyTime   time.Time `gorm:"column:MODIFY_TIME"`
	MaxCpu       float64   `gorm:"column:MAX_CPU"`
	MinCpu       float64   `gorm:"column:MIN_CPU"`
	MemorySize   int64     `gorm:"column:MEMORY_SIZE"`
	LogDiskSize  int64     `gorm:"column:LOG_DISK_SIZE"`
	MaxIops      int64     `gorm:"column:MAX_IOPS"`
	MinIops      int64     `gorm:"column:MIN_IOPS"`
	IopsWeight   int64     `gorm:"column:IOPS_WEIGHT"`
}

type DbaObResourcePools struct {
	ResourcePoolID int64     `gorm:"column:RESOURCE_POOL_ID"`
	Name           string    `gorm:"column:NAME"`
	TenantID       int64     `gorm:"column:TENANT_ID"`
	CreateTime     time.Time `gorm:"column:CREATE_TIME"`
	ModifyTime     time.Time `gorm:"column:MODIFY_TIME"`
	UnitCount      int64     `gorm:"column:UNIT_COUNT"`
	UnitConfigID   int64     `gorm:"column:UNIT_CONFIG_ID"`
	ZoneList       string    `gorm:"column:ZONE_LIST"`
	ReplicaType    string    `gorm:"column:REPLICA_TYPE"`
}
