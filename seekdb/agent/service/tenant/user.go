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

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/param"
	"gorm.io/gorm"
)

func (t *TenantService) GrantDbPrivilegesWithHost(db *gorm.DB, userName, hostName string, privilege param.DbPrivilegeParam) error {
	if len(privilege.Privileges) > 0 {
		realPrivileges := convertPrivileges(privilege.Privileges)
		sql := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`@`%s`", strings.Join(realPrivileges, ","), privilege.DbName, userName, hostName)
		return db.Exec(sql).Error
	}
	return nil
}

func convertPrivileges(privileges []string) []string {
	realPrivileges := make([]string, 0)
	for _, privilege := range privileges {
		realPrivileges = append(realPrivileges, strings.ReplaceAll(privilege, "_", " "))
	}
	return realPrivileges
}

func (t *TenantService) GrantDbPrivileges(userName string, param *param.DbPrivilegeParam) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	if param != nil && len(param.Privileges) > 0 {
		realPrivileges := convertPrivileges(param.Privileges)
		sql := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`", strings.Join(realPrivileges, ", "), param.DbName, userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *TenantService) RevokeDbPrivileges(userName string, param *param.DbPrivilegeParam) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	if param != nil && len(param.Privileges) > 0 {
		realPrivileges := convertPrivileges(param.Privileges)
		sql := fmt.Sprintf("REVOKE %s ON %s.* FROM `%s`", strings.Join(realPrivileges, ", "), param.DbName, userName)
		return db.Exec(sql).Error
	}
	return nil
}

func (t *TenantService) LockUser(userName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER USER '%s' ACCOUNT LOCK", userName)
	return db.Exec(sql).Error
}

func (t *TenantService) GetUserSessionStats(userName string) ([]oceanbase.SessionStats, error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, err
	}
	sessionStats := make([]oceanbase.SessionStats, 0)
	result := oceanbaseDb.Table(GV_OB_SESSION).Where("user=? and tenant=?", userName, constant.TENANT_SYS).Select("COUNT(*) as COUNT, STATE").Group("STATE").Scan(&sessionStats)
	return sessionStats, result.Error
}

func (t *TenantService) UnlockUser(userName string) error {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER USER '%s' ACCOUNT UNLOCK", userName)
	return db.Exec(sql).Error
}
