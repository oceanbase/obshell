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

package task

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/constant"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

var isClusterStatusInit bool = false

type clusterStatusMaintainer struct {
}

func (maintainer *clusterStatusMaintainer) setStatus(tx *gorm.DB, newStatus int, oldStatus int) error {
	clusterStatus := oceanbase.ClusterStatus{
		Id:     1,
		Status: newStatus,
	}
	resp := tx.Model(&clusterStatus).Where("id=? and status=?", 1, oldStatus).Updates(&clusterStatus)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return fmt.Errorf("failed to start maintenance: cluster is not %d", oldStatus)
	}
	return nil
}

func (maintainer *clusterStatusMaintainer) StartMaintenance(tx *gorm.DB) error {
	return maintainer.setStatus(tx, constant.CLUSTER_UNDER_MAINTENANCE, constant.CLUSTER_RUNNING)
}

func (maintainer *clusterStatusMaintainer) StopMaintenance(tx *gorm.DB) error {
	return maintainer.setStatus(tx, constant.CLUSTER_RUNNING, constant.CLUSTER_UNDER_MAINTENANCE)
}

func (maintainer *clusterStatusMaintainer) IsRunning() (bool, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return false, err
	}
	var status int
	if err := db.Model(&oceanbase.ClusterStatus{}).Select("status").Where("id=?", 1).First(&status).Error; err != nil {
		return false, err
	}
	return status == constant.CLUSTER_RUNNING, nil
}

func (maintainer *clusterStatusMaintainer) IsInited() (bool, error) {
	if !isClusterStatusInit {
		db, err := oceanbasedb.GetOcsInstance()
		if err != nil {
			return false, err
		}
		var clusterStatus *oceanbase.ClusterStatus
		if err := db.Model(&oceanbase.ClusterStatus{}).Where("id = 1").Scan(&clusterStatus).Error; err != nil {
			return false, err
		}
		isClusterStatusInit = (clusterStatus != nil)
	}
	return isClusterStatusInit, nil
}

type agentStatusMaintainer struct {
}

func (maintainer *agentStatusMaintainer) setStatus(tx *gorm.DB, newStatus int, oldStatus int) error {
	resp := tx.Model(&sqlite.OcsInfo{}).Where("name=? and value=?", constant.OCS_INFO_STATUS, oldStatus).Update("value", strconv.Itoa(newStatus))
	if resp.Error != nil {
		return resp.Error
	}
	if resp.RowsAffected == 0 {
		return fmt.Errorf("failed to start maintenance: agent status is not %d", oldStatus)
	}
	return nil
}

func (maintainer *agentStatusMaintainer) StartMaintenance(tx *gorm.DB) error {
	return maintainer.setStatus(tx, constant.AGENT_UNDER_MAINTENANCE, constant.AGENT_RUNNING)
}

func (maintainer *agentStatusMaintainer) StopMaintenance(tx *gorm.DB) error {
	return maintainer.setStatus(tx, constant.AGENT_RUNNING, constant.AGENT_UNDER_MAINTENANCE)
}

func (maintainer *agentStatusMaintainer) IsRunning() (bool, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return false, err
	}
	var status int
	if err := db.Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", constant.OCS_INFO_STATUS).First(&status).Error; err != nil {
		return false, err
	}
	return status == constant.AGENT_RUNNING, nil
}

func (maintainer *agentStatusMaintainer) IsInited() (bool, error) {
	return true, nil
}
