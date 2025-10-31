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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/api/common"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	taskservice "github.com/oceanbase/obshell/seekdb/agent/service/task"
)

// get dag detail by id
//
// @ID getDagDetail
// @Summary get dag detail by id
// @Description get dag detail by id
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path string true "id"
// @Param showDetails query param.TaskQueryParams true "show details"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/dag/{id} [get]
func GetDagDetail(c *gin.Context) {
	var dagDTOParam task.DagDetailDTO
	var dagDetailDTO *task.DagDetailDTO
	var service taskservice.TaskServiceInterface

	if err := c.BindUri(&dagDTOParam); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dagID, agent, err := task.ConvertGenericID(dagDTOParam.GenericID)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound))
		return
	}

	param := getTaskQueryParams(c)
	if agent == nil {
		service = clusterTaskService
	} else {
		service = localTaskService
	}

	dag, err := service.GetDagInstance(dagID)
	if err != nil {
		common.SendResponse(c, nil, errors.WrapRetain(errors.ErrTaskNotFound, err))
		return
	}

	dagDetailDTO, err = convertDagDetailDTO(dag, *param.ShowDetails)
	common.SendResponse(c, dagDetailDTO, err)
}

// get all dags in cluster
//
// @ID getAllClusterDags
// @Summary get all dags in cluster
// @Description get all dags in cluster
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header	header string true "Authorization"
// @Param show_details query param.TaskQueryParams true "show details"
// @Success 200 object	http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure 400 object	http.OcsAgentResponse
// @Failure 404 object	http.OcsAgentResponse
// @Failure 500 object	http.OcsAgentResponse
// @Router /api/v1/task/dags/ob [get]
func GetAllClusterDags(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
		return
	}

	param := getTaskQueryParams(c)
	dags, err := clusterTaskService.GetAllDagInstances()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dagDetailDTOs := make([]*task.DagDetailDTO, 0)
	for _, dag := range dags {
		dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
		if err != nil {
			log.WithContext(common.NewContextWithTraceId(c)).Errorf("convert dag detail dto failed: %v", err)
			continue
		}
		dagDetailDTOs = append(dagDetailDTOs, dagDetailDTO)
	}
	common.SendResponse(c, dagDetailDTOs, nil)
}

// get all dags in agent
//
// @ID getAllAgentDags
// @Summary get all dags in the agent
// @Description	get all dags in the agent
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param show_details query param.TaskQueryParams true "show details"
// @Success 200 object http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router	/api/v1/task/dags/agent [get]
func GetAllAgentDags(c *gin.Context) {
	param := getTaskQueryParams(c)
	dags, err := localTaskService.GetAllDagInstances()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dagDetailDTOs := make([]*task.DagDetailDTO, 0)
	for _, dag := range dags {
		dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
		if err != nil {
			log.WithContext(common.NewContextWithTraceId(c)).Errorf("convert dag detail dto failed: %v", err)
			continue
		}
		dagDetailDTOs = append(dagDetailDTOs, dagDetailDTO)
	}
	common.SendResponse(c, dagDetailDTOs, nil)
}

func convertDagDetailDTO(dag *task.Dag, fillDeatil bool) (dagDetailDTO *task.DagDetailDTO, err error) {
	dagDetailDTO = task.NewDagDetailDTO(dag)

	if fillDeatil {
		var nodes []*task.Node
		var service taskservice.TaskServiceInterface
		var nodeDetailDTO *task.NodeDetailDTO

		if dag.IsLocalTask() {
			service = localTaskService
		} else {
			service = clusterTaskService
		}

		nodes, err = service.GetNodes(dag)
		if err != nil {
			return
		}

		n := len(nodes)
		for i := 0; i < n; i++ {
			if _, err = service.GetSubTasks(nodes[i]); err != nil {
				return
			}

			nodeDetailDTO, err = getNodeDetail(service, nodes[i], dag.GetDagType())
			if err != nil {
				return
			}
			dagDetailDTO.Nodes = append(dagDetailDTO.Nodes, nodeDetailDTO)
		}
		dagDetailDTO.SetVisible(true)
	}
	return
}

// dag handler
//
// @ID dagHandler
// @Summary operate dag
// @Description operate dag
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path string true "dag id"
// @Param body body task.DagOperator true "dag operator, supported values are (rollback/retry/cancel/pass)"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/dag/{id} [post]
func DagHandler(c *gin.Context) {
	var dagOperator task.DagOperator
	var service taskservice.TaskServiceInterface

	if err := c.BindUri(&dagOperator.DagDetailDTO); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	dagID, agent, err := task.ConvertGenericID(dagOperator.GenericID)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if err := c.BindJSON(&dagOperator); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound))
		return
	}

	if agent == nil {
		service = clusterTaskService
	} else {
		service = localTaskService
	}

	dag, err := service.GetDagInstance(dagID)
	if err != nil {
		common.SendResponse(c, nil, errors.WrapRetain(errors.ErrTaskNotFound, err))
		return
	}

	switch strings.ToUpper(dagOperator.Operator) {
	case task.ROLLBACK_STR:
		err = service.SetDagRollback(dag)
	case task.RETRY_STR:
		err = service.SetDagRetryAndReady(dag)
	case task.CANCEL_STR:
		err = service.CancelDag(dag)
	case task.PASS_STR:
		err = service.PassDag(dag)
	default:
		err = errors.Occur(errors.ErrTaskDagOperatorNotSupport, dagOperator.Operator)
	}

	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, nil)
}

