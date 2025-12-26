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

import (
	"strings"
)

type InspectionParam struct {
	Scenario string `json:"scenario" binding:"required"`
}

type QueryInspectionHistoryParam struct {
	Page      uint    `json:"page" form:"page"`
	Size      uint    `json:"size" form:"size"`
	Scenario  *string `json:"scenario" form:"scenario"`
	Sort      string  `json:"sort" form:"sort"`
	SortBy    string  `json:"sort_by" form:"-"`
	SortOrder string  `json:"sort_order" form:"-"`
}

func (p *QueryInspectionHistoryParam) Format() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Size == 0 {
		p.Size = 10
	}
	if p.Sort != "" {
		parts := strings.Split(p.Sort, ",")
		if len(parts) == 2 {
			p.SortBy = parts[0]
			p.SortOrder = parts[1]
		} else {
			p.SortBy = parts[0]
		}
	}
	if p.SortBy == "" {
		p.SortBy = "start_time"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}
