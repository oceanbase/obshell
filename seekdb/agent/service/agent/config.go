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

package agent

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

// Serialize port persistence and publication across all AgentService instances.
var mysqlPortMu sync.Mutex

func (s *AgentService) UpdatePort(mysqlPort int) error {
	if mysqlPort == 0 {
		return nil
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	mysqlPortMu.Lock()
	defer mysqlPortMu.Unlock()

	err = sqliteDb.Transaction(func(tx *gorm.DB) error {
		return s.updateOBConfig(tx, constant.CONFIG_MYSQL_PORT, fmt.Sprint(mysqlPort))
	})
	if err == nil {
		meta.SetMysqlPort(mysqlPort)
	}
	return err
}

func (s *AgentService) SyncMysqlPort() error {
	hasLocalStore, err := path.HasLocalStore()
	if err != nil || !hasLocalStore {
		return err
	}
	obDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	var mysqlPort int
	if err := obDb.Raw("SELECT MYSQL_PORT()").Scan(&mysqlPort).Error; err != nil {
		return err
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return s.syncMysqlPort(sqliteDb, mysqlPort)
}

func (s *AgentService) syncMysqlPort(db *gorm.DB, mysqlPort int) error {
	if mysqlPort < 0 || mysqlPort > 65535 {
		return fmt.Errorf("invalid seekdb MySQL port: %d", mysqlPort)
	}
	mysqlPortMu.Lock()
	defer mysqlPortMu.Unlock()

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := s.updateOBConfig(tx, constant.CONFIG_MYSQL_PORT, fmt.Sprint(mysqlPort)); err != nil {
			return err
		}
		return tx.Model(&sqlite.AllAgent{}).
			Where("ip = ? AND port = ?", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()).
			Update("mysql_port", mysqlPort).Error
	})
	if err == nil {
		meta.SetMysqlPort(mysqlPort)
	}
	return err
}

func (s *AgentService) updateObserverUser(db *gorm.DB, user string) error {
	return s.updateOBConfig(db, constant.CONFIG_USER, user)
}

func (s *AgentService) updateObserverDataDir(db *gorm.DB, dataDir string) error {
	return s.updateOBConfig(db, constant.CONFIG_DATA_DIR, dataDir)
}

func (s *AgentService) updateObserverRedoDir(db *gorm.DB, redoDir string) error {
	return s.updateOBConfig(db, constant.CONFIG_REDO_DIR, redoDir)
}

func (s *AgentService) updateObserverObVersion(db *gorm.DB, obVersion string) error {
	return s.updateOBConfig(db, constant.CONFIG_OB_VERSION, obVersion)
}

func (s *AgentService) updateOBConfig(db *gorm.DB, name string, value string) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&sqlite.ObConfig{
		Name:  name,
		Value: value}).Error
}

func (s *AgentService) getOBConifg(db *gorm.DB, name string, value interface{}) error {
	err := db.Model(&sqlite.ObConfig{}).Select("value").Where("name = ?", name).First(value).Error
	if err == gorm.ErrRecordNotFound {
		if old, exist := constant.OB_CONFIG_COMPATIBLE_MAP[name]; exist {
			return db.Model(&sqlite.ObConfig{}).Select("value").Where("name = ?", old).Scan(value).Error
		}
	}
	return err
}

func (s *AgentService) GetObConfig(key string, value interface{}) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	err = db.Model(&sqlite.ObConfig{}).Select("value").Where("name = ?", key).First(value).Error
	return err
}
