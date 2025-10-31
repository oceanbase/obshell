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

type ModifyWhitelistParam struct {
	Whitelist *string `json:"whitelist" binding:"required"`
}

type SetParametersParam struct {
	Parameters map[string]interface{} `json:"parameters" binding:"required"`
}

type SetVariablesParam struct {
	Variables map[string]interface{} `json:"variables" binding:"required"`
}

type CreateUserParam struct {
	UserName         string             `json:"user_name" binding:"required"`
	Password         string             `json:"password" binding:"required"`
	GlobalPrivileges []string           `json:"global_privileges"`
	DbPrivileges     []DbPrivilegeParam `json:"db_privileges"`
	HostName         string             `json:"host_name"`
}

type DbPrivilegeParam struct {
	DbName     string   `json:"db_name" binding:"required"`
	Privileges []string `json:"privileges" binding:"required"`
}
