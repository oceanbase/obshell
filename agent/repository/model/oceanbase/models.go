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
	"time"

	"github.com/oceanbase/obshell/agent/repository/model/bo"
)

type ObParameters struct {
	SvrIp        string `gorm:"column:SVR_IP"`
	SvrPort      int    `gorm:"column:SVR_PORT"`
	Zone         string `gorm:"column:ZONE"`
	Scope        string `gorm:"column:SCOPE"`
	Name         string `gorm:"column:NAME"`
	Value        string `gorm:"column:VALUE"`
	TenantId     int    `gorm:"column:TENANT_ID"`
	EditLevel    string `gorm:"column:EDIT_LEVEL"`
	DefaultValue string `gorm:"column:DEFAULT_VALUE"`
	Section      string `gorm:"column:SECTION"`
	Info         string `gorm:"column:INFO"`
	DataType     string `gorm:"column:DATA_TYPE"`
}

func (ObParameters) TableName() string {
	return "oceanbase.GV$OB_PARAMETERS"
}

type DbaObZones struct {
	Zone   string `gorm:"column:ZONE"`
	Status string `gorm:"column:STATUS"`
	Region string `gorm:"column:REGION"`
	Idc    string `gorm:"column:IDC"`
}

type OBServer struct {
	Zone               string    `gorm:"column:ZONE"`
	Id                 int64     `gorm:"column:ID"`
	SvrIp              string    `gorm:"column:SVR_IP"`
	SvrPort            int       `gorm:"column:SVR_PORT"`
	SqlPort            int       `gorm:"column:SQL_PORT"`
	StopTime           time.Time `gorm:"column:STOP_TIME"`
	StartServiceTime   time.Time `gorm:"column:START_SERVICE_TIME"`
	WithRs             string    `gorm:"column:WITH_ROOTSERVER"`
	Status             string    `gorm:"column:STATUS"`
	BuildVersion       string    `gorm:"column:BUILD_VERSION"`
	BlockMigrateInTime time.Time `gorm:"column:BLOCK_MIGRATE_IN_TIME"`
}

func (OBServer) TableName() string {
	return "oceanbase.DBA_OB_SERVERS"
}

func (observer *OBServer) ToBo() bo.Observer {
	return bo.Observer{
		Id:             observer.Id,
		Ip:             observer.SvrIp,
		SvrPort:        observer.SvrPort,
		SqlPort:        observer.SqlPort,
		Version:        observer.BuildVersion,
		InnerStatus:    observer.Status,
		StartTime:      observer.StartServiceTime,
		StopTime:       observer.StopTime,
		WithRootserver: observer.WithRs == "YES",
	}
}

type ObLogStat struct {
	TenantId          int    `gorm:"column:TENANT_ID"`
	LsId              int    `gorm:"column:LS_ID"`
	SvrIp             string `gorm:"column:SVR_IP"`
	SvrPort           int    `gorm:"column:SVR_PORT"`
	Role              string `gorm:"column:ROLE"`
	ProposalId        int64  `gorm:"column:PROPOSAL_ID"`
	ConfigVersion     string `gorm:"column:CONFIG_VERSION"`
	AccessMode        string `gorm:"column:ACCESS_MODE"`
	PaxosMemberList   string `gorm:"column:PAXOS_MEMBER_LIST"`
	PaxosReplicaNum   int    `gorm:"column:PAXOS_REPLICA_NUM"`
	InSync            int    `gorm:"column:IN_SYNC"`
	BaseLsn           int64  `gorm:"column:BASE_LSN"`
	BeginLsn          int64  `gorm:"column:BEGIN_LSN"`
	BeginScn          int64  `gorm:"column:BEGIN_SCN"`
	EndLsn            int64  `gorm:"column:END_LSN"`
	EndScn            int64  `gorm:"column:END_SCN"`
	MaxLsn            int64  `gorm:"column:MAX_LSN"`
	MaxScn            int64  `gorm:"column:MAX_SCN"`
	ArbitrationMember string `gorm:"column:ARBITRATION_MEMBER"`
	DegradedList      string `gorm:"column:DEGRADED_LIST"`
	LearnerList       string `gorm:"column:LEARNER_LIST"`
}

func (ObLogStat) TableName() string {
	return "oceanbase.GV$OB_LOG_STAT"
}

type RootServer struct {
	Zone    string `gorm:"column:ZONE"`
	Role    string `gorm:"column:ROLE"`
	SvrIp   string `gorm:"column:SVR_IP"`
	SvrPort int    `gorm:"column:SVR_PORT"`
}

func (r *RootServer) ToBO() bo.RootServer {
	return bo.RootServer{
		Ip:      r.SvrIp,
		Role:    r.Role,
		SvrPort: r.SvrPort,
		Zone:    r.Zone,
	}
}

type SysStat struct {
	ConId       int64  `gorm:"column:CON_ID"`
	SvrIp       string `gorm:"column:SVR_IP"`
	SvrPort     int    `gorm:"column:SVR_PORT"`
	StatisticId int64  `gorm:"column:STATISTIC#"`
	Name        string `gorm:"column:NAME"`
	Class       int64  `gorm:"column:CLASS"`
	Value       int64  `gorm:"column:VALUE"`
	ValueType   string `gorm:"column:VALUE_TYPE"`
	StatId      int64  `gorm:"column:STAT_ID"`
}

func (SysStat) TableName() string {
	return "oceanbase.GV$sysstat"
}
