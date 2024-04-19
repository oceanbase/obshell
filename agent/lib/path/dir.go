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

	"github.com/oceanbase/obshell/agent/constant"
)

var mypath string

func init() {
	if mypath == "" {
		execPath, err := os.Executable()
		if err != nil {
			panic(err)
		}

		realPath, err := filepath.EvalSymlinks(execPath)
		if err != nil {
			panic(err)
		}

		mypath = realPath
	}
}

func AgentDir() string {
	binDir := filepath.Dir(mypath)
	return filepath.Dir(binDir)
}

func RunDir() string {
	return filepath.Join(AgentDir(), constant.DIR_RUN)
}

func BinDir() string {
	return filepath.Join(AgentDir(), constant.DIR_BIN)
}

func LogDir() string {
	return filepath.Join(AgentDir(), constant.DIR_LOG_OBSHELL)
}

func EtcDir() string {
	return filepath.Join(AgentDir(), constant.OB_DIR_ETC)
}

func CertificateDir() string {
	return filepath.Join(AgentDir(), constant.DIR_CA)
}
