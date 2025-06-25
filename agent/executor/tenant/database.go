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
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/param"
)

func DeleteDatabase(tenantName, databaseName string, password *string) error {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	err = tenantService.DropDatabase(db, databaseName)
	if err != nil {
		return errors.Wrapf(err, "Failed to drop database %s of tenant %s", databaseName, tenantName)
	}
	return nil
}

func CreateDatabase(tenantName string, param *param.CreateDatabaseParam) error {
	db, err := GetConnectionWithPassword(tenantName, param.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	err = tenantService.CreateDatabase(db, param)
	if err != nil {
		return errors.Wrapf(err, "Failed to create database %s of tenant %s", param.DbName, tenantName)
	}
	return nil
}

func AlterDatabase(tenantName string, databaseName string, param *param.ModifyDatabaseParam) error {
	if param.Collation == nil && param.ReadOnly == nil {
		return nil
	}
	db, err := GetConnectionWithPassword(tenantName, param.RootPassword)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	err = tenantService.AlterDatabase(db, databaseName, param)
	if err != nil {
		return errors.Wrapf(err, "Failed to modify database %s of tenant %s", databaseName, tenantName)
	}
	return nil
}

func GetDatabase(tenantName, databaseName string, password *string) (*bo.Database, error) {
	databases, err := ListDatabases(tenantName, password)
	if err != nil {
		return nil, err
	}
	for _, database := range databases {
		if database.DbName == databaseName {
			return &database, nil
		}
	}
	return nil, errors.Occur(errors.ErrObDatabaseNotExist, databaseName, tenantName)
}

func ListDatabases(tenantName string, password *string) ([]bo.Database, error) {
	db, err := GetConnectionWithPassword(tenantName, password)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get db connection of tenant %s", tenantName)
	}
	databases, err := tenantService.ListDatabases(db)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to list databases of tenant %s", tenantName)
	}
	obdatabases := make([]bo.Database, 0)
	for _, database := range databases {
		connectionUrl := bo.ObproxyAndConnectionString{
			Type:             constant.OB_CONNECTION_TYPE_DIRECT,
			ConnectionString: fmt.Sprintf("jdbc:mysql://%s:%d/%s", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT, database.Name),
		}
		obdatabase := bo.Database{
			DbName:         database.Name,
			Charset:        database.CharSetName,
			Collation:      database.CollationName,
			ReadOnly:       strings.HasPrefix(strings.ToUpper(database.ReadOnly), "Y"),
			CreateTime:     database.CreateTimestamp,
			ConnectionUrls: []bo.ObproxyAndConnectionString{connectionUrl},
		}
		obdatabases = append(obdatabases, obdatabase)
	}
	return obdatabases, nil
}
