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

package bo

import "time"

type TenantSession struct {
	Id               uint64  `json:"id"`
	SvrIp            string  `json:"svr_ip"`
	SvrPort          int64   `json:"svr_port"`
	SqlPort          int64   `json:"sql_port"`
	User             string  `json:"user"`
	Db               string  `json:"db"`
	Tenant           string  `json:"tenant"`
	Host             string  `json:"host"`
	Command          string  `json:"command"`
	Time             float64 `json:"time"`
	State            string  `json:"state"`
	Info             string  `json:"info"`
	ProxySessId      uint64  `json:"proxy_sess_id"`
	ProxyIp          string  `json:"proxy_ip"` // parse from host when ProxySessId is not null
	Action           string  `json:"action"`
	Module           string  `json:"module"`
	ClientInfo       string  `json:"client_info"`
	Level            int64   `json:"level"`
	SamplePercentage int     `json:"sample_percentage"`
	RecordPolicy     string  `json:"record_policy"`
	SqlId            string  `json:"sql_id"`
	TotalCpuTime     int64   `json:"total_cpu_time"`
	MemoryUsage      uint64  `json:"memory_usage"`
}

type PaginatedTenantSessions struct {
	Page     CustomPage      `json:"page"`
	Contents []TenantSession `json:"contents"`
}

type TenantSessionStats struct {
	TotalCount    int                        `json:"total_count"`
	ActiveCount   int                        `json:"active_count"`
	MaxActiveTime float64                    `json:"max_active_time"`
	DbStats       []TenantSessionDbStats     `json:"db_stats"`
	UserStats     []TenantSessionUserStats   `json:"user_stats"`
	ClientStats   []TenantSessionClientStats `json:"client_stats"`
}

type TenantSessionDbStats struct {
	DbName      string `json:"db_name"`
	TotalCount  int64  `json:"total_count"`
	ActiveCount int64  `json:"active_count"`
}

type TenantSessionUserStats struct {
	UserName    string `json:"user_name"`
	TotalCount  int64  `json:"total_count"`
	ActiveCount int64  `json:"active_count"`
}

type TenantSessionClientStats struct {
	ClientIp    string `json:"client_ip"`
	TotalCount  int64  `json:"total_count"`
	ActiveCount int64  `json:"active_count"`
}

type DeadLock struct {
	EventId    string         `json:"event_id"`
	ReportTime time.Time      `json:"report_time"`
	Size       int64          `json:"size"`
	Nodes      []DeadLockNode `json:"nodes"`
}

type DeadLockNode struct {
	SvrIp           string `json:"svr_ip"`
	SvrPort         int    `json:"svr_port"`
	Idx             int    `json:"idx"`
	TransactionHash string `json:"transaction_hash"`
	RollBacked      bool   `json:"roll_backed"`
	Resource        string `json:"resource"`
	Sql             string `json:"sql"`
}

type PaginatedDeadLocks struct {
	Page     CustomPage `json:"page"`
	Contents []DeadLock `json:"contents"`
}
