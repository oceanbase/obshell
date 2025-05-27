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

package secure

import (
	"github.com/go-resty/resty/v2"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/meta"
)

var skipBodyEncryptRoutes = []string{}

func AddSkipBodyEncryptRoutes(routes ...string) {
	skipBodyEncryptRoutes = append(skipBodyEncryptRoutes, routes...)
}

func GetSkipBodyEncryptRoutes() []string {
	return skipBodyEncryptRoutes
}

// SendGetRequest will send http get request to the agent.
// If ret is not nil, it should be a pointer.
func SendGetRequest(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) error {
	return sendRequestAndBuildReturn(agentInfo, uri, http.GET, param, ret)
}

// SendPutRequest will send http put request to the agent.
// If ret is not nil, it should be a pointer.
func SendPutRequest(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) error {
	return sendRequestAndBuildReturn(agentInfo, uri, http.PUT, param, ret)
}

// SendPostRequest will send http post request to the agent.
// If ret is not nil, it should be a pointer.
func SendPostRequest(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) error {
	return sendRequestAndBuildReturn(agentInfo, uri, http.POST, param, ret)
}

// SendPatchRequest will send http patch request to the agent.
// If ret is not nil, it should be a pointer.
func SendPatchRequest(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) error {
	return sendRequestAndBuildReturn(agentInfo, uri, http.PATCH, param, ret)
}

// SendDeleteRequest will send http delete request to the agent.
// If ret is not nil, it should be a pointer.
func SendDeleteRequest(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) error {
	return sendRequestAndBuildReturn(agentInfo, uri, http.DELETE, param, ret)
}

// SendGetRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendGetRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, http.GET, param, ret)
}

// SendPutRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendPutRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, http.PUT, param, ret)
}

// SendPostRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendPostRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, http.POST, param, ret)
}

// SendPatchRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendPatchRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, http.PATCH, param, ret)
}

// SendDeleteRequestAndReturnResponse will return http response and error.
// If ret is not nil, it should be a pointer.
func SendDeleteRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, param interface{}, ret interface{}) (*resty.Response, error) {
	return SendRequestAndReturnResponse(agentInfo, uri, http.DELETE, param, ret)
}

func SendRequestAndReturnResponse(agentInfo meta.AgentInfoInterface, uri string, method string, param interface{}, ret interface{}) (*resty.Response, error) {
	for _, route := range skipBodyEncryptRoutes {
		if route == uri {
			return http.SendRequestAndReturnResponse(agentInfo, uri, method, param, ret, BuildHeader(agentInfo, uri, false))
		}
	}
	encryptedBody, header, err := BuildBodyAndHeader(agentInfo, uri, param)
	if err != nil {
		return nil, err
	}
	return http.SendRequestAndReturnResponse(agentInfo, uri, method, encryptedBody, ret, header)
}

func sendRequestAndBuildReturn(agentInfo meta.AgentInfoInterface, uri string, method string, param interface{}, ret interface{}) error {
	for _, route := range skipBodyEncryptRoutes {
		if route == uri {
			return http.SendRequestAndBuildReturn(agentInfo, uri, method, param, ret, BuildHeader(agentInfo, uri, false))
		}
	}
	encryptedBody, header, err := BuildBodyAndHeader(agentInfo, uri, param)
	if err != nil {
		return err
	}
	return http.SendRequestAndBuildReturn(agentInfo, uri, method, encryptedBody, ret, header)
}

func BuildBodyAndHeader(agentInfo meta.AgentInfoInterface, uri string, param interface{}) (encryptedBody interface{}, header map[string]string, err error) {
	encryptedBody, Key, Iv, err := BuildBody(agentInfo, param)
	if err != nil {
		return nil, nil, errors.Wrap(err, "build body failed")
	}
	header = BuildHeader(agentInfo, uri, false, Key, Iv)
	return encryptedBody, header, nil
}

func BuildBody(agentInfo meta.AgentInfoInterface, param interface{}) (encryptedBody interface{}, Key, Iv []byte, err error) {
	encryptedBody, Key, Iv, err = nil, nil, nil, nil
	if encryptMethod == "rsa" {
		encryptedBody, err = EncryptBodyWithRsa(agentInfo, param)
	} else if encryptMethod == "aes" {
		encryptedBody, Key, Iv, err = EncryptBodyWithAes(param)
	} else if encryptMethod == "sm4" {
		encryptedBody, Key, Iv, err = EncryptBodyWithSm4(param)
	}
	return
}

func BuildHeaderForForward(agentInfo meta.AgentInfoInterface, uri string, keys ...[]byte) map[string]string {
	return BuildHeader(agentInfo, uri, true, keys...)
}

func SendRequestWithPassword(agentInfo meta.AgentInfoInterface, uri string, method string, agentPassword string, param interface{}, ret interface{}) error {
	encryptedBody, Key, Iv, err := BuildBody(agentInfo, param)
	if err != nil {
		return errors.Wrap(err, "build body failed")
	}
	header := BuildAgentHeader(agentInfo, agentPassword, uri, false, Key, Iv)
	return http.SendRequestAndBuildReturn(agentInfo, uri, method, encryptedBody, ret, header)
}
