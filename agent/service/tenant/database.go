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

	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
	"gorm.io/gorm"
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

func (t *TenantService) AlterDatabase(db *gorm.DB, databaseName string, modifyDatabaseParam *param.ModifyDatabaseParam) error {
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

func (t *TenantService) IsDatabaseExist(db *gorm.DB, databaseName string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM oceanbase.DBA_OB_DATABASES WHERE DATABASE_NAME = ?", databaseName).Scan(&count).Error
	return count > 0, err
}

func (t *TenantService) CreateDatabase(db *gorm.DB, createDatabaseParam *param.CreateDatabaseParam) error {
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

func (t *TenantService) DropDatabase(db *gorm.DB, databaseName string) error {
	sql := fmt.Sprintf("DROP DATABASE `%s`", databaseName)
	return db.Exec(sql).Error
}

func (t *TenantService) ListDatabases(db *gorm.DB) ([]oceanbase.Database, error) {
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

func (t *TenantService) GetDatabase(db *gorm.DB, databaseName string) (*oceanbase.Database, error) {
	var database oceanbase.Database
	result := db.Raw(GET_DATABASE_SQL, databaseName).Scan(&database)
	return &database, result.Error
}

func (t *TenantService) ListDatabasePrivileges(db *gorm.DB) ([]oceanbase.MysqlDb, error) {
	dbs := make([]oceanbase.MysqlDb, 0)
	result := db.Table(MYSQL_DB).Find(&dbs)
	return dbs, result.Error
}

func (t *TenantService) ListDatabasePrivilegesOfUser(db *gorm.DB, userName string) ([]oceanbase.MysqlDb, error) {
	dbs := make([]oceanbase.MysqlDb, 0)
	result := db.Table(MYSQL_DB).Find(&dbs, "user=?", userName)
	return dbs, result.Error
}
