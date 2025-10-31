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

package global

import (
	"time"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
)

type VirtualClock struct {
	dbTime    time.Time
	localTime time.Time
	db        func() (*gorm.DB, error)
	syncSql   string
}

func (vc *VirtualClock) Now() (time.Time, error) {
	// current time  - local time
	offset := time.Since(vc.localTime)
	if offset > constant.DEFAULT_VIRTUAL_CLOCK_SYNC_INTERVAL {
		err := vc.Sync()
		return vc.dbTime, err
	}
	return vc.dbTime.Add(offset), nil
}

func (vc *VirtualClock) Sync() (err error) {
	db, err := vc.db()
	if err != nil {
		return
	}

	err = db.Raw(vc.syncSql).Scan(&vc.dbTime).Error
	if err != nil {
		return
	}
	vc.localTime = time.Now()
	return nil
}

type Time struct {
	ob_now VirtualClock
}

var TIME Time = Time{
	ob_now: VirtualClock{
		db:      oceanbase.GetRestrictedInstance,
		syncSql: "SELECT NOW(6)",
	},
}

func (t *Time) ObNow() (time.Time, error) {
	return t.ob_now.Now()
}

func (t *Time) SqliteNow() (time.Time, error) {
	return time.Now(), nil
}
