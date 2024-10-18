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

package param

type CreateTenantParam struct {
	Name         *string                `json:"name" binding:"required"`      // Tenant name.
	ZoneList     []ZoneParam            `json:"zone_list" binding:"required"` // Tenant zone list with unit config.
	Mode         string                 `json:"mode"`                         // Tenant mode, "MYSQL"(default) or "ORACLE".
	PrimaryZone  string                 `json:"primary_zone"`                 // Tenant primary_zone.
	Whitelist    *string                `json:"whitelist"`                    // Tenant whitelist.
	RootPassword string                 `json:"root_password"`                // Root password.
	Charset      string                 `json:"charset"`
	Collation    string                 `json:"collation"`
	ReadOnly     bool                   `json:"read_only"`     // Default to false.
	Comment      string                 `json:"comment"`       // Messages.
	Variables    map[string]interface{} `json:"variables"`     // Teantn global variables.
	Parameters   map[string]interface{} `json:"parameters"`    // Tenant parameters.
	Scenario     string                 `json:"scenario"`      // Tenant scenario.
	ImportScript bool                   `json:"import_script"` // whether to import script.
	TimeZone     string                 `json:"-"`
}

type DropTenantParam struct {
	Name        string `json:"-"`            // Tenant name will be ignored in request body.
	NeedRecycle *bool  `json:"need_recycle"` // Whether to recycle tenant(can be flashback).
}

type RenameTenantParam struct {
	Name    string  `json:"-"`                           // Tenant name will be ignored in request body.
	NewName *string `json:"new_name" binding:"required"` // New tenant name.
}

type ScaleOutTenantReplicasParam struct {
	ZoneList []ZoneParam `json:"zone_list" binding:"required"` // Tenant zone list with unit config.
}

type ScaleInTenantReplicasParam struct {
	Zones []string `json:"zones" binding:"required"`
}

type ModifyReplicasParam struct {
	ZoneList []ModifyReplicaZoneParam `json:"zone_list" binding:"required"` // Tenant zone list with unit config.
}

type ModifyReplicaZoneParam struct {
	Name           string  `json:"zone_name" binding:"required"`
	ReplicaType    *string `json:"replica_type"` // Replica type, "FULL"(default) or "READONLY".
	UnitConfigName *string `json:"unit_config_name"`
	UnitNum        *int    `json:"unit_num"`
}

type ZoneParam struct {
	Name        string `json:"name" binding:"required"`
	ReplicaType string `json:"replica_type"` // Replica type, "FULL"(default) or "READONLY".
	PoolParam
}

type PoolParam struct {
	UnitConfigName string `json:"unit_config_name" binding:"required"`
	UnitNum        int    `json:"unit_num" binding:"required"`
}

type ModifyTenantWhitelistParam struct {
	Whitelist *string `json:"whitelist" binding:"required"`
}

type ModifyTenantRootPasswordParam struct {
	OldPwd string  `json:"old_password"`
	NewPwd *string `json:"new_password" binding:"required"`
}

type ModifyTenantPrimaryZoneParam struct {
	PrimaryZone *string `json:"primary_zone" binding:"required"`
}

type SetTenantParametersParam struct {
	Parameters map[string]interface{} `json:"parameters" binding:"required"`
}

type SetTenantVariablesParam struct {
	Variables map[string]interface{} `json:"variables" binding:"required"`
}

// Task Param
type CreateResourcePoolTaskParam struct {
	PoolName       string
	ZoneName       string // a resource pool is bound to a zone.
	UnitConfigName string
	UnitNum        int
}
