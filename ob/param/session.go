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

type QueryTenantSessionParam struct {
	User         string   `form:"user"`
	Db           string   `form:"db"`
	Host         string   `form:"host"`
	SessionId    int64    `form:"id"`
	ObserverList []string `form:"-"`
	Observers    string   `form:"observer_list"`
	ActiveOnly   bool     `form:"active_only"`
	Sort         string   `form:"sort"`
	SortBy       string   `form:"-"`
	SortOrder    string   `form:"-"`
	CustomPageQuery
}

func (p *QueryTenantSessionParam) Format() {
	p.CustomPageQuery.Format()
	if p.Sort != "" {
		parts := strings.Split(p.Sort, ",")
		if len(parts) == 2 {
			p.SortBy = parts[0]
			p.SortOrder = parts[1]
		} else {
			p.SortBy = parts[0]
		}
	}
	if p.Observers != "" {
		p.ObserverList = strings.Split(p.Observers, ",")
	}
}

type KillTenantSessionsParam struct {
	SessionIds []int `json:"session_ids"`
}

type KillTenantSessionQueryParam struct {
	SessionIds []int `json:"session_ids"`
}

type QueryTenantDeadLocksParam struct {
	CustomPageQuery
}

func (p *QueryTenantDeadLocksParam) Format() {
	p.CustomPageQuery.Format()
}
