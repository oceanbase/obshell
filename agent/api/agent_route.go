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

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	http2 "github.com/oceanbase/obshell/agent/lib/http"
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
		common.PostHandlers("/debug/pprof", "/swagger",
			// get all obcluster parameters
			constant.URI_API_V1+constant.URI_OBCLUSTER_GROUP+constant.URI_PARAMETERS,
		),
		common.HeaderDecrypt(),
		common.BodyDecrypt(constant.URI_API_V1+constant.URI_PACKAGE), // decrypt request body
		common.PaddingBody(), // if the response body is empty, the response body is padded with "{}"
		common.PreHandlers(
			constant.URI_API_V1+constant.URI_UPGRADE+constant.URI_PACKAGE,
			// upload obproxy package
			constant.URI_OBPROXY_API_PREFIX+constant.URI_PACKAGE,
			constant.URI_API_V1+constant.URI_PACKAGE,

			constant.URI_TASK_RPC_PREFIX+constant.URI_SUB_TASK,
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

	ob := v1.Group(constant.URI_OB_GROUP)
	agent := v1.Group(constant.URI_AGENT_GROUP)
	agents := v1.Group(constant.URI_AGENTS_GROUP)
	obcluster := v1.Group(constant.URI_OBCLUSTER_GROUP)
	observer := v1.Group(constant.URI_OBSERVER_GROUP)
	upgrade := v1.Group(constant.URI_UPGRADE)
	unit := v1.Group(constant.URI_UNIT_GROUP)
	units := v1.Group(constant.URI_UNITS_GROUP)
	pools := v1.Group(constant.URI_POOLS_GROUP)
	pool := v1.Group(constant.URI_POOL_GROUP)
	recyclebin := v1.Group(constant.URI_RECYCLEBIN_GROUP)
	zone := v1.Group(constant.URI_ZONE_GROUP)
	pkg := v1.Group(constant.URI_PACKAGE)

	if !isLocalRoute {
		ob.Use(common.Verify())
		agent.Use(common.Verify())
		agents.Use(common.Verify())
		obcluster.Use(common.Verify())
		observer.Use(common.Verify())
		upgrade.Use(common.Verify())
		unit.Use(common.Verify())
		units.Use(common.Verify())
		pools.Use(common.Verify())
		pool.Use(common.Verify())
		recyclebin.Use(common.Verify())
	}

	v1.GET(constant.URI_TIME, TimeHandler)
	v1.GET(constant.URI_INFO, InfoHandler(s))
	v1.GET(constant.URI_GIT_INFO, GitInfoHandler)
	v1.GET(constant.URI_STATUS, StatusHandler(s))
	v1.POST(constant.URI_STATUS, StatusHandler(s))
	v1.GET(constant.URI_SECRET, secretHandler)

	InitTaskRoutes(v1, isLocalRoute)
	InitTenantRoutes(v1, isLocalRoute)
	InitBackupRoutes(v1, isLocalRoute)
	InitRestoreRoutes(v1, isLocalRoute)
	InitObproxyRoutes(v1, isLocalRoute)
	InitMetricRoutes(v1, isLocalRoute)
	InitAlarmRoutes(v1, isLocalRoute)

	system := v1.Group(constant.URI_SYSTEM_GROUP)
	InitExternalRoutes(system, isLocalRoute)

	// ob routes
	ob.POST(constant.URI_INIT, obInitHandler)
	ob.POST(constant.URI_STOP, obStopHandler)
	ob.POST(constant.URI_START, obStartHandler)
	ob.GET(constant.URI_INFO, obInfoHandler)
	ob.POST(constant.URI_SCALE_OUT, obClusterScaleOutHandler)
	ob.POST(constant.URI_SCALE_IN, obClusterScaleInHandler)
	ob.POST(constant.URI_UPGRADE, obUpgradeHandler)
	ob.POST(constant.URI_UPGRADE+constant.URI_CHECK, obUpgradeCheckHandler)
	ob.GET(constant.URI_AGENTS, obAgentsHandler)

	// agent routes
	agent.GET(constant.URI_HOST_INFO, GetHostInfo)
	agent.POST(constant.URI_JOIN, agentJoinHandler)
	agent.POST("", agentJoinHandler)
	agent.DELETE("", agentRemoveHandler)
	agent.POST(constant.URI_REMOVE, agentRemoveHandler)
	agent.POST(constant.URI_UPGRADE, agentUpgradeHandler)
	agent.POST(constant.URI_UPGRADE+constant.URI_CHECK, agentUpgradeCheckHandler)
	agent.POST(constant.URI_PASSWORD, agentSetPasswordHandler)

	// agents routes
	agents.GET(constant.URI_STATUS, GetAllAgentStatus(s))

	// obcluster routes
	obcluster.PUT(constant.URI_CONFIG, obclusterConfigHandler(true))
	obcluster.POST(constant.URI_CONFIG, obclusterConfigHandler(true))
	obcluster.GET(constant.URI_INFO, obclusterInfoHandler)
	obcluster.GET(constant.URI_PARAMETERS, obclusterParametersHandler)
	obcluster.PATCH(constant.URI_PARAMETERS, obclusterSetParametersHandler)
	obcluster.GET(constant.URI_CHARSETS, getObclusterCharsets)
	obcluster.GET(constant.URI_STATISTICS, GetStatistics)
	obcluster.GET(constant.URI_UNIT_CONFIG_LIMIT, checkClusterAgentWrapper(getUnitConfigLimitHandler))

	// observer routes
	observer.PUT(constant.URI_CONFIG, obServerConfigHandler(true))
	observer.POST(constant.URI_CONFIG, obServerConfigHandler(true))
	observer.DELETE("", obClusterScaleInHandler)

	// zone routes
	zone.DELETE(constant.URI_PATH_PARAM_NAME, zoneDeleteHandler)

	// upgrade routes
	upgrade.POST(constant.URI_PACKAGE, pkgUploadHandler)
	upgrade.GET(constant.URI_PACKAGE+constant.URI_INFO, pkgInfoHandler)
	upgrade.POST(constant.URI_PARAMS+constant.URI_BACKUP, paramsBackupHandler)
	upgrade.POST(constant.URI_PARAMS+constant.URI_RESTORE, paramsRestoreHandler)

	pkg.POST("", common.VerifyFile(), newPkgUploadHandler)

	// unit routes
	unit.POST("", unitConfigCreateHandler)
	unit.DELETE(constant.URI_PATH_PARAM_NAME, unitConfigDropHandler)
	unit.GET(constant.URI_PATH_PARAM_NAME, unitConfigGetHandler)
	units.GET("", unitConfigListHandler)

	// pool routes
	pools.GET("", poolListHandler)
	pool.DELETE(constant.URI_PATH_PARAM_NAME, poolDropHandler)

	// recyclebin routes
	recyclebin.GET(constant.URI_TENANTS, recyclebinListTenantHandler)
	recyclebin.DELETE(constant.URI_TENANT_GROUP+constant.URI_PATH_PARAM_NAME, recyclebinPurgeTenantHandler)
	recyclebin.POST(constant.URI_TENANT_GROUP+constant.URI_PATH_PARAM_NAME, recyclebinFlashbackTenantHandler)
	r.NoRoute(func(c *gin.Context) {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonNotFound, "404 not found"))
	})
}
