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

package driver

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

const (
	ALIAS_TYPE_BOOL             = "tinyint(1)" // tinyint(1) is alias of boolean in OceanBase
	VALUE_CURRENT_TIMESTAMP     = "CURRENT_TIMESTAMP"
	ON_UPDATE_CURRENT_TIMESTAMP = "ON UPDATE CURRENT_TIMESTAMP"
)

type Migrator struct {
	mysql.Migrator
}

func (m Migrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columnTypes, err := m.Migrator.ColumnTypes(value)
	if err != nil {
		return nil, err
	}
	columnExtra := make(map[string]string)
	err = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var (
			currentDatabase, table = m.CurrentSchema(stmt, stmt.Table)
			columnTypeSQL          = "SELECT column_name, extra FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ORDINAL_POSITION "
		)
		columns, rowErr := m.DB.Raw(columnTypeSQL, currentDatabase, table).Rows()
		if rowErr != nil {
			return rowErr
		}
		defer columns.Close()
		i := 0
		for columns.Next() {
			var (
				columnName string
				extra      string
			)
			if scanErr := columns.Scan(&columnName, &extra); scanErr != nil {
				return scanErr
			}
			columnExtra[columnName] = strings.ToUpper(extra)
			i++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, columnType := range columnTypes {
		// If the MySQL driver overrides the ColumnType method,
		// modifications are needed here.
		if column, ok := columnType.(migrator.ColumnType); ok {
			if defaultValue, ok := columnType.DefaultValue(); ok {
				if strings.HasPrefix(defaultValue, VALUE_CURRENT_TIMESTAMP) {
					// Add precision for CURRENT_TIMESTAMP.
					if precision, _, ok := columnType.DecimalSize(); ok && precision > 0 {
						column.DefaultValueValue.String = strings.Replace(defaultValue, VALUE_CURRENT_TIMESTAMP, fmt.Sprintf("%s(%d)", VALUE_CURRENT_TIMESTAMP, precision), 1)
					}
					if extra, ok := columnExtra[columnType.Name()]; ok && strings.Contains(extra, ON_UPDATE_CURRENT_TIMESTAMP) {
						// Add 'on update current_timestamp' to default value.
						column.DefaultValueValue.String += fmt.Sprintf(" %s", ON_UPDATE_CURRENT_TIMESTAMP)
					}
				} else {
					if typeName, ok := columnType.ColumnType(); ok && strings.EqualFold(typeName, ALIAS_TYPE_BOOL) {
						if defaultValue == "1" {
							column.DefaultValueValue.String = "true"
						} else if defaultValue == "0" {
							column.DefaultValueValue.String = "false"
						}
					}
				}
				columnTypes[i] = column
			}
		}
	}

	return columnTypes, nil
}
