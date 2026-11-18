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
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
)

func HandleOBMeta(ctx context.Context) {
	log.Info("start to handle ob meta")
	var count int
	for {
		if ctx.Err() != nil {
			return
		}
		if oceanbase.IsConnecting() {
			// If the connection is being established, wait for it to complete
		} else if _, err := oceanbase.GetInstance(); err == nil {
			break
		} else {
			// Because initConnection() only returns when there's a password error or on success
			// and the password must be correct at this point, so we can wait here until the connection is successful
			if count%100 == 0 {
				log.WithError(err).Error("get ocs db connection failed")
			}
		}
		count++
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}

	if err := retryOBMeta(ctx, time.Second, agentService.SyncMysqlPort, log.NewEntry(log.StandardLogger())); err != nil {
		log.WithError(err).Debug("stopped handling ob meta")
	}
}

func retryOBMeta(ctx context.Context, interval time.Duration, syncMeta func() error, logger *log.Entry) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := syncMeta(); err == nil {
			return nil
		} else {
			logger.WithError(err).Error("sync seekdb MySQL port failed, retrying")
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
