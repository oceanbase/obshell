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
	"bytes"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	ocshttp "github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/lib/trace"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
)

type UNIX_CONNECT_TYPE string

const UNIX_CONNECT UNIX_CONNECT_TYPE = "unix_conn"

const (
	statusURI = constant.URI_API_V1 + constant.URI_STATUS

	localRouteKey = "localRoute"
	apiRouteKey   = "apiRoute"

	originalBody = "ORIGINAL_BODY"
)

var (
	emptyRe          = regexp.MustCompile(`\s+`)
	recoveryResponse = ocshttp.BuildResponse(nil, errors.Occur(errors.ErrUnexpected, "Internal Server Error"))
)

func SetLocalRouteFlag(c *gin.Context) {
	ctx := NewContextWithTraceId(c)
	log.WithContext(ctx).Debug("set local route flag")
	c.Set(localRouteKey, true)
	c.Next()
}

func SetApiFlag(c *gin.Context) {
	ctx := NewContextWithTraceId(c)
	log.WithContext(ctx).Debug("set api flag")
	c.Set(apiRouteKey, true)
	c.Next()
}

// UnixSocketMiddleware creates a Gin middleware to enforce authorization for requests coming from Unix domain sockets.
func UnixSocketMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		r := c.Request
		var err error
		peerCred := getPeerCred(r) // Obtain the Unix user credentials from the socket.
		if peerCred != nil {
			userId := peerCred.Uid

			// Attempt to obtain the UID we want to compare against.
			// This can be taken from the observer process, the ownership of the OB etc directory, or the current process.
			var compareUid uint32
			if pidStr, _ := process.GetObserverPid(); pidStr != "" {
				pid, _ := strconv.Atoi(pidStr)
				compareUid, err = process.GetUidFromPid(pid)
			} else if path.IsEtcDirExist() {
				compareUid, err = path.EtcDirOwnerUid()
			} else {
				compareUid = process.Uid()
			}
			if err != nil {
				log.WithContext(NewContextWithTraceId(c)).Errorf("get uid failed, err: %v", err)
				resp := ocshttp.BuildResponse(nil, errors.Occur(errors.ErrUserPermissionDenied))
				c.JSON(resp.Status, resp)
				c.Abort()
			}

			if userId == compareUid || userId == 0 {
				c.Next()
				return
			}
		}

		// If authorization fails or credentials are not available, respond with 'permission denied'.
		resp := ocshttp.BuildResponse(nil, errors.Occur(errors.ErrUserPermissionDenied))
		c.JSON(resp.Status, resp)
		c.Abort()
	}
}

// getPeerCred retrieves the Unix credentials (UID, GID, PID) of the peer process of a Unix Domain Socket connection.
// It uses the 'SO_PEERCRED' socket option to get the credentials from an HTTP request object.
func getPeerCred(r *http.Request) (ucred *syscall.Ucred) {
	iconn := r.Context().Value(UNIX_CONNECT)
	if iconn == nil {
		return
	}

	conn, ok := iconn.(net.Conn)
	if !ok {
		return
	}

	rawConn, err := conn.(*net.UnixConn).SyscallConn()
	if err != nil {
		return
	}

	err = rawConn.Control(func(fd uintptr) {
		var err error
		ucred, err = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err != nil {
			return
		}
	})
	if err != nil {
		return
	}
	return ucred
}

func compareRequestUri(c *gin.Context, maskUri string) bool {
	requestUriArr := strings.Split(c.Request.RequestURI, "/")
	maskUriArr := strings.Split(maskUri, "/")
	if len(requestUriArr) != len(maskUriArr) {
		return false
	}
	for i := range requestUriArr {
		if requestUriArr[i] == maskUriArr[i] {
			continue
		}
		if strings.HasPrefix(maskUriArr[i], ":") {
			continue
		}
		return false
	}
	return true
}

