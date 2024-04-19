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
	"fmt"

	"golang.org/x/text/language"
)

// OcsAgentError defines ocsagent error and implements error interface.
type OcsAgentError struct {
	ErrorCode ErrorCode     // error code
	Args      []interface{} // args for error message formatting
}

// Message will return error message composed of errorcode and args.
func (e OcsAgentError) Message(lang language.Tag) string {
	return GetMessage(lang, e.ErrorCode, e.Args)
}

// DefaultMessage will return err default message.
func (e OcsAgentError) DefaultMessage() string {
	return e.Message(defaultLanguage)
}

// Error will return error string.
func (e OcsAgentError) Error() string {
	return fmt.Sprintf("OcsAgentError: code = %d, message = %s", e.ErrorCode.Code, e.DefaultMessage())
}

// Occur returns *OcsAgentError composed of errorcode and args
func Occur(errorCode ErrorCode, args ...interface{}) *OcsAgentError {
	err := &OcsAgentError{
		ErrorCode: errorCode,
		Args:      args,
	}
	return err
}

// Occurf formats according to a format specifier (The first one of the `args`)
// and returns the resulting string as a value that satisfies `OcsAgentError.Args`.
func Occurf(errorCode ErrorCode, format string, args ...interface{}) *OcsAgentError {
	err := &OcsAgentError{
		ErrorCode: errorCode,
		Args:      []interface{}{fmt.Sprintf(format, args...)},
	}
	return err
}
