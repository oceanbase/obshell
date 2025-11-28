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

package oceanbase

import (
	"strings"
	"time"

	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type TenantSession struct {
	Id               int64   `gorm:"column:ID"`
	SvrIp            string  `gorm:"column:SVR_IP"`
	SvrPort          int64   `gorm:"column:SVR_PORT"`
	SqlPort          int64   `gorm:"column:SQL_PORT"`
	User             string  `gorm:"column:USER"`
	Db               string  `gorm:"column:DB"`
	Tenant           string  `gorm:"column:TENANT"`
	Host             string  `gorm:"column:HOST"`
	ClientIp         string  `gorm:"column:USER_CLIENT_IP"`
	Command          string  `gorm:"column:COMMAND"`
	Time             float64 `gorm:"column:TIME"`
	State            string  `gorm:"column:STATE"`
	Info             string  `gorm:"column:INFO"`
	ProxySessId      int64   `gorm:"column:PROXY_SESSID"`
	Action           string  `gorm:"column:ACTION"`
	Module           string  `gorm:"column:MODULE"`
	ClientInfo       string  `gorm:"column:CLIENT_INFO"`
	Level            int64   `gorm:"column:LEVEL"`
	SamplePercentage int     `gorm:"column:SAMPLE_PERCENTAGE"`
	RecordPolicy     string  `gorm:"column:RECORD_POLICY"`
	SqlId            string  `gorm:"column:SQL_ID"`
	TotalCpuTime     int64   `gorm:"column:TOTAL_CPU_TIME"`
	MemoryUsage      int64   `gorm:"column:MEMORY_USAGE"`
}

func (t *TenantSession) TableName() string {
	return "oceanbase.GV$OB_PROCESSLIST"
}

func (t *TenantSession) ToBo() *bo.TenantSession {
	proxyIp := ""
	if t.ProxySessId != 0 {
		proxyIp = strings.Split(t.Host, ":")[0]
	}
	host := t.ClientIp
	if t.ProxySessId == 0 && t.Host != "" {
		ip := strings.Split(t.Host, ":")[0]
		if ip == host {
			host = t.Host
		}
	}

	return &bo.TenantSession{
		Id:               t.Id,
		SvrIp:            t.SvrIp,
		SvrPort:          t.SvrPort,
		SqlPort:          t.SqlPort,
		User:             t.User,
		Db:               t.Db,
		Tenant:           t.Tenant,
		Host:             host,
		Command:          t.Command,
		Time:             t.Time,
		State:            t.State,
		Info:             t.Info,
		ProxySessId:      t.ProxySessId,
		ProxyIp:          proxyIp,
		Action:           t.Action,
		Module:           t.Module,
		ClientInfo:       t.ClientInfo,
		Level:            t.Level,
		SamplePercentage: t.SamplePercentage,
		RecordPolicy:     t.RecordPolicy,
		SqlId:            t.SqlId,
		TotalCpuTime:     t.TotalCpuTime,
		MemoryUsage:      t.MemoryUsage,
	}
}

type DeadLockEvent struct {
	TenantId    int       `gorm:"column:TENANT_ID"`
	EventId     string    `gorm:"column:EVENT_ID"`
	SvrIp       string    `gorm:"column:SVR_IP"`
	SvrPort     int       `gorm:"column:SVR_PORT"`
	ReportTime  time.Time `gorm:"column:REPORT_TIME"`
	CycleIdx    int       `gorm:"column:CYCLE_IDX"`
	CycleSize   int       `gorm:"column:CYCLE_SIZE"`
	Role        string    `gorm:"column:ROLE"`
	CreateTime  time.Time `gorm:"column:CREATE_TIME"`
	Module      string    `gorm:"column:MODULE"`
	Visitor     string    `gorm:"column:VISITOR"`
	Object      string    `gorm:"column:OBJECT"`
	ExtraName1  string    `gorm:"column:EXTRA_NAME1"`
	ExtraValue1 string    `gorm:"column:EXTRA_VALUE1"`
}

func (d *DeadLockEvent) TableName() string {
	return "oceanbase.CDB_OB_DEADLOCK_EVENT_HISTORY"
}

func (d *DeadLockEvent) ToDeadLockNode() *bo.DeadLockNode {
	return &bo.DeadLockNode{
		SvrIp:           d.SvrIp,
		SvrPort:         d.SvrPort,
		Idx:             d.CycleIdx,
		TransactionHash: d.parseTransactionHash(),
		RollBacked:      strings.Compare(strings.ToLower(d.Role), "victim") == 0,
		Resource:        d.Object,
		Sql:             d.parseDql(),
	}
}

func (d *DeadLockEvent) parseTransactionHash() string {
	visitor := strings.ToLower(d.Visitor)
	idx := strings.Index(visitor, "hash:")
	var hash string
	if idx >= 0 {
		endIdx := strings.Index(visitor[idx:], ",")
		if endIdx >= 0 {
			endIdx = idx + endIdx
			hash = visitor[idx+5 : endIdx]
		}
	} else {
		idx = strings.Index(visitor, "txid:")
		if idx >= 0 {
			endIdx := strings.Index(visitor[idx:], "}")
			if endIdx >= 0 {
				endIdx = idx + endIdx
				hash = visitor[idx+5 : endIdx]
			}
		}
	}
	return hash
}

func (d *DeadLockEvent) parseDql() string {
	if strings.Contains(strings.TrimSpace(strings.ToLower(d.ExtraName1)), "current sql") ||
		strings.Contains(strings.TrimSpace(strings.ToLower(d.ExtraName1)), "wait_sql") {
		return strings.TrimSpace(d.ExtraValue1)
	}
	return ""
}
