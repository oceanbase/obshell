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

import "time"

type ObproxyAndConnectionString struct {
	Type             string `json:"type"`
	ObProxyAddress   string `json:"obproxy_address"`
	ObProxyPort      int    `json:"obproxy_port"`
	ConnectionString string `json:"connection_string"`
}

type DbPrivilege struct {
	DbName     string   `json:"db_name"`
	Privileges []string `json:"privileges"`
}

type ObUserSessionStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type ObUserStats struct {
	Session *ObUserSessionStats `json:"session"`
}

type ObUser struct {
	UserName            string                       `json:"user_name"`
	IsLocked            bool                         `json:"is_locked"`
	CreateTime          time.Time                    `json:"create_time"`
	ConnectionStrings   []ObproxyAndConnectionString `json:"connection_strings"`
	AccessibleDatabases []string                     `json:"accessible_databases"`
	GrantedRoles        []string                     `json:"granted_roles"`
	GlobalPrivileges    []string                     `json:"global_privileges"`
	DbPrivileges        []DbPrivilege                `json:"db_privileges"`
	ObjectPrivileges    []ObjectPrivilege            `json:"object_privileges"` // only for oracle tenant
}

type ObRole struct {
	Role        string    `json:"role"`
	GmtCreate   time.Time `json:"gmt_create"`
	GmtModified time.Time `json:"gmt_modified"`

	GlobalPrivileges []string          `json:"global_privileges"`
	ObjectPrivileges []ObjectPrivilege `json:"object_privileges"`
	GrantedRoles     []string          `json:"granted_roles"`
	UserGrantees     []string          `json:"user_grantees"`
	RoleGrantees     []string          `json:"role_grantees"`
}

type DbaObjectBo struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	FullName string `json:"full_name"`
}

type ObjectPrivilege struct {
	Object     DbaObjectBo `json:"object"`
	Privileges []string    `json:"privileges"`
}
