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
	"strings"
	"time"

	"github.com/oceanbase/obshell/ob/agent/executor/inspection/constant"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

type InspectionService struct{}

func (s *InspectionService) SaveInspectionReport(report *oceanbase.InspectionReport) error {
	if report == nil {
		return nil
	}
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Model(oceanbase.InspectionReport{}).Create(report).Error
}

func (s *InspectionService) QueryInspectionReports(scenario *string, sortBy, sortOrder string, offset, limit int) ([]oceanbase.InspectionReport, int64, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, 0, err
	}

	query := oceanbaseDb.Model(oceanbase.InspectionReport{})

	if scenario != nil && *scenario != "" {
		scenarios := strings.Split(*scenario, ",")
		// Trim spaces from each scenario value
		for i, s := range scenarios {
			scenarios[i] = strings.TrimSpace(s)
		}
		if len(scenarios) == 1 {
			query = query.Where("scenario = ?", scenarios[0])
		} else {
			query = query.Where("scenario IN ?", scenarios)
		}
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if sortBy != "" {
		order := sortBy
		if sortOrder != "" {
			order += " " + sortOrder
		} else {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("start_time DESC")
	}

	query = query.Offset(offset).Limit(limit)

	var reports []oceanbase.InspectionReport
	if err := query.Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, totalCount, nil
}

func (s *InspectionService) GetInspectionReportById(id string) (*oceanbase.InspectionReport, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	var report oceanbase.InspectionReport
	if err := oceanbaseDb.Model(oceanbase.InspectionReport{}).Where("id = ?", id).First(&report).Error; err != nil {
		return nil, err
	}

	return &report, nil
}

func (s *InspectionService) GetInspectionReportByLocalTaskId(localTaskId string) (*oceanbase.InspectionReport, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	var report oceanbase.InspectionReport
	if err := oceanbaseDb.Model(oceanbase.InspectionReport{}).
		Where("local_task_id = ? AND status != ?", localTaskId, constant.INSPECTION_STATUS_SUCCEED).
		Order("start_time DESC").
		First(&report).Error; err != nil {
		return nil, err
	}

	return &report, nil
}

// GetInspectionReportByLocalTaskIdIncludeSucceed gets inspection report by local task id, including SUCCEED status
func (s *InspectionService) GetInspectionReportByLocalTaskIdIncludeSucceed(localTaskId string) (*oceanbase.InspectionReport, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	var report oceanbase.InspectionReport
	if err := oceanbaseDb.Model(oceanbase.InspectionReport{}).
		Where("local_task_id = ?", localTaskId).
		Order("start_time DESC").
		First(&report).Error; err != nil {
		return nil, err
	}

	return &report, nil
}

// UpdateInspectionReport updates inspection report with the provided report object
// GORM's Updates will ignore zero values, so only non-zero fields will be updated
// If status is FAILED or SUCCEED and finish_time is zero, it will be automatically set to now
func (s *InspectionService) UpdateInspectionReport(report *oceanbase.InspectionReport) error {
	if report == nil {
		return nil
	}
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}

	// Truncate error message to 64KB if set
	if len(report.ErrorMessage) > 65535 {
		report.ErrorMessage = report.ErrorMessage[:65535]
	}

	// Auto-update finish_time when status is FAILED or SUCCEED (unless explicitly set)
	if (report.Status == constant.INSPECTION_STATUS_FAILED || report.Status == constant.INSPECTION_STATUS_SUCCEED) && report.FinishTime.IsZero() {
		report.FinishTime = time.Now()
	}

	return oceanbaseDb.Model(oceanbase.InspectionReport{}).Where("id = ?", report.Id).Updates(report).Error
}

// UpdateInspectionReportStatus updates only the status field
func (s *InspectionService) UpdateInspectionReportStatus(id int, status string) error {
	return s.UpdateInspectionReport(&oceanbase.InspectionReport{
		Id:     id,
		Status: status,
	})
}

// UpdateInspectionReportStatusAndError updates status and error message fields
func (s *InspectionService) UpdateInspectionReportStatusAndError(id int, status string, errorMessage string) error {
	return s.UpdateInspectionReport(&oceanbase.InspectionReport{
		Id:           id,
		Status:       status,
		ErrorMessage: errorMessage,
	})
}

// UpdateInspectionReportComplete updates all fields of the inspection report
func (s *InspectionService) UpdateInspectionReportComplete(report *oceanbase.InspectionReport) error {
	return s.UpdateInspectionReport(report)
}
