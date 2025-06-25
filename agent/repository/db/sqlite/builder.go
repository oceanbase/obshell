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

package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	sqliteDriver "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

func createGormDbByConfig(datasourceConfig config.SqliteDataSourceConfig) (*gorm.DB, error) {
	datasourceConfig, err := InitSqliteDataSourceConfig(datasourceConfig)
	if err != nil {
		log.WithError(err).Info("initialize sqlite config failed")
		return nil, err
	}
	err = CheckFilepath(datasourceConfig.DataDir)
	if err != nil {
		log.WithError(err).Info("check sqlite data dir failed")
		return nil, err
	}
	dsn := GenerateSqliteDsn(datasourceConfig)
	db, err := gorm.Open(sqliteDriver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(datasourceConfig.GetLoggerLevel()),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: constant.DB_SINGULAR_TABLE,
		},
	})
	if err != nil {
		log.WithError(err).Info("open sqlite failed")
		return nil, err
	}
	sqliteDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqliteDb.SetMaxIdleConns(datasourceConfig.MaxIdleConns)
	sqliteDb.SetMaxOpenConns(datasourceConfig.MaxOpenConns)
	sqliteDb.SetConnMaxLifetime(time.Duration(datasourceConfig.ConnMaxLifetime))
	log.Info("open sqlite succeed")
	return db, nil
}

func CheckFilepath(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.WithError(err).Info("dir not exist")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.WithError(err).Info("create dir failed")
			return err
		}
	}
	return nil
}

func InitSqliteDataSourceConfig(dsConfig config.SqliteDataSourceConfig) (config.SqliteDataSourceConfig, error) {
	if dsConfig.DataDir == "" {
		log.Error("sqlite data dir cannot be empty")
		return dsConfig, errors.Occur(errors.ErrCommonUnexpected, "sqlite data dir is empty")
	}
	defaultDsConfig := config.DefaultSqliteDataSourceConfig()
	defaultDsConfigValue := reflect.ValueOf(defaultDsConfig)
	dsConfigType := reflect.TypeOf(dsConfig)
	dsConfigValue := reflect.ValueOf(&dsConfig).Elem()
	for i := 0; i < dsConfigType.NumField(); i++ {
		value := dsConfigValue.Field(i).Interface()
		vType := reflect.TypeOf(value)
		switch vType.Kind() {
		case reflect.Int:
			if dsConfigValue.Field(i).IsZero() {
				dsConfigValue.Field(i).SetInt(defaultDsConfigValue.Field(i).Int())
			}
		case reflect.String:
			if dsConfigValue.Field(i).String() == "" {
				dsConfigValue.Field(i).SetString(defaultDsConfigValue.Field(i).String())
			}
		}
	}
	return dsConfig, nil
}

func GenerateSqliteDsn(datasourceConfig config.SqliteDataSourceConfig) string {
	return fmt.Sprintf(datasourceConfig.DsnTemplate,
		datasourceConfig.DataDir,
		datasourceConfig.Cache,
		datasourceConfig.FK,
	)
}

var SqliteTables = []interface{}{
	sqlite.AllAgent{},
	sqlite.ObSysParameter{},
	sqlite.OcsInfo{},
	sqlite.ObproxyInfo{},
	sqlite.ObGlobalConfig{},
	sqlite.ObZoneConfig{},
	sqlite.ObServerConfig{},
	sqlite.ObConfig{},
	sqlite.OcsConfig{},
	sqlite.OcsToken{},
	sqlite.TaskMapping{},
	sqlite.SubtaskInstance{},
	sqlite.SubTaskLog{},
	sqlite.DagInstance{},
	sqlite.NodeInstance{},
	sqlite.UpgradePkgInfo{},
	sqlite.UpgradePkgChunk{},
}

// MigrateSqliteTables will check if the sqlite tables exist, if not, it will create them.
// Sqlite tables migration is only required for first start and upgrade
// The first start is based on whether the `ip` of sqlite's ocs_info table is "".
func MigrateSqliteTables(forUpgrade bool) (err error) {
	if ocs_db_sqlite == nil {
		return errors.Occur(errors.ErrAgentSqliteDBNotInit)
	}

	// Check if sqlite tables exist.
	for _, table := range SqliteTables {
		if !ocs_db_sqlite.Migrator().HasTable(table) {
			return ocs_db_sqlite.AutoMigrate(SqliteTables...)
		}
	}

	// Check if agent has been initialized.
	var ip string
	err = ocs_db_sqlite.Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", constant.OCS_INFO_IP).Scan(&ip).Error
	if err != nil {
		return err
	}
	if ip == "" || forUpgrade {
		log.Info("register sqlite tables")
		return ocs_db_sqlite.AutoMigrate(SqliteTables...)
	}

	return nil
}
