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

package observer

import (
	"strings"
)

const (
	defaultWhitelist = "127.0.0.1"
)

func ModifyWhitelist(whitelist string) error {
	if err := tenantService.ModifyWhitelist(mergeWhitelist(whitelist)); err != nil {
		return err
	}
	return nil
}

// mergeWhitelist merge s
func mergeWhitelist(specific string) string {
	if specific == "" {
		return defaultWhitelist
	}
	splits := strings.Split(specific, ",")
	splits = append(splits, defaultWhitelist)
	whitelistMap := make([]string, 0)
	// 去重
	unique := make(map[string]struct{})
	for _, item := range splits {
		if item == "%" {
			return "%"
		}
		if _, ok := unique[item]; !ok {
			unique[item] = struct{}{}
			whitelistMap = append(whitelistMap, item)
		}
	}
	return strings.Join(whitelistMap, ",")
}
