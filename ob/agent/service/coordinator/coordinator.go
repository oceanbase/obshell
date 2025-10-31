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

package coordinator

import (
	"time"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

type CoordinatorService struct{}

// MaintainerDO is the data object of maintainer
// Gap is the time difference between the current time and the last active time of the maintainer, in seconds
// IsActive is whether the maintainer is active. If the gap is less than MAINTAINER_MAX_ACTIVE_TIME_SEC, it is active
type MaintainerDO struct {
	oceanbase.TaskMaintainer
	Gap      float64
	IsActive bool
}

func (s *CoordinatorService) GetMaintainerFromOb() (m MaintainerDO, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("select *, TimeStampDiff(Microsecond, active_time, now(6)) as gap,  TimeStampDiff(second, active_time, now(6)) < ? as is_active from task_maintainer where id = 1", constant.MAINTAINER_MAX_ACTIVE_TIME_SEC).Scan(&m).Error
	return
}

func (s *CoordinatorService) UpdateMaintainerToOb(maintainer oceanbase.TaskMaintainer) error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	maintainer.AgentTime = time.Now().Unix()
	resp := oceanbaseDb.Model(&oceanbase.TaskMaintainer{}).Where("id = 1 and now(6) - active_time >= ?", constant.MAINTAINER_MAX_ACTIVE_TIME_SEC).Updates(&maintainer)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.Occur(errors.ErrGormNoRowAffected, "update maintainer failed")
	}
	return nil
}

func (s *CoordinatorService) RenewalMaintainer(maintainer oceanbase.TaskMaintainer) error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	maintainer.AgentTime = time.Now().Unix()
	maintainer.ActiveTime = time.Time{}
	resp := oceanbaseDb.Model(&oceanbase.TaskMaintainer{}).
		Where("id = 1 and agent_ip = ? and agent_port = ?", maintainer.AgentIp, maintainer.AgentPort).
		Updates(&maintainer)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return errors.Occur(errors.ErrGormNoRowAffected, "update maintainer failed")
	}
	return nil

}
