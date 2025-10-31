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

	"github.com/oceanbase/obshell/seekdb/agent/api/common"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/executor/observer"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/param"
)

// @ID stopObserver
// @Summary stop observer
// @Description stop observer
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param body body param.ObStopParam true "stop observer params"
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/stop [post]
func obStopHandler(c *gin.Context) {
	var param param.ObStopParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	emergencyMode, err := isEmergencyMode(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if emergencyMode {
		data, err := observer.EmergencyStop()
		common.SendResponse(c, data, err)
	} else {
		data, err := observer.CreateStopDag(param)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		if data == nil {
			common.SendNoContentResponse(c, nil)
			return
		}
		common.SendResponse(c, data, nil)
	}
}

// @ID startObserver
// @Summary start observers
// @Description start observers
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/start [post]
func obStartHandler(c *gin.Context) {
	var param param.StartObParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	emergencyMode, err := isEmergencyMode(c)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	hasStart, e := observer.HasStarted()
	if e != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if !hasStart {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObClusterNotInitialized))
		return
	}

	if emergencyMode {
		data, err := observer.EmergencyStart()
		common.SendResponse(c, data, err)
	} else {
		data, err := observer.CreateStartDag()
		common.SendResponse(c, data, err)
	}
}

// @ID restartObserver
// @Summary restart observer
// @Description restart observer
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ObRestartParam true "restart observer params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/restart [post]
func obRestartHandler(c *gin.Context) {
	var param param.ObRestartParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	data, err := observer.CreateRestartDag(param)
	common.SendResponse(c, data, err)
}

// @ID GetObInfo
// @Summary get ob and agent info
// @Description get ob and agent info
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=observer.ObserverInfo}
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/info [get]
func observerInfoHandler(c *gin.Context) {
	data := observer.GetObserverInfo()
	common.SendResponse(c, data, nil)
}

func isEmergencyMode(c *gin.Context) (bool, error) {
	if common.IsLocalRoute(c) && !meta.OCS_AGENT.IsClusterAgent() {
		return true, nil
	}
	return false, nil
}

// @ID				getCompaction
// @Summary		get major compaction info
// @Description	get major compaction info
// @Tags			observer
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=bo.TenantCompaction}
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/observer/compaction [get]
func getCompactionHandler(c *gin.Context) {
	compaction, err := observer.GetCompaction()
	common.SendResponse(c, compaction, err)
}

// @ID				majorCompaction
// @Summary		trigger  major compaction
// @Description	trigger  major compaction
// @Tags			observer
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/observer/compact [post]
func majorCompactionHandler(c *gin.Context) {
	common.SendResponse(c, nil, observer.MajorCompaction())
}

// @ID				clearCompactionError
// @Summary		clear major compaction error
// @Description	clear major compaction error
// @Tags			observer
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/observer/compaction-error [delete]
func clearCompactionErrorHandler(c *gin.Context) {
	common.SendResponse(c, nil, observer.ClearCompactionError())
}

// @ID modifyWhitelist
// @Summary modify whitelist
// @Description modify whitelist
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ModifyWhitelistParam true "modify whitelist params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/whitelist [put]
func modifyWhitelistHandler(c *gin.Context) {
	var param param.ModifyWhitelistParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	var err error
	if param.Whitelist == nil {
		err = observer.ModifyWhitelist("")
	} else {
		err = observer.ModifyWhitelist(*param.Whitelist)
	}
	common.SendResponse(c, nil, err)
}

// @ID setParameters
// @Summary set parameters
// @Description set parameters
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.SetParametersParam true "set parameters params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/parameters [PATCH]
func setParametersHandler(c *gin.Context) {
	var param param.SetParametersParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, observer.SetParameters(param.Parameters))
}

// @ID setVariables
// @Summary set variables
// @Description set variables
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.SetVariablesParam true "set global variables params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/variables [PATCH]
func setVariablesHandler(c *gin.Context) {
	var param param.SetVariablesParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, observer.SetVariables(param))
}

// @ID getParameters
// @Summary get parameters
// @Description get parameters
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param filter query string false "filter format"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.GvObParameter}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/parameters [get]
func getParameters(c *gin.Context) {
	format := c.Query("filter")
	parameters, err := observer.GetParameters(format)
	common.SendResponse(c, parameters, err)
}

// @ID getVariables
// @Summary get variables
// @Description get variables
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param filter query string false "filter format"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.DbaObSysVariable}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/variables [get]
func getVariables(c *gin.Context) {
	format := c.Query("filter")
	variables, err := observer.GetVariables(format)
	common.SendResponse(c, variables, err)
}

// StatisticsHandler returns the statistics data
//
// @ID GetStatistics
// @Summary get statistics data
// @Description get statistics data
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=bo.ObclusterStatisticInfo}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/statistics [GET]
func GetStatistics(c *gin.Context) {
	statisticsData := observer.GetStatisticsInfo()
	common.SendResponse(c, statisticsData, nil)
}

// @ID getObserverCharsets
// @Summary get observer charsets
// @Description get observer charsets
// @Tags observer
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]bo.CharsetInfo}
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/charsets [get]
func getObserverCharsets(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	charsets, err := observer.GetObserverCharsets()
	common.SendResponse(c, charsets, err)
}
