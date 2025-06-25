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

package task

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
)

// get node detail by node_id
//
//	@ID				getNodeDetail
//	@Summary		get node detail by node_id
//	@Description	get node detail by node_id
//	@Tags			task
//	@Accept			application/json
//	@Produce		application/json
//	@Param			X-OCS-Header	header	string					true	"Authorization"
//	@Param			id				path	string					true	"id"
//	@Param			showDetails		query	param.TaskQueryParams	true	"show details"
//
//	@Success		200				object	http.OcsAgentResponse{data=task.NodeDetailDTO}
//	@Failure		400				object	http.OcsAgentResponse
//	@Failure		404				object	http.OcsAgentResponse
//	@Failure		500				object	http.OcsAgentResponse
//	@Router			/api/v1/task/node/{id} [get]
func GetNodeDetail(c *gin.Context) {
	var nodeDTOParam task.NodeDetailDTO
	var nodeDetailDTO *task.NodeDetailDTO
	var service taskservice.TaskServiceInterface

	if err := c.BindUri(&nodeDTOParam); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	nodeID, agent, err := task.ConvertGenericID(nodeDTOParam.GenericID)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		if task.IsObproxyTask(nodeDTOParam.GenericID) {
			common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFoundWithReason, "obproxy task not found"))
		}
		if meta.OCS_AGENT.IsFollowerAgent() {
			// forward request to master
			master := agentService.GetMasterAgentInfo()
			if master == nil {
				common.SendResponse(c, nil, errors.Occur(errors.ErrAgentNoMaster))
				return
			}
			common.ForwardRequest(c, master, nil)
		} else {
			common.ForwardRequest(c, agent, nil)
		}
		return
	}

	param := getTaskQueryParams(c)
	if agent == nil {
		service = clusterTaskService
	} else {
		service = localTaskService
	}

	node, err := service.GetNodeByNodeId(nodeID)
	if err != nil {
		common.SendResponse(c, nil, errors.WrapRetain(errors.ErrTaskNotFound, err))
		return
	}

	if *param.ShowDetails {
		_, err = service.GetSubTasks(node)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}

		dag, err := service.GetDagInstance(int64(node.GetDagId()))
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		if task.ConvertToGenericID(dag, dag.GetDagType())[0] != nodeDTOParam.GenericID[0] {
			common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFoundWithReason, "node type not match"))
			return
		}
		nodeDetailDTO, err = getNodeDetail(service, node, dag.GetDagType())
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
	}
	nodeDetailDTO.SetVisible(true)
	common.SendResponse(c, nodeDetailDTO, nil)
}

func getNodeDetail(service taskservice.TaskServiceInterface, node *task.Node, dagType string) (nodeDetailDTO *task.NodeDetailDTO, err error) {
	nodeDetailDTO = task.NewNodeDetailDTO(node, dagType)
	subTasks := node.GetSubTasks()
	n := len(subTasks)
	for i := 0; i < n; i++ {
		taskDetailDTO, err := getSubTaskDetail(service, subTasks[i], dagType)
		if err != nil {
			return nil, err
		}
		nodeDetailDTO.SubTasks = append(nodeDetailDTO.SubTasks, taskDetailDTO)
	}
	return
}

// node handler
//
// @ID nodeHandler
// @Summary operate node
// @Description operate node
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path string true "node id"
// @Param body body task.NodeOperator true "node operator, supported values is (pass) only"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/node/{id} [post]
func NodeHandler(c *gin.Context) {
	var nodeDTOParam task.NodeDetailDTO
	var service taskservice.TaskServiceInterface
	var nodeOperator task.DagOperator

	if err := c.BindUri(&nodeDTOParam); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	nodeID, agent, err := task.ConvertGenericID(nodeDTOParam.GenericID)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if err := c.BindJSON(&nodeOperator); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if strings.ToUpper(nodeOperator.Operator) != task.PASS_STR {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNodeOperatorNotSupport, nodeOperator.Operator))
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		if task.IsObproxyTask(nodeDTOParam.GenericID) {
			common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFoundWithReason, "obproxy task not found"))
		}
		if meta.OCS_AGENT.IsFollowerAgent() {
			// forward request to master
			master := agentService.GetMasterAgentInfo()
			if master == nil {
				common.SendResponse(c, nil, errors.Occur(errors.ErrAgentNoMaster))
				return
			}
			common.ForwardRequest(c, master, nil)
		} else {
			common.ForwardRequest(c, agent, nil)
		}
		return
	}

	if agent == nil {
		service = clusterTaskService
	} else {
		service = localTaskService
	}

	node, err := service.GetNodeByNodeId(nodeID)
	if err != nil {
		common.SendResponse(c, nil, errors.WrapRetain(errors.ErrTaskNotFound, err))
		return
	}

	dag, err := service.GetDagInstance(int64(node.GetDagId()))
	if err != nil {
		common.SendResponse(c, nil, errors.WrapRetain(errors.ErrTaskNotFound, err))
		return
	}

	common.SendResponse(c, nil, service.PassNode(node, dag))
}
