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
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	logconfig "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

var table_list = []interface{}{
	oceanbase.AllAgent{},
	oceanbase.TaskMaintainer{},
	oceanbase.DagInstance{},
	oceanbase.NodeInstance{},
	oceanbase.SubtaskInstance{},
	oceanbase.SubTaskLog{},
	oceanbase.UpgradePkgInfo{},
	oceanbase.UpgradePkgChunk{},
	oceanbase.ClusterStatus{},
	oceanbase.PartialMaintenance{},
}

// createGormDbByConfig will create an ob db instance according to the configuration and
// set the specifications of the connection pool.
func createGormDbByConfig(datasourceConfig *config.ObDataSourceConfig) (db *gorm.DB, err error) {
	atomic.AddInt32(&connectingCount, 1)
	defer atomic.AddInt32(&connectingCount, -1)
	dsn := datasourceConfig.GetDSN()
	gormConfig := gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(logconfig.GetDBLoggerLevel())),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: constant.DB_SINGULAR_TABLE,
		}}

	times := datasourceConfig.GetTryTimes()
	updateTimes := func() {
		if times > -1 {
			times--
		}
	}

	for ; times != 0; updateTimes() {
		log.Info("try connect oceanbase: ", times)
		db, err = gorm.Open(mysqlDriver.Open(dsn), &gormConfig)
		hasAttemptedConnection = true
		if err == nil {
			break
		}

		log.WithError(err).Info("open oceanbase failed")
		if !datasourceConfig.GetSkipPwdCheck() && isPasswordError(err.Error()) {
			log.WithError(err).Info("password error")
			return
		}

		if err := CheckObserverProcess(); err != nil {
			return nil, err
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return
	}
	oceanbaseDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	oceanbaseDb.SetMaxIdleConns(datasourceConfig.GetMaxIdleConns())
	oceanbaseDb.SetMaxOpenConns(datasourceConfig.GetMaxOpenConns())
	oceanbaseDb.SetConnMaxLifetime(time.Duration(datasourceConfig.GetConnMaxLifetime()))
	return db, nil
}

func isPasswordError(errMsg string) bool {
	return strings.Contains(errMsg, "Access denied")
}

// CreateDataBase will query whether the ocs db exists, create it if it does not exist
func CreateDataBase(dBname string) (err error) {
	if dbInstance == nil {
		return errors.New("oceanbase db is nil")
	}
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s READ WRITE", dBname)
	err = dbInstance.Exec(sql).Error
	if err != nil {
		log.WithError(err).Infof("create database %s failed", dBname)
		return err
	}
	log.Infof("create database %s succeed", dBname)
	return nil
}

func IsTableAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	// Error 1050: Table already exists
	return strings.Contains(err.Error(), "Error 1050")
}

func IsDuplicateColumn(err error) bool {
	if err == nil {
		return false
	}
	// Error 1060: Duplicate column name
	return strings.Contains(err.Error(), "Error 1060")
}

func IsTableNotExists(err error) bool {
	if err == nil {
		return false
	}
	// Error 1146: Table doesn't exist
	return strings.Contains(err.Error(), "Error 1146")
}

func AutoMigrateObTables() (err error) {
	if dbInstance == nil {
		return errors.New("oceanbase db is nil")
	}
	// When the ob db instance exists, do ob table migration
	return dbInstance.AutoMigrate(table_list...)
}

func ParallelAutoMigrateObTables() (err error) {
	if dbInstance == nil {
		return errors.New("oceanbase db is nil")
	}
	for _, table := range table_list {
		for i := 0; i < 10; i++ {
			err = dbInstance.AutoMigrate(table)
			if err == nil || IsTableAlreadyExists(err) || IsDuplicateColumn(err) {
				break
			}
			if IsTableNotExists(err) {
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				log.WithError(err).Errorf("auto migrate ob table %s failed", reflect.TypeOf(table).Name())
				return err
			}
		}
	}
	log.Info("auto migrate ob tables succeed")
	return nil
}
