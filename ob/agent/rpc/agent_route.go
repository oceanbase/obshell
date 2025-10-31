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

package rpc

import (
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/constant"
	http2 "github.com/oceanbase/obshell/ob/agent/lib/http"
)

func InitOcsAgentRpcRoutes(s *http2.State, r *gin.Engine, isLocalRoute bool) {
	if isLocalRoute {
		r.Use(common.UnixSocketMiddleware())
	}
	v1 := r.Group(constant.URI_RPC_V1)

	v1.Use(
		common.Verify(),
	)

	agent := v1.Group(constant.URI_AGENT_GROUP)
	agent.POST("", agentJoinHandler)
	agent.POST(constant.URI_TOKEN, agentAddTokenHandler)
	agent.DELETE("", agentRemoveHandler)
	agent.POST(constant.URI_UPDATE, agentUpdateHandler)
	agent.POST(constant.URI_SYNC_BIN, takeOverAgentUpdateBinaryHandler)

	InitTaskRoutes(v1)

	ob := v1.Group(constant.URI_OB_GROUP)
	ob.POST(constant.URI_START, obStartHandler)
	ob.POST(constant.URI_STOP, obStopHandler)
	ob.POST(constant.URI_DEPLOY, obServerDeployHandler)
	ob.POST(constant.URI_DESTROY, obServerDestroyHandler)
	ob.POST(constant.URI_SCALE_OUT, obLocalScaleOutHandler)

	observer := v1.Group(constant.URI_OBSERVER_GROUP)
	observer.POST(constant.URI_DEPLOY, obServerDeployHandler)
	observer.DELETE("", killObserverHandler)
	observer.POST("", startObserverHandler)

	maintainer := v1.Group(constant.URI_MAINTAINER)
	maintainer.GET("", getMaintainerHandler)
	maintainer.POST(constant.URI_UPDATE, updateAllAgentsHandler)
}