// PreHandlers returns a Gin middleware function to extract and log
// trace IDs from incoming HTTP requests, and to log request details.
func PreHandlers(maskBodyRoutes ...string) func(*gin.Context) {
	return func(c *gin.Context) {
		if c.Request.RequestURI == statusURI {
			c.Next()
			return
		}
		traceId := trace.GetTraceId(c.Request)
		c.Set(TraceIdKey, traceId)

		ctx := NewContextWithTraceId(c)

		masked := false
		for _, it := range maskBodyRoutes {
			if compareRequestUri(c, it) {
				masked = true
				break
			}
		}
		if masked {
			log.WithContext(ctx).Infof("API request: [%v %v, client=%v, traceId=%v]",
				c.Request.Method, c.Request.URL, c.ClientIP(), traceId)
		} else {
			body := readRequestBody(c)
			log.WithContext(ctx).Infof("API request: [%v %v, client=%v, traceId=%v, body=%v]",
				c.Request.Method, c.Request.URL, c.ClientIP(), traceId, body)
		}

		c.Next()
	}
}

func readRequestBody(c *gin.Context) string {
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return emptyRe.ReplaceAllString(string(body), "")
}

func getOcsResponseFromContext(c *gin.Context) ocshttp.OcsAgentResponse {
	ctx := NewContextWithTraceId(c)

	if len(c.Errors) > 0 {
		var subErrors []interface{}
		for _, e := range c.Errors {
			switch e.Type {
			case gin.ErrorTypeBind:
				return ocshttp.BuildResponse(nil, errors.Occur(errors.ErrIllegalArgument, e.Err))
			default:
				subErrors = append(subErrors, ocshttp.ApiUnknownError{Error: e.Err})
			}
		}
		return ocshttp.NewSubErrorsResponse(subErrors)
	}

	if r, ok := c.Get(OcsAgentResponseKey); ok {
		if resp, ok := r.(ocshttp.OcsAgentResponse); ok {
			return resp
		}
	}
	log.WithContext(ctx).Error("ocsagent found no response object from gin context")
	return ocshttp.BuildResponse(nil, errors.Occur(errors.ErrUnexpected, "ocsagent cannot build response body"))
}

func PaddingBody() func(*gin.Context) {
	return func(c *gin.Context) {
		if c.Request.ContentLength == 0 {
			c.Request.Body = io.NopCloser(strings.NewReader("{}"))
			c.Request.ContentLength = 2
		}

		c.Next()
	}
}

// PostHandlers returns a Gin middleware function that logs the response and duration of API requests.
func PostHandlers(excludeRoutes ...string) func(*gin.Context) {
	return func(c *gin.Context) {
		for _, it := range excludeRoutes {
			if strings.HasPrefix(c.Request.RequestURI, it) {
				c.Next()
				return
			}
		}

		startTime := time.Now()
		c.Set(constant.REQUEST_RECEIVED_TIME, startTime.Unix())

		c.Next()

		if _, ok := c.Get(needForwardedFlag); ok {
			return
		}

		ctx := NewContextWithTraceId(c)
		resp := getOcsResponseFromContext(c)

		resp.Duration = time.Since(startTime).Milliseconds()

		if v, ok := c.Get(TraceIdKey); ok {
			if traceId, ok := v.(string); ok {
				resp.TraceId = traceId
			}
		}

		if resp.Successful {
			if c.Request.RequestURI != statusURI {
				log.WithContext(ctx).Infof("API response OK: [%v %v, client=%v, traceId=%v, duration=%v, status=%v, data=%+v]",
					c.Request.Method, c.Request.URL, c.ClientIP(), resp.TraceId, resp.Duration, resp.Status, resp.Data)
			} else {
				log.WithContext(ctx).Debugf("API response OK: [%v %v, client=%v, traceId=%v, duration=%v, status=%v, data=%+v]",
					c.Request.Method, c.Request.URL, c.ClientIP(), resp.TraceId, resp.Duration, resp.Status, resp.Data)
			}
		} else {
			log.WithContext(ctx).Infof("API response error: [%v %v, client=%v, traceId=%v, duration=%v, status=%v,data=%+v, error=%v]",
				c.Request.Method, c.Request.URL, c.ClientIP(), resp.TraceId, resp.Duration, resp.Status, resp.Data, resp.Error.String())
		}
		c.JSON(resp.Status, resp)
	}
}

