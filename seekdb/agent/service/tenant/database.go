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

	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/param"
)

const (
	LIST_DATABASE_SQL = `
        SELECT
            floor(time_to_usec(o.CREATED) / 1000000) AS CREATE_TIMESTAMP,
            o.OBJECT_ID AS DATABASE_ID,
            d.DATABASE_NAME AS NAME,
            c.ID AS COLLATION_TYPE,
            c.COLLATION_NAME AS COLLATION_NAME,
            c.CHARACTER_SET_NAME AS CHARACTER_SET_NAME,
            c.ID AS COLLATION_TYPE,
            d.READ_ONLY as READ_ONLY
        FROM
            oceanbase.DBA_OB_DATABASES d
        JOIN
            oceanbase.DBA_OBJECTS o
        JOIN
            information_schema.collations c
        ON
            d.DATABASE_NAME = o.OBJECT_NAME
        AND
            d.COLLATION = c.COLLATION_NAME
        WHERE
            o.OBJECT_TYPE = 'DATABASE'
    `
	GET_DATABASE_SQL = LIST_DATABASE_SQL + " and d.DATABASE_NAME = ?"
)

func (t *TenantService) AlterDatabase(databaseName string, modifyDatabaseParam *param.ModifyDatabaseParam) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER DATABASE `%s`", databaseName)
	if modifyDatabaseParam.Collation != nil {
		sql = fmt.Sprintf("%s DEFAULT COLLATE = '%s'", sql, *modifyDatabaseParam.Collation)
	}
	if modifyDatabaseParam.ReadOnly != nil {
		if *modifyDatabaseParam.ReadOnly {
			sql = sql + " READ ONLY"
		} else {
			sql = sql + " READ WRITE"
		}
	}
	return db.Exec(sql).Error
}

func (t *TenantService) IsDatabaseExist(databaseName string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM oceanbase.DBA_OB_DATABASES WHERE DATABASE_NAME = ?", databaseName).Scan(&count).Error
	return count > 0, err
}

func (t *TenantService) CreateDatabase(createDatabaseParam *param.CreateDatabaseParam) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("CREATE DATABASE `%s`", createDatabaseParam.DbName)
	if createDatabaseParam.Collation != nil {
		sql = fmt.Sprintf("%s DEFAULT COLLATE = '%s'", sql, *createDatabaseParam.Collation)
	}
	if createDatabaseParam.ReadOnly != nil {
		if *createDatabaseParam.ReadOnly {
			sql = sql + " READ ONLY"
		} else {
			sql = sql + " READ WRITE"
		}
	}
	return db.Exec(sql).Error
}

func (t *TenantService) DropDatabase(databaseName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP DATABASE `%s`", databaseName)
	return db.Exec(sql).Error
}

func (t *TenantService) ListDatabases() ([]oceanbase.Database, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	dbNames := make([]oceanbase.DatabaseName, 0)
	dbs := make([]oceanbase.Database, 0)
	result := db.Raw("SHOW DATABASES").Scan(&dbNames)
	if result.Error != nil {
		return dbs, result.Error
	}
	result = db.Raw(LIST_DATABASE_SQL).Scan(&dbs)
	availableDbs := make([]oceanbase.Database, 0, len(dbs))
	for i := range dbs {
		for _, dbName := range dbNames {
			if dbs[i].Name == dbName.Database {
				availableDbs = append(availableDbs, dbs[i])
			}
		}
	}
	return availableDbs, result.Error
}

func (t *TenantService) GetDatabase(databaseName string) (*oceanbase.Database, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var database oceanbase.Database
	result := db.Raw(GET_DATABASE_SQL, databaseName).Scan(&database)
	return &database, result.Error
}

func (t *TenantService) ListDatabasePrivileges() ([]oceanbase.MysqlDb, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	dbs := make([]oceanbase.MysqlDb, 0)
	result := db.Table(MYSQL_DB).Find(&dbs)
	return dbs, result.Error
}

func (t *TenantService) GetDatabaseCount() (int, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}

	dbNames := make([]oceanbase.DatabaseName, 0)
	err = db.Raw("SHOW DATABASES").Scan(&dbNames).Error
	if err != nil {
		return 0, err
	}
	return len(dbNames), nil
}

func (t *TenantService) ListDatabasePrivilegesOfUser(userName string) ([]oceanbase.MysqlDb, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	dbs := make([]oceanbase.MysqlDb, 0)
	result := db.Table(MYSQL_DB).Find(&dbs, "user=?", userName)
	return dbs, result.Error
}
