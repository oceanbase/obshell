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

package http

import (
	"net"
	"net/http"
	"os"
	"time"

	resty "github.com/go-resty/resty/v2"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/json"
)

const UNIX_SOCKET_DEFAULT_TIME_OUT = 30 * time.Second

func SendGetRequestViaUnixSocket(socketPath string, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturnViaUnixSocket(socketPath, uri, GET, param, ret, nil)
}

func SendPutRequestViaUnixSocket(socketPath string, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturnViaUnixSocket(socketPath, uri, PUT, param, ret, nil)
}

func SendPostRequestViaUnixSocket(socketPath string, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturnViaUnixSocket(socketPath, uri, POST, param, ret, nil)
}

func SendPatchRequestViaUnixSocket(socketPath string, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturnViaUnixSocket(socketPath, uri, PATCH, param, ret, nil)
}

func SendDeleteRequestViaUnixSocket(socketPath string, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturnViaUnixSocket(socketPath, uri, DELETE, param, ret, nil)
}

func SendRequestAndReturnResponseViaUnixSocket(socketPath string, uri string, method string, param, ret interface{}, headers map[string]string) (*resty.Response, error) {
	response, err := sendRequestViaUnixSocket(socketPath, uri, method, param, ret, nil)
	if err != nil {
		return response.response, err
	}
	err = buildReturn(response, ret)
	if err != nil {
		return response.response, err
	}
	return response.response, nil
}

func SendRequestAndBuildReturnViaUnixSocket(socketPath, uri, method string, param, ret interface{}, headers map[string]string) error {
	response, err := sendRequestViaUnixSocket(socketPath, uri, method, param, ret, headers)
	if err != nil {
		return err
	}
	return buildRequestReturnForUnixSocket(response, ret)
}

func UploadFileViaUnixSocket(socketPath, uri, filePath string, ret interface{}) (err error) {
	var agentResp OcsAgentResponse
	var response *resty.Response
	transport := http.Transport{
		DisableKeepAlives: true,
		Dial: func(_, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	client := resty.New()
	client.SetTransport(&transport).SetScheme("http").SetBaseURL("localhost")
	client.JSONUnmarshal = json.Unmarshal
	request := client.R()
	if ret != nil {
		request.SetResult(&agentResp)
	}
	request.SetError(&agentResp)

	response, err = request.SetFile("file", filePath).Post(uri)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}

	return buildRequestReturnForUnixSocket(ocsAgentResponse{
		response:  response,
		agentResp: agentResp,
	}, ret)
}

// buildRequestReturnForUnixSocket will parses the return value, if specified.
// If the response is wrong, response-related error will be returned.
// If an error occurred in the parsing process, parsing-related error will be returned only if the request err is nil.
func buildRequestReturnForUnixSocket(agentResponse ocsAgentResponse, ret interface{}) error {
	if agentResponse.response.IsError() {
		return agentResponse.agentResp.Error
	}
	buildRetErr := buildReturn(agentResponse, ret)
	return buildRetErr
}

func sendRequestViaUnixSocket(socketPath, uri, method string, param interface{}, ret interface{}, headers map[string]string) (agentResponse ocsAgentResponse, err error) {
	var agentResp OcsAgentResponse
	var response *resty.Response
	transport := http.Transport{
		DisableKeepAlives: true,
		Dial: func(_, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	client := resty.New()
	client.SetTransport(&transport).SetScheme("http").SetBaseURL("localhost").SetTimeout(UNIX_SOCKET_DEFAULT_TIME_OUT)
	client.JSONUnmarshal = json.Unmarshal
	request := client.R()
	if ret != nil {
		request.SetResult(&agentResp)
	}
	request.
		SetHeader("Content-Type", "application/json").
		SetError(&agentResp)

	if method != GET {
		request.SetBody(param)
	}
	if method == GET {
		query_params, ok := param.(map[string]string)
		if ok {
			for k, v := range query_params {
				request.SetQueryParam(k, v)
			}
		}
	}

	for k, v := range headers {
		request.SetHeader(k, v)
	}

	switch method {
	case GET:
		response, err = request.Get(uri)
	case PUT:
		response, err = request.Put(uri)
	case POST:
		response, err = request.Post(uri)
	case PATCH:
		response, err = request.Patch(uri)
	case DELETE:
		response, err = request.Delete(uri)
	default:
		return agentResponse, errors.Occur(errors.ErrRequestMethodNotSupport, method)
	}
	return ocsAgentResponse{
		response:  response,
		agentResp: agentResp,
	}, err
}

func SocketIsActive(socketPath string) bool {
	if !IsSocketFile(socketPath) {
		return false
	}
	return SocketCanConnect("unix", socketPath, time.Second)
}

func SocketCanConnect(network, addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout(network, addr, timeout)
	if conn != nil {
		conn.Close()
	}
	return err == nil
}

func IsSocketFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && (stat.Mode()&os.ModeSocket != 0)
}
