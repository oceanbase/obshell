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
	"errors"
	"fmt"

	obdriver "github.com/oceanbase/go-oceanbase-driver"
	"golang.org/x/text/language"
)

type OcsAgentErrorInterface interface {
	ErrorCode() ErrorCode
	Error() string
	ErrorMessage() string
	LocaleMessage(lang language.Tag) string
}

// OcsAgentError defines ocsagent error and implements error interface.
type OcsAgentError struct {
	errorCode ErrorCode     // error code
	args      []interface{} // args for error message formatting
}

type OcsAgentErrorWrapper struct {
	cause error
	code  ErrorCode
	OcsAgentError
}

// Message will return error message composed of errorcode and args.
func (e OcsAgentError) message(lang language.Tag) string {
	return GetMessage(lang, e.errorCode.key, e.args)
}

// DefaultMessage will return err default message.
func (e OcsAgentError) Message() string {
	return e.message(defaultLanguage)
}

func (e OcsAgentError) Error() string {
	return e.Message()
}

func (e OcsAgentError) ErrorCode() ErrorCode {
	return e.errorCode
}

func (e OcsAgentError) Args() []interface{} {
	return e.args
}

func (e OcsAgentError) ErrorMessage() string {
	return fmt.Sprintf("[%s]: %s", e.errorCode.Code, e.Error())
}

func (e OcsAgentError) LocaleMessage(lang language.Tag) string {
	return GetMessage(lang, e.errorCode.key, e.args)
}

func (e OcsAgentErrorWrapper) ErrorCode() ErrorCode {
	return e.code
}

func (e OcsAgentErrorWrapper) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.OcsAgentError.Error(), e.cause.Error())
	}
	return e.OcsAgentError.Error()
}

func (e OcsAgentErrorWrapper) ErrorMessage() string {
	return fmt.Sprintf("[%s]: %s", e.code.Code, e.Error())
}

func (e OcsAgentErrorWrapper) Args() []interface{} {
	return e.OcsAgentError.Args()
}

func (e OcsAgentErrorWrapper) Unwrap() error {
	return e.cause
}

func (e OcsAgentErrorWrapper) LocaleMessage(lang language.Tag) string {
	if e.cause != nil {
		if ocsAgentError, ok := e.cause.(OcsAgentErrorInterface); ok {
			return fmt.Sprintf("%s: %s", e.OcsAgentError.LocaleMessage(lang), ocsAgentError.LocaleMessage(lang))
		} else {
			return fmt.Sprintf("%s: %s", e.OcsAgentError.LocaleMessage(lang), e.cause.Error())
		}
	}
	return e.OcsAgentError.LocaleMessage(lang)
}

// OccurWithError returns *OcsAgentError composed of errorcode and error message.
func OccurWithMessage(message string, errorCode ErrorCode, args ...interface{}) *OcsAgentErrorWrapper {
	return &OcsAgentErrorWrapper{
		cause: errors.New(message),
		code:  errorCode,
		OcsAgentError: OcsAgentError{
			errorCode: errorCode,
			args:      args,
		},
	}
}

// Occur returns *OcsAgentError composed of errorcode and args
func Occur(errorCode ErrorCode, args ...interface{}) *OcsAgentError {
	return &OcsAgentError{
		errorCode: errorCode,
		args:      args,
	}
}

// Occurf formats according to a format specifier (The first one of the `args`)
// and returns the resulting string as a value that satisfies `OcsAgentError.Args`.
func Occurf(errorCode ErrorCode, format string, args ...interface{}) *OcsAgentError {
	return Occur(errorCode, fmt.Sprintf(format, args...))
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func IsMysqlError(err error) bool {
	if _, ok := err.(*obdriver.MySQLError); ok {
		return true
	}
	if x, ok := err.(interface{ Unwrap() error }); ok {
		return IsMysqlError(x.Unwrap())
	}
	return false
}
