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

package unit

import (
	"fmt"

	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	oceanbaseModel "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

type UnitService struct{}

const (
	DBA_OB_UNIT_CONFIGS = "oceanbase.DBA_OB_UNIT_CONFIGS"
)

const (
	SQL_CREATE_UNIT_CONFIG = "CREATE RESOURCE UNIT `%s`  MEMORY_SIZE '%s', MAX_CPU %f"
	DROP_UNIT_CONFIG       = "DROP RESOURCE UNIT `%s`"
)

func (u *UnitService) CreateUnit(param param.CreateResourceUnitConfigParams) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(SQL_CREATE_UNIT_CONFIG, *param.Name, *param.MemorySize, *param.MaxCpu)
	if param.MinCpu != nil {
		sql += fmt.Sprintf(", MIN_CPU %f", *param.MinCpu)
	}
	if param.MaxIops != nil {
		sql += fmt.Sprintf(", MAX_IOPS %d", *param.MaxIops)
	}
	if param.MinIops != nil {
		sql += fmt.Sprintf(", MIN_IOPS %d", *param.MinIops)
	}
	if param.LogDiskSize != nil {
		sql += fmt.Sprintf(", LOG_DISK_SIZE '%s'", *param.LogDiskSize)
	}
	return db.Exec(sql).Error
}

func (u *UnitService) IsUnitConfigExist(unit_name string) (bool, error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(DBA_OB_UNIT_CONFIGS).Where("NAME = ?", unit_name).Count(&count).Error
	return count > 0, err
}

func (u *UnitService) DropUnit(unit_name string) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(DROP_UNIT_CONFIG, unit_name)
	return db.Exec(sql).Error
}

func (u *UnitService) GetAllUnitConfig() (units []oceanbaseModel.DbaObUnitConfig, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_UNIT_CONFIGS).Select("create_time,modify_time,unit_config_id,name,max_cpu,min_cpu,memory_size,log_disk_size,max_iops,min_iops").Scan(&units).Error
	return
}

func (u *UnitService) GetUnitConfigByName(name string) (unit *oceanbaseModel.DbaObUnitConfig, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_UNIT_CONFIGS).Select("create_time,modify_time,unit_config_id,name,max_cpu,min_cpu,memory_size,log_disk_size,max_iops,min_iops").Where("NAME = ?", name).Scan(&unit).Error
	return
}

func (u *UnitService) GetUnitConfigById(id int) (unit *oceanbaseModel.DbaObUnitConfig, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_UNIT_CONFIGS).Select("create_time,modify_time,unit_config_id,name,max_cpu,min_cpu,memory_size,log_disk_size,max_iops,min_iops").Where("UNIT_CONFIG_ID = ?", id).Scan(&unit).Error
	return
}

func (u *UnitService) GetUnitConfigNameById(id int) (name string, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return "", err
	}
	err = db.Table(DBA_OB_UNIT_CONFIGS).Select("NAME").Where("UNIT_CONFIG_ID = ?", id).Scan(&name).Error
	return
}
