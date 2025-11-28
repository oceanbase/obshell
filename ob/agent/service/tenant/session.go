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

	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func (s *TenantService) GetSessions(tenantName string, p *param.QueryTenantSessionParam) (sessions []oceanbase.TenantSession, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	query := oceanbaseDb.Model(&oceanbase.TenantSession{}).Where("TENANT = ?", tenantName)
	if p.User != "" {
		query = query.Where("USER = ?", p.User)
	}
	if p.Db != "" {
		query = query.Where("DB = ?", p.Db)
	}
	if p.Host != "" {
		query = query.Where("USER_CLIENT_IP = ?", p.Host)
	}
	if p.SessionId != 0 {
		query = query.Where("ID = ?", p.SessionId)
	}
	if len(p.ObserverList) > 0 {
		optionStr := make([]string, 0)
		for _, observer := range p.ObserverList {
			parts := strings.Split(observer, ":")
			if len(parts) != 2 {
				continue
			}
			optionStr = append(optionStr, fmt.Sprintf("SVR_IP = '%s' && SVR_PORT = '%s'  ", parts[0], parts[1]))
		}
		query = query.Where(strings.Join(optionStr, " OR "))
	}
	if p.ActiveOnly {
		query = query.Where("STATE = 'ACTIVE'")
	}
	if p.SortBy != "" && p.SortOrder != "" {
		query = query.Order(fmt.Sprintf("%s %s", p.SortBy, p.SortOrder))
	}
	offset := (p.Page - 1) * p.Size
	query = query.Offset(int(offset)).Limit(int(p.Size))
	err = query.Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *TenantService) GetSession(tenantName string, sessionId string) (session *oceanbase.TenantSession, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	query := oceanbaseDb.Model(&oceanbase.TenantSession{}).Where("TENANT = ?", tenantName).Where("ID = ?", sessionId)
	err = query.Scan(&session).Error
	if err != nil {
		return nil, err
	}
	return
}

func (s *TenantService) ListSessions(tenantName string) (sessions []oceanbase.TenantSession, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	query := oceanbaseDb.Model(&oceanbase.TenantSession{}).Where("TENANT = ?", tenantName)
	err = query.Scan(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *TenantService) KillSession(sessionId int) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	query := oceanbaseDb.Exec("KILL  ?", sessionId)
	return query.Error
}

func (s *TenantService) KillSessionQuery(sessionId int) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	query := oceanbaseDb.Exec("KILL QUERY ?", sessionId)
	return query.Error
}

func (s *TenantService) ListDeadLockEvents(tenantId int) (deadlocks []oceanbase.DeadLockEvent, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	query := oceanbaseDb.Model(&oceanbase.DeadLockEvent{}).Where("TENANT_ID = ?", tenantId)
	err = query.Scan(&deadlocks).Error
	if err != nil {
		return nil, err
	}
	return deadlocks, nil
}
