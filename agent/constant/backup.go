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

import "time"

const (
	BINDING_MODE_OPTIONAL  = "OPTIONAL"
	BINDING_MODE_MANDATORY = "MANDATORY"

	LOG_ARCHIVE_CONCURRENCY_LOW  = 0
	LOG_ARCHIVE_CONCURRENCY_HIGH = 100

	HA_LOW_THREAD_SCORE_LOW  = 0
	HA_LOW_THREAD_SCORE_HIGH = 100

	PIECE_SWITCH_INTERVAL_LOW  = time.Hour * 24
	PIECE_SWITCH_INTERVAL_HIGH = time.Hour * 24 * 7

	ARCHIVE_LAG_TARGET_HIGH       = time.Hour * 2
	ARCHIVE_LAG_TARGET_LOW_FOR_S3 = time.Minute

	DELETE_POLICY_DEFAULT = "default"

	LOG_MODE_NOARCHIVELOG = "NOARCHIVELOG"
	LOG_MODE_ARCHIVELOG   = "ARCHIVELOG"

	ARCHIVELOG_STATUS_NULL        = ""
	ARCHIVELOG_STATUS_PREPARE     = "PREPARE"
	ARCHIVELOG_STATUS_BEGINNING   = "BEGINNING"
	ARCHIVELOG_STATUS_DOING       = "DOING"
	ARCHIVELOG_STATUS_INTERRUPTED = "INTERRUPTED"
	ARCHIVELOG_STATUS_STOP        = "STOP"
	ARCHIVELOG_STATUS_STOPPING    = "STOPPING"
	ARCHIVELOG_STATUS_SUSPENDING  = "SUSPENDING"
	ARCHIVELOG_STATUS_SUSPEND     = "SUSPEND"

	TIME_UNIT_MICROSECOND = "us"
	TIME_UNIT_MILLISECOND = "ms"
	TIME_UNIT_SECOND      = "s"
	TIME_UNIT_MINUTE      = "m"
	TIME_UNIT_HOUR        = "h"
	TIME_UNIT_DAY         = "d"

	PREFIX_OSS  = "oss://"
	PREFIX_COS  = "cos://"
	PREFIX_S3   = "s3://"
	PREFIX_FILE = "file://"

	PROTOCOL_OSS  = "oss"
	PROTOCOL_COS  = "cos"
	PROTOCOL_S3   = "s3"
	PROTOCOL_FILE = "file"

	BACKUP_DIR_CLOG = "clog"
	BACKUP_DIR_DATA = "data"

	BACKUP_MODE_FULL        = "full"
	BACKUP_MODE_INCREMENTAL = "incremental"

	BACKUP_CANCELED = "canceled"
)

const (
	RESTORE_UNIT_NUM_DEFAULT = 1

	HA_HIGH_THREAD_SCORE_DEFAULT = 10
)
