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

	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/param"
)

type ShowOverviewData struct {
	Name             string
	ID               int
	Version          string
	UnderMaintenance bool
	AgentVersion     string
	Connected        bool
}

type ShowRowData struct {
	meta.AgentInstance
	param.ServerConfig
	UnderMaintenance bool
	OBState          string
}

const (
	COL_ZONE            = "ZONE"
	COL_IP              = "IP"
	COL_OPT_PORT        = "OPT PORT"
	COL_IDENTITY        = "IDENTITY"
	COL_SQL_PORT        = "SQL PORT"
	COL_SVR_PORT        = "SVR PORT"
	COL_STATUS          = "STATUS"
	COL_ROOT_SVR        = "ROOT SERVER"
	COL_MAINTENANCE     = "MAINTENANCE"
	COL_OB_STATE        = "OB STATE"
	COL_OBSEVER_VERSION = "OBSERVER VERSION"
	COL_OBSHELL_VERSION = "OBSHELL VERSION"
)

var (
	normalHeader         = []string{COL_ZONE, COL_IP, COL_OPT_PORT, COL_IDENTITY, COL_SQL_PORT, COL_SVR_PORT, COL_STATUS, COL_ROOT_SVR}
	detailedHeader       = []string{COL_ZONE, COL_IP, COL_OPT_PORT, COL_IDENTITY, COL_MAINTENANCE, COL_SQL_PORT, COL_SVR_PORT, COL_STATUS, COL_ROOT_SVR, COL_OB_STATE}
	obVersionHeader      = []string{COL_ZONE, COL_IP, COL_OPT_PORT, COL_IDENTITY, COL_SQL_PORT, COL_SVR_PORT, COL_OBSEVER_VERSION}
	obshellVersionheader = []string{COL_IP, COL_OPT_PORT, COL_IDENTITY, COL_OBSHELL_VERSION}
)

func PrintShowTable(overviewData ShowOverviewData, agent2row map[meta.AgentInfo]ShowRowData, detail bool) {
	if detail {
		doPrintShowTableInDetail(overviewData, agent2row)
		return
	}
	doPrintShowTable(overviewData, agent2row)
}

func doPrintShowTable(overviewData ShowOverviewData, agent2row map[meta.AgentInfo]ShowRowData) {
	rows := make([][]string, 0, len(agent2row))
	for _, data := range agent2row {
		row := []string{
			data.Zone,
			data.Ip,
			fmt.Sprint(data.Port),
			string(data.Identity),
			fmt.Sprint(data.SqlPort),
			fmt.Sprint(data.SvrPort),
			data.Status,
			data.WithRootSvr,
		}
		checkEmptyStrInRow(row)
		rows = append(rows, row)
	}
	sortRows(rows, 0, 1, 2)
	stdio.PrintTableWithTitle(newTitle(&overviewData, false), normalHeader, rows)
}

func doPrintShowTableInDetail(overviewData ShowOverviewData, agent2row map[meta.AgentInfo]ShowRowData) {
	rows := make([][]string, 0, len(agent2row))
	for _, data := range agent2row {
		row := []string{
			data.Zone,
			data.Ip,
			fmt.Sprint(data.Port),
			string(data.Identity),
			strings.ToUpper(strconv.FormatBool(data.UnderMaintenance)),
			fmt.Sprint(data.SqlPort),
			fmt.Sprint(data.SvrPort),
			data.Status,
			data.WithRootSvr,
			data.OBState,
		}
		checkEmptyStrInRow(row)
		rows = append(rows, row)
	}
	sortRows(rows, 0, 1, 2)
	stdio.PrintTableWithTitle(newTitle(&overviewData, true), detailedHeader, rows)
}

func newTitle(overviewData *ShowOverviewData, showDetails bool) string {
	maxLen := 0
	lines := make([]string, 0, 4)
	if overviewData.Connected {
		lines = append(lines, fmt.Sprintf("CLUSTER INFO: %s (ID:%d)", overviewData.Name, overviewData.ID))
		lines = append(lines, fmt.Sprintf("MAINTENANCE: %s", strings.ToUpper(strconv.FormatBool(overviewData.UnderMaintenance))))
		lines = append(lines, fmt.Sprintf("CLUSTER VERSION: %s", overviewData.Version))
	} else {
		lines = append(lines, "CLUSTER INFO: N/A")
		lines = append(lines, "MAINTENANCE: N/A")
		lines = append(lines, "CLUSTER VERSION: N/A")
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

func PrintWarnForBuildVersion(agent2row map[meta.AgentInfo]ShowRowData) {
	rows := make([][]string, 0, len(agent2row))
	for _, data := range agent2row {
		row := []string{
			data.Zone,
			data.Ip,
			fmt.Sprint(data.Port),
			string(data.Identity),
			fmt.Sprint(data.SqlPort),
			fmt.Sprint(data.SvrPort),
			data.BuildVersion,
		}
		checkEmptyStrInRow(row)
		rows = append(rows, row)
	}
	stdio.Print("")
	stdio.Warn("Inconsistent versions detected among observers.")
	sortRows(rows, 0, 1, 2)
	stdio.PrintTable(obVersionHeader, rows)
}

func PrintWarnForAgentVersion(agents []meta.AgentInstance) {
	rows := make([][]string, 0, len(agents))
	for _, data := range agents {
		row := []string{
			data.Ip,
			fmt.Sprint(data.Port),
			string(data.Identity),
			data.Version,
		}
		checkEmptyStrInRow(row)
		rows = append(rows, row)
	}

	stdio.Print("")
	stdio.Warn("Inconsistent versions detected among obshell agents.")
	sortRows(rows, 0, 1)
	stdio.PrintTable(obshellVersionheader, rows)
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
