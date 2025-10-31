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
)

type ObParameters struct {
	SvrIp        string `gorm:"column:SVR_IP"`
	SvrPort      int    `gorm:"column:SVR_PORT"`
	Zone         string `gorm:"column:ZONE"`
	Scope        string `gorm:"column:SCOPE"`
	Name         string `gorm:"column:NAME"`
	Value        string `gorm:"column:VALUE"`
	EditLevel    string `gorm:"column:EDIT_LEVEL"`
	DefaultValue string `gorm:"column:DEFAULT_VALUE"`
	Section      string `gorm:"column:SECTION"`
	Info         string `gorm:"column:INFO"`
	DataType     string `gorm:"column:DATA_TYPE"`
}

func (ObParameters) TableName() string {
	return "oceanbase.GV$OB_PARAMETERS"
}

type OBServer struct {
	Id                 int64     `gorm:"column:ID"`
	SvrIp              string    `gorm:"column:SVR_IP"`
	SqlPort            int       `gorm:"column:SQL_PORT"`
	StopTime           time.Time `gorm:"column:STOP_TIME"`
	StartServiceTime   time.Time `gorm:"column:START_SERVICE_TIME"`
	CreateTime         time.Time `gorm:"column:CREATE_TIME"`
	Status             string    `gorm:"column:STATUS"`
	BuildVersion       string    `gorm:"column:BUILD_VERSION"`
	BlockMigrateInTime time.Time `gorm:"column:BLOCK_MIGRATE_IN_TIME"`
}

func (OBServer) TableName() string {
	return "oceanbase.DBA_OB_SERVERS"
}
