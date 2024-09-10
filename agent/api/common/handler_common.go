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

package common

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	libhttp "github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/log"
)

const (
	OcsAgentResponseKey = "ocsAgentResponse"
	TraceIdKey          = "traceId"
)

// NewContextWithTraceId extracts the traceId value from the Gin context
// and embeds it into a new standard context, which can be used in
// subsequent operations that require tracing.
func NewContextWithTraceId(c *gin.Context) context.Context {
	traceId := ""
	if t, ok := c.Get(TraceIdKey); ok {
		if ts, ok := t.(string); ok {
			traceId = ts
		}
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, log.TraceIdKey{}, traceId)
	return ctx
}

// SendResponse constructs a standardized response object and attaches it to the Gin context.
// It is typically used to ensure that all HTTP responses have a consistent format.
func SendResponse(c *gin.Context, data interface{}, err error) {
	var resp libhttp.OcsAgentResponse
	if c.Writer.Status() == http.StatusNoContent {
		resp = libhttp.BuildNoContentResponse()
	} else {
		resp = libhttp.BuildResponse(data, err)
	}
	c.Set(OcsAgentResponseKey, resp)
}

func SendNoContentResponse(c *gin.Context, err error) {
	c.Status(http.StatusNoContent)
	SendResponse(c, nil, err)
}

func IsLocalRoute(c *gin.Context) bool {
	_, isLocalRoute := c.Get(localRouteKey)
	return isLocalRoute
}

func IsApiRoute(c *gin.Context) bool {
	_, isApiRoute := c.Get(apiRouteKey)
	return isApiRoute
}
