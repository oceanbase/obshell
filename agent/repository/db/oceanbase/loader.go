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
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
)

var (
	connectionIniting = false
	connectingCount   int32
)

func Init() {
	go initOnce.Do(initConnection)
}

func IsConnecting() bool {
	return connectionIniting || connectingCount > 0
}

func HasAttemptedConnection() bool {
	return hasAttemptedConnection
}

func initConnection() {
	connectionIniting = true
	defer func() {
		connectionIniting = false
	}()

	if meta.MYSQL_PORT != 0 {
		loadOceanbaseInstanceWithoutDBNameUntilSucc()
		loadOceanbaseUntilSucc()
	} else {
		log.Info("mysql port has not been initialized, no need to reload oceanbase instance")
	}
}

func loadOceanbaseInstanceWithoutDBNameUntilSucc() {
	var err error
	log.Info("initialzie oceanbase instance without db name ...")
	dsConfig := getDataSourceConfig().SetTryTimes(1).SetDBName("")
	for {
		if dbInstance != nil {
			return
		}

		dsConfig.SetPort(meta.MYSQL_PORT).SetPassword(meta.OCEANBASE_PWD)
		err = loadGormOceanbase(dsConfig)
		if err == nil {
			log.Info("init oceanbase instance without db name success")
			return
		}
		log.Info("loadOceanbaseInstanceWithoutDBNameUntilSucc last error is ", err)
		lastInitError = err
		if isPasswordError(err.Error()) {
			log.Warnf("init oceanbase instance without db name failed: %v", err)
			return
		}
		log.Info("init oceanbase instance without db name failed")
		time.Sleep(time.Second)
	}
}

func loadOceanbaseUntilSucc() {
	var err error
	log.Info("initialzie oceanbase instance ...")
	dsConfig := getDataSourceConfig().SetTryTimes(1)
	for {
		if dbInstance != nil && currentConfig.GetDBName() == dsConfig.GetDBName() {
			return
		}

		dsConfig.SetPort(meta.MYSQL_PORT).SetPassword(meta.OCEANBASE_PWD)
		err = loadGormOceanbase(dsConfig)
		if err == nil {
			log.Info("init oceanbase instance success")
			return
		}
		log.Info("loadOceanbaseUntilSucc last error is ", err)
		lastInitError = err
		if isPasswordError(err.Error()) {
			log.Warnf("init oceanbase instance without db name failed: %v", err)
			return
		}
		log.Info("init oceanbase instance")
		time.Sleep(time.Second)
	}
}

// LoadOceanbaseInstance creates a db instance according to the configuration.
// The `config` is a variable-length parameter, and the corresponding operation is selected
// according to whether the parameter is set or not. If there are no special requirements, do not set config.
func LoadOceanbaseInstance(config ...*config.ObDataSourceConfig) error {
	if len(config) > 0 {
		return loadGormOceanbase(config[0])
	} else {
		return loadOceanbaseInstanceNormal()
	}
}

// loadOceanbaseInstanceNormal will generate default configuration and create ob instance.
func loadOceanbaseInstanceNormal() error {
	log.Info("load oceanbase instance")
	dsConfig := getDataSourceConfig()
	if err := loadGormOceanbase(dsConfig); err != nil {
		return err
	}
	return nil
}

// loadGormOceanbase will create a db instance according to the configuration,
// if the instance currently exists, will close it and replace it.
func loadGormOceanbase(dsConfig *config.ObDataSourceConfig) error {
	if err := fillConfigPort(dsConfig); err != nil {
		return errors.Wrap(err, "get port failed")
	}
	db, err := createGormDbByConfig(dsConfig)
	if err != nil {
		return errors.Wrap(err, "initialize oceanbase failed")
	}
	setDB(db, dsConfig)
	return nil
}

func showUpdateCurrentConfig(dsConfig *config.ObDataSourceConfig) bool {
	if currentConfig == nil {
		log.Info("current config is nil, update db instance")
		return true
	}
	if currentConfig.GetPassword() != meta.OCEANBASE_PWD {
		log.Info("password changed, update db instance")
		return true
	}
	if currentConfig.GetPort() != meta.MYSQL_PORT {
		log.Info("port changed, update db instance")
		return true
	}
	if !isOcs && dsConfig.GetDBName() != currentConfig.GetDBName() {
		log.Info("db name changed, update db instance")
		return true
	}

	_, err := GetRestrictedInstance()
	return err != nil
}

func setDB(db *gorm.DB, dsConfig *config.ObDataSourceConfig) {
	// The lock is used to ensure that the concurrent access to the db is safe.
	dbLock.Lock()
	defer dbLock.Unlock()

	if !showUpdateCurrentConfig(dsConfig) {
		releaseDB(db)
		return
	}

	// Free previous db.
	releaseDB(dbInstance)

	// Set new db.
	dbInstance = db
	currentConfig = dsConfig
	isOcs = dsConfig.GetDBName() == constant.DB_OCS
}

func releaseDB(preDB *gorm.DB) {
	// Delay release db
	if preDB != nil {
		oceanbaseDB, err := preDB.DB()
		if err != nil {
			log.WithError(err).Warn("release pre db failed")
		}

		go func() {
			defer func() {
				err := recover()
				if err != nil {
					log.WithError(err.(error)).Warn("release pre db failed")
				}
			}()

			for oceanbaseDB.Stats().InUse != 0 {
				log.Debug("pre db is using, wait for release")
				time.Sleep(time.Second)
			}
			oceanbaseDB.Close()
		}()
	}
}

func fillConfigPort(dsConfig *config.ObDataSourceConfig) error {
	if dsConfig.GetPort() != 0 {
		return nil
	}
	// Get mysql port.
	if meta.MYSQL_PORT == 0 {
		return errors.New("mysql port has not been initialized")
	}
	dsConfig.SetPort(meta.MYSQL_PORT)
	return nil
}

// getDataSourceConfig will generate default connection configuration.
func getDataSourceConfig() *config.ObDataSourceConfig {
	dsConfig := config.NewObDataSourceConfig().SetPassword(meta.GetOceanbasePwd()).SetParseTime(true)
	return dsConfig
}

// LoadOceanbaseInstanceForTest will try connecting ob with the given configuration to verify that the configuration is correct.
func LoadOceanbaseInstanceForTest(dsConfig *config.ObDataSourceConfig) error {
	if err := fillConfigPort(dsConfig); err != nil {
		log.WithError(err).Error("get port failed")
		return err
	}
	log.Info("load oceanbase instance for test")
	if err := loadObGormForTest(dsConfig); err != nil {
		log.WithError(err).Error("load gorm oceanbase for test failed")
		return err
	}
	return nil
}

// loadObGormForTest will try to connect to the db according to the configuration and close it.
func loadObGormForTest(dsConfig *config.ObDataSourceConfig) error {
	db, err := gorm.Open(mysqlDriver.Open(dsConfig.GetDSN()))
	defer func() {
		if db != nil {
			oceanbaseDB, _ := db.DB()
			oceanbaseDB.Close()
		}
	}()
	if err != nil {
		log.WithError(err).Error("open ob db failed")
		return err
	}
	return nil
}
