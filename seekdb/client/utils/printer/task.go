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
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
)

func PrintDagStruct(dag *task.DagDetailDTO, detail bool) {
	bytes, err := yaml.Marshal(convertDag2MapSlice(dag, detail))
	if err != nil {
		stdio.Verbosef("print dag struct failed, err: %s", err.Error())
		log.Error(err)
		return
	}
	stdio.Print(postprocessDagStructText(string(bytes)))
}

func convertDag2MapSlice(dag *task.DagDetailDTO, detail bool) (data yaml.MapSlice) {
	if dag == nil {
		return
	}
	data = append(data,
		yaml.MapItem{Key: "id", Value: dag.GenericID},
		yaml.MapItem{Key: "dag_id", Value: dag.DagID},
		yaml.MapItem{Key: "name", Value: dag.Name},
		yaml.MapItem{Key: "stage", Value: dag.Stage},
		yaml.MapItem{Key: "max_stage", Value: dag.MaxStage},
		yaml.MapItem{Key: "state", Value: dag.State},
		yaml.MapItem{Key: "operator", Value: dag.Operator},
		yaml.MapItem{Key: "start_time", Value: dag.StartTime},
		yaml.MapItem{Key: "end_time", Value: dag.EndTime},
	)
	if detail {
		data = append(data, yaml.MapItem{
			Key: "nodes", Value: convertNodes2MapSlice(dag.Nodes),
		})
	}
	return
}

func convertNodes2MapSlice(nodes []*task.NodeDetailDTO) (data yaml.MapSlice) {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		data = append(data,
			yaml.MapItem{Key: "id", Value: node.GenericID},
			yaml.MapItem{Key: "node_id", Value: node.NodeID},
			yaml.MapItem{Key: "name", Value: node.Name},
			yaml.MapItem{Key: "state", Value: node.State},
			yaml.MapItem{Key: "operator", Value: node.Operator},
			yaml.MapItem{Key: "start_time", Value: node.StartTime},
			yaml.MapItem{Key: "end_time", Value: node.EndTime},
			yaml.MapItem{Key: "subtasks", Value: convertSubTasks2MapSlice(node.SubTasks)},
		)
	}
	return
}

func convertSubTasks2MapSlice(subTasks []*task.TaskDetailDTO) (data yaml.MapSlice) {
	for _, subTask := range subTasks {
		if subTask == nil {
			continue
		}
		data = append(data,
			yaml.MapItem{Key: "id", Value: subTask.GenericID},
			yaml.MapItem{Key: "task_id", Value: subTask.TaskID},
			yaml.MapItem{Key: "name", Value: subTask.Name},
			yaml.MapItem{Key: "state", Value: subTask.State},
			yaml.MapItem{Key: "operator", Value: subTask.Operator},
			yaml.MapItem{Key: "start_time", Value: subTask.StartTime},
			yaml.MapItem{Key: "end_time", Value: subTask.EndTime},
			yaml.MapItem{Key: "execute_times", Value: subTask.ExecuteTimes},
			yaml.MapItem{Key: "execute_agent", Value: yaml.MapSlice{
				yaml.MapItem{Key: "ip", Value: subTask.ExecuteAgent.Ip},
				yaml.MapItem{Key: "port", Value: subTask.ExecuteAgent.Port},
			}},
			yaml.MapItem{Key: "task_logs", Value: subTask.TaskLogs},
		)
	}
	return
}

func postprocessDagStructText(text string) string {
	return strings.ReplaceAll(text, "\"", "")
}
