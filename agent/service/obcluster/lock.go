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

package obcluster

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/agent/errors"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

type TxLock struct {
	tx *gorm.DB
}

func (obclusterService *ObclusterService) GetClusterStatusLock() (*TxLock, error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, err
	}
	if err = oceanbaseDb.Model(&oceanbase.ClusterStatus{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{}),
	}).Create(&oceanbase.ClusterStatus{
		Id:     0,
		Status: 0,
	}).Error; err != nil {
		return nil, err
	}

	if err := obclusterService.setSessionObQueryTimeout(oceanbaseDb, 20000000); err != nil {
		return nil, errors.Wrap(err, "set session ob query timeout failed")
	}
	return &TxLock{tx: oceanbaseDb}, nil
}

func (lk *TxLock) Lock() error {
	tx := lk.tx.Begin()
	// Lock all_agent by select for update cluster_status's entry (0,0).
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", 0).Find(&oceanbase.ClusterStatus{}).Error; err != nil {
		return errors.Wrap(err, "lock cluster status failed")
	}
	lk.tx = tx
	return nil
}

func (lk *TxLock) Unlock() (err error) {
	if lk.tx == nil {
		return
	}
	if err := lk.tx.Commit().Error; err != nil {
		return err
	}
	lk.tx = nil
	return nil
}

func (obclusterService *ObclusterService) InitializeClusterStatusForLock() error {
	// Insert (0,0) into ocs.cluster_status just for lock.
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Model(&oceanbase.ClusterStatus{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{}),
	}).Create(&oceanbase.ClusterStatus{
		Id:     0,
		Status: 0,
	}).Error
}
