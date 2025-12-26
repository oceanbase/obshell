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
)

type InspectionResultStatistics struct {
	CriticalCount int `json:"critical_count"`
	FailedCount   int `json:"failed_count"`
	WarningCount  int `json:"warning_count"`
	PassCount     int `json:"pass_count"`
}

type InspectionReportBriefInfo struct {
	Id                         int    `json:"id,omitempty"`
	Scenario                   string `json:"scenario" binding:"required"`
	InspectionResultStatistics `json:",inline"`
	StartTime                  time.Time `json:"start_time,omitempty"`
	FinishTime                 time.Time `json:"finish_time,omitempty"`
	LocalTaskId                string    `json:"local_task_id,omitempty"`
	Status                     string    `json:"status,omitempty"`
	ErrorMessage               string    `json:"error_message,omitempty"`
}

type InspectionItem struct {
	Name    string   `json:"name" binding:"required"`
	Results []string `json:"results,omitempty"`
}

type ResultDetail struct {
	CriticalItems []InspectionItem `json:"critical_items,omitempty"`
	WarningItems  []InspectionItem `json:"warning_items,omitempty"`
	FailedItems   []InspectionItem `json:"failed_items,omitempty"`
	PassItems     []InspectionItem `json:"pass_items,omitempty"`
}

type InspectionReport struct {
	InspectionReportBriefInfo `json:",inline"`
	ResultDetail              ResultDetail `json:"result_detail,omitempty"`
}

// PaginatedInspectionHistoryResponse represents paginated inspection history
type PaginatedInspectionHistoryResponse struct {
	Contents []InspectionReportBriefInfo `json:"contents"`
	Page     CustomPage                  `json:"page"`
}
