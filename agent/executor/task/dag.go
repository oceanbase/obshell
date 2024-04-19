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
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
	taskservice "github.com/oceanbase/obshell/agent/service/task"
	"github.com/oceanbase/obshell/param"
)

// get dag detail by id
//
//	@ID				getDagDetail
//	@Summary		get dag detail by id
//	@Description	get dag detail by id
//	@Tags			task
//	@Accept			application/json
//	@Produce		application/json
//	@Param			X-OCS-Header	header	string					true	"Authorization"
//	@Param			id				path	string					true	"id"
//	@Param			showDetails		query	param.TaskQueryParams	true	"show details"
//	@Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
//	@Failure		400				object	http.OcsAgentResponse
//	@Failure		404				object	http.OcsAgentResponse
//	@Failure		500				object	http.OcsAgentResponse
//	@Router			/api/v1/task/dag/{id} [get]
func GetDagDetail(c *gin.Context) {
	var dagDTOParam task.DagDetailDTO
	var dagDetailDTO *task.DagDetailDTO
	var service taskservice.TaskServiceInterface

	if err := c.BindUri(&dagDTOParam); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	dagID, agent, err := task.ConvertGenericID(dagDTOParam.GenericID)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		// forward request to agent
		common.ForwardRequest(c, agent)
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
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound, err))
		return
	}

	dagDetailDTO, err = convertDagDetailDTO(dag, *param.ShowDetails)
	common.SendResponse(c, dagDetailDTO, err)
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

			nodeDetailDTO, err = getNodeDetail(service, nodes[i])
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
//	@ID				dagHandler
//	@Summary		operate dag
//	@Description	operate dag
//	@Tags			task
//	@Accept			application/json
//	@Produce		application/json
//	@Param			X-OCS-Header	header	string	true	"Authorization"
//	@Param			id				path	string	true	"dag id"
//	@Param			operator		body	string	true	"operator(rollback/retry/cancel/pass)"	example({"operator": "rollback"})
//	@Success		200				object	http.OcsAgentResponse
//	@Failure		400				object	http.OcsAgentResponse
//	@Failure		404				object	http.OcsAgentResponse
//	@Failure		500				object	http.OcsAgentResponse
//	@Router			/api/v1/task/dag/{id} [post]
func DagHandler(c *gin.Context) {
	var dagOperator task.DagOperator
	var service taskservice.TaskServiceInterface

	if err := c.BindUri(&dagOperator.DagDetailDTO); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	dagID, agent, err := task.ConvertGenericID(dagOperator.GenericID)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	if err := c.BindJSON(&dagOperator); err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
		return
	}

	if agent != nil && !meta.OCS_AGENT.Equal(agent) {
		common.ForwardRequest(c, agent, dagOperator)
		return
	}

	if agent == nil {
		service = clusterTaskService
	} else {
		service = localTaskService
	}

	dag, err := service.GetDagInstance(dagID)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound, err))
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
		err = fmt.Errorf("invalid operator: %s", dagOperator.Operator)
	}

	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrKnown, err))
		return
	}
	common.SendResponse(c, nil, nil)
}

// @ID				GetObLastMaintenanceDag
// @Summary		get ob last maintenance dag
// @Description	get ob last maintenance dag
// @Tags			task
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		404				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/task/dag/maintain/ob [get]
func GetObLastMaintenanceDag(c *gin.Context) {
	param := getTaskQueryParams(c)
	dag, err := clusterTaskService.FindLastMaintenanceDag()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if dag == nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound, "Cluster is not under maintenance"))
		return
	}
	dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
	// No need to wrap with errors.Occur as the err != nil will be auto-wrapped into errors.Occur(errors.ErrUnexpected, err).
	common.SendResponse(c, dagDetailDTO, err)
}

// @ID				GetAgentLastMaintenanceDag
// @Summary		get agent last maintenance dag
// @Description	get agent last maintenance dag
// @Tags			task
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure		404				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/task/dag/maintain/agent [get]
func GetAgentLastMaintenanceDag(c *gin.Context) {
	param := getTaskQueryParams(c)
	dag, err := localTaskService.FindLastMaintenanceDag()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if dag == nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound, "Agent is not under maintenance"))
		return
	}
	// construct dagDetailDTO
	dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
	// No need to wrap with errors.Occur as the err != nil will be auto-wrapped into errors.Occur(errors.ErrUnexpected, err).
	common.SendResponse(c, dagDetailDTO, err)
}

