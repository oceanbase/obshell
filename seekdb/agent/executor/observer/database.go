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

package observer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/bo"
	"github.com/oceanbase/obshell/seekdb/param"
)

func DeleteDatabase(databaseName string) error {
	exist, err := tenantService.IsDatabaseExist(databaseName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if database %s exists", databaseName)
	}
	if !exist {
		return nil
	}
	err = tenantService.DropDatabase(databaseName)
	if err != nil {
		return errors.Wrapf(err, "Failed to drop database %s", databaseName)
	}
	return nil
}

func CreateDatabase(param *param.CreateDatabaseParam) error {
	// check the database name is valid
	if !regexp.MustCompile(constant.DATABASE_PATTERN).MatchString(param.DbName) {
		return errors.Occur(errors.ErrObDatabaseNameInvalid, param.DbName)
	}
	err := tenantService.CreateDatabase(param)
	if err != nil {
		return errors.Wrapf(err, "Failed to create database %s", param.DbName)
	}
	return nil
}

func AlterDatabase(databaseName string, param *param.ModifyDatabaseParam) error {
	if param.Collation == nil && param.ReadOnly == nil {
		return nil
	}
	exist, err := tenantService.IsDatabaseExist(databaseName)
	if err != nil {
		return errors.Wrapf(err, "Failed to check if database %s exists", databaseName)
	}
	if !exist {
		return errors.Occur(errors.ErrObDatabaseNotExist, databaseName)
	}
	err = tenantService.AlterDatabase(databaseName, param)
	if err != nil {
		return errors.Wrapf(err, "Failed to modify database %s", databaseName)
	}
	return nil
}

func GetDatabase(databaseName string) (*bo.Database, error) {
	databases, err := ListDatabases()
	if err != nil {
		return nil, err
	}
	for _, database := range databases {
		if database.DbName == databaseName {
			return &database, nil
		}
	}
	return nil, errors.Occur(errors.ErrObDatabaseNotExist, databaseName)
}

func ListDatabases() ([]bo.Database, error) {
	databases, err := tenantService.ListDatabases()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list databases")
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
