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
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/inspection"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/param"
)

// @ID triggerInspection
// @Summary trigger cluster inspection
// @Description trigger cluster inspection with specified scenario (basic or performance)
// @Tags obcluster
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.InspectionParam true "inspection parameters"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/inspection [post]
func triggerInspectionHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	var params param.InspectionParam
	if err := c.BindJSON(&params); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	data, err := inspection.TriggerInspection(&params)
	common.SendResponse(c, data, err)
}

// @ID getInspectionHistory
// @Summary get inspection history
// @Description get paginated inspection history with filters
// @Tags obcluster
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param scenario query string false "Scenario filter ('basic' or 'performance' or 'basic,performance')"
// @Param sort query string false "Sort parameter (format: field,order)" default(start_time,desc)
// @Success 200 object http.OcsAgentResponse{data=bo.PaginatedInspectionHistoryResponse}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/inspection/reports [get]
func getInspectionHistoryHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	p := &param.QueryInspectionHistoryParam{}
	if err := c.BindQuery(p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	p.Format()

	data, err := inspection.GetInspectionHistory(p)
	common.SendResponse(c, data, err)
}

// @ID getInspectionReport
// @Summary get inspection report by id
// @Description get detailed inspection report by report id
// @Tags obcluster
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path string true "Inspection report id"
// @Success 200 object http.OcsAgentResponse{data=bo.InspectionReport}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/inspection/report/{id} [get]
func getInspectionReportHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	id := c.Param("id")
	if id == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonIllegalArgument, "id"))
		return
	}

	data, err := inspection.GetInspectionReport(id)
	common.SendResponse(c, data, err)
}
