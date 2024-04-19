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
	"sync"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/config"
)

const (
	TEST_DATABASE_SQL         = "SELECT 1"
	TEST_OCEANBASE_SQL        = "SHOW DATABASES"
	TEST_OCEANBASE_TABLES_SQL = "SHOW TABLES"
)

var (
	dbLock        sync.Mutex
	dbInstance    *gorm.DB
	currentConfig *config.ObDataSourceConfig
	isOcs         bool
	initOnce      sync.Once

	lastInitError          error
	hasAttemptedConnection bool
)