// Recovery is a utility function meant to be used with the Gin middleware for panic recovery.
func Recovery(c *gin.Context, err interface{}) {
	log.WithContext(NewContextWithTraceId(c)).Errorf("request context %+v, err:%+v", c, err)
	c.JSON(recoveryResponse.Status, recoveryResponse)
}

// SetContentType is a Gin middleware that forces the Content-Type of all responses to "application/json".
// This is particularly useful to handle cases where gin might set the Content-Type to "text/plain"
// due to certain types of errors, such as JSON binding errors.
func SetContentType(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")

	c.Next()
}

func BodyDecrypt(skipRoutes ...string) func(*gin.Context) {
	return func(c *gin.Context) {
		for _, route := range skipRoutes {
			if route == c.Request.RequestURI {
				c.Next()
				return
			}
		}

		var err error
		if c.Request.Header.Get(constant.OCS_HEADER) == "" && c.Request.Header.Get(constant.OCS_AGENT_HEADER) == "" {
			c.Next()
			return
		}
		var header secure.HttpHeader
		if c.Request.Header.Get(constant.OCS_AGENT_HEADER) != "" {
			header, err = secure.DecryptHeader(c.Request.Header.Get(constant.OCS_AGENT_HEADER))
			if err != nil {
				log.WithContext(NewContextWithTraceId(c)).Errorf("header decrypt failed, err: %v", err)
				c.Abort()
				SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, "header decrypt failed"))
				return
			}
			c.Set(constant.OCS_AGENT_HEADER, header)
		} else {
			header, err = secure.DecryptHeader(c.Request.Header.Get(constant.OCS_HEADER))
			if err != nil {
				log.WithContext(NewContextWithTraceId(c)).Errorf("header decrypt failed, err: %v", err)
				c.Abort()
				SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, "header decrypt failed"))
				return
			}
			c.Set(constant.OCS_HEADER, header)
		}

		for _, route := range secure.GetSkipBodyEncryptRoutes() {
			if route == c.Request.RequestURI {
				c.Next()
				return
			}
		}

		// Decrypts the request body on routes where encryption is expected.
		encryptedBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.WithContext(NewContextWithTraceId(c)).Errorf("read body failed, err: %v", err)
			c.Abort()
			SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, "read body failed"))
			return
		}
		if len(encryptedBody) == 0 {
			c.Next()
			return
		}
		body, err := secure.BodyDecrypt(encryptedBody, string(header.Keys))
		if err != nil {
			log.WithContext(NewContextWithTraceId(c)).Errorf("body decrypt failed, err: %v", err)
			c.Abort()
			SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, "body decrypt failed"))
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		c.Set(originalBody, string(encryptedBody))
		c.Next()
	}
}

