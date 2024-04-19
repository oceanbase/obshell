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

	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/engine/task"
)

func PrintDagsTable(dags []*task.DagDetailDTO) {
	header := []string{"ID", "Name", "State", "Stage", "Max Stage", "Start Time", "End Time"}
	rows := make([][]string, 0, len(dags))
	for _, dag := range dags {
		rows = append(rows, []string{
			dag.GenericID,
			dag.Name,
			dag.State,
			fmt.Sprint(dag.Stage),
			fmt.Sprint(dag.MaxStage),
			dag.StartTime.String(),
			dag.EndTime.String(),
		})
	}
	stdio.PrintTable(header, rows)
}
