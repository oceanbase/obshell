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
	"sort"
	"strings"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/common"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func GetTenantSessions(tenantName string, p *param.QueryTenantSessionParam) (*bo.PaginatedTenantSessions, error) {
	sessions, err := tenantService.GetSessions(tenantName, p)
	if err != nil {
		return nil, err
	}
	var sessionBos []bo.TenantSession
	for _, session := range sessions {
		sessionBos = append(sessionBos, *session.ToBo())
	}
	return &bo.PaginatedTenantSessions{
		Contents: sessionBos,
		Page: bo.CustomPage{
			Number:        p.Page,
			Size:          p.Size,
			TotalPages:    common.CalculateTotalPages(uint64(len(sessions)), p.Size),
			TotalElements: uint64(len(sessions)),
		},
	}, nil
}

func GetTenantSession(tenantName string, sessionId string) (*bo.TenantSession, error) {
	session, err := tenantService.GetSession(tenantName, sessionId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.Occur(errors.ErrObTenantSessionNotExist, sessionId)
	}
	return session.ToBo(), nil
}

func GetTenantSessionStats(tenantName string) (*bo.TenantSessionStats, error) {
	sessions, err := tenantService.ListSessions(tenantName)
	if err != nil {
		return nil, err
	}
	var sessionStats bo.TenantSessionStats
	sessionStats.TotalCount = len(sessions)
	for _, session := range sessions {
		if session.State == "ACTIVE" {
			sessionStats.ActiveCount += 1
		}
		if session.Time > sessionStats.MaxActiveTime {
			sessionStats.MaxActiveTime = session.Time
		}
	}
	sessionStats.DbStats = make([]bo.TenantSessionDbStats, 0)
	sessionStats.UserStats = make([]bo.TenantSessionUserStats, 0)
	sessionStats.ClientStats = make([]bo.TenantSessionClientStats, 0)
	dbStats := make(map[string]*bo.TenantSessionDbStats)
	userStats := make(map[string]*bo.TenantSessionUserStats)
	clientStats := make(map[string]*bo.TenantSessionClientStats)
	for _, session := range sessions {
		// Aggregation db stats.
		if dbStats[session.Db] == nil {
			dbStats[session.Db] = &bo.TenantSessionDbStats{
				DbName:      session.Db,
				TotalCount:  0,
				ActiveCount: 0,
			}
		}
		dbStats[session.Db].TotalCount += 1
		if session.State == "ACTIVE" {
			dbStats[session.Db].ActiveCount += 1
		}

		// Aggregation user stats.
		if userStats[session.User] == nil {
			userStats[session.User] = &bo.TenantSessionUserStats{
				UserName:    session.User,
				TotalCount:  0,
				ActiveCount: 0,
			}
		}
		userStats[session.User].TotalCount += 1
		if session.State == "ACTIVE" {
			userStats[session.User].ActiveCount += 1
		}

		// Aggregation client stats.
		clientIp := strings.Split(session.Host, ":")[0]
		if clientStats[clientIp] == nil {
			clientStats[clientIp] = &bo.TenantSessionClientStats{
				ClientIp:    clientIp,
				TotalCount:  0,
				ActiveCount: 0,
			}
		}
		clientStats[clientIp].TotalCount += 1
		if session.State == "ACTIVE" {
			clientStats[clientIp].ActiveCount += 1
		}
	}

	// Convert map to slice.
	sessionStats.DbStats = make([]bo.TenantSessionDbStats, 0, len(dbStats))
	for _, dbStat := range dbStats {
		sessionStats.DbStats = append(sessionStats.DbStats, *dbStat)
	}
	for _, userStat := range userStats {
		sessionStats.UserStats = append(sessionStats.UserStats, *userStat)
	}
	for _, clientStat := range clientStats {
		sessionStats.ClientStats = append(sessionStats.ClientStats, *clientStat)
	}
	return &sessionStats, nil
}

func KillTenantSessions(tenantName string, sessionId []int) error {
	for _, id := range sessionId {
		err := tenantService.KillSession(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func KillTenantSessionQueries(tenantName string, sessionId []int) error {
	for _, id := range sessionId {
		err := tenantService.KillSessionQuery(id)
		if err != nil {
			return err
		}
	}
	return nil
}

// Compatible with OceanBase versions 4.2.5.x, 4.3.x, and later.
func ListTenantDeadLocks(tenantName string, p *param.QueryTenantDeadLocksParam) (*bo.PaginatedDeadLocks, error) {
	tenantId, err := tenantService.GetTenantId(tenantName)
	if err != nil {
		return nil, err
	}
	deadlockEvents, err := tenantService.ListDeadLockEvents(tenantId)
	if err != nil {
		return nil, err
	}
	deadLockEventMap := buildDeadLockEventMap(deadlockEvents)
	var deadlockBos []bo.DeadLock
	for _, events := range deadLockEventMap {
		deadlock := bo.DeadLock{
			EventId:    events[0].EventId,
			ReportTime: events[0].ReportTime,
			Size:       int64(len(events)),
			Nodes:      make([]bo.DeadLockNode, 0),
		}
		for _, event := range events {
			deadlock.Nodes = append(deadlock.Nodes, *event.ToDeadLockNode())
		}
		deadlockBos = append(deadlockBos, deadlock)
	}
	sort.Slice(deadlockBos, func(i, j int) bool {
		return deadlockBos[i].ReportTime.After(deadlockBos[j].ReportTime)
	})
	offset := (p.Page - 1) * p.Size
	if offset >= uint64(len(deadlockBos)) {
		return &bo.PaginatedDeadLocks{
			Page: bo.CustomPage{
				Number:        p.Page,
				Size:          p.Size,
				TotalPages:    0,
				TotalElements: 0,
			},
		}, nil
	}
	end := offset + p.Size
	if end > uint64(len(deadlockBos)) {
		end = uint64(len(deadlockBos))
	}
	return &bo.PaginatedDeadLocks{
		Page: bo.CustomPage{
			Number:        p.Page,
			Size:          p.Size,
			TotalPages:    common.CalculateTotalPages(uint64(len(deadlockBos)), p.Size),
			TotalElements: uint64(len(deadlockBos)),
		},
		Contents: deadlockBos[offset:end],
	}, nil
}

func buildDeadLockEventMap(deadlockEvents []oceanbase.DeadLockEvent) map[string][]*oceanbase.DeadLockEvent {
	deadLockEventMap := make(map[string][]*oceanbase.DeadLockEvent)
	for _, deadlockEvent := range deadlockEvents {
		deadLockEventMap[deadlockEvent.EventId] = append(deadLockEventMap[deadlockEvent.EventId], &deadlockEvent)
	}
	// filter map
	for eventId, event := range deadLockEventMap {
		if event == nil || (len(event) != event[0].CycleSize) {
			delete(deadLockEventMap, eventId)
		}
	}
	return deadLockEventMap
}
