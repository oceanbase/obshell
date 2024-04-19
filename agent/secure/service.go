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

package secure

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	oceanbasemodel "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	sqlitemodel "github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

type SecureConfig struct {
	AuthExpiredDuration time.Duration
}

var (
	// secureConfig holds the security configurations used for secure communication and data handling.
	secureConfig = SecureConfig{}

	// These Models represent the SQLite data model for storing data.
	obConfigModel    = &sqlitemodel.ObConfig{}
	ocsConfigModel   = &sqlitemodel.OcsConfig{}
	ocsInfoModel     = &sqlitemodel.OcsInfo{}
	agentModelSqlite = &sqlitemodel.AllAgent{}
	agentModelOB     = &oceanbasemodel.AllAgent{}
)

func getOCSConfig(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	if err = db.Model(ocsConfigModel).Where("name=?", key).First(&value).Error; err != nil {
		return
	}
	return
}

func getOBConifg(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(obConfigModel).Select("value").Where("name=?", key).First(value).Error
	return
}

func updateOBConifg(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return updateOBConifgInTransaction(db, key, value)
}

func updateOBConifgInTransaction(tx *gorm.DB, key string, value interface{}) (err error) {
	data := map[string]interface{}{
		"name":  key,
		"value": value,
	}
	err = tx.Model(obConfigModel).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(data).Error
	return
}

func getOCSInfo(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(ocsInfoModel).Select("value").Where("name=?", key).First(value).Error
	return
}

func updateOCSInfo(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	data := map[string]interface{}{
		"name":  key,
		"value": value,
	}
	err = db.Model(ocsInfoModel).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(data).Error
	return
}

func getCipherPassword() (cipherPassword string, err error) {
	err = getOBConifg(constant.CONFIG_ROOT_PWD, &cipherPassword)
	return
}

func getPrivateKey() (key string, err error) {
	err = getOCSInfo(constant.AGENT_PRIVATE_KEY, &key)
	return
}

func UpdateObPassword(password string) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return UpdateObPasswordInTransaction(db, password)
}

func UpdateObPasswordInTransaction(tx *gorm.DB, password string) (err error) {
	val, err := Decrypt(password)
	if err != nil {
		return
	}
	if err = updateOBConifgInTransaction(tx, constant.CONFIG_ROOT_PWD, password); err != nil {
		return
	}
	meta.SetOceanbasePwd(val)
	return
}

func getAgentByPublicKey(publicKey string) (agentInfo meta.AgentInfo, err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = db.Model(agentModelOB).Where("public_key = ?", publicKey).Find(&agentInfo).Error
	return
}

func getAllAgentsInfo() (agents []meta.AgentInfo, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(agentModelSqlite).Find(&agents).Error
	return
}

func getPublicKeyByAgentInfo(agent meta.AgentInfoInterface) (pk string, err error) {
	db, _ := oceanbasedb.GetOcsInstance()
	if db == nil {
		db, err = sqlitedb.GetSqliteInstance()
	}
	if err != nil {
		return
	}
	err = db.Model(agentModelSqlite).Select("public_key").Where("ip=? and port=?", agent.GetIp(), agent.GetPort()).Find(&pk).Error
	return
}

func updateAgentPublicKey(agent meta.AgentInfoInterface, publicKey string) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Model(agentModelSqlite).Where("ip=? and port=?", agent.GetIp(), agent.GetPort()).Update("public_key", publicKey).Error
}

func getTokenByAgentInfo(agent meta.AgentInfoInterface) (token string, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlitemodel.OcsToken{}).Select("token").Where("ip=? and port=?", agent.GetIp(), agent.GetPort()).Find(&token).Error
	return
}

func updateToken(agent meta.AgentInfoInterface, token string) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	ocsToken := sqlitemodel.OcsToken{
		Ip:    agent.GetIp(),
		Port:  agent.GetPort(),
		Token: token,
	}
	return sqliteDb.Model(&ocsToken).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ip"}, {Name: "port"}},
		DoUpdates: clause.AssignmentColumns([]string{"token"}),
	}).Create(&ocsToken).Error
}

func DeleteToken(agent meta.AgentInfoInterface) (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return sqliteDb.Model(&sqlitemodel.OcsToken{}).Delete("ip=? and port=?", agent.GetIp(), agent.GetPort()).Error
}

func getAuthExpiredDuration() time.Duration {
	if secureConfig.AuthExpiredDuration == 0 {
		var config string
		err := getOCSConfig(constant.AGENT_AUTH_EXPIRED_DURATION, &config)
		if err != nil {
			secureConfig.AuthExpiredDuration = constant.DEFAULT_AUTH_EXPIRED_DURATION
		} else {
			duration, err := time.ParseDuration(config)
			if err != nil {
				secureConfig.AuthExpiredDuration = constant.DEFAULT_AUTH_EXPIRED_DURATION
			} else {
				secureConfig.AuthExpiredDuration = duration
			}
		}
	}
	return secureConfig.AuthExpiredDuration
}
