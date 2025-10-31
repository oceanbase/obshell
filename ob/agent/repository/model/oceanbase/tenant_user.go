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
	"fmt"
	"strings"
	"time"

	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type SessionStats struct {
	Count int64  `gorm:"column:COUNT"`
	State string `gorm:"column:STATE"`
}

type ObUser interface {
	ToUserBo() bo.ObUser
	Name() string
}

type MysqlUser struct {
	Host                   string `gorm:"column:host"`
	User                   string `gorm:"column:user"`
	Password               string `gorm:"column:password"`
	SelectPriv             string `gorm:"column:select_priv"`
	InsertPriv             string `gorm:"column:insert_priv"`
	UpdatePriv             string `gorm:"column:update_priv"`
	DeletePriv             string `gorm:"column:delete_priv"`
	CreatePriv             string `gorm:"column:create_priv"`
	DropPriv               string `gorm:"column:drop_priv"`
	ReloadPriv             string `gorm:"column:reload_priv"`
	ShutdownPriv           string `gorm:"column:shutdown_priv"`
	ProcessPriv            string `gorm:"column:process_priv"`
	FilePriv               string `gorm:"column:file_priv"`
	GrantPriv              string `gorm:"column:grant_priv"`
	ReferencesPriv         string `gorm:"column:references_priv"`
	IndexPriv              string `gorm:"column:index_priv"`
	AlterPriv              string `gorm:"column:alter_priv"`
	ShowDbPriv             string `gorm:"column:show_db_priv"`
	SuperPriv              string `gorm:"column:super_priv"`
	CreateTmpTablePriv     string `gorm:"column:create_tmp_table_priv"`
	LockTablesPriv         string `gorm:"column:lock_tables_priv"`
	ExecutePriv            string `gorm:"column:execute_priv"`
	ReplSlavePriv          string `gorm:"column:repl_slave_priv"`
	ReplClientPriv         string `gorm:"column:repl_client_priv"`
	CreateViewPriv         string `gorm:"column:create_view_priv"`
	ShowViewPriv           string `gorm:"column:show_view_priv"`
	CreateRoutinePriv      string `gorm:"column:create_routine_priv"`
	AlterRoutinePriv       string `gorm:"column:alter_routine_priv"`
	CreateUserPriv         string `gorm:"column:create_user_priv"`
	EventPriv              string `gorm:"column:event_priv"`
	TriggerPriv            string `gorm:"column:trigger_priv"`
	CreateTablespacePriv   string `gorm:"column:create_tablespace_priv"`
	SslType                string `gorm:"column:ssl_type"`
	SslCipher              string `gorm:"column:ssl_cipher"`
	X509Issuer             string `gorm:"column:x509_issuer"`
	X509Subject            string `gorm:"column:x509_subject"`
	MaxQuestions           int64  `gorm:"column:max_questions"`
	MaxUpdates             int64  `gorm:"column:max_updates"`
	MaxConnections         int64  `gorm:"column:max_connections"`
	MaxUserConnections     int64  `gorm:"column:max_user_connections"`
	Plugin                 string `gorm:"column:plugin"`
	AuthenticationString   string `gorm:"column:authentication_string"`
	PasswordExpired        string `gorm:"column:password_expired"`
	AccountLocked          string `gorm:"column:account_locked"`
	DropDatabaseLinkPriv   string `gorm:"column:drop_database_link_priv"`
	CreateDatabaseLinkPriv string `gorm:"column:create_database_link_priv"`
	CreateRolePriv         string `gorm:"column:create_role_priv"`
	DropRolePriv           string `gorm:"column:drop_role_priv"`
}

func (t MysqlUser) ToUserBo() bo.ObUser {
	return bo.ObUser{
		UserName: t.User,
		IsLocked: strings.HasPrefix(strings.ToUpper(t.AccountLocked), "Y"),
	}
}

func (t MysqlUser) Name() string {
	return t.User
}

type OracleUser struct {
	UserName      string    `gorm:"column:USERNAME"`
	Created       time.Time `gorm:"column:CREATED"`
	AccountStatus string    `gorm:"column:ACCOUNT_STATUS"` // OPEN means unlocked, else means locked
}

func (t OracleUser) ToUserBo() bo.ObUser {
	return bo.ObUser{
		UserName:   t.UserName,
		IsLocked:   t.AccountStatus != "OPEN",
		CreateTime: t.Created,
	}
}

func (t OracleUser) Name() string {
	return t.UserName
}

type Role struct {
	Role string `gorm:"column:ROLE"`
}

func (t Role) ToRoleBo() bo.ObRole {
	return bo.ObRole{
		Role: t.Role,
	}
}

type RolePrivilege struct {
	Grantee     string `gorm:"column:GRANTEE"`
	GrantedRole string `gorm:"column:GRANTED_ROLE"`
	AdminOption string `gorm:"column:ADMIN_OPTION"`
	DefaultRole string `gorm:"column:DEFAULT_ROLE"`
}

type GlobalPrivilege struct {
	Grantee   string `gorm:"column:GRANTEE"`
	Privilege string `gorm:"column:PRIVILEGE"`
}

type ObjectPrivilege struct {
	Grantee   string `gorm:"column:GRANTEE"`
	Owner     string `gorm:"column:OWNER"`
	Type      string `gorm:"column:OBJECT_TYPE"`
	Name      string `gorm:"column:OBJECT_NAME"`
	Privilege string `gorm:"column:PRIVILEGE"`
}

type DbaObject struct {
	Type  string `gorm:"column:OBJECT_TYPE"`
	Name  string `gorm:"column:OBJECT_NAME"`
	Owner string `gorm:"column:OWNER"`
}

func (t DbaObject) ToDbObjectBo() bo.DbaObjectBo {
	return bo.DbaObjectBo{
		Type:     t.Type,
		Name:     t.Name,
		Owner:    t.Owner,
		FullName: fmt.Sprintf("%s.%s", t.Owner, t.Name),
	}
}
