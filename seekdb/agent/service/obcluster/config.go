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

package obcluster

import (
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
)

func (s *ObserverService) GetOBParatemerByName(name string, value interface{}) (err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return
	}

	err = db.Table(ob_parameters_view).Select("value").Where("name = ?", name).Scan(value).Error
	return
}

func (s *ObserverService) GetOBStringParatemerByName(name string) (value string, err error) {
	err = s.GetOBParatemerByName(name, &value)
	return
}

func (s *ObserverService) GetOBIntParatemerByName(name string) (value int, err error) {
	err = s.GetOBParatemerByName(name, &value)
	return
}