func VerifyObRouters(c *gin.Context, curTs int64, header *secure.HttpHeader, passwordType secure.VerifyType) {
	pass := false
	var err error
	switch meta.OCS_AGENT.GetIdentity() {
	case meta.SINGLE:
		if err = secure.VerifyToken(header.Token); err == nil {
			pass = true
			break
		}
		if meta.AGENT_PWD.Inited() {
			if passwordType == secure.AGENT_PASSWORD {
				if err = secure.VerifyAuth(header.Auth, header.Ts, curTs, secure.AGENT_PASSWORD); err == nil {
					pass = true
				} else {
					decryptAgentPassword, err1 := secure.Decrypt(header.Auth)
					if err1 == nil && secure.VerifyAuth(decryptAgentPassword, header.Ts, curTs, secure.AGENT_PASSWORD) == nil {
						pass = true
					}
				}
			} else {
				err = errors.New("agent password has been set, use agent password to verify")
			}
		} else {
			pass = true
		}
	case meta.FOLLOWER:
		// Follower verify token only.
		if err = secure.VerifyToken(header.Token); err == nil {
			pass = true
		} else {
			if IsApiRoute(c) && header.ForwardType != secure.ManualForward {
				// If the request is api and is not manual forwarded, auto forward it.
				autoForward(c)
				c.Abort()
				return
			}
		}
	case meta.MASTER:
		if header.ForwardType == secure.ManualForward {
			// When a request is manually forwarded, it must have a valid follower token.
			if err = secure.VerifyTokenByAgentInfo(header.Token, header.ForwardAgent); err == nil {
				pass = true
			}
			break
		} else if header.ForwardType == secure.AutoForward {
			// If the request is auto-forwarded, set IsAutoForwardedFlag to true for parse password.
			c.Set(IsAutoForwardedFlag, true)
			c.Set(FollowerAgentOfForward, header.ForwardAgent)
		}
		fallthrough
	default:
		if passwordType == secure.OCEANBASE_PASSWORD {
			if !meta.OCEANBASE_PASSWORD_INITIALIZED && meta.AGENT_PWD.Inited() {
				err = errors.New("oceanbase password is not initialized, use agent password to verify")
			} else {
				if err = secure.VerifyAuth(header.Auth, header.Ts, curTs, secure.OCEANBASE_PASSWORD); err == nil {
					pass = true
				}
			}
		} else {
			if meta.OCEANBASE_PASSWORD_INITIALIZED && !meta.AGENT_PWD.Inited() {
				err = errors.New("agent password is not initialized, use oceanbase password to verify")
			} else if !meta.AGENT_PWD.Inited() {
				pass = true
			} else if err = secure.VerifyAuth(header.Auth, header.Ts, curTs, secure.AGENT_PASSWORD); err == nil {
				pass = true
			}
		}
	}
	if !pass {
		log.WithContext(NewContextWithTraceId(c)).Errorf("Security verification failed: %s", err.Error())
		c.Abort()
		SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, err.Error()))
		return
	}
}

func VerifyForSetAgentPassword(c *gin.Context, curTs int64, header *secure.HttpHeader, passwordType secure.VerifyType) {
	pass := false
	var err error
	if meta.AGENT_PWD.Inited() {
		if passwordType == secure.AGENT_PASSWORD {
			if err = secure.VerifyAuth(header.Auth, header.Ts, curTs, secure.AGENT_PASSWORD); err == nil {
				pass = true
			}
		} else {
			err = errors.New("agent password has been set, use agent password to verify")
		}
	} else if meta.OCS_AGENT.IsClusterAgent() {
		if passwordType == secure.OCEANBASE_PASSWORD {
			if err = secure.VerifyAuth(header.Auth, header.Ts, curTs, secure.OCEANBASE_PASSWORD); err == nil {
				pass = true
			}
		} else {
			err = errors.New("oceanbase password has been set, use oceanbase password to verify")
		}
	} else if meta.OCS_AGENT.IsSingleAgent() {
		pass = true
	}
	if !pass {
		log.WithContext(NewContextWithTraceId(c)).Errorf("Security verification failed: %s", err.Error())
		c.Abort()
		SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, err.Error()))
		return
	}
}

