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
package inspection

import (
	"github.com/oceanbase/obshell/ob/agent/executor/inspection/constant"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/param"
)

func GetInspectionHistory(p *param.QueryInspectionHistoryParam) (*bo.PaginatedInspectionHistoryResponse, error) {
	p.Format()

	offset := int((p.Page - 1) * p.Size)
	limit := int(p.Size)

	reports, totalCount, err := inspectionService.QueryInspectionReports(
		p.Scenario,
		p.SortBy,
		p.SortOrder,
		offset,
		limit,
	)
	if err != nil {
		return nil, err
	}

	contents := make([]bo.InspectionReportBriefInfo, 0, len(reports))
	for _, report := range reports {
		status := report.Status
		// If status is RUNNING, check task status
		if status == constant.INSPECTION_STATUS_RUNNING {
			taskStatus, err := getTaskStatusForInspectionReport(report.LocalTaskId)
			if err != nil {
				// Task not found, set status to DELETED
				status = constant.INSPECTION_STATUS_DELETED
				inspectionService.UpdateInspectionReportStatus(report.Id, constant.INSPECTION_STATUS_DELETED)
			} else {
				status = taskStatus
				if taskStatus != constant.INSPECTION_STATUS_RUNNING {
					inspectionService.UpdateInspectionReportStatus(report.Id, taskStatus)
				}
			}
		}

		contents = append(contents, bo.InspectionReportBriefInfo{
			Id:       report.Id,
			Scenario: report.Scenario,
			InspectionResultStatistics: bo.InspectionResultStatistics{
				CriticalCount: report.CriticalCount,
				FailedCount:   report.FailCount,
				WarningCount:  report.WarningCount,
				PassCount:     report.PassCount,
			},
			StartTime:    report.StartTime,
			FinishTime:   report.FinishTime,
			LocalTaskId:  report.LocalTaskId,
			Status:       status,
			ErrorMessage: report.ErrorMessage,
		})
	}

	totalPages := uint(totalCount) / p.Size
	if uint(totalCount)%p.Size > 0 {
		totalPages++
	}

	return &bo.PaginatedInspectionHistoryResponse{
		Contents: contents,
		Page: bo.CustomPage{
			Number:        uint64(p.Page),
			Size:          uint64(p.Size),
			TotalPages:    uint64(totalPages),
			TotalElements: uint64(totalCount),
		},
	}, nil
}

func GetInspectionReport(id string) (*bo.InspectionReport, error) {
	report, err := inspectionService.GetInspectionReportById(id)
	if err != nil {
		return nil, err
	}

	// If status is RUNNING, check task status
	if report.Status == constant.INSPECTION_STATUS_RUNNING {
		status, err := getTaskStatusForInspectionReport(report.LocalTaskId)
		if err != nil {
			// Task not found, set status to DELETED
			report.Status = constant.INSPECTION_STATUS_DELETED
			inspectionService.UpdateInspectionReportStatus(report.Id, constant.INSPECTION_STATUS_DELETED)
		} else {
			report.Status = status
			if status != constant.INSPECTION_STATUS_RUNNING {
				inspectionService.UpdateInspectionReportStatus(report.Id, status)
			}
		}
	}

	return parseInspectionReportFromJSON(
		report.Report,
		report.Scenario,
		report.StartTime,
		report.FinishTime,
		report.CriticalCount,
		report.FailCount,
		report.WarningCount,
		report.PassCount,
		report.LocalTaskId,
		report.Status,
		report.ErrorMessage,
	)
}
