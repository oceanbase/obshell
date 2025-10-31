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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/config"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
)

func VerifyTimeStamp(ts string, curTs int64) error {
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return errors.Occur(errors.ErrSecurityAuthenticationTimestampInvalid, ts, err.Error())
	}
	if curTs > int64(tsInt) {
		log.Warnf("auth expired at: %v, current: %v", ts, curTs)
		return errors.Occur(errors.ErrSecurityAuthenticationExpired)
	}
	return nil
}

func VerifyAuth(pwd string, ts string, curTs int64) error {
	if pwd != "" {
		if err := VerifyTimeStamp(ts, curTs); err != nil {
			return err
		}
	}

	if pwd != meta.OCEANBASE_PWD {
		if oceanbase.HasOceanbaseInstance() {
			if err := VerifyOceanbasePassword(pwd); err != nil {
				return err
			}
			if err := dumpPassword(); err != nil {
				log.WithError(err).Error("dump password failed")
				return errors.Wrap(err, "dump password failed")
			}
		} else {
			return errors.Occur(errors.ErrSecurityAuthenticationIncorrectOceanbasePassword)
		}
	}
	return nil
}

func VerifyOceanbasePassword(password string) error {
	if err := oceanbase.LoadOceanbaseInstanceForTest(config.NewObMysqlDataSourceConfig().SetPassword(password)); err != nil {
		if strings.Contains(err.Error(), "Access denied") {
			return errors.Occur(errors.ErrSecurityAuthenticationIncorrectOceanbasePassword)
		}
		if meta.OCEANBASE_PWD != password {
			return errors.Occur(errors.ErrSecurityAuthenticationIncorrectOceanbasePassword)
		}
		log.WithError(err).Error("unexpected db error")
		return nil
	}
	meta.SetOceanbasePwd(password)
	if err := oceanbase.LoadOceanbaseInstance(); err != nil {
		log.WithError(err).Error("unexpected db error")
		return err
	}
	dumpPassword()
	return nil
}
