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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/meta"
	sqlitedb "github.com/oceanbase/obshell/ob/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

func (s *AgentService) UpdatePort(mysqlPort, rpcPort int) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err := s.updateMysqlPort(tx, mysqlPort); err != nil {
			return err
		}
		return s.updateRpcPort(tx, rpcPort)
	})
}

func (s *AgentService) UpdatePortAndZone(mysqlPort int, rpcPort int, zone string) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err := s.updateMysqlPort(tx, mysqlPort); err != nil {
			return err
		}
		if err := s.updateRpcPort(tx, rpcPort); err != nil {
			return err
		}
		return s.updateZone(tx, zone)
	})
}

func (s *AgentService) updateMysqlPort(db *gorm.DB, port int) error {
	if port == 0 {
		return nil
	}
	if err := s.updateOBConfig(db, constant.CONFIG_MYSQL_PORT, fmt.Sprint(port)); err != nil {
		return err
	}
	meta.MYSQL_PORT = port
	return nil
}

func (s *AgentService) updateRpcPort(db *gorm.DB, port int) error {
	if port == 0 {
		return nil
	}
	if err := s.updateOBConfig(db, constant.CONFIG_RPC_PORT, fmt.Sprint(port)); err != nil {
		return err
	}
	meta.RPC_PORT = port
	return nil
}

func (s *AgentService) updateZone(db *gorm.DB, zone string) error {
	if err := s.updateOBConfig(db, constant.CONFIG_ZONE, zone); err != nil {
		return err
	}
	ocsAgent.Zone = zone
	return nil
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

func (s *AgentService) SetAgentPassword(password string) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	encrptyPassword, err := secure.Encrypt(password)
	if err != nil {
		return err
	}

	if err := sqliteDb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&sqlite.OcsInfo{
		Name:  constant.CONFIG_AGENT_PASSWORD,
		Value: encrptyPassword}).Error; err != nil {
		return err
	}

	meta.AGENT_PWD.SetPassword(password)
	secure.InvalidateAllSessions()
	return nil
}
