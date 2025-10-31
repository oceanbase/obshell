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

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
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
		if newStatus != task.GLOBAL_MAINTENANCE {
			var nowStatus int
			if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&oceanbase.ClusterStatus{}).Select("status").Where("id=?", 1).First(&nowStatus).Error; err != nil {
				return err
			} else if nowStatus == newStatus {
				return nil
			}
		}

		var taskService ClusterTaskService
		if dag, _ := taskService.GetLastMaintenanceDag(); dag != nil {
			// Try to get the maintenance dag name
			return errors.Occur(errors.ErrObClusterUnderMaintenanceWithDag, dag.GetName())
		}
		return errors.Occur(errors.ErrObClusterUnderMaintenance)
	}
	return nil
}

func (maintainer *clusterStatusMaintainer) StartMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	if !dag.IsMaintenance() {
		return nil
	}
	if err := maintainer.setStatus(tx, dag.GetMaintenanceType(), task.NOT_UNDER_MAINTENANCE); err != nil {
		return err
	}
	return nil
}

func (maintainer *clusterStatusMaintainer) UpdateMaintenanceTask(tx *gorm.DB, dag *task.Dag) error {
	if !dag.IsMaintenance() || dag.GetMaintenanceType() == task.GLOBAL_MAINTENANCE {
		return nil
	}

	return nil
}

func (maintainer *clusterStatusMaintainer) StopMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	switch dag.GetMaintenanceType() {
	case task.GLOBAL_MAINTENANCE:
		return maintainer.setStatus(tx, task.NOT_UNDER_MAINTENANCE, task.GLOBAL_MAINTENANCE)
	default:
		return nil
	}
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
	return status == task.NOT_UNDER_MAINTENANCE, nil
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
		if newStatus == task.NOT_UNDER_MAINTENANCE {
			var nowStatus int
			if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", "status").First(&nowStatus).Error; err != nil {
				return err
			} else if nowStatus == newStatus {
				return nil
			}
		}
		return errors.Occur(errors.ErrAgentUnderMaintenance, meta.OCS_AGENT.String())
	}
	return nil
}

func (maintainer *agentStatusMaintainer) StartMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	if !dag.IsMaintenance() {
		return nil
	}
	fmt.Println("StartMaintenance", dag.GetMaintenanceType())
	switch dag.GetMaintenanceType() {
	case task.GLOBAL_MAINTENANCE:
		return maintainer.setStatus(tx, dag.GetMaintenanceType(), task.NOT_UNDER_MAINTENANCE)
	default:
		return nil
	}
}

func (maintainer *agentStatusMaintainer) UpdateMaintenanceTask(tx *gorm.DB, dag *task.Dag) error {
	return nil
}

func (maintainer *agentStatusMaintainer) StopMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	switch dag.GetMaintenanceType() {
	case task.GLOBAL_MAINTENANCE:
		return maintainer.setStatus(tx, task.NOT_UNDER_MAINTENANCE, task.GLOBAL_MAINTENANCE)
	default:
		return nil
	}
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
	return status == task.NOT_UNDER_MAINTENANCE, nil
}

func (maintainer *agentStatusMaintainer) IsInited() (bool, error) {
	return true, nil
}
