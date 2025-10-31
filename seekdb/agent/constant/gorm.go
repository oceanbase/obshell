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

package constant

const (
	// source config default value
	DB_USERNAME = "root"
	LOCAL_IP    = "127.0.0.1"
	LOCAL_IP_V6 = "::1"

	DB_DEFAULT_CHARSET  = "utf8mb4"
	DB_DEFAULT_LOCATION = "Local"

	// gorm config default value
	DB_DEFAULT_MAX_IDLE_CONNS    = 1
	DB_DEFAULT_MAX_OPEN_CONNS    = 5
	DB_DEFAULT_CONN_MAX_LIFETIME = 0

	// create table default value
	DB_SINGULAR_TABLE = true
)

const (
	MAX_GET_INSTANCE_RETRIES      = 600 // TODO: smaller
	GET_INSTANCE_RETRY_INTERVAL   = 1
	UPGRADE_BINARY_RETRY_INTERVAL = 10
)
