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

package stdio

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

func newTableRow(items []string) (tableRow table.Row) {
	for _, item := range items {
		tableRow = append(tableRow, item)
	}
	return
}

func (io *IO) newTable(header []string, data [][]string) table.Writer {
	table := table.NewWriter()
	table.AppendHeader(newTableRow(header))
	for _, row := range data {
		table.AppendRow(newTableRow(row))
	}
	table.Style().Options.SeparateRows = true
	table.SetOutputMirror(io.currOutStream)
	return table
}

func (io *IO) PrintTable(header []string, data [][]string) {
	table := io.newTable(header, data)
	table.Render()
}

func (io *IO) PrintTableWithTitle(title string, header []string, data [][]string) {
	table := io.newTable(header, data)
	table.SetTitle(title)
	table.Render()
}
