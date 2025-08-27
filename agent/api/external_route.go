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

func InitExternalRoutes(r *gin.RouterGroup, isLocalRoute bool) {
	external := r.Group(constant.URI_EXTERNAL_GROUP)

	if !isLocalRoute {
		external.Use(common.Verify())
	}

	external.PUT(constant.URI_PROMETHEUS, SetPrometheusConfig)
	external.GET(constant.URI_PROMETHEUS, GetPrometheusConfig)
	external.PUT(constant.URI_ALERTMANAGER, SetAlertmanagerConfig)
	external.GET(constant.URI_ALERTMANAGER, GetAlertmanagerConfig)
}
