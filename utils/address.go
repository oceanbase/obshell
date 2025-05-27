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
	"net"
	"strconv"
)

func IsValidIp(ip string) bool {
	return net.ParseIP(ip) != nil
}

func IsValidPort(port string) bool {
	if port == "" {
		return true
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return IsValidPortValue(p)
}

func IsValidPortValue(p int) bool {
	return p > 1024 && p < 65536
}
