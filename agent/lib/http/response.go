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
	"fmt"
	"reflect"
	"time"

	"github.com/oceanbase/obshell/agent/errors"
)

// OcsAgentResponseData is the response data struct of ocsagent.
type OcsAgentResponseData struct {
	TaskId int
	// TaskStage int
}

// OcsAgentResponse is the response struct of ocsagent.
type OcsAgentResponse struct {
	Successful bool        `json:"successful"`      // Whether request successful or not
	Timestamp  time.Time   `json:"timestamp"`       // Request handling timestamp (server time)
	Duration   int64       `json:"duration"`        // Request handling time cost (ms)
	Status     int         `json:"status"`          // HTTP status code
	TraceId    string      `json:"traceId"`         // Request trace ID, contained in server logs
	Data       interface{} `json:"data,omitempty"`  // Data payload when response is successful
	Error      *ApiError   `json:"error,omitempty"` // Error payload when response is failed
}

// ApiError is the api error struct of ocsagent.
type ApiError struct {
	Code      int           `json:"code"`                // Error code
	Message   string        `json:"message"`             // Error message
	SubErrors []interface{} `json:"subErrors,omitempty"` // Sub errors
}

// ApiFieldError is the api field error struct of ocsagent.
type ApiFieldError struct {
	Tag     string `json:"tag"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (a ApiError) String() string {
	if len(a.SubErrors) == 0 {
		return fmt.Sprintf("{Code:%v, Message:%v}", a.Code, a.Message)
	} else {
		return fmt.Sprintf("{Code:%v, Message:%v, SubErrors:%+v}", a.Code, a.Message, a.SubErrors)
	}
}

type IterableData struct {
	Contents interface{} `json:"contents"`
}

type ApiUnknownError struct {
	Error error `json:"error"`
}

func BuildResponse(data interface{}, err error) OcsAgentResponse {
	response := buildSuccessResponse()
	response.buildResponseData(data)
	if err != nil {
		response = buildErrorResponse(err, response)
	}
	return *response
}

func buildErrorResponse(err error, response *OcsAgentResponse) *OcsAgentResponse {
	agenterr, ok := err.(*errors.OcsAgentError)
	if !ok && err != nil {
		agenterr = errors.Occur(errors.ErrUnexpected, err)
	}
	if agenterr != nil {
		response.Successful = false
		response.Status = agenterr.ErrorCode.Kind
		response.Error = &ApiError{
			Code:    agenterr.ErrorCode.Code,
			Message: agenterr.DefaultMessage()}
	}
	return response
}

func buildSuccessResponse() *OcsAgentResponse {
	return &OcsAgentResponse{
		Successful: true,
		Timestamp:  time.Now(),
		Status:     200,
		Error:      nil,
	}
}

func (response *OcsAgentResponse) buildResponseData(data interface{}) {
	if data != nil && reflect.TypeOf(data).Kind() == reflect.Slice {
		iterableData := IterableData{Contents: data}
		response.Data = iterableData
	} else {
		response.Data = data
	}
}

func NewSubErrorsResponse(subErrors []interface{}) OcsAgentResponse {
	allValidationError := true
	for _, subError := range subErrors {
		if _, ok := subError.(ApiFieldError); !ok {
			allValidationError = false
		}
	}

	var status int
	var code int
	var message string
	if allValidationError {
		status = errors.ErrBadRequest.Kind
		code = errors.ErrBadRequest.Code
		message = "validation error"
	} else {
		status = errors.ErrUnexpected.Kind
		code = errors.ErrUnexpected.Code
		message = "unhandled error"
	}

	return OcsAgentResponse{
		Successful: false,
		Timestamp:  time.Now(),
		Status:     status,
		Data:       nil,
		Error: &ApiError{
			Code:      code,
			Message:   message,
			SubErrors: subErrors,
		},
	}
}
