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

package obproxy

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/driver/mysql"
	"github.com/oceanbase/obshell/agent/repository/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	obproxyInstance *gorm.DB

	WAIT_OBPROXY_CONNECTED_MAX_TIMES    = 100
	WAIT_OBPROXY_CONNECTED_MAX_INTERVAL = 10 * time.Second
)

func LoadObproxyInstance() (db *gorm.DB, err error) {
	if meta.OBPROXY_SQL_PORT == 0 {
		return nil, errors.Occur(errors.ErrCommonUnexpected, "obproxy sql port has not been initialized")
	}
	dsConfig := config.NewObproxyDataSourceConfig().SetPort(meta.OBPROXY_SQL_PORT).SetPassword(meta.OBPROXY_SYS_PWD)

	gormConfig := gorm.Config{
		Logger: logger.OBDefault.LogMode(dsConfig.GetLoggerLevel()),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: constant.DB_SINGULAR_TABLE,
		}}
	db, err = gorm.Open(mysql.OpenObproxy(dsConfig.GetDSN()), &gormConfig)
	if err == nil {
		releaseDB(obproxyInstance)
		obproxyInstance = db
	} else {
		return nil, errors.Wrap(err, "load obproxy instance failed")
	}
	return obproxyInstance, nil
}

func LoadObproxyInstanceForHealthCheck(dsConfig *config.ObDataSourceConfig) (err error) {
	db, err := LoadTempObproxyInstance(dsConfig)
	if err != nil {
		return errors.Wrap(err, "load obproxy instance failed")
	}
	if err := db.Exec("show proxyconfig").Error; err != nil {
		return errors.Wrap(err, "check obproxy instance failed")
	}
	meta.OBPROXY_SYS_PWD = dsConfig.GetPassword()
	releaseDB(db)
	return err
}

func LoadTempObproxyInstance(dsConfig *config.ObDataSourceConfig) (db *gorm.DB, err error) {
	gormConfig := gorm.Config{
		Logger: logger.OBDefault.LogMode(dsConfig.GetLoggerLevel()),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: constant.DB_SINGULAR_TABLE,
		}}
	db, err = gorm.Open(mysql.OpenObproxy(dsConfig.GetDSN()), &gormConfig)
	if err != nil {
		return nil, errors.Wrap(err, "load temp obproxy instance failed")
	}
	return db, nil
}

func GetObproxyInstance() (*gorm.DB, error) {
	if obproxyInstance == nil {
		log.Info("obproxy instance is nil, load obproxy instance")
		if _, err := LoadObproxyInstance(); err != nil {
			return nil, err
		}
	}
	// health check
	if err := obproxyInstance.Exec("show proxyconfig").Error; err != nil {
		log.WithError(err).Warn("obproxy instance is not available")
		return nil, err
	}
	return obproxyInstance, nil
}

func releaseDB(preDB *gorm.DB) {
	// Delay release db
	if preDB != nil {
		db, err := preDB.DB()
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

			for db.Stats().InUse != 0 {
				log.Debug("pre db is using, wait for release")
				time.Sleep(time.Second)
			}
			db.Close()
		}()
	}
}
