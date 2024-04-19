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

package errors

import "net/http"

type ErrorKind = int

const (
	badRequest      ErrorKind = http.StatusBadRequest
	illegalArgument ErrorKind = http.StatusBadRequest
	unauthorized    ErrorKind = http.StatusUnauthorized
	notFound        ErrorKind = http.StatusNotFound
	unexpected      ErrorKind = http.StatusInternalServerError
	known           ErrorKind = http.StatusInternalServerError
)

// ErrorCode includes code, kind and key.
type ErrorCode struct {
	Code int
	Kind ErrorKind
	key  string
}

var errorCodes []ErrorCode

// NewErrorCode will create a new ErrorCode and append it to errorCodes
func NewErrorCode(code int, kind ErrorKind, key string) ErrorCode {
	errorCode := ErrorCode{
		Code: code,
		Kind: kind,
		key:  key,
	}
	errorCodes = append(errorCodes, errorCode)
	return errorCode
}

var (
	// general error codes, range: 1000 ~ 1999
	ErrBadRequest      = NewErrorCode(1000, badRequest, "err.bad.request")
	ErrIllegalArgument = NewErrorCode(1001, illegalArgument, "err.illegal.argument")
	ErrUnexpected      = NewErrorCode(1002, unexpected, "err.unexpected")
	ErrKnown           = NewErrorCode(1010, known, "err.known")

	// ob operation error codes, range: 10000 ~ 10999
	ErrUserPermissionDenied = NewErrorCode(10000, unauthorized, "err.user.permission.denied")
	ErrUnauthorized         = NewErrorCode(10008, unauthorized, "err.unauthorized")
	ErrObclusterNotFound    = NewErrorCode(10009, known, "err.obcluster.not.found")

	// task error codes, range: 2300 ~ 2399
	ErrTaskNotFound     = NewErrorCode(2300, notFound, "err.task.not.found")
	ErrTaskCreateFailed = NewErrorCode(2301, known, "err.task.create.failed")
)
