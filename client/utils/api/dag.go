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

package api

import (
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
)

func CallApiAndPrintStage(uri string, param interface{}) (dag *task.DagDetailDTO, err error) {
	dag, err = CallApi(uri, param)
	if err != nil {
		return
	}
	dagHandler := NewDagHandler(dag)
	if err = dagHandler.PrintDagStage(); err != nil {
		return
	}
	return dag, nil
}

func callApiHelper(sendRequest func(uri string, param interface{}, res interface{}) error, uri string, param interface{}) (*task.DagDetailDTO, error) {
	res := &task.DagDetailDTO{}
	stdio.Verbosef("Calling API %s", uri)
	stdio.Verbosef("Param is %+v", param)

	err := sendRequest(uri, param, &res)
	if err != nil {
		stdio.Verbosef("Call API %s failed, error is %s", uri, err)
		return nil, err
	}

	stdio.Printf("Task '%s' has been created successfully.", res.Name)
	stdio.Printf("You can view the task details by '%s/bin/obshell task show -i %s -d'.", global.HomePath, res.GenericID)
	return res, nil
}

func CallApi(uri string, param interface{}) (*task.DagDetailDTO, error) {
	sendRequest := func(uri string, param interface{}, res interface{}) error {
		return http.SendPostRequestViaUnixSocket(path.ObshellSocketPath(), uri, param, res)
	}
	return callApiHelper(sendRequest, uri, param)
}

func CallApiViaTCP(agentInfo meta.AgentInfoInterface, uri string, param interface{}) (*task.DagDetailDTO, error) {
	sendRequest := func(uri string, param interface{}, res interface{}) error {
		return secure.SendPostRequest(agentInfo, uri, param, res)
	}
	return callApiHelper(sendRequest, uri, param)
}

func GetFailedDagLastLog(currentDag *task.DagDetailDTO) (res string) {
	nodes := currentDag.Nodes

	var subTask *task.TaskDetailDTO
	var currentNode *task.NodeDetailDTO
	var logs []string
	for i := 0; i < len(nodes); i++ {
		currentNode = nodes[i]
		if !currentNode.IsFailed() {
			continue
		}

		if currentNode.Operator == task.OPERATOR_MAP[task.CANCEL] {
			return fmt.Sprintf("Sorry, Task '%s' was cancelled", currentDag.Name)
		}

		for j := 0; j < len(currentNode.SubTasks); j++ {
			subTask = currentNode.SubTasks[j]
			if subTask.IsFailed() {
				lastLog := subTask.TaskLogs[len(subTask.TaskLogs)-1]
				logs = append(logs, fmt.Sprintf("%s %s", subTask.ExecuteAgent.String(), lastLog))
			}
		}
		return strings.Join(logs, " ")
	}
	return "No failed task log found, please check the task details"
}

func GetDagDetail(id string) (res *task.DagDetailDTO, err error) {
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+id, nil, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "Get %s detail failed", id)
	}
	return res, nil
}

func GetDagDetailForUpgrade(id string) (res *task.DagDetailDTO, err error) {
	stdio.Verbose("Get dag detail by tmp socket")
	err = http.SendGetRequestViaUnixSocket(path.ObshellTmpSocketPath(), constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+id, nil, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "Get %s detail by tmp socket failed", id)
	}
	return res, nil
}

func GetDagDetailViaTCP(agentInfo meta.AgentInfoInterface, id string) (res *task.DagDetailDTO, err error) {
	err = secure.SendGetRequest(agentInfo, constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+id, nil, &res)
	return
}

func PassDag(id string) (err error) {
	return sendDagOperatorRequest(task.PASS, id)
}

func CancelDag(id string) (err error) {
	return sendDagOperatorRequest(task.CANCEL, id)
}

func sendDagOperatorRequest(operator int, id string) error {
	dagOperator := task.DagOperator{Operator: task.OPERATOR_MAP[operator]}
	return http.SendPostRequestViaUnixSocket(path.ObshellSocketPath(), constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+id, dagOperator, nil)
}

func IsEmecTypeDag(dag *task.DagDetailDTO) (id string, res bool) {
	if dag.AdditionalData != nil {
		data := *(dag.AdditionalData)
		mainDagid, ok := data[ob.ADDL_KEY_MAIN_DAG_ID].(string)
		stdio.Verbosef(" %s is emec type dag %v, main dag %s", dag.GenericID, ok, mainDagid)
		if ok {
			return mainDagid, true
		}
	}
	return "", false
}

func GetMainDagID(dag *task.DagDetailDTO) (id string, res bool) {
	if dag.AdditionalData != nil {
		data := *(dag.AdditionalData)
		mainDagid, ok := data[ob.ADDL_KEY_MAIN_DAG_ID].(string)
		if ok {
			return mainDagid, true
		}
	}
	return "", false
}

func GetSubDagIDs(dag *task.DagDetailDTO) (ids []string, res bool) {
	if dag.AdditionalData != nil {
		data := *(dag.AdditionalData)
		subDags := data[ob.ADDL_KEY_SUB_DAGS]
		subDagIDs, ok := subDags.(map[string]interface{})
		if ok {
			for _, v := range subDagIDs {
				id := fmt.Sprintf("%v", v)
				ids = append(ids, id)
			}
			return ids, true
		}
	}
	return nil, false
}

func GetObLastMaintenanceDag() (dag *task.DagDetailDTO, err error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_MAINTAIN + constant.URI_OB_GROUP
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &dag)
	if err != nil {
		return nil, err
	}
	return dag, nil
}

func GetAllUnfinishedDags() (dags []*task.DagDetailDTO, err error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_UNFINISH
	return getDags(uri)
}

func GetOBUnfinishedDags() (dags []*task.DagDetailDTO, err error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_OB_GROUP + constant.URI_UNFINISH
	return getDags(uri)
}

func GetAgentUnfinishedDags() (dags []*task.DagDetailDTO, err error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_AGENT_GROUP + constant.URI_UNFINISH
	return getDags(uri)
}

func getDags(uri string) (dags []*task.DagDetailDTO, err error) {
	stdio.Verbosef("Calling API %s", uri)
	err = http.SendGetRequestViaUnixSocket(path.ObshellSocketPath(), uri, nil, &dags)
	if err != nil {
		return nil, errors.Wrap(err, "Get dags failed")
	}
	return dags, nil
}
