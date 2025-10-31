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

type MysqlDb struct {
	Host               string `gorm:"column:host"`
	Db                 string `gorm:"column:db"`
	User               string `gorm:"column:user"`
	SelectPriv         string `gorm:"column:select_priv"`
	InsertPriv         string `gorm:"column:insert_priv"`
	UpdatePriv         string `gorm:"column:update_priv"`
	DeletePriv         string `gorm:"column:delete_priv"`
	CreatePriv         string `gorm:"column:create_priv"`
	DropPriv           string `gorm:"column:drop_priv"`
	GrantPriv          string `gorm:"column:grant_priv"`
	ReferencesPriv     string `gorm:"column:references_priv"`
	IndexPriv          string `gorm:"column:index_priv"`
	AlterPriv          string `gorm:"column:alter_priv"`
	CreateTmpTablePriv string `gorm:"column:create_tmp_table_priv"`
	LockTablesPriv     string `gorm:"column:lock_tables_priv"`
	CreateViewPriv     string `gorm:"column:create_view_priv"`
	ShowViewPriv       string `gorm:"column:show_view_priv"`
	CreateRoutinePriv  string `gorm:"column:create_routine_priv"`
	AlterRoutinePriv   string `gorm:"column:alter_routine_priv"`
	ExecutePriv        string `gorm:"column:execute_priv"`
	EventPriv          string `gorm:"column:event_priv"`
	TriggerPriv        string `gorm:"column:trigger_priv"`
}

type DatabaseName struct {
	Database string `gorm:"column:Database"`
}

type Database struct {
	CreateTimestamp int64  `gorm:"column:CREATE_TIMESTAMP"`
	DatabaseID      string `gorm:"column:DATABASE_ID"`
	Name            string `gorm:"column:NAME"`
	CollationType   string `gorm:"column:COLLATION_TYPE"`
	CollationName   string `gorm:"column:COLLATION_NAME"`
	CharSetName     string `gorm:"column:CHARACTER_SET_NAME"`
	ReadOnly        string `gorm:"column:READ_ONLY"`
}
