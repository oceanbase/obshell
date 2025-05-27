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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

type Dialector struct {
	mysql.Dialector
}

func Open(dsn string) gorm.Dialector {
	mysqlDialector := mysql.Open(dsn).(*mysql.Dialector)
	return Dialector{
		Dialector: *mysqlDialector,
	}
}

func OpenObproxy(dsn string) gorm.Dialector {
	mysqlDialector := mysql.Open(dsn).(*mysql.Dialector)
	mysqlDialector.Config.SkipInitializeWithVersion = true
	return Dialector{
		Dialector: *mysqlDialector,
	}
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	mysqlMigrator := mysql.Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: dialector,
			},
		},
		Dialector: dialector.Dialector,
	}
	return Migrator{
		Migrator: mysqlMigrator,
	}
}
