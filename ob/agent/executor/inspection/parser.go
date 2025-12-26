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
	"encoding/json"
	"strings"
	"time"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

func parseResult(result string) (*ObdiagResult, error) {
	var report ObdiagResult
	if err := json.Unmarshal([]byte(result), &report); err != nil {
		return nil, errors.Wrap(err, "failed to parse inspection result json")
	}
	return &report, nil
}

func parseReportBriefInfo(report *ObdiagResult) (*bo.InspectionReportBriefInfo, error) {
	criticalCount := 0
	failedCount := 0
	warningCount := 0
	passCount := 0

	for _, v := range report.Data.Observer.Critical {
		criticalCount += len(v)
	}
	for _, v := range report.Data.Observer.Fail {
		failedCount += len(v)
	}
	for _, v := range report.Data.Observer.Warning {
		warningCount += len(v)
	}
	for _, results := range report.Data.Observer.All {
		for _, value := range results {
			if strings.ToLower(value) == "all pass" {
				passCount++
			}
		}
	}
	return &bo.InspectionReportBriefInfo{
		InspectionResultStatistics: bo.InspectionResultStatistics{
			CriticalCount: criticalCount,
			FailedCount:   failedCount,
			WarningCount:  warningCount,
			PassCount:     passCount,
		},
	}, nil
}

func parseInspectionReportFromJSON(reportJSON string, scenario string, startTime, finishTime time.Time, criticalCount, failCount, warningCount, passCount int, localTaskId string, status string, errorMessage string) (*bo.InspectionReport, error) {
	briefInfo := bo.InspectionReportBriefInfo{
		Scenario: scenario,
		InspectionResultStatistics: bo.InspectionResultStatistics{
			CriticalCount: criticalCount,
			FailedCount:   failCount,
			WarningCount:  warningCount,
			PassCount:     passCount,
		},
		StartTime:    startTime,
		FinishTime:   finishTime,
		LocalTaskId:  localTaskId,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	var resultDetail bo.ResultDetail
	if reportJSON != "" {
		obdiagResult, err := parseResult(reportJSON)
		if err == nil {
			resultDetail = bo.ResultDetail{
				CriticalItems: newInspectionItemsFromMap(obdiagResult.Data.Observer.Critical),
				WarningItems:  newInspectionItemsFromMap(obdiagResult.Data.Observer.Warning),
				FailedItems:   newInspectionItemsFromMap(obdiagResult.Data.Observer.Fail),
				PassItems:     newPassInspectionItemsFromMap(obdiagResult.Data.Observer.All),
			}
		}
	}

	return &bo.InspectionReport{
		InspectionReportBriefInfo: briefInfo,
		ResultDetail:              resultDetail,
	}, nil
}

func newInspectionItemsFromMap(m map[string][]string) []bo.InspectionItem {
	items := make([]bo.InspectionItem, 0, len(m))
	for name, results := range m {
		items = append(items, bo.InspectionItem{
			Name:    name,
			Results: results,
		})
	}
	return items
}

func newPassInspectionItemsFromMap(m map[string][]string) []bo.InspectionItem {
	items := make([]bo.InspectionItem, 0)
	for name, results := range m {
		// Only include items where at least one result is "all pass"
		hasPass := false
		for _, value := range results {
			if strings.ToLower(value) == "all pass" {
				hasPass = true
				break
			}
		}
		if hasPass {
			items = append(items, bo.InspectionItem{
				Name:    name,
				Results: results,
			})
		}
	}
	return items
}
