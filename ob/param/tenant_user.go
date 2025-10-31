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

import "strings"

type ModifyUserGlobalPrivilegeParam struct {
	TenantRootPasswordParam
	GlobalPrivileges []string `json:"global_privileges"`
}

type ModifyUserDbPrivilegeParam struct {
	TenantRootPasswordParam
	DbPrivileges []DbPrivilegeParam `json:"db_privileges" binding:"required"`
}

type ChangeUserPasswordParam struct {
	TenantRootPasswordParam
	NewPassword string `json:"new_password"`
}

type ListUsersQueryParam struct {
	Sort string `form:"sort"`
	CustomPageQuery
	SortBy    string `form:"-"`
	SortOrder string `form:"-"`
}

func (p *ListUsersQueryParam) Format() {
	if p.Sort != "" {
		parts := strings.Split(p.Sort, ",")
		if len(parts) == 2 {
			p.SortBy = parts[0]
			p.SortOrder = parts[1]
		} else {
			p.SortBy = parts[0]
		}
	}
	if p.SortBy != "create_time" {
		p.SortBy = "create_time"
	}
	if p.SortOrder != "asc" && p.SortOrder != "desc" {
		p.SortOrder = "asc"
	}
	p.CustomPageQuery.Format()
}
