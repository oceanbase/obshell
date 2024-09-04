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
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	libhttp "github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
	agentservice "github.com/oceanbase/obshell/agent/service/agent"
)

const (
	needForwardedFlag   = "forward"       // needForwardedFlag marks whether the current request should be forwarded
	IsAutoForwardedFlag = "IsAutoForward" // IsAutoForwardedFlag marks whether the current request is auto forwarded
)

// autoForward is used by middleware to forward the request to master agent.
func autoForward(c *gin.Context) {
	agentService := agentservice.AgentService{}
	master := agentService.GetMasterAgentInfo()
	if master == nil {
		SendResponse(c, nil, errors.Occur(errors.ErrUnauthorized))
		return
	}

	ctx := NewContextWithTraceId(c)
	log.WithContext(ctx).Infof("Forward request: [%v %v, client=%v, agent=%s]", c.Request.Method, c.Request.URL, c.ClientIP(), master.String())

	// OriginalBody only would be set in api request
	// Follower agent forward request to master agent, use the original encrypted body.
	// Repackage the request header
	var headers map[string]string
	var body interface{}
	if originalBody, exist := c.Get(originalBody); exist {
		body = originalBody
	}

	headerByte, exist := c.Get(constant.OCS_HEADER)
	if headerByte == nil || !exist {
		SendResponse(c, nil, errors.Occur(errors.ErrUnauthorized))
		return
	}

	header, ok := headerByte.(secure.HttpHeader)
	if !ok {
		SendResponse(c, nil, errors.Occur(errors.ErrUnauthorized))
		return
	}

	headers, err := secure.RepackageHeaderForAutoForward(&header, master)
	if err != nil {
		SendResponse(c, nil, err)
		return
	}

	sendRequsetForForward(c, ctx, master, headers, body)
}

// ForwardRequest is used by handler to forward the request to other agent.
func ForwardRequest(c *gin.Context, agentInfo meta.AgentInfoInterface, param ...interface{}) {
	ctx := NewContextWithTraceId(c)
	log.WithContext(ctx).Infof("Forward request: [%v %v, client=%v, agent=%s]", c.Request.Method, c.Request.URL, c.ClientIP(), agentInfo.String())

	var body interface{}
	if len(param) > 0 {
		body = param[0]
	}

	// forward for local route or cluster agent
	body, headers, err := buildForwardBodyAndHeader(agentInfo, c.Request.RequestURI, body)
	if err != nil {
		SendResponse(c, nil, err)
		return
	}

	sendRequsetForForward(c, ctx, agentInfo, headers, body)
}

func sendRequsetForForward(c *gin.Context, ctx context.Context, agentInfo meta.AgentInfoInterface, headers map[string]string, body interface{}) {
	startTime := time.Now()
	request := libhttp.NewClient().R()
	for k, v := range headers {
		request.SetHeader(k, v)
	}
	request.SetBody(body)

	uri := fmt.Sprintf("%s://%s:%d%s", global.Protocol, agentInfo.GetIp(), agentInfo.GetPort(), c.Request.URL)
	response, err := request.Execute(c.Request.Method, uri)
	if err != nil {
		log.WithError(err).Errorf("API response failed : [%v %v, client=%v, agent=%v]", c.Request.Method, c.Request.URL, c.ClientIP(), agentInfo.String())
		SendResponse(c, nil, err)
		return
	}

	for k, v := range response.Header() {
		c.Header(k, v[0])
	}

	c.Set(needForwardedFlag, true)
	c.Status(response.StatusCode())
	c.Writer.Write(response.Body())
	duration := time.Since(startTime).Milliseconds()
	traceId, _ := c.Get(TraceIdKey)
	log.WithContext(ctx).Infof("API response OK: [%v %v, client=%v, agent=%v, traceId=%v, duration=%v, status=%v]",
		c.Request.Method, c.Request.URL, c.ClientIP(), agentInfo.String(), traceId, duration, response.StatusCode())
}

func buildForwardBodyAndHeader(agentInfo meta.AgentInfoInterface, uri string, body interface{}) (interface{}, map[string]string, error) {
	var headers = map[string]string{}

	for _, route := range secure.GetSkipBodyEncryptRoutes() {
		if route == uri {
			headers = secure.BuildHeaderForForward(agentInfo, uri)
			return body, headers, nil
		}
	}

	body, Key, Iv, err := secure.BuildBody(agentInfo, body)
	if err != nil {
		return nil, nil, err
	}
	headers = secure.BuildHeaderForForward(agentInfo, uri, Key, Iv)
	return body, headers, nil
}
