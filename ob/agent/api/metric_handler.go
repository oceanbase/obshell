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
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	metricexecutor "github.com/oceanbase/obshell/ob/agent/executor/metric"
	metricconstant "github.com/oceanbase/obshell/ob/agent/executor/metric/constant"
	"github.com/oceanbase/obshell/ob/model/metric"
	log "github.com/sirupsen/logrus"
)

// @ID ListAllMetrics
// @Summary list all metrics
// @Description list all metrics meta info, return by groups
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Param scope query string true "metrics scope" Enums(OBCLUSTER, OBTENANT)
// @Success 200 object http.OcsAgentResponse{data=[]metric.MetricClass}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/metrics [GET]
// @Security ApiKeyAuth
func ListMetricMetas(c *gin.Context) {
	language := c.GetHeader(constant.ACCEPT_LANGUAGE)
	scope := c.Query(metricconstant.PARAM_SCOPE)
	switch scope {
	case metricconstant.SCOPE_CLUSTER,
		metricconstant.SCOPE_TENANT,
		metricconstant.SCOPE_CLUSTER_OVERVIEW,
		metricconstant.SCOPE_TENANT_OVERVIEW,
		metricconstant.SCOPE_OBPROXY:
	default:
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonBadRequest, "invalid scope"))
		return
	}
	metricClasses, err := metricexecutor.ListMetricClasses(scope, language)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	log.Debugf("List metric classes: %+v", metricClasses)
	common.SendResponse(c, metricClasses, nil)
}

// @ID QueryMetrics
// @Summary query metrics
// @Description query metric data
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Param body body metric.MetricQuery true "metric query request body"
// @Success 200 object http.OcsAgentResponse{data=[]metric.MetricData}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/metrics/query [POST]
// @Security ApiKeyAuth
func QueryMetrics(c *gin.Context) {
	queryParam := &metric.MetricQuery{}
	err := c.Bind(queryParam)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonBadRequest, err.Error()))
		return
	}
	log.Infof("Query metric data with param: %+v", queryParam)
	metricDatas := metricexecutor.QueryMetricData(queryParam)
	log.Debugf("Query metric data: %+v", metricDatas)
	common.SendResponse(c, metricDatas, nil)
}
