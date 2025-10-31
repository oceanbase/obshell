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

package config

import "github.com/oceanbase/obshell/ob/agent/constant"

const (
	SQLITE_DSN_TEMPLATE = "%s?parseTime=true&cache=%s&_fk=%d"
	CACHE_SHARED        = "shared"
	FK_1                = 1
)

type SqliteDataSourceConfig struct {
	DsnTemplate     string
	DataDir         string
	Cache           string
	FK              int
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int

	// logger config
	LoggerConfig
}

func DefaultSqliteDataSourceConfig() SqliteDataSourceConfig {
	config := SqliteDataSourceConfig{
		Cache:           CACHE_SHARED,
		FK:              FK_1,
		MaxIdleConns:    constant.DB_DEFAULT_MAX_IDLE_CONNS,
		MaxOpenConns:    constant.DB_DEFAULT_MAX_OPEN_CONNS,
		ConnMaxLifetime: constant.DB_DEFAULT_CONN_MAX_LIFETIME,

		DsnTemplate: SQLITE_DSN_TEMPLATE,
	}
	return config
}
