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

package utils

import (
	"regexp"
	"strings"
)

var validParamKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ShellQuote wraps a string in single quotes with proper escaping for safe
// use in shell commands. Single quotes inside the string are escaped as '\''.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// IsValidParamKey checks whether a parameter key contains only safe characters
// (letters, digits, and underscores).
func IsValidParamKey(key string) bool {
	return validParamKeyRegex.MatchString(key)
}
