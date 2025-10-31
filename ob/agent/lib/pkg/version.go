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

package pkg

import (
	"math"
	"strconv"
	"strings"
)

// CompareVersion compares two version-release strings,
// if vr1 > vr2, return 1, if vr1 < vr2, return -1, else return 0.
func CompareVersion(vr1, vr2 string) int {
	ver1s, rel1 := getVersionAndRelease(vr1)
	len1 := len(ver1s)

	ver2s, rel2 := getVersionAndRelease(vr2)
	len2 := len(ver2s)

	length := int(math.Min(float64(len1), float64(len2)))
	for i := 0; i < length; i++ {
		v1, _ := strconv.Atoi(ver1s[i])
		v2, _ := strconv.Atoi(ver2s[i])
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
	}
	if len1 > len2 {
		return 1
	} else if len1 < len2 {
		return -1
	}

	if rel1 > rel2 {
		return 1
	} else if rel1 < rel2 {
		return -1
	}
	return 0
}

func getVersionAndRelease(version string) ([]string, int) {
	info := strings.Split(version, "-")
	if len(info) == 1 {
		return strings.Split(info[0], "."), 0
	}
	rel, _ := strconv.Atoi(info[1])
	return strings.Split(info[0], "."), rel
}
