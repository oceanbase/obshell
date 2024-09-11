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
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"time"

	resty "github.com/go-resty/resty/v2"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/meta"
)

const (
	GET    = "GET"    // http get method
	PUT    = "PUT"    // http put method
	POST   = "POST"   // http post method
	PATCH  = "PATCH"  // http patch method
	DELETE = "DELETE" // http delete method

	TCP_DEFAULT_TIME_OUT = 3 * time.Minute // default timeout
)

// SendGetRequest will send http get request to the agent.
// If ret is not nil, it should be a pointer.
func SendGetRequest(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturn(agentInfo, uri, GET, param, ret, nil)
}

// SendPutRequest will send http put request to the agent.
// If ret is not nil, it should be a pointer.
func SendPutRequest(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturn(agentInfo, uri, PUT, param, ret, nil)
}

// SendPostRequest will send http post request to the agent.
// If ret is not nil, it should be a pointer.
func SendPostRequest(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturn(agentInfo, uri, POST, param, ret, nil)
}

// SendPatchRequest will send http patch request to the agent.
// If ret is not nil, it should be a pointer.
func SendPatchRequest(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturn(agentInfo, uri, PATCH, param, ret, nil)
}

// SendDeleteRequest will send http delete request to the agent.
// If ret is not nil, it should be a pointer.
func SendDeleteRequest(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) error {
	return SendRequestAndBuildReturn(agentInfo, uri, DELETE, param, ret, nil)
}

// SendGetRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendGetRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, GET, param, ret, nil)
}

// SendPutRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendPutRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, PUT, param, ret, nil)
}

// SendPostRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendPostRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, POST, param, ret, nil)
}

// SendPatchRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendPatchRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, PATCH, param, ret, nil)
}

// SendDeleteRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendDeleteRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, DELETE, param, ret, nil)
}

type ocsAgentResponse struct {
	response  *resty.Response  // http response
	agentResp OcsAgentResponse // ocsagent response
}

// SendRequestAndReturnResponse will return http response and error.
func SendRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, method string, param, ret interface{}, headers map[string]string) (*resty.Response, error) {
	ocsAgentResponse, err := sendHttpRequest(agentInfo, uri, method, param, ret, headers)
	if err != nil {
		return ocsAgentResponse.response, err
	}
	err = buildReturn(ocsAgentResponse, ret)
	if err != nil {
		return ocsAgentResponse.response, err
	}
	return ocsAgentResponse.response, nil
}

func SendRequestAndBuildReturn(agentInfo meta.AgentInfoInterface, uri string, method string, param, ret interface{}, headers map[string]string) error {
	ocsAgentResponse, err := sendHttpRequest(agentInfo, uri, method, param, ret, headers)
	if err != nil {
		return err
	}
	return buildRequestReturn(ocsAgentResponse, ret)
}

func NewClient() *resty.Client {
	client := resty.New().SetTimeout(TCP_DEFAULT_TIME_OUT)

	if global.EnableHTTPS {
		tlsConfig := &tls.Config{
			RootCAs:            global.CaCertPool,
			InsecureSkipVerify: global.SkipVerify,
		}
		client.SetTLSClientConfig(tlsConfig)
	}
	client.JSONUnmarshal = json.Unmarshal
	return client
}

// sendHttpRequest will execute the http request according to the type of the method,
// return ocsAgentResponse and the error occurred during sending the request
func sendHttpRequest(agentInfo meta.AgentInfoInterface, uri string, method string, param, ret interface{}, headers map[string]string) (agentResponse ocsAgentResponse, err error) {
	var agentResp OcsAgentResponse
	var response *resty.Response
	targetUrl := fmt.Sprintf("%s://%s:%d%s", global.Protocol, agentInfo.GetIp(), agentInfo.GetPort(), uri)
	request := NewClient().R()
	if ret != nil {
		request.SetResult(&agentResp)
	}
	request.
		SetHeader("Content-Type", "application/json").
		SetBody(param).
		SetError(&agentResp)

	for k, v := range headers {
		request.SetHeader(k, v)
	}

	switch method {
	case GET:
		response, err = request.Get(targetUrl)
	case PUT:
		response, err = request.Put(targetUrl)
	case POST:
		response, err = request.Post(targetUrl)
	case PATCH:
		response, err = request.Patch(targetUrl)
	case DELETE:
		response, err = request.Delete(targetUrl)
	default:
		return agentResponse, fmt.Errorf("%s method not support", method)
	}
	return ocsAgentResponse{
		response:  response,
		agentResp: agentResp,
	}, err
}

// buildRequestReturn will parses the return value, if specified.
// If the response is wrong, response-related error will be returned.
// If an error occurred in the parsing process, parsing-related error will be returned only if the request err is nil.
func buildRequestReturn(agentResponse ocsAgentResponse, ret interface{}) error {
	buildRetErr := buildReturn(agentResponse, ret)
	if agentResponse.response.IsError() {
		return fmt.Errorf("%s", agentResponse.agentResp.Error)
	}
	return buildRetErr
}

// buildReturn is used to deserialize the response Data into the specified Ret
func buildReturn(agentResponse ocsAgentResponse, ret interface{}) error {
	if ret != nil && agentResponse.response.StatusCode() != http.StatusNoContent {
		if agentResponse.agentResp.Data == nil {
			return errors.New("response data is nil")
		}
		responseMap, ok := agentResponse.agentResp.Data.(map[string]interface{})
		if !ok {
			return errors.New("response data is not map")
		}
		if len(responseMap) == 0 {
			return errors.New("response data is empty")
		}
		data, err := json.Marshal(responseMap)
		if err != nil {
			return err
		}
		if err = unmarshal(data, ret); err != nil {
			return err
		}
	}
	return nil
}

func unmarshal(data []byte, ret interface{}) error {
	if reflect.TypeOf(ret).Elem().Kind() == reflect.Slice {
		iterableData := IterableData{}
		err := json.Unmarshal(data, &iterableData)
		if err != nil {
			return err
		}
		data, err = json.Marshal(iterableData.Contents)
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(data, ret)
}
