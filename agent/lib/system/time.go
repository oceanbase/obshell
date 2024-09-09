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

package system

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeUnit string

const (
	Nanosecond  time.Duration = 1
	Microsecond               = 1000 * Nanosecond
	Millisecond               = 1000 * Microsecond
	Second                    = 1000 * Millisecond
	Minute                    = 60 * Second
	Hour                      = 60 * Minute
	Day                       = 24 * Hour

	TIME_UNIT_NANOSECOND  TimeUnit = "ns"
	TIME_UNIT_MICROSECOND TimeUnit = "us"
	TIME_UNIT_MILLISECOND TimeUnit = "ms"
	TIME_UNIT_SECOND      TimeUnit = "s"
	TIME_UNIT_MINUTE      TimeUnit = "m"
	TIME_UNIT_HOUR        TimeUnit = "h"
	TIME_UNIT_DAY         TimeUnit = "d"
)

var unitMap = map[TimeUnit]time.Duration{
	TIME_UNIT_NANOSECOND:  Nanosecond,
	TIME_UNIT_MICROSECOND: Microsecond,
	TIME_UNIT_MILLISECOND: Millisecond,
	TIME_UNIT_SECOND:      Second,
	TIME_UNIT_MINUTE:      Minute,
	TIME_UNIT_HOUR:        Hour,
	TIME_UNIT_DAY:         Day,
}

func ParseTimeWithRange(s string, minUnit, maxUnit time.Duration) (duration time.Duration, err error) {
	re := regexp.MustCompile(`^(\d+)([a-zA-Z]+)$`)
	match := re.FindStringSubmatch(s)

	if match == nil {
		err = fmt.Errorf("invalid time duration %s", s)
		return
	}

	value := match[1]
	unit := strings.ToLower(match[2])

	valInt, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("invalid time duration %s", s)
		return
	}

	unitValue, ok := unitMap[TimeUnit(unit)]
	if !ok {
		err = fmt.Errorf("invalid time unit %s", unit)
		return
	}

	duration = time.Duration(valInt) * unitValue
	if duration < minUnit {
		err = fmt.Errorf("time duration %s is less than min unit %s", s, minUnit)
		return
	}

	if duration > maxUnit {
		err = fmt.Errorf("time duration %s is greater than max unit %s", s, maxUnit)
		return
	}

	return
}

func ParseTime(s string) (duration time.Duration, err error) {
	return ParseTimeWithRange(s, Nanosecond, 365*Day)
}
