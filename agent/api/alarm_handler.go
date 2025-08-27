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

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/executor/alarm"
	"github.com/oceanbase/obshell/model/alarm/alert"
	"github.com/oceanbase/obshell/model/alarm/rule"
	"github.com/oceanbase/obshell/model/alarm/silence"
)

// ListAlerts godoc
// @ID ListAlerts
// @Summary List all alerts
// @Description List all alerts
// @Tags alarm
// @Accept json
// @Produce json
// @Param filter body alert.AlertFilter false "alert filter"
// @Success 200 {object} http.OcsAgentResponse{data=[]alert.Alert}
// @Router /api/v1/alarm/alerts [post]
func ListAlerts(ctx *gin.Context) {
	filter := &alert.AlertFilter{}
	err := ctx.Bind(filter)
	if err != nil {
		common.SendResponse(ctx, nil, err)
		return
	}
	data, err := alarm.ListAlerts(ctx, filter)
	common.SendResponse(ctx, data, err)
}

// ListSilencers godoc
// @ID ListSilencers
// @Summary List all silencers
// @Description List all silencers
// @Tags alarm
// @Accept json
// @Produce json
// @Param filter body silence.SilencerFilter false "silencer filter"
// @Success 200 {object} http.OcsAgentResponse{data=[]silence.SilencerResponse}
// @Router /api/v1/alarm/silencers [post]
func ListSilencers(ctx *gin.Context) {
	filter := &silence.SilencerFilter{}
	err := ctx.Bind(filter)
	if err != nil {
		common.SendResponse(ctx, nil, err)
		return
	}
	data, err := alarm.ListSilencers(ctx, filter)
	common.SendResponse(ctx, data, err)
}

// GetSilencer godoc
// @ID GetSilencer
// @Summary Get a silencer
// @Description Get a silencer by id
// @Tags alarm
// @Accept json
// @Produce json
// @Param id path string true "silencer id"
// @Success 200 {object} http.OcsAgentResponse{data=silence.SilencerResponse}
// @Router /api/v1/alarm/silencer/{id} [get]
func GetSilencer(ctx *gin.Context) {
	id := ctx.Param("id")
	data, err := alarm.GetSilencer(ctx, id)
	common.SendResponse(ctx, data, err)
}

// CreateOrUpdateSilencer godoc
// @ID CreateOrUpdateSilencer
// @Summary Create or update a silencer
// @Description Create or update a silencer
// @Tags alarm
// @Accept json
// @Produce json
// @Param silencer body silence.SilencerParam true "silencer"
// @Success 200 {object} http.OcsAgentResponse{data=silence.SilencerResponse}
// @Router /api/v1/alarm/silencer [put]
func CreateOrUpdateSilencer(ctx *gin.Context) {
	param := &silence.SilencerParam{}
	err := ctx.Bind(param)
	if err != nil {
		common.SendResponse(ctx, nil, err)
		return
	}
	data, err := alarm.CreateOrUpdateSilencer(ctx, param)
	common.SendResponse(ctx, data, err)
}

// DeleteSilencer godoc
// @ID DeleteSilencer
// @Summary Delete a silencer
// @Description Delete a silencer by id
// @Tags alarm
// @Accept json
// @Produce json
// @Param id path string true "silencer id"
// @Success 200 {object} http.OcsAgentResponse
// @Router /api/v1/alarm/silencer/{id} [delete]
func DeleteSilencer(ctx *gin.Context) {
	id := ctx.Param("id")
	err := alarm.DeleteSilencer(ctx, id)
	common.SendResponse(ctx, nil, err)
}

// ListRules godoc
// @ID ListRules
// @Summary List all rules
// @Description List all rules
// @Tags alarm
// @Accept json
// @Produce json
// @Param filter body rule.RuleFilter false "rule filter"
// @Success 200 {object} http.OcsAgentResponse{data=[]rule.RuleResponse}
// @Router /api/v1/alarm/rules [post]
func ListRules(ctx *gin.Context) {
	filter := &rule.RuleFilter{}
	err := ctx.Bind(filter)
	if err != nil {
		common.SendResponse(ctx, nil, err)
		return
	}
	data, err := alarm.ListRules(ctx, filter)
	common.SendResponse(ctx, data, err)
}

// GetRule godoc
// @ID GetRule
// @Summary Get a rule
// @Description Get a rule by name
// @Tags alarm
// @Accept json
// @Produce json
// @Param name path string true "rule name"
// @Success 200 {object} http.OcsAgentResponse{data=rule.RuleResponse}
// @Router /api/v1/alarm/rule/{name} [get]
func GetRule(ctx *gin.Context) {
	name := ctx.Param(constant.URI_PARAM_NAME)
	data, err := alarm.GetRule(ctx, name)
	common.SendResponse(ctx, data, err)
}
