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
	if isLocalRoute {
		r.Use(common.UnixSocketMiddleware())
	}
	r.Use(
		gin.CustomRecovery(common.Recovery), // gin's crash-free middleware
		common.PostHandlers("/debug/pprof", "/swagger"),
		common.BodyDecrypt(constant.URI_API_V1+constant.URI_OB_GROUP+constant.URI_INFO), // decrypt request body
		common.PreHandlers(constant.URI_API_V1+constant.URI_UPGRADE+constant.URI_PACKAGE,
			constant.URI_API_V1+constant.URI_OBCLUSTER_GROUP+constant.URI_CONFIG),
		common.SetContentType,
	)

	if isLocalRoute {
		r.Use(common.SetLocalRouteFlag)
	} else {
		initSwagger(r)
	}

	// groups
	v1 := r.Group(constant.URI_API_V1)
	ob := v1.Group(constant.URI_OB_GROUP)
	agent := v1.Group(constant.URI_AGENT_GROUP)
	agents := v1.Group(constant.URI_AGENTS_GROUP)
	obcluster := v1.Group(constant.URI_OBCLUSTER_GROUP)
	observer := v1.Group(constant.URI_OBSERVER_GROUP)
	upgrade := v1.Group(constant.URI_UPGRADE)

	if !isLocalRoute {
		ob.Use(common.Verify(constant.URI_API_V1 + constant.URI_OB_GROUP + constant.URI_INFO))
		agent.Use(common.Verify())
		agents.Use(common.Verify())
		obcluster.Use(common.Verify())
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

	// ob routes
	ob.POST(constant.URI_INIT, obInitHandler)
	ob.POST(constant.URI_STOP, obStopHandler)
	ob.POST(constant.URI_START, obStartHandler)
	ob.GET(constant.URI_INFO, obInfoHandler)
	ob.POST(constant.URI_SCALE_OUT, obClusterScaleOutHandler)
	ob.POST(constant.URI_UPGRADE, obUpgradeHandler)
	ob.POST(constant.URI_UPGRADE+constant.URI_CHECK, obUpgradeCheckHandler)
	ob.GET(constant.URI_AGENTS, obAgentsHandler)

	// agent routes
	agent.POST(constant.URI_JOIN, agentJoinHandler)
	agent.POST("", agentJoinHandler)
	agent.DELETE("", agentRemoveHandler)
	agent.POST(constant.URI_REMOVE, agentRemoveHandler)
	agent.POST(constant.URI_UPGRADE, agentUpgradeHandler)
	agent.POST(constant.URI_UPGRADE+constant.URI_CHECK, agentUpgradeCheckHandler)

	// agents routes
	agents.GET(constant.URI_STATUS, GetAllAgentStatus(s))

	// obcluster routes
	obcluster.PUT(constant.URI_CONFIG, obclusterConfigHandler(true))
	obcluster.POST(constant.URI_CONFIG, obclusterConfigHandler(true))

	// observer routes
	observer.PUT(constant.URI_CONFIG, obServerConfigHandler(true))
	observer.POST(constant.URI_CONFIG, obServerConfigHandler(true))

	// upgrade routes
	upgrade.POST(constant.URI_PACKAGE, pkgUploadHandler)
	upgrade.POST(constant.URI_PARAMS+constant.URI_BACKUP, paramsBackupHandler)
	upgrade.POST(constant.URI_PARAMS+constant.URI_RESTORE, paramsRestoreHandler)

	r.NoRoute(func(c *gin.Context) {
		err := errors.Occur(errors.ErrBadRequest, "404 not found")
		common.SendResponse(c, nil, err)
	})
}
