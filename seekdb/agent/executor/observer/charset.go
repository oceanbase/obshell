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
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
)

func supportMysql(charset *oceanbase.ObCharset) bool {
	return charset.Charset != "utf16"
}

// GetObserverCharsets retrieves all character sets supporting MYSQL and their collations from the OceanBase cluster.
func GetObserverCharsets() ([]bo.CharsetInfo, error) {
	charsets, err := observerService.GetAllCharsets()
	if err != nil {
		return nil, err
	}
	collations, err := observerService.GetAllCollations()
	if err != nil {
		return nil, err
	}
	charsetInfoMap := make(map[string]*bo.CharsetInfo, len(charsets))
	for i := range charsets {
		if !supportMysql(&charsets[i]) {
			continue
		}
		charsetInfoMap[charsets[i].Charset] = &bo.CharsetInfo{
			Name:        charsets[i].Charset,
			Description: charsets[i].Description,
			Maxlen:      charsets[i].MaxLen,
			Collations:  make([]bo.CollationInfo, 0),
		}
	}
	for i := range collations {
		if _, ok := charsetInfoMap[collations[i].Charset]; !ok {
			continue
		}
		charsetInfoMap[collations[i].Charset].Collations = append(charsetInfoMap[collations[i].Charset].Collations, bo.CollationInfo{
			Name:      collations[i].Collation,
			IsDefault: collations[i].IsDefault == "Yes",
		})
	}
	charsetinfos := make([]bo.CharsetInfo, 0, len(charsetInfoMap))
	for _, charsetInfo := range charsetInfoMap {
		charsetinfos = append(charsetinfos, *charsetInfo)
	}
	return charsetinfos, nil
}
