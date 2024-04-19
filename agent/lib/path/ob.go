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

	"github.com/oceanbase/obshell/agent/constant"
)

func ObConfigPath() string {
	return filepath.Join(EtcDir(), constant.OB_CONFIG_FILE)
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
