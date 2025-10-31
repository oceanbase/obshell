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
	"errors"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/lib/crypto"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

var (
	Crypter *crypto.RSACrypto
)

// Init will initialize secure module.
func Init() (err error) {
	if Crypter == nil {
		if err = RestoreKey(); err != nil {
			return New()
		}
	}
	return nil
}

// New will generate new RSA crypto.
func New() (err error) {
	Crypter, err = crypto.NewRSACrypto()
	if err != nil {
		return err
	}
	return Dump()
}

// Dump will dump private key into sqlite.
func Dump() error {
	return updateOCSInfo(constant.AGENT_PRIVATE_KEY, Crypter.Private())
}

// RestoreKey will restore key from sqlite.
func RestoreKey() error {
	key, err := getPrivateKey()
	if err != nil {
		return err
	}
	Crypter, err = crypto.NewRSACryptoFromKey(key)
	if err != nil {
		return err
	}
	log.Info("restore private key from sqlite successed")
	return nil
}

// Public will return thecurrent public key.
func Public() string {
	return Crypter.Public()
}

// LoadOceanbasePassword will load password from environment variable or sqlite.
func LoadOceanbasePassword(password *string) error {
	if password == nil {
		rootPwd, isSet := syscall.Getenv(constant.OB_ROOT_PASSWORD)
		if !isSet {
			return CheckObPasswordInSqlite()
		}
		log.Info("get password from environment variable")
		password = &rootPwd
	} else {
		log.Infof("get password from command line.")
	}

	// clear root password, avoid to cover sqlite when agent restart
	syscall.Unsetenv(constant.OB_ROOT_PASSWORD)
	meta.SetOceanbasePwd(*password)
	go dumpTempObPassword(*password)
	return nil
}

func dumpTempObPassword(pwd string) {
	log.Info("current password is temporary, will dump it into sqlite")
	for meta.OCEANBASE_PWD == pwd {
		if oceanbase.HasOceanbaseInstance() {
			if err := dumpPassword(); err != nil {
				log.WithError(err).Error("dump temporary password into sqlite failed")
			} else {
				log.Info("dump temporary password into sqlite successed")
			}
			break
		}
		time.Sleep(time.Second)
	}
}

// CheckObPasswordInSqlite will try connecting ob using password stored in sqlite.
func CheckObPasswordInSqlite() error {
	log.Info("retore password of root@sys from sqlite")
	password, err := getCipherPassword()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithError(err).Errorf("get '%s' failed", constant.CONFIG_ROOT_PWD)
			return err
		}
		log.Info("no password in sqlite")
		// No password in sqlite, no need to check.
		return nil
	}
	if password != "" {
		password, err = Decrypt(password)
		if err != nil {
			log.WithError(err).Error("decrypt password failed")
			return err
		}
	}
	meta.SetOceanbasePwd(password)
	return nil
}

// dumpPassword will dump encrypted password into sqlite.
func dumpPassword() error {
	passwrod := meta.OCEANBASE_PWD
	if meta.OCEANBASE_PWD != "" {
		cipherPassword, err := Crypter.Encrypt(meta.OCEANBASE_PWD)
		if err != nil {
			log.WithError(err).Error("encrypt password failed")
			return err
		}
		passwrod = cipherPassword
	}
	return updateOBConifg(constant.CONFIG_ROOT_PWD, passwrod)
}

func EncryptPwdInObConfigs(configs []sqlite.ObConfig) (err error) {
	for i := range configs {
		if configs[i].Name == constant.CONFIG_ROOT_PWD && configs[i].Value != "" {
			configs[i].Value, err = Crypter.Encrypt(configs[i].Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Decrypt(value string) (res string, err error) {
	if value != "" {
		res, err = Crypter.Decrypt(value)
	}
	return
}

func Encrypt(value string) (res string, err error) {
	res, err = Crypter.Encrypt(value)
	return
}
