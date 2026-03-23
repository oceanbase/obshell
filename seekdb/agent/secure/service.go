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

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	sqlitemodel "github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

type SecureConfig struct {
	AuthExpiredDuration time.Duration
}

var (
	// secureConfig holds the security configurations used for secure communication and data handling.
	secureConfig = SecureConfig{}

	// These Models represent the SQLite data model for storing data.
	obConfigModel  = &sqlitemodel.ObConfig{}
	ocsConfigModel = &sqlitemodel.OcsConfig{}
	ocsInfoModel   = &sqlitemodel.OcsInfo{}
)

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
	return updateOCSInfoInTransaction(db, key, value)
}

func updateOCSInfoInTransaction(tx *gorm.DB, key string, value interface{}) (err error) {
	data := map[string]interface{}{
		"name":  key,
		"value": value,
	}
	return tx.Model(ocsInfoModel).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(data).Error
}

// reEncryptAndSavePasswordsInTransaction encrypts root password with current Crypter and saves via tx.
func reEncryptAndSavePasswordsInTransaction(tx *gorm.DB, rootPwd string) error {
	if rootPwd != "" {
		cipher, err := Crypter.Encrypt(rootPwd)
		if err != nil {
			return err
		}
		return updateOBConifgInTransaction(tx, constant.CONFIG_ROOT_PWD, cipher)
	}
	return nil
}

// reGenerateRSACryptoAndreEncrypt persists the current Crypter (private key, public key in sqlite) and
// re-encrypted root password in a single sqlite transaction. Caller must set Crypter to the new key before calling.
func reGenerateRSACryptoAndreEncrypt(rootPwdPlain string) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) error {
		if err := updateOCSInfoInTransaction(tx, constant.AGENT_PRIVATE_KEY, Crypter.Private()); err != nil {
			return err
		}
		if err := reEncryptAndSavePasswordsInTransaction(tx, rootPwdPlain); err != nil {
			return err
		}
		if rootPwdPlain == "" {
			_ = updateOBConifgInTransaction(tx, constant.CONFIG_ROOT_PWD, "")
		}
		return nil
	})
}

func getCipherPassword() (cipherPassword string, err error) {
	err = getOBConifg(constant.CONFIG_ROOT_PWD, &cipherPassword)
	return
}

func getPrivateKey() (key string, err error) {
	err = getOCSInfo(constant.AGENT_PRIVATE_KEY, &key)
	return
}
