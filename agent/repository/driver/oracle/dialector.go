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

package oracle

import (
	"database/sql"

	"github.com/oceanbase/obshell/agent/constant"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

type Dialector struct {
	mysql.Dialector
}

func Open(dsn string) gorm.Dialector {
	return Dialector{
		Dialector: mysql.Dialector{
			Config: &mysql.Config{
				DSN: dsn,
			},
		},
	}
}

func (d Dialector) Name() string {
	return constant.ORACLE_MODE
}

func (d Dialector) Initialize(db *gorm.DB) (err error) {
	d.DefaultStringSize = 1024
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{})

	d.DriverName = "oracle"

	if d.Conn != nil {
		db.ConnPool = d.Conn
	} else {
		db.ConnPool, err = sql.Open(d.DriverName, d.DSN)
		if err != nil {
			return err
		}
	}
	return
}
