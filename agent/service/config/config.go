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

package config

import (
	"github.com/oceanbase/obshell/agent/errors"
	obdb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	obmodel "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"gorm.io/gorm"
)

func SaveOcsConfig(name, value, info string) error {
	db, err := obdb.GetOcsInstance()
	if err != nil {
		return errors.Wrap(err, "Get oceanbase instance failed")
	}
	cfg := obmodel.OcsConfig{
		Name:  name,
		Value: value,
		Info:  info,
	}
	return db.Save(&cfg).Error
}

func GetOcsConfig(name string) (*obmodel.OcsConfig, error) {
	db, err := obdb.GetOcsInstance()
	if err != nil {
		return nil, errors.Wrap(err, "Get oceanbase instance failed")
	}
	var cfg obmodel.OcsConfig
	err = db.Where("name = ?", name).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Get ocs config failed")
	}
	return &cfg, nil
}
