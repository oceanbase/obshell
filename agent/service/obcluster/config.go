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

package obcluster

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/agent/secure"
)

func (s *ObserverService) GetObConfigByName(name string) (config sqlite.ObConfig, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.ObConfig{}).Where("name=?", name).Find(&config).Error
	return
}

func (s *ObserverService) GetOBParatemerByName(name string, value interface{}) (err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return
	}

	err = db.Table(ob_parameters_view).Select("value").Where("name = ?", name).Scan(value).Error
	return
}

func (s *ObserverService) GetOBStringParatemerByName(name string) (value string, err error) {
	err = s.GetOBParatemerByName(name, &value)
	return
}

func (s *ObserverService) GetOBIntParatemerByName(name string) (value int, err error) {
	err = s.GetOBParatemerByName(name, &value)
	return
}

func (s *ObserverService) GetObConfigMap() (res map[string]string, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	config := make([]sqlite.ObConfig, 0)
	if err = sqliteDb.Model(&sqlite.ObConfig{}).Find(&config).Error; err != nil {
		return
	}
	res = make(map[string]string, 0)
	for _, item := range config {
		res[item.Name] = item.Value
	}
	return
}

func (s *ObserverService) GetObServerConfig(agentInfo meta.AgentInfoInterface) (config []sqlite.ObServerConfig, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObServerConfig{}).Where("agent_ip=? and agent_port=?", agentInfo.GetIp(), agentInfo.GetPort()).Find(&config).Error
	return
}

func (s *ObserverService) GetObZoneConfig(zone string) (config []sqlite.ObZoneConfig, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObZoneConfig{}).Where("zone=?", zone).Find(&config).Error
	return
}

func (s *ObserverService) GetObGlobalConfig() (config []sqlite.ObGlobalConfig, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObGlobalConfig{}).Find(&config).Error
	return
}

func (s *ObserverService) GetObGlobalConfigByName(name string) (config sqlite.ObGlobalConfig, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObGlobalConfig{}).Where("name=?", name).Find(&config).Error
	return
}

func (s *ObclusterService) UpdateClusterConfig(config map[string]string, deleteAll bool) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err = tx.Delete(&sqlite.ObGlobalConfig{}, "is_cluster=true").Error; err != nil {
			return err
		}

		configs := make([]*sqlite.ObGlobalConfig, 0)
		for name, value := range config {
			configs = append(configs, &sqlite.ObGlobalConfig{
				Name:      name,
				IsCluster: true,
				Value:     value,
			})
		}
		if err = tx.Create(configs).Error; err != nil {
			return err
		}

		if value, ok := config[constant.CONFIG_ROOT_PWD]; ok {
			if err = secure.UpdateObPasswordInTransaction(tx, value); err != nil {
				return errors.Wrap(err, "secure dump root password failed")
			}
		}
		return nil
	})
}

func (s *ObserverService) UpdateGlobalConfig(config map[string]string, deleteAll bool) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err = tx.Delete(&sqlite.ObGlobalConfig{}, "is_cluster=false").Error; err != nil {
			return err
		}

		configs := make([]*sqlite.ObGlobalConfig, 0)
		for name, value := range config {
			configs = append(configs, &sqlite.ObGlobalConfig{
				Name:      name,
				IsCluster: false,
				Value:     value,
			})
		}
		if err = tx.Create(configs).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *ObserverService) UpdateZoneConfig(config map[string]string, zoneList []string, deleteAll bool) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if deleteAll {
			if err = tx.Delete(&sqlite.ObZoneConfig{}, "zone in ?", zoneList).Error; err != nil {
				return err
			}
		}

		configs := make([]*sqlite.ObZoneConfig, 0)
		for name, value := range config {
			for _, zone := range zoneList {
				configs = append(configs, &sqlite.ObZoneConfig{
					Zone:  zone,
					Name:  name,
					Value: value,
				})
			}
		}
		if err = tx.Create(configs).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *ObserverService) UpdateServerConfig(config map[string]string, agentList []meta.AgentInfoInterface, deleteAll bool) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if deleteAll {
			model := &sqlite.ObServerConfig{}
			for _, agent := range agentList {
				if err = tx.Delete(&model, "agent_ip = ? and agent_port = ?", agent.GetIp(), agent.GetPort()).Error; err != nil {
					return err
				}
			}
		}

		configs := make([]*sqlite.ObServerConfig, 0)
		for name, value := range config {
			for _, agent := range agentList {
				configs = append(configs, &sqlite.ObServerConfig{
					AgentIp:   agent.GetIp(),
					AgentPort: agent.GetPort(),
					Name:      name,
					Value:     value,
				})
			}
		}
		if err = tx.Create(configs).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *ObserverService) PutObConfig(params map[string]string) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err := s.clearObConfig(tx); err != nil {
			return err
		}
		for name, value := range params {
			if err = tx.Create(&sqlite.ObConfig{Name: name, Value: value}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ObserverService) PatchObConfig(params map[string]string) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		for name, value := range params {
			if err = tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&sqlite.ObConfig{
				Name:  name,
				Value: value}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ObserverService) PatchObserverConfig(configs []*sqlite.ObServerConfig) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObServerConfig{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "agent_port"}, {Name: "agent_ip"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&configs).Error
	return
}

