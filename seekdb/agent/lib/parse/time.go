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

package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
)

const (
	TIME_SECOND = "S"
	TIME_MINUTE = "M"
	TIME_HOUR   = "H"
	TIME_DAY    = "D"
)

func TimeParse(input string) (int, error) {
	// Compile a regular expression to match the input format
	pattern := regexp.MustCompile(`^([0-9]+)([a-zA-Z]?)$`)
	matches := pattern.FindStringSubmatch(input)

	// Check if the input matches the pattern
	if matches == nil {
		return 0, errors.Occur(errors.ErrCommonInvalidTimeDuration, input, "invalid format")
	}

	// Convert the captured numeric part of the input to an integer
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, errors.Occur(errors.ErrCommonInvalidTimeDuration, input, "invalid number")
	}

	// Get the unit character (if any) and determine the conversion factor
	unit := matches[2]
	switch strings.ToUpper(unit) {
	case "":
		// Default unit is microseconds, so convert to seconds
		return num / 1000 / 1000, nil
	case TIME_SECOND:
		return num, nil
	case TIME_MINUTE:
		return num * 60, nil
	case TIME_HOUR:
		return num * 60 * 60, nil
	case TIME_DAY:
		return num * 24 * 60 * 60, nil
	default:
		return 0, errors.Occur(errors.ErrCommonInvalidTimeDuration, input, "invalid time unit")
	}
}

// ParseOBDateTime parses datetime strings returned by OceanBase (e.g. V$OB_SERVER_STAT),
// such as "2006-01-02 15:04:05.000000" or "2006-01-02 15:04:05".
// Returns zero time and nil error for empty string.
func ParseOBDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid ob datetime: %s", s)
}
