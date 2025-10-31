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
	"github.com/gin-gonic/gin/binding"

	"github.com/oceanbase/obshell/seekdb/agent/api/common"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	http2 "github.com/oceanbase/obshell/seekdb/agent/lib/http"
)

// @title			obshell API
// @version		1.0
// @description	This is a set of operation and maintenance management API interfaces developed based on observer.
// @BasePath		/
// @contact.name	obshell API Support
// @contact.url	https://open.oceanbase.com
// @license.name	Apache - 2.0
// @license.url	http://www.apache.org/licenses/
func InitOcsAgentRoutes(s *http2.State, r *gin.Engine, isLocalRoute bool) {
	binding.EnableDecoderUseNumber = true
	if isLocalRoute {
		r.Use(common.UnixSocketMiddleware())
	}
	r.Use(
		gin.CustomRecovery(common.Recovery), // gin's crash-free middleware
		common.PostHandlers("/debug/pprof", "/swagger"),
		common.HeaderDecrypt(),
		common.BodyDecrypt(constant.URI_API_V1+constant.URI_PACKAGE), // decrypt request body
		common.PaddingBody(),                                         // if the response body is empty, the response body is padded with "{}"
		common.PreHandlers(
			constant.URI_API_V1+constant.URI_UPGRADE+constant.URI_PACKAGE,
			constant.URI_API_V1+constant.URI_PACKAGE,
		),
		common.SetContentType,
	)

	if isLocalRoute {
		r.Use(common.SetLocalRouteFlag)
	} else {
		initSwagger(r)
		initFrontendRouter(r)
	}

	// groups
	v1 := r.Group(constant.URI_API_V1)
	v1.Use(common.SetApiFlag)

	agent := v1.Group(constant.URI_AGENT_GROUP)
	observer := v1.Group(constant.URI_OBSERVER_GROUP)
	upgrade := v1.Group(constant.URI_UPGRADE)
	pkg := v1.Group(constant.URI_PACKAGE)

	if !isLocalRoute {
		agent.Use(common.Verify())
		observer.Use(common.Verify())
		upgrade.Use(common.Verify())
	}

	v1.GET(constant.URI_TIME, TimeHandler)
	v1.GET(constant.URI_INFO, InfoHandler(s))
	v1.GET(constant.URI_GIT_INFO, GitInfoHandler)
	v1.GET(constant.URI_STATUS, StatusHandler(s))
	v1.POST(constant.URI_STATUS, StatusHandler(s))
	v1.GET(constant.URI_SECRET, secretHandler)

	InitTaskRoutes(v1, isLocalRoute)
	InitMetricRoutes(v1, isLocalRoute)
	InitAlarmRoutes(v1, isLocalRoute)

	system := v1.Group(constant.URI_SYSTEM_GROUP)
	InitExternalRoutes(system, isLocalRoute)

	// observer routes
	observer.POST(constant.URI_STOP, obStopHandler)
	observer.POST(constant.URI_START, obStartHandler)
	observer.POST(constant.URI_RESTART, obRestartHandler)
	observer.GET(constant.URI_INFO, observerInfoHandler)
	observer.GET(constant.URI_STATISTICS, GetStatistics)
	// for compaction
	observer.GET(constant.URI_COMPACTION, getCompactionHandler)
	observer.POST(constant.URI_COMPACT, majorCompactionHandler)
	observer.DELETE(constant.URI_COMPACTION_ERROR, clearCompactionErrorHandler)
	// for whitelist
	observer.PUT(constant.URI_WHITELIST, modifyWhitelistHandler)
	// for variables
	observer.GET(constant.URI_VARIABLES, getVariables)
	observer.PATCH(constant.URI_VARIABLES, setVariablesHandler)
	// for parameters
	observer.GET(constant.URI_PARAMETERS, getParameters)
	observer.PATCH(constant.URI_PARAMETERS, setParametersHandler)
	// for charsets
	observer.GET(constant.URI_CHARSETS, getObserverCharsets)

	InitUserRoutes(observer, isLocalRoute)
	InitDatabaseRoutes(observer, isLocalRoute)

	// agent routes
	agent.POST(constant.URI_UPGRADE, agentUpgradeHandler)
	agent.POST(constant.URI_UPGRADE+constant.URI_CHECK, agentUpgradeCheckHandler)

	// upgrade routes
	upgrade.POST(constant.URI_PACKAGE, pkgUploadHandler)
	upgrade.DELETE(constant.URI_PACKAGE, pkgDeleteHandler)
	upgrade.GET(constant.URI_PACKAGE+constant.URI_INFO, pkgInfoHandler)

	pkg.POST("", common.VerifyFile(), pkgUploadHandler)
	r.NoRoute(func(c *gin.Context) {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonNotFound, "404 not found"))
	})
}
