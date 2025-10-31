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

package oceanbase

import (
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
)

var (
	ERR_OBSERVER_NOT_EXIST = errors.Occur(errors.ErrObServerProcessNotExist)
)

// GetOcsInstance will return a connection to the OCS database.
// If the connection cannot execute the SQL command 'SHOW DATABASES', it will return an error.
func GetOcsInstance() (db *gorm.DB, err error) {
	db, err = getSqlExecutableInstance(TEST_OCEANBASE_SQL)
	if err != nil {
		return nil, err
	}

	if isOcs {
		return db, nil
	}
	return nil, errors.Occur(errors.ErrAgentOceanbaseDBNotOcs)
}

// GetInstance will return the current connection regardless of the database it is connected with.
// If the connection cannot execute the SQL command 'SHOW DATABASES', it will return an error.
func GetInstance() (db *gorm.DB, err error) {
	return getSqlExecutableInstance(TEST_OCEANBASE_SQL)
}

func GetRawInstance() (db *gorm.DB) {
	return dbInstance
}

func ClearInstance() {
	dbInstance = nil
}

// GetRestrictedInstance will return the connection which not specify any database
// and this connection can only execute the SQL command 'SELECT 1'.
func GetRestrictedInstance() (db *gorm.DB, err error) {
	return getSqlExecutableInstance(TEST_DATABASE_SQL)
}

func checkObAvailable() (bool, error) {
	var count int = 0
	err := dbInstance.Raw("select count(*) from oceanbase.GV$OB_SERVER_SCHEMA_INFO where (svr_ip, svr_port) in (select svr_ip, svr_port from oceanbase.GV$OB_LOG_STAT where tenant_id = 1 and role = 'LEADER') and tenant_id = 1 and refreshed_schema_version = received_schema_version").Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetAvailableInstance() (db *gorm.DB, err error) {
	// If the ob instance currently in memory is nil, it will return an error.
	if dbInstance == nil {
		return nil, errors.Occur(errors.ErrAgentOceanbaseNotHold)
	}

	if canUse, err := checkObAvailable(); err == nil && canUse {
		return dbInstance, nil
	}

	// If the test sql execution fails, it will check the observer process
	if err := CheckObserverProcess(); err != nil {
		return nil, err
	}

	// If the above checks pass, the current db is unavailable
	return nil, errors.Occur(errors.ErrAgentOceanbaseUesless)
}

// GetSqlExecutableInstance will return the connection which can execute the specified sql command.
func getSqlExecutableInstance(sql string) (db *gorm.DB, err error) {
	// If the ob instance currently in memory is nil, it will return an error.
	if dbInstance == nil {
		return nil, errors.Occur(errors.ErrAgentOceanbaseNotHold)
	}

	// If the db instance in the current memory is not nil,
	// the specified SQL command is executed to confirm if the instance is available.
	if err := dbInstance.Exec(sql).Error; err == nil {
		return dbInstance, nil
	}

	// If the test sql execution fails, it will check the observer process
	if err := CheckObserverProcess(); err != nil {
		return nil, err
	}

	// If the above checks pass, the current db is unavailable
	return nil, errors.Occur(errors.ErrAgentOceanbaseUesless)
}

// if observer process not exist, return error
func CheckObserverProcess() error {
	exist, err := process.CheckObserverProcess()
	if err != nil {
		return errors.Occur(errors.ErrObServerProcessCheckFailed, err.Error())
	}
	if !exist {
		return ERR_OBSERVER_NOT_EXIST
	}
	return nil
}

func HasOceanbaseInstance() bool {
	return dbInstance != nil
}

func GetLastInitError() error {
	if dbInstance != nil {
		return nil
	}
	return lastInitError
}

func IsInitPasswordError() bool {
	err := GetLastInitError()
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return isPasswordError(errMsg)
}
