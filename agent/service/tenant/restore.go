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

package tenant

import (
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/agent/engine/task"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func (s *TenantService) GetRunningRestoreTask(tenantName string) (*bo.CdbObRestoreProgress, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	res := new(oceanbase.CdbObRestoreProgress)
	err = oceanbaseDb.Table(CDB_OB_RESTORE_PROGRESS).Where("RESTORE_TENANT_NAME = ? AND TENANT_ID = 1", tenantName).Scan(&res).Error
	if err != nil {
		return nil, err
	}
	if res.RestoreTenantName == "" {
		return nil, nil
	}
	return res.ToBO(), nil
}

func (s *TenantService) GetLastRestoreTask(tenantName string) (*bo.CdbObRestoreHistory, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	res := &oceanbase.CdbObRestoreHistory{}
	err = oceanbaseDb.Table(CDB_OB_RESTORE_HISTORY).Where("RESTORE_TENANT_NAME = ?", tenantName).Order("START_TIMESTAMP desc").Limit(1).Scan(&res).Error
	if err != nil {
		return nil, err
	}
	if res.RestoreTenantName == "" {
		return nil, nil
	}
	return res.ToBO(), nil
}

func (s *TenantService) CancelRestore(tenantName string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("ALTER SYSTEM CANCEL RESTORE `%s`;", tenantName)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) Restore(c *param.RestoreParam, locality, poolList string, scn int64) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}

	var sql string
	if c.Decryption != nil && len(*c.Decryption) > 0 {
		length := len(*c.Decryption)
		for i := 0; i < length; i++ {
			(*c.Decryption)[i] = strings.ReplaceAll((*c.Decryption)[i], "\"", "\\\"")
			(*c.Decryption)[i] = fmt.Sprintf("'%s'", (*c.Decryption)[i])
		}
		sql = fmt.Sprintf("SET DECRYPTION IDENTIFIED BY %s;", strings.Join(*c.Decryption, ","))
	}

	if c.KmsEncryptInfo != nil {
		sql = fmt.Sprintf("%s SET @kms_encrypt_info =\"%s\";", sql, *c.KmsEncryptInfo)
	}

	restoreSql := fmt.Sprintf("ALTER SYSTEM RESTORE %s FROM \"%s, %s\"", c.TenantName, c.DataBackupUri, *c.ArchiveLogUri)
	if c.Timestamp != nil {
		restoreSql = fmt.Sprintf("%s UNTIL TIME= \"%s\"", restoreSql, c.Timestamp.Format("2006-01-02 15:04:05.00"))
	}
	if scn != 0 {
		restoreSql = fmt.Sprintf("%s UNTIL SCN=%d", restoreSql, scn)
	}

	restoreOption := fmt.Sprintf("pool_list=%s", poolList)
	if locality != "" {
		restoreOption = fmt.Sprintf("%s&locality=%s", restoreOption, locality)
	}
	if c.PrimaryZone != nil {
		restoreOption = fmt.Sprintf("%s&primary_zone=%s", restoreOption, *c.PrimaryZone)
	}
	if c.Concurrency != nil {
		restoreOption = fmt.Sprintf("%s&concurrency=%d", restoreOption, *c.Concurrency)
	}
	restoreSql = fmt.Sprintf("%s WITH '%s';", restoreSql, restoreOption)

	sql = fmt.Sprintf("%s %s", sql, restoreSql)
	return oceanbaseDb.Exec(sql).Error
}

func (s *TenantService) GetTenantLevelDagIDByTenantName(name string) (id *int64, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.PartialMaintenance{}).Select("dag_id").Where("lock_name = ? and lock_type = ?", name, task.TENANT_MAINTENANCE).Scan(&id).Error
	return
}
