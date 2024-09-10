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
	SvrIp      string `gorm:"column:SVR_IP"`
	SvrPort    int    `gorm:"column:SVR_PORT"`
	Zone       string `gorm:"column:ZONE"`
	Scope      string `gorm:"column:SCOPE"`
	TenantId   int    `gorm:"column:TENANT_ID"`
	Name       string `gorm:"column:NAME"`
	Value      string `gorm:"column:VALUE"`
	TenantName string
}

type DbaObZones struct {
	Zone   string `gorm:"column:ZONE"`
	Status string `gorm:"column:STATUS"`
	Region string `gorm:"column:REGION"`
}

type OBServer struct {
	Zone             string    `gorm:"column:ZONE"`
	SvrIp            string    `gorm:"column:SVR_IP"`
	SvrPort          int       `gorm:"column:SVR_PORT"`
	SqlPort          int       `gorm:"column:SQL_PORT"`
	StopTime         time.Time `gorm:"column:STOP_TIME"`
	StartServiceTime time.Time `gorm:"column:START_SERVICE_TIME"`
	WithRs           string    `gorm:"column:WITH_ROOTSERVER"`
	Status           string    `gorm:"column:STATUS"`
	BuildVersion     string    `gorm:"column:BUILD_VERSION"`
}

type DbaOBTenants struct {
	TenantID    int64  `gorm:"column:TENANT_ID"`
	TenantName  string `gorm:"column:TENANT_NAME"`
	PrimaryZone string `gorm:"column:PRIMARY_ZONE"`
	Locality    string `gorm:"column:LOCALITY"`
	UnitNum     int64  `gorm:"column:UNIT_NUM"`
	LogMode     string `gorm:"column:LOG_MODE"`
}

type CdbObBackupDeletePolicy struct {
	TenantID       int64  `gorm:"column:TENANT_ID"`
	PolicyName     string `gorm:"column:POLICY_NAME"`
	RecoveryWindow string `gorm:"column:RECOVERY_WINDOW"`
}

type CdbOBArchivelog struct {
	TenantID int64  `gorm:"column:TENANT_ID"`
	Status   string `gorm:"column:STATUS"`
	Path     string `gorm:"column:PATH"`
}

type CdbObBackupTask struct {
	TenantID              int64     `gorm:"column:TENANT_ID" json:"tenant_id"`
	TaskID                int64     `gorm:"column:TASK_ID" json:"task_id"`
	JobID                 int64     `gorm:"column:JOB_ID" json:"job_id"`
	Incarnation           int64     `gorm:"column:INCARNATION" json:"incarnation"`
	BackupSetID           int64     `gorm:"column:BACKUP_SET_ID" json:"backup_set_id"`
	StartTimestamp        time.Time `gorm:"column:START_TIMESTAMP" json:"start_timestamp"`
	EndTimestamp          time.Time `gorm:"column:END_TIMESTAMP" json:"end_timestamp"`
	Status                string    `gorm:"column:STATUS" json:"status"`
	StartScn              int64     `gorm:"column:START_SCN" json:"start_scn"`
	EndScn                int64     `gorm:"column:END_SCN" json:"end_scn"`
	UserLsStartScn        int64     `gorm:"column:USER_LS_START_SCN" json:"user_ls_start_scn"`
	EncryptionMode        string    `gorm:"column:ENCRYPTION_MODE" json:"encryption_mode"`
	InputBytes            int64     `gorm:"column:INPUT_BYTES" json:"input_bytes"`
	OutputBytes           int64     `gorm:"column:OUTPUT_BYTES" json:"output_bytes"`
	OutputRateBytes       float64   `gorm:"column:OUTPUT_RATE_BYTES" json:"output_rate_bytes"`
	ExtraMetaBytes        int64     `gorm:"column:EXTRA_META_BYTES" json:"extra_meta_bytes"`
	TabletCount           int64     `gorm:"column:TABLET_COUNT" json:"tablet_count"`
	FinishTabletCount     int64     `gorm:"column:FINISH_TABLET_COUNT" json:"finish_tablet_count"`
	MacroBlockCount       int64     `gorm:"column:MACRO_BLOCK_COUNT" json:"macro_block_count"`
	FinishMacroBlockCount int64     `gorm:"column:FINISH_MACRO_BLOCK_COUNT" json:"finish_macro_block_count"`
	FileCount             int64     `gorm:"column:FILE_COUNT" json:"file_count"`
	MetaTurnID            int64     `gorm:"column:META_TURN_ID" json:"meta_turn_id"`
	DataTurnID            int64     `gorm:"column:DATA_TURN_ID" json:"data_turn_id"`
	Result                int64     `gorm:"column:RESULT" json:"result"`
	Comment               string    `gorm:"column:COMMENT" json:"comment"`
	Path                  string    `gorm:"column:PATH" json:"path"`
}

type DbaObUnitConfigs struct {
	UnitConfigID int64     `gorm:"column:UNIT_CONFIG_ID"`
	Name         string    `gorm:"column:NAME"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME"`
	ModifyTime   time.Time `gorm:"column:MODIFY_TIME"`
	MaxCpu       float64   `gorm:"column:MAX_CPU"`
	MinCpu       float64   `gorm:"column:MIN_CPU"`
	MemorySize   int64     `gorm:"column:MEMORY_SIZE"`
	LogDiskSize  int64     `gorm:"column:LOG_DISK_SIZE"`
	MaxIops      int64     `gorm:"column:MAX_IOPS"`
	MinIops      int64     `gorm:"column:MIN_IOPS"`
	IopsWeight   int64     `gorm:"column:IOPS_WEIGHT"`
}

type DbaObResourcePools struct {
	ResourcePoolID int64     `gorm:"column:RESOURCE_POOL_ID"`
	Name           string    `gorm:"column:NAME"`
	TenantID       int64     `gorm:"column:TENANT_ID"`
	CreateTime     time.Time `gorm:"column:CREATE_TIME"`
	ModifyTime     time.Time `gorm:"column:MODIFY_TIME"`
	UnitCount      int64     `gorm:"column:UNIT_COUNT"`
	UnitConfigID   int64     `gorm:"column:UNIT_CONFIG_ID"`
	ZoneList       string    `gorm:"column:ZONE_LIST"`
	ReplicaType    string    `gorm:"column:REPLICA_TYPE"`
}