// @ID GetObLastMaintenanceDag
// @Summary get ob last maintenance dag
// @Description get ob last maintenance dag
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/dag/maintain/ob [get]
func GetObLastMaintenanceDag(c *gin.Context) {
	param := getTaskQueryParams(c)
	dag, err := clusterTaskService.FindLastMaintenanceDag()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if dag == nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFoundWithReason, "Cluster is not under maintenance"))
		return
	}
	dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
	// No need to wrap with errors.Occur as the err != nil will be auto-wrapped into errors.Occur(errors.ErrUnexpected, err).
	common.SendResponse(c, dagDetailDTO, err)
}

// @ID GetAgentLastMaintenanceDag
// @Summary get agent last maintenance dag
// @Description get agent last maintenance dag
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/dag/maintain/agent [get]
func GetAgentLastMaintenanceDag(c *gin.Context) {
	param := getTaskQueryParams(c)
	dag, err := localTaskService.FindLastMaintenanceDag()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if dag == nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFoundWithReason, "Agent is not under maintenance"))
		return
	}
	// construct dagDetailDTO
	dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
	// No need to wrap with errors.Occur as the err != nil will be auto-wrapped into errors.Occur(errors.ErrUnexpected, err).
	common.SendResponse(c, dagDetailDTO, err)
}

// @ID GetUnfinishedDags
// @Summary get unfinished dags
// @Description get unfinished dags
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param showDetails query param.TaskQueryParams true "show details"
// @Success 200 object http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Router /api/v1/task/dag/unfinish [get]
func GetUnfinishedDags(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	switch meta.OCS_AGENT.GetIdentity() {
	case meta.CLUSTER_AGENT:
		param := getTaskQueryParams(c)
		agentsDags, err := getAgentUnfinishDags(*param.ShowDetails)
		if err != nil {
			log.WithContext(ctx).Errorf("get agent unfinished dags failed: %v", err)
		}
		clusterDags, err := getClusterUnfinishDags(*param.ShowDetails)
		if err != nil {
			log.WithContext(ctx).Errorf("get cluster unfinished dags failed: %v", err)
		}
		dags := append(agentsDags, clusterDags...)
		common.SendResponse(c, dags, nil)
	case meta.SINGLE, meta.MASTER, meta.TAKE_OVER_MASTER:
		GetAgentUnfinishDags(c)
	default:
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyUnknown, meta.OCS_AGENT.GetIdentity()))
	}
}

// @ID GetClusterUnfinishDags
// @Summary get cluster unfinished dags
// @Description get cluster unfinished dags
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param showDetails query param.TaskQueryParams true "show details"
// @Success 200 object http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/dag/ob/unfinish [get]
func GetClusterUnfinishDags(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
		return
	}

	param := getTaskQueryParams(c)
	dagDetailDTOs, err := getClusterUnfinishDags(*param.ShowDetails)
	common.SendResponse(c, dagDetailDTOs, err)
}

// @ID GetAgentUnfinishDags
// @Summary get agent unfinished dags
// @Description get agent unfinished dags
// @Tags task
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param showDetails query param.TaskQueryParams true "show details"
// @Success 200 object http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/task/dag/agent/unfinish [get]
func GetAgentUnfinishDags(c *gin.Context) {
	param := getTaskQueryParams(c)
	dagDetailDTOs, err := getAgentUnfinishDags(*param.ShowDetails)
	common.SendResponse(c, dagDetailDTOs, err)
}

func getClusterUnfinishDags(fillDetails bool) (dagDetailDTOs []*task.DagDetailDTO, err error) {
	dags, err := clusterTaskService.GetAllUnfinishedDagInstance()
	if err != nil {
		return
	}
	dagDetailDTOs, err = convertDagDetailDTOs(dags, fillDetails)
	return
}

func getAgentUnfinishDags(fillDetails bool) (dagDetailDTOs []*task.DagDetailDTO, err error) {
	dags, err := localTaskService.GetAllUnfinishedDagInstance()
	if err != nil {
		return
	}
	dagDetailDTOs, err = convertDagDetailDTOs(dags, fillDetails)
	return
}

func convertDagDetailDTOs(dags []*task.Dag, fillDetails bool) (dagDetailDTOs []*task.DagDetailDTO, err error) {
	dagDetailDTOs = make([]*task.DagDetailDTO, 0, len(dags))
	for _, dag := range dags {
		dagDetailDTO, err := convertDagDetailDTO(dag, fillDetails)
		if err != nil {
			return nil, err
		}
		dagDetailDTOs = append(dagDetailDTOs, dagDetailDTO)
	}
	return
}
