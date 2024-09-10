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

	"github.com/oceanbase/obshell/agent/constant"
)

type DbaObTenantJobBo struct {
	JobId         int
	JobType       string
	JobStatus     string
	ResultCode    int
	Progress      int
	TenantId      int
	SqlText       string
	ExtraInfo     string
	StartTime     time.Time
	ModifyTime    time.Time
	CurrentTarget interface{}
}

func (a *DbaObTenantJobBo) TargetIs(target interface{}) bool {
	switch a.JobType {
	case constant.ALTER_TENANT_LOCALITY:
		current, ok := a.CurrentTarget.(map[string]string)
		if !ok {
			return false
		}
		target, ok := target.(map[string]string)
		if !ok {
			return false
		}
		if len(current) != len(target) {
			return false
		}
		for zone, locality := range current {
			if locality != target[zone] {
				return false
			}
		}
	case constant.ALTER_RESOURCE_TENANT_UNIT_NUM:
		current, ok := a.CurrentTarget.(int)
		if !ok {
			return false
		}
		target, ok := target.(int)
		if !ok {
			return false
		}
		if current != target {
			return false
		}
	case constant.ALTER_TENANT_PRIMARY_ZONE:
		current, ok := a.CurrentTarget.([]string)
		if !ok {
			return false
		}
		target, ok := target.([]string)
		if !ok {
			return false
		}
		if len(current) < len(target) || len(current)-len(target) > 1 {
			return false
		}
		for i, zonesStr := range target {
			if current[i] != zonesStr {
				return false
			}
		}
	}
	return true
}

type ObUnitConfig struct {
	GmtCreate    time.Time `json:"create_time"`
	GmtModified  time.Time `json:"modify_time"`
	UnitConfigId int       `json:"unit_config_id"`
	Name         string    `json:"name"`
	MaxCpu       float64   `json:"max_cpu"`
	MinCpu       float64   `json:"min_cpu"`
	MemorySize   int       `json:"memory_size"`
	LogDiskSize  int       `json:"log_disk_size"`
	MaxIops      int       `json:"max_iops"`
	MinIops      int       `json:"min_iops"`
}

type ResourcePoolWithUnit struct {
	Name     string        `json:"pool_name"`
	Id       int           `json:"pool_id"`
	ZoneList string        `json:"zone_list"`
	UnitNum  int           `json:"unit_num"`
	Unit     *ObUnitConfig `json:"unit_config"`
}

type TenantInfo struct {
	Name         string                  `json:"tenant_name"`
	Id           int                     `json:"tenant_id"`
	CreatedTime  time.Time               `json:"created_time"`
	Mode         string                  `json:"mode"`
	Status       string                  `json:"status"`
	Locked       string                  `json:"locked"`
	PrimaryZone  string                  `json:"primary_zone"`
	Locality     string                  `json:"locality"`
	InRecyclebin string                  `json:"in_recyclebin"`
	Charset      string                  `json:"charset"`   // Only for ORACLE tenant
	Collation    string                  `json:"collation"` // Only for ORACLE tenant
	Whitelist    string                  `json:"whitelist"`
	Pools        []*ResourcePoolWithUnit `json:"pools"`
}