func VerifyAgentRoutes(c *gin.Context, curTs int64, header *secure.HttpHeader, passwordType secure.VerifyType) {
	pass := false
	var err error
	if passwordType != secure.AGENT_PASSWORD {
		err = errors.New("Please use agent password to verify")
	} else {
		if meta.AGENT_PWD.Inited() {
			if err = secure.VerifyAuth(header.Auth, header.Ts, curTs, secure.AGENT_PASSWORD); err == nil {
				pass = true
			}
		} else {
			err = errors.New("agent password is not initialized")
		}
	}
	if !pass {
		log.WithContext(NewContextWithTraceId(c)).Errorf("Security verification failed: %s", err.Error())
		c.Abort()
		SendResponse(c, nil, errors.Occurf(errors.ErrUnauthorized, err.Error()))
		return
	}
}

func VerifyTaskRoutes(c *gin.Context, curTs int64, header *secure.HttpHeader, passwordType secure.VerifyType) {
	id := c.Param("id")
	if id == "" {
		VerifyObRouters(c, curTs, header, passwordType)
		return
	}
	if task.IsObproxyTask(id) {
		VerifyAgentRoutes(c, curTs, header, passwordType)
		return
	} else {
		VerifyObRouters(c, curTs, header, passwordType)
		return
	}
}

func Verify(routeType ...secure.RouteType) func(*gin.Context) {
	return func(c *gin.Context) {
		log.WithContext(NewContextWithTraceId(c)).Infof("verfiy request: %s", c.Request.RequestURI)
		var header secure.HttpHeader
		obHeaderByte, _ := c.Get(constant.OCS_HEADER)
		agentHeaderByte, _ := c.Get(constant.OCS_AGENT_HEADER)
		var headerByte any
		var passwordType secure.VerifyType
		if agentHeaderByte != nil {
			passwordType = secure.AGENT_PASSWORD
			headerByte = agentHeaderByte
		} else if obHeaderByte != nil {
			passwordType = secure.OCEANBASE_PASSWORD
			headerByte = obHeaderByte
		} else {
			log.WithContext(NewContextWithTraceId(c)).Error("header not found")
			c.Abort()
			SendResponse(c, nil, errors.Occur(errors.ErrUnauthorized, "header not found"))
			return
		}
		if passwordType != secure.AGENT_PASSWORD && len(routeType) != 0 && routeType[0] == secure.ROUTE_OBPROXY {
			log.WithContext(NewContextWithTraceId(c)).Error("agent header not found")
			c.Abort()
			SendResponse(c, nil, errors.Occur(errors.ErrUnauthorized, "aegnt header not found"))
			return
		}

		header, ok := headerByte.(secure.HttpHeader)
		if !ok {
			log.WithContext(NewContextWithTraceId(c)).Error("header type error")
			c.Abort()
			SendResponse(c, nil, errors.Occur(errors.ErrUnauthorized, "header type error"))
			return
		}

		// Verify the URI in the header matches the URI of the request.
		if header.Uri != c.Request.RequestURI {
			log.WithContext(NewContextWithTraceId(c)).Errorf("verify uri failed, uri: %s, expect: %s", header.Uri, c.Request.RequestURI)
			authErr := errors.Occurf(errors.ErrUnauthorized, "uri mismatch")
			c.Abort()
			SendResponse(c, nil, authErr)
			return
		}

		curTs := time.Now().Unix()
		if r, ok := c.Get(constant.REQUEST_RECEIVED_TIME); ok {
			if receivedTs, ok := r.(int64); ok {
				curTs = receivedTs
			}
		}

		if c.Request.RequestURI == constant.URI_AGENT_API_PREFIX+constant.URI_PASSWORD {
			VerifyForSetAgentPassword(c, curTs, &header, passwordType)
		} else if len(routeType) != 0 && routeType[0] == secure.ROUTE_OBPROXY {
			VerifyAgentRoutes(c, curTs, &header, passwordType)
		} else if len(routeType) != 0 && routeType[0] == secure.ROUTE_TASK {
			VerifyTaskRoutes(c, curTs, &header, passwordType)
		} else {
			VerifyObRouters(c, curTs, &header, passwordType)
		}
		// Verification succeeded, continue to the next middleware.
		c.Next()
	}
}
