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
	"net"
	"os"
	"strings"
)

func GetAddressList() ([]string, error) {
	addressList, err := GetAllAddressList()
	if err != nil {
		return nil, err
	}
	return RemoveDuplicates(addressList), nil
}

func GetAllAddressList() ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	address, err := net.LookupHost(hostname)
	if err != nil {
		return nil, err
	}
	targetAddress := address[0:]
	for _, add := range address {
		if !strings.HasPrefix(add, "127") {
			targetAddress = append(targetAddress, add)
		}
	}
	return targetAddress, nil
}

func RemoveDuplicates(items []string) []string {
	slice := make([]string, 0)
	tmp := make(map[string]struct{}, 0)
	for _, item := range items {
		if _, ok := tmp[item]; !ok {
			tmp[item] = struct{}{}
			slice = append(slice, item)
		}
	}
	return slice
}
