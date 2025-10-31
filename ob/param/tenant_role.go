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

type ModifyRoleParam struct {
	TenantRootPasswordParam
	Roles []string `json:"roles" binding:"required"`
}

type ModifyRoleGlobalPrivilegeParam struct {
	TenantRootPasswordParam
	GlobalPrivileges []string `json:"global_privileges"`
}

type DbObjectParam struct {
	ObjectType string `json:"object_type"`
	ObjectName string `json:"object_name" binding:"required"`
	Owner      string `json:"owner" binding:"required"`
}

type ObjectPrivilegeParam struct {
	ObjectType string   `json:"object_type"`
	ObjectName string   `json:"object_name" binding:"required"`
	Owner      string   `json:"owner" binding:"required"`
	Privileges []string `json:"privileges" binding:"required"`
}

type ObjectPrivilegeOperationParam struct {
	TenantRootPasswordParam
	ObjectPrivileges []ObjectPrivilegeParam `json:"object_privileges" binding:"required"`
}

type ModifyObjectPrivilegeParam = ObjectPrivilegeOperationParam
type RevokeObjectPrivilegeParam = ObjectPrivilegeOperationParam
type GrantObjectPrivilegeParam = ObjectPrivilegeOperationParam

type CreateRoleParam struct {
	TenantRootPasswordParam
	RoleName         string   `json:"role_name" binding:"required"`
	GlobalPrivileges []string `json:"global_privileges"`
	Roles            []string `json:"roles"`
}

type DropRoleParam struct {
	TenantRootPasswordParam
}
