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

	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
)

type DbaObSysVariable struct {
	Name  string `gorm:"column:NAME" json:"name"`
	Value string `gorm:"column:VALUE" json:"value"`
	Info  string `gorm:"column:INFO" json:"info"`
}

type GvObParameter struct {
	Name      string `gorm:"column:NAME" json:"name"`
	Value     string `gorm:"column:VALUE" json:"value"`
	DataType  string `gorm:"column:DATA_TYPE" json:"data_type"`
	Info      string `gorm:"column:INFO" json:"info"`
	EditLevel string `gorm:"column:EDIT_LEVEL" json:"edit_level"`
}

type ObServerResource struct {
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

func (ObServerResource) TableName() string {
	return "oceanbase.GV$OB_SERVERS"
}

type CdbObMajorCompaction struct {
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