func (s *ObserverService) PatchObZoneConfig(configs []*sqlite.ObZoneConfig) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObZoneConfig{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "zone"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&configs).Error
	return
}

func (s *ObserverService) PatchGlobalConfig(globalConfig []*sqlite.ObGlobalConfig) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.ObGlobalConfig{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(globalConfig).Error
	return
}

func (s *ObserverService) ClearObConfig() error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return s.clearObConfig(db)
}

func (s *ObserverService) clearObConfig(db *gorm.DB) error {
	return db.Delete(&sqlite.ObConfig{}, "1=1").Error
}

func (s *ObserverService) deleteObServerConfig(db *gorm.DB, agent meta.AgentInfoInterface) error {
	if agent == nil {
		return db.Delete(&sqlite.ObServerConfig{}, "1=1").Error
	} else {
		return db.Delete(&sqlite.ObServerConfig{}, "agent_ip = ? and agent_port = ?", agent.GetIp(), agent.GetPort()).Error
	}
}

func (s *ObserverService) deleteObZoneConifg(db *gorm.DB, zone string) error {
	if zone == "" {
		return db.Delete(&sqlite.ObZoneConfig{}, "1=1").Error
	} else {
		return db.Delete(&sqlite.ObZoneConfig{}, "zone = ?", zone).Error
	}
}

func (s *ObserverService) ClearObServerConfig() error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return s.deleteObServerConfig(db, nil)
}

func (s *ObserverService) DeleteObServerConfig(agent meta.AgentInfoInterface) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return s.deleteObServerConfig(db, agent)
}

func (s *ObserverService) ClearObZoneConifg() error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return s.deleteObZoneConifg(db, "")
}

func (s *ObserverService) DeleteObZoneConifg(zone string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return s.deleteObZoneConifg(db, zone)
}

func (s *ObserverService) ClearObserverAndZoneConfig(agent meta.AgentInfoWithZoneInterface) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.deleteObServerConfig(tx, agent); err != nil {
			return err
		}
		if err := s.deleteObZoneConifg(tx, agent.GetZone()); err != nil {
			return err
		}
		return tx.Delete(&sqlite.ObConfig{}, "1=1").Error
	})
}

func (s *ObserverService) ClearGlobalConfig() error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.deleteObServerConfig(tx, nil); err != nil {
			return err
		}
		if err := s.deleteObZoneConifg(tx, ""); err != nil {
			return err
		}
		return tx.Delete(&sqlite.ObGlobalConfig{}, "1=1").Error
	})
}

func (s *ObserverService) GetObServerConfigMap(agentInfo meta.AgentInfoInterface) (map[string]sqlite.ObConfig, error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	var data []sqlite.ObServerConfig
	if err = sqliteDb.Model(&sqlite.ObServerConfig{}).Where("agent_ip=? and agent_port=?", agentInfo.GetIp(), agentInfo.GetPort()).Find(&data).Error; err != nil {
		return nil, err
	}

	config := make(map[string]sqlite.ObConfig, 0)
	for _, item := range data {
		config[item.Name] = sqlite.ObConfig{
			Value:     item.Value,
			GmtModify: item.GmtModify,
		}
	}
	return config, nil
}

func (s *ObserverService) GetObZoneConfigMap(zone string) (map[string]sqlite.ObConfig, error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	var data []sqlite.ObZoneConfig
	if err = sqliteDb.Model(&sqlite.ObZoneConfig{}).Where("zone=?", zone).Find(&data).Error; err != nil {
		return nil, err
	}

	config := make(map[string]sqlite.ObConfig, 0)
	for _, item := range data {
		config[item.Name] = sqlite.ObConfig{
			Value:     item.Value,
			GmtModify: item.GmtModify,
		}
	}
	return config, nil
}

func (s *ObserverService) GetObGlobalConfigMap() (map[string]sqlite.ObConfig, error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	var data []sqlite.ObGlobalConfig
	if err = sqliteDb.Model(&sqlite.ObGlobalConfig{}).Find(&data).Error; err != nil {
		return nil, err
	}

	config := make(map[string]sqlite.ObConfig, 0)
	for _, item := range data {
		config[item.Name] = sqlite.ObConfig{
			Value:     item.Value,
			GmtModify: item.GmtModify,
		}
	}
	return config, nil
}
