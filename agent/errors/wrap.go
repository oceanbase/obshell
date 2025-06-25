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

import (
	"github.com/pkg/errors"
)

type OcsAgentErrorInterface interface {
	ErrorCode() ErrorCode
	Error() string
	ErrorMessage() string
}

// New returns an error with the supplied message.
func New(message string) error {
	return errors.New(message)
}

// Errorf returns an error with the supplied formatted message.
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

// Wrap returns an error annotating err with a stack trace
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	if ocsAgentErr, ok := err.(OcsAgentErrorInterface); ok {
		return OcsAgentErrorWrapper{
			cause: err,
			OcsAgentErrorExporter: OcsAgentErrorExporter{
				errorCode: ocsAgentErr.ErrorCode(),
				err:       errors.New(message),
			},
		}
	} else {
		return errors.Wrap(err, message)
	}
}

// Wrapf returns an error annotating err with a stack trace
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	if ocsAgentErr, ok := err.(OcsAgentErrorInterface); ok {
		return OcsAgentErrorWrapper{
			cause: err,
			OcsAgentErrorExporter: OcsAgentErrorExporter{
				errorCode: ocsAgentErr.ErrorCode(),
				err:       errors.Errorf(format, args...),
			},
		}
	} else {
		return errors.Wrapf(err, format, args...)
	}
}

// WrapRetain returns an error annotating err with a stack trace
// WrapRetain will retain the original errorCode if it's OcsAgentError
func WrapRetain(errorCode ErrorCode, err error, args ...interface{}) *OcsAgentErrorWrapper {
	if err == nil {
		return nil
	}
	newErr := &OcsAgentErrorWrapper{
		cause: err,
		OcsAgentErrorExporter: OcsAgentErrorExporter{
			errorCode: errorCode,
			err: ocsAgentError{
				ErrorCode: errorCode,
				Args:      args,
			},
		},
	}
	if originalErr, ok := err.(OcsAgentErrorInterface); ok && originalErr.ErrorCode().Code != ErrCommonUnexpected.Code {
		newErr.OcsAgentErrorExporter.SetErrorCode(originalErr.ErrorCode())
	}
	return newErr
}

// WrapOverride returns an error annotating err with a stack trace
// WrapOverride will override the error code if it's OcsAgentError
func WrapOverride(errorCode ErrorCode, err error, args ...interface{}) *OcsAgentErrorWrapper {
	if err == nil {
		return nil
	}
	newErr := &OcsAgentErrorWrapper{
		cause: err,
		OcsAgentErrorExporter: OcsAgentErrorExporter{
			errorCode: errorCode,
			err: ocsAgentError{
				ErrorCode: errorCode,
				Args:      args,
			},
		},
	}
	return newErr
}
