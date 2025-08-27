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
)

func InitAlarmRoutes(parentGroup *gin.RouterGroup, isLocalRoute bool) {
	alarm := parentGroup.Group(constant.URI_ALARM_GROUP)

	if !isLocalRoute {
		alarm.Use(common.Verify())
	}

	// alerts
	alarm.POST(constant.URI_ALERTS, ListAlerts)

	// silencers
	alarm.POST(constant.URI_SILENCERS, ListSilencers)
	alarm.GET(constant.URI_SILENCER+constant.URI_PATH_PARAM_ID, GetSilencer)
	alarm.PUT(constant.URI_SILENCER, CreateOrUpdateSilencer)
	alarm.DELETE(constant.URI_SILENCER+constant.URI_PATH_PARAM_ID, DeleteSilencer)

	// rules
	alarm.POST(constant.URI_RULES, ListRules)
	alarm.GET(constant.URI_RULE+constant.URI_PATH_PARAM_NAME, GetRule)
}