// @ID				GetAllAgentsLastMaintenanceDag
// @Summary		get agent last maintenance dag
// @Description	get agent last maintenance dag
// @Tags			task
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure		404				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/task/dag/maintain/agents [get]
func GetAllAgentLastMaintenanceDag(c *gin.Context) {
	param := getTaskQueryParams(c)
	dagDetailDTOs := make([]*task.DagDetailDTO, 0)
	dag, err := localTaskService.FindLastMaintenanceDag()
	if err == nil {
		dagDetailDTO, err := convertDagDetailDTO(dag, *param.ShowDetails)
		if err != nil {
			log.WithContext(common.NewContextWithTraceId(c)).Errorf("convert dag detail dto failed: %v", err)
		} else {
			dagDetailDTOs = append(dagDetailDTOs, dagDetailDTO)
		}
	} else {
		log.WithContext(common.NewContextWithTraceId(c)).Errorf("get agent last maintenance dag failed: %v", err)
	}

	agents, _ := agentService.GetAllAgentsInfo()
	for _, agent := range agents {
		if agent.Equal(meta.OCS_AGENT) {
			continue
		}

		var dagDetailDTO *task.DagDetailDTO
		url := fmt.Sprintf("%s%s%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, constant.URI_AGENT_GROUP)
		err = secure.SendGetRequest(&agent, url, nil, &dagDetailDTO)
		if err != nil {
			log.WithContext(common.NewContextWithTraceId(c)).Errorf("get agent last maintenance dag failed: %v", err)
		} else {
			dagDetailDTOs = append(dagDetailDTOs, dagDetailDTO)
		}
	}
	common.SendResponse(c, dagDetailDTOs, nil)
}

// @ID				GetUnfinishedDags
// @Summary		get unfinished dags
// @Description	get unfinished dags
// @Tags			task
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string					true	"Authorization"
// @Param			showDetails		query	param.TaskQueryParams	true	"show details"
// @Success		200				object	http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Router			/api/v1/task/dag/unfinish [get]
func GetUnfinishedDags(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	switch meta.OCS_AGENT.GetIdentity() {
	case meta.FOLLOWER:
		master := agentService.GetMasterAgentInfo()
		if master == nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "master is not found"))
			return
		}
		common.ForwardRequest(c, master)
	case meta.CLUSTER_AGENT:
		param := getTaskQueryParams(c)
		agentsDags, errs := getAllAgentUnfinishDags(param)
		if len(errs) != 0 {
			log.WithContext(ctx).Errorf("get agent unfinished dags failed: %v", errs)
		}
		clusterDags, err := getClusterUnfinishDags(*param.ShowDetails)
		if err != nil {
			log.WithContext(ctx).Errorf("get cluster unfinished dags failed: %v", err)
		}
		dags := append(agentsDags, clusterDags...)
		common.SendResponse(c, dags, nil)
	case meta.SINGLE, meta.MASTER, meta.TAKE_OVER_MASTER, meta.TAKE_OVER_FOLLOWER:
		GetAgentUnfinishDags(c)
	default:
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "unknown agent identity: %s", meta.OCS_AGENT.GetIdentity()))
	}
}

// @ID				GetClusterUnfinishDags
// @Summary		get cluster unfinished dags
// @Description	get cluster unfinished dags
// @Tags			task
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string					true	"Authorization"
// @Param			showDetails		query	param.TaskQueryParams	true	"show details"
// @Success		200				object	http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/task/dag/ob/unfinish [get]
func GetClusterUnfinishDags(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "only cluster agent can get cluster unfinished dags"))
		return
	}

	param := getTaskQueryParams(c)
	dagDetailDTOs, err := getClusterUnfinishDags(*param.ShowDetails)
	common.SendResponse(c, dagDetailDTOs, err)
}

// @ID				GetAgentUnfinishDags
// @Summary		get agent unfinished dags
// @Description	get agent unfinished dags
// @Tags			task
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string					true	"Authorization"
// @Param			showDetails		query	param.TaskQueryParams	true	"show details"
// @Success		200				object	http.OcsAgentResponse{data=[]task.DagDetailDTO}
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/task/dag/agent/unfinish [get]
func GetAgentUnfinishDags(c *gin.Context) {
	param := getTaskQueryParams(c)
	dagDetailDTOs, err := getAgentUnfinishDags(*param.ShowDetails)
	common.SendResponse(c, dagDetailDTOs, err)
}

func getAllAgentUnfinishDags(param *param.TaskQueryParams) (dags []*task.DagDetailDTO, errs []error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		errs = append(errs, err)
		return
	}

	for _, agent := range agents {
		var dagDetailDTOs []*task.DagDetailDTO
		if agent.Equal(meta.OCS_AGENT) {
			dagDetailDTOs, err = getAgentUnfinishDags(*param.ShowDetails)
		} else {
			url := fmt.Sprintf("%s%s%s%s?show_details=%v", constant.URI_TASK_API_PREFIX, constant.URI_DAG, constant.URI_AGENT_GROUP, constant.URI_UNFINISH, *param.ShowDetails)
			err = secure.SendGetRequest(&agent, url, nil, &dagDetailDTOs)
		}
		if err != nil {
			errs = append(errs, err)
		}
		dags = append(dags, dagDetailDTOs...)
	}
	return
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
