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

package printer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	obmodel "github.com/oceanbase/obshell/seekdb/model/observer"
)

type ShowOverviewData struct {
	Name             string
	ID               string
	Version          string
	UnderMaintenance bool
	AgentVersion     string
	Connected        bool
}

type ShowRowData struct {
	AgentPort int
	obmodel.ObserverInfo
	UnderMaintenance bool
	OBState          string
}

const (
	COL_OPT_PORT       = "OPT PORT"
	COL_SQL_PORT       = "PORT"
	COL_MAINTENANCE    = "MAINTENANCE"
	COL_OB_STATE       = "OB STATE"
	COL_DATA_DIR       = "DATA DIR"
	COL_REDO_DIR       = "REDO DIR"
	COL_LOG_DIR        = "LOG DIR"
	COL_BASE_DIR       = "BASE DIR"
	COL_BIN_PATH       = "BIN PATH"
	COL_CPU_COUNT      = "CPU COUNT"
	COL_MEMORY_SIZE    = "MEMORY SIZE"
	COL_LOG_DISK_SIZE  = "LOG DISK SIZE"
	COL_DATA_DISK_SIZE = "DATA DISK SIZE"
)

var (
	normalHeader   = []string{COL_OPT_PORT, COL_SQL_PORT, COL_OB_STATE}
	detailedHeader = []string{COL_OPT_PORT, COL_MAINTENANCE, COL_SQL_PORT, COL_OB_STATE, COL_DATA_DIR, COL_REDO_DIR, COL_LOG_DIR, COL_BASE_DIR, COL_BIN_PATH, COL_CPU_COUNT, COL_MEMORY_SIZE, COL_LOG_DISK_SIZE, COL_DATA_DISK_SIZE}
)

func PrintShowTable(overviewData ShowOverviewData, agentRowData ShowRowData, detail bool) {
	if detail {
		doPrintShowTableInDetail(overviewData, agentRowData)
		return
	}
	doPrintShowTable(overviewData, agentRowData)
}

func doPrintShowTable(overviewData ShowOverviewData, agentRowData ShowRowData) {
	rows := make([][]string, 0, 1)
	row := []string{
		fmt.Sprint(agentRowData.AgentPort),
		fmt.Sprint(agentRowData.Port),
		agentRowData.OBState,
	}
	checkEmptyStrInRow(row)
	rows = append(rows, row)
	stdio.PrintTableWithTitle(newTitle(&overviewData, false), normalHeader, rows)
}

func doPrintShowTableInDetail(overviewData ShowOverviewData, agentRowData ShowRowData) {
	rows := make([][]string, 0, 1)
	rows = append(rows, []string{
		fmt.Sprint(agentRowData.AgentPort),
		strings.ToUpper(strconv.FormatBool(agentRowData.UnderMaintenance)),
		fmt.Sprint(agentRowData.Port),
		agentRowData.OBState,
		agentRowData.DataDir,
		agentRowData.RedoDir,
		agentRowData.LogDir,
		agentRowData.BaseDir,
		agentRowData.BinPath,
		fmt.Sprint(agentRowData.CpuCount),
		agentRowData.MemorySize,
		agentRowData.LogDiskSize,
		agentRowData.DataDiskSize,
	})

	checkEmptyStrInRow(rows[0])
	sortRows(rows, 0, 1, 2)
	stdio.PrintTableWithTitle(newTitle(&overviewData, true), detailedHeader, rows)
}

func newTitle(overviewData *ShowOverviewData, showDetails bool) string {
	maxLen := 0
	lines := make([]string, 0, 4)
	if overviewData.Connected {
		lines = append(lines, fmt.Sprintf("SEEKDB INFO: %s", overviewData.Name))
		lines = append(lines, fmt.Sprintf("MAINTENANCE: %s", strings.ToUpper(strconv.FormatBool(overviewData.UnderMaintenance))))
		lines = append(lines, fmt.Sprintf("OCEANBASE SEEKDB VERSION: %s", overviewData.Version))
	} else {
		lines = append(lines, "SEEKDB INFO: N/A")
		lines = append(lines, "MAINTENANCE: N/A")
		lines = append(lines, "OCEANBASE SEEKDB VERSION: N/A")
	}
	lines = append(lines, fmt.Sprintf("OBSHELL VERSION: %s", overviewData.AgentVersion))

	if !showDetails {
		return fmt.Sprintf("%s\n%s", lines[0], lines[2])
	}

	if len(lines[0]) > len(lines[1]) {
		maxLen = len(lines[0])
	} else {
		maxLen = len(lines[1])
	}
	maxLen += 3
	lines[0] = fmt.Sprintf("%-*s", maxLen, lines[0])
	lines[1] = fmt.Sprintf("%-*s", maxLen, lines[1])
	return fmt.Sprintf("%s%s\n%s%s", lines[0], lines[2], lines[1], lines[3])
}

func checkEmptyStrInRow(row []string) {
	for i := range row {
		if row[i] == "" || row[i] == "0" {
			row[i] = "N/A"
		}
	}
}

func sortRows(rows [][]string, fieldsToCompare ...int) {
	sort.Sort(showRowsToSort{rows, fieldsToCompare})
}

type showRowsToSort struct {
	rows            [][]string
	fieldsToCompare []int
}

func (s showRowsToSort) Less(i, j int) bool {
	if len(s.fieldsToCompare) > 0 {
		for _, f := range s.fieldsToCompare {
			if s.rows[i][f] != s.rows[j][f] {
				return s.rows[i][f] < s.rows[j][f]
			}
		}
	}
	return false
}

func (s showRowsToSort) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

func (s showRowsToSort) Len() int {
	return len(s.rows)
}
