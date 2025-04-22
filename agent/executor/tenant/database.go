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

func DeleteDatabase(tenantName, databaseName string) *errors.OcsAgentError {
	db, err := GetConnection(tenantName)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	err = tenantService.DropDatabase(db, databaseName)
	if err != nil {
		errors.Occurf(errors.ErrUnexpected, "Failed to drop database %s of tenant %s", databaseName, tenantName)
	}
	return nil
}

func CreateDatabase(tenantName string, param *param.CreateDatabaseParam) *errors.OcsAgentError {
	db, err := GetConnection(tenantName)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	err = tenantService.CreateDatabase(db, param)
	if err != nil {
		errors.Occurf(errors.ErrUnexpected, "Failed to create database %s of tenant %s", param.DbName, tenantName)
	}
	return nil
}

func AlterDatabase(tenantName string, databaseName string, param *param.ModifyDatabaseParam) *errors.OcsAgentError {
	if param.Collation == nil && param.ReadOnly == nil {
		return nil
	}
	db, err := GetConnection(tenantName)
	defer CloseDbConnection(db)
	if err != nil {
		return errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	err = tenantService.AlterDatabase(db, databaseName, param)
	if err != nil {
		errors.Occurf(errors.ErrUnexpected, "Failed to modify database %s of tenant %s", databaseName, tenantName)
	}
	return nil
}

func GetDatabase(tenantName, databaseName string) (*bo.Database, *errors.OcsAgentError) {
	databases, err := ListDatabases(tenantName)
	if err != nil {
		return nil, err
	}
	for _, database := range databases {
		if database.DbName == databaseName {
			return &database, nil
		}
	}
	return nil, errors.Occurf(errors.ErrNotFound, "Database %s of tenant %s", databaseName, tenantName)
}

func ListDatabases(tenantName string) ([]bo.Database, *errors.OcsAgentError) {
	db, err := GetConnection(tenantName)
	defer CloseDbConnection(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to get db connection of tenant %s, err: %s", tenantName, err.Error())
	}
	databases, err := tenantService.ListDatabases(db)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "Failed to list databases of tenant %s, err: %s", tenantName, err.Error())
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
