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

package path

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
)

// MetaDbPath returns the path to seekdb's meta.db (config store: ./etc/meta.db).
// Config items are stored in table __all_sys_parameter instead of seekdb.config.bin.
func MetaDbPath() string {
	return filepath.Join(EtcDir(), constant.OB_META_DB)
}

func IsEtcDirExist() bool {
	_, err := os.Stat(EtcDir())
	return err == nil
}

func EtcDirOwnerUid() (uint32, error) {
	fi, err := os.Stat(EtcDir())
	if err != nil {
		return 0, err
	}
	return fi.Sys().(*syscall.Stat_t).Uid, nil
}

func ImportTimeZoneInfoScriptPath() string {
	return filepath.Join(BinDir(), constant.OB_IMPORT_TIME_ZONE_INFO_SCRIPT)
}

func ImportTimeZoneInfoFilePath() string {
	return filepath.Join(EtcDir(), constant.OB_IMPORT_TIME_ZONE_INFO_FILE)
}

func ImportSrsDataScriptPath() string {
	return filepath.Join(BinDir(), constant.OB_IMPORT_SRS_DATA_SCRIPT)
}

func ImportSrsDataFilePath() string {
	return filepath.Join(EtcDir(), constant.OB_IMPORT_SRS_DATA_FILE)
}

func OBAdmin() string {
	return filepath.Join(BinDir(), constant.OB_ADMIN)
}
