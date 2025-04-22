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

	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

type DbaObUnitConfig struct {
	UnitConfigId int       `gorm:"column:unit_config_id" json:"unit_config_id"`
	Name         string    `gorm:"column:name" json:"name"`
	MaxCpu       float64   `gorm:"column:max_cpu" json:"max_cpu"`
	MinCpu       float64   `gorm:"column:min_cpu" json:"min_cpu"`
	MemorySize   int64     `gorm:"column:memory_size" json:"memory_size"`
	LogDiskSize  int64     `gorm:"column:log_disk_size" json:"log_disk_size"`
	MaxIops      uint      `gorm:"column:max_iops" json:"max_iops"`
	MinIops      uint      `gorm:"column:min_iops" json:"min_iops"`
	GmtCreate    time.Time `gorm:"column:create_time" json:"create_time"`
	GmtModified  time.Time `gorm:"column:modify_time" json:"modify_time"`
}

func ConvertDbaObUnitConfigToObUnit(unit *DbaObUnitConfig) *bo.ObUnitConfig {
	return &bo.ObUnitConfig{
		GmtCreate:    unit.GmtCreate,
		GmtModified:  unit.GmtModified,
		UnitConfigId: unit.UnitConfigId,
		Name:         unit.Name,
		MaxCpu:       unit.MaxCpu,
		MinCpu:       unit.MinCpu,
		MemorySize:   unit.MemorySize,
		LogDiskSize:  unit.LogDiskSize,
		MaxIops:      unit.MaxIops,
		MinIops:      unit.MinIops,
	}
}
