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
)

const (
	CAP_KB_BIT = 10
	CAP_MB_BIT = 20
	CAP_GB_BIT = 30
	CAP_TB_BIT = 40
	CAP_PB_BIT = 50

	CAP_K  = "K"
	CAP_M  = "M"
	CAP_G  = "G"
	CAP_T  = "T"
	CAP_P  = "P"
	CAP_KB = "KB"
	CAP_MB = "MB"
	CAP_GB = "GB"
	CAP_TB = "TB"
	CAP_PB = "PB"

	CAPACITY_PATTERN = `^([123456789]\d*)([KMGTP][B]?)$`
)

func CapacityParser(capacity string) (int, bool) {
	cap := strings.ToUpper(capacity)
	re := regexp.MustCompile(CAPACITY_PATTERN)
	if re.MatchString(cap) {
		match := re.FindStringSubmatch(cap)
		num, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, false
		}
		switch match[2] {
		case CAP_K, CAP_KB:
			return num << CAP_KB_BIT, true
		case CAP_M, CAP_MB:
			return num << CAP_MB_BIT, true
		case CAP_G, CAP_GB:
			return num << CAP_GB_BIT, true
		case CAP_T, CAP_TB:
			return num << CAP_TB_BIT, true
		case CAP_P, CAP_PB:
			return num << CAP_PB_BIT, true
		}
	}
	return 0, false
}

func FormatCapacity(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
		PB = TB * 1024
		EB = PB * 1024
	)

	switch {
	case bytes >= EB:
		return fmt.Sprintf("%.2f EB", float64(bytes)/float64(EB))
	case bytes >= PB:
		return fmt.Sprintf("%.2f EB", float64(bytes)/float64(PB))
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
