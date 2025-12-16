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

package tenant

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
)

func (t *TenantService) SetParameters(parameters map[string]interface{}) error {
	if len(parameters) == 0 {
		return nil
	}
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	items := make([]string, 0)
	for k, v := range parameters {
		items = append(items, fmt.Sprintf("`%s` = \"%v\"", k, v))
	}
	sql := SQL_SET_PARAMETER_BASIC + strings.Join(items, ",")
	return db.Exec(sql).Error
}

func (t *TenantService) SetVariables(variables map[string]interface{}) error {
	if len(variables) == 0 {
		return nil
	}
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	variablesSql := ""
	for k, v := range variables {
		if val, ok := v.(string); ok {
			if number, err := strconv.Atoi(val); err == nil {
				variablesSql += fmt.Sprintf(", GLOBAL "+k+"= %v", number)
			} else if float, err := strconv.ParseFloat(val, 64); err == nil {
				variablesSql += fmt.Sprintf(", GLOBAL "+k+"= %v", float)
			} else {
				variablesSql += fmt.Sprintf(", GLOBAL "+k+"= `%v`", val)
			}
		} else {
			variablesSql += fmt.Sprintf(", GLOBAL "+k+"= %v", v)
		}
	}
	sqlText := fmt.Sprintf("SET %s", variablesSql[1:])
	return db.Exec(sqlText).Error
}

func (t *TenantService) ModifyWhitelist(whitelist string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_ALTER_TENANT_WHITELIST, whitelist)).Error
}

func (t *TenantService) GetParameters(filter string) (parameters []oceanbase.GvObParameter, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(GV_OB_PARAMETERS).
		Select("DISTINCT NAME, VALUE, DATA_TYPE, INFO, EDIT_LEVEL").
		Where("NAME LIKE ?", filter).
		Scan(&parameters).Error
	return
}

func (t *TenantService) GetParameter(parameterName string) (parameter *oceanbase.GvObParameter, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(GV_OB_PARAMETERS).Select("DISTINCT NAME, VALUE, DATA_TYPE, INFO, EDIT_LEVEL").
		Where("NAME = ?", parameterName).
		Scan(&parameter).Error
	// retry for bad case for virtual table
	if parameter == nil && err == nil {
		err = db.Table(GV_OB_PARAMETERS).Where("NAME = ?", parameterName).Scan(&parameter).Error
	}
	return
}

func (t *TenantService) GetVariables(filter string) (variables []oceanbase.DbaObSysVariable, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_SYS_VARIABLES).
		Where("NAME LIKE ?", filter).Scan(&variables).Error
	if err != nil {
		return nil, err
	}

	return
}

func (t *TenantService) IsVariableExist(variableName string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(DBA_OB_SYS_VARIABLES).Where("NAME = ?", variableName).Count(&count).Error
	return count > 0, err
}

func (t *TenantService) GetTenantVariable(variableName string) (variable *oceanbase.DbaObSysVariable, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(DBA_OB_SYS_VARIABLES).
		Where("NAME = ?", variableName).Scan(&variable).Error
	if err != nil {
		return nil, err
	}

	return
}

func (s *TenantService) GetCompaction() (compaction *oceanbase.CdbObMajorCompaction, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Model(oceanbase.CdbObMajorCompaction{}).Where("tenant_id = ?", constant.TENANT_SYS_ID).Scan(&compaction).Error
	return
}

func (s *TenantService) MajorCompaction() error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec("ALTER SYSTEM MAJOR FREEZE").Error
}

func (s *TenantService) ClearCompactionError() error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec("ALTER SYSTEM CLEAR MERGE ERROR").Error
}
