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

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	bo "github.com/oceanbase/obshell/agent/repository/model/bo"
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
		if newStatus != task.GLOBAL_MAINTENANCE {
			var nowStatus int
			if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&oceanbase.ClusterStatus{}).Select("status").Where("id=?", 1).First(&nowStatus).Error; err != nil {
				return err
			} else if nowStatus == newStatus {
				return nil
			}
		}
		return fmt.Errorf("failed to start maintenance: agent status is not %d", oldStatus)
	}
	return nil
}

func (maintainer *clusterStatusMaintainer) getPartialLockForUpdate(tx *gorm.DB, paritialLock *bo.PartialMaintenance) error {
	return tx.Set("gorm:query_option", "FOR UPDATE").Model(paritialLock).Where("lock_name=? and lock_type=?", paritialLock.LockName, paritialLock.LockType).First(paritialLock).Error
}

func (maintainer *clusterStatusMaintainer) releasePartialLock(tx *gorm.DB, paritialLock *bo.PartialMaintenance) error {
	return tx.Model(paritialLock).Where("lock_name=? and lock_type=?", paritialLock.LockName, paritialLock.LockType).Update("gmt_locked", ZERO_TIME).Error
}

func (maintainer *clusterStatusMaintainer) holdPartialLock(tx *gorm.DB, dag task.Maintainer) error {
	if dag.GetMaintenanceType() == task.GLOBAL_MAINTENANCE {
		return nil
	}

	paritialLock := bo.PartialMaintenance{
		LockName: dag.GetMaintenanceKey(),
		LockType: dag.GetMaintenanceType(),
	}
	err := tx.Model(&paritialLock).Create(&paritialLock).Error
	if err != nil {
		if dbErr, ok := err.(*mysql.MySQLError); ok {
			if dbErr.Number == 1062 {
				// If there is a unique key conflict (error number 1062), get the lock for update
				if err1 := maintainer.getPartialLockForUpdate(tx, &paritialLock); err1 != nil {
					return err1
				} else if paritialLock.GmtLocked.After(ZERO_TIME) {
					return fmt.Errorf("failed to start maintenance: %s is already under maintenance", paritialLock.LockName)
				}
				return nil
			}
		}
		return err
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
	return maintainer.holdPartialLock(tx, dag)
}

func (maintainer *clusterStatusMaintainer) UpdateMaintenanceTask(tx *gorm.DB, dag *task.Dag) error {
	if !dag.IsMaintenance() || dag.GetMaintenanceType() == task.GLOBAL_MAINTENANCE {
		return nil
	}

	lock := bo.PartialMaintenance{
		LockName: dag.GetMaintenanceKey(),
		LockType: dag.GetMaintenanceType(),
	}
	if err := maintainer.getPartialLockForUpdate(tx, &lock); err != nil {
		return err
	}

	if lock.DagID > 0 && dag.IsFail() && dag.GetID() != lock.DagID {
		gid := task.ConvertIDToGenericID(dag.GetID(), false)
		oldGid := task.ConvertIDToGenericID(lock.DagID, false)
		return fmt.Errorf("%s has already executed task %s. '%s: %s' cannot be executed. Please submit a new request", lock.LockName, oldGid, gid, dag.GetName())
	}

	lockDo := oceanbase.PartialMaintenance{
		Id:    lock.Id,
		DagID: dag.GetID(),
		Count: lock.Count + 1,
	}
	return tx.Model(&lockDo).Where("id=?", lock.Id).Updates(&lockDo).Error
}

func (maintainer *clusterStatusMaintainer) StopMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	switch dag.GetMaintenanceType() {
	case task.GLOBAL_MAINTENANCE:
		return maintainer.setStatus(tx, task.NOT_UNDER_MAINTENANCE, task.GLOBAL_MAINTENANCE)
	case task.TENANT_MAINTENANCE:
		lock := bo.PartialMaintenance{
			LockName: dag.GetMaintenanceKey(),
			LockType: dag.GetMaintenanceType(),
		}
		if err := maintainer.releasePartialLock(tx, &lock); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&bo.PartialMaintenance{}).Where("lock_type=? and gmt_locked >= '1970-01-02'", lock.LockType).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return maintainer.setStatus(tx, task.NOT_UNDER_MAINTENANCE, task.TENANT_MAINTENANCE)
		}
		return nil
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
		return fmt.Errorf("failed to start maintenance: agent status is not %d", oldStatus)
	}
	return nil
}

func (maintainer *agentStatusMaintainer) StartMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	if !dag.IsMaintenance() {
		return nil
	}
	return maintainer.setStatus(tx, task.GLOBAL_MAINTENANCE, task.NOT_UNDER_MAINTENANCE)
}

func (maintainer *agentStatusMaintainer) UpdateMaintenanceTask(tx *gorm.DB, dag *task.Dag) error {
	return nil
}

func (maintainer *agentStatusMaintainer) StopMaintenance(tx *gorm.DB, dag task.Maintainer) error {
	return maintainer.setStatus(tx, task.NOT_UNDER_MAINTENANCE, task.GLOBAL_MAINTENANCE)
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
