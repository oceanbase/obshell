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

package observer

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
)

func HandleOBMeta() (err error) {
	log.Info("start to handle ob meta")
	var count int
	for {
		if oceanbase.IsConnecting() {
			// If the connection is being established, wait for it to complete
		} else if _, err = oceanbase.GetOcsInstance(); err == nil {
			break
		} else {
			// Because initConnection() only returns when there's a password error or on success
			// and the password must be correct at this point, so we can wait here until the connection is successful
			if count%100 == 0 {
				log.WithError(err).Error("get ocs db connection failed")
			}
		}
		count++
		time.Sleep(10 * time.Millisecond)
	}

	log.Info("try to start migrate table")
	if err = oceanbase.AutoMigrateObTables(true); err != nil {
		log.WithError(err).Error("auto migrate ob tables failed")
		return
	}

	return nil
}
