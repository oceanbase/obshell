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
	"fmt"
	"path/filepath"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
)

func ObproxyEtcDir() string {
	return filepath.Join(meta.OBPROXY_HOME_PATH, constant.OBPROXY_DIR_ETC)
}

func ObproxyLibDir() string {
	return filepath.Join(meta.OBPROXY_HOME_PATH, constant.OBPROXY_DIR_LIB)
}

func ObproxyLogDir() string {
	return filepath.Join(meta.OBPROXY_HOME_PATH, constant.OBPROXY_DIR_LIB)
}

func ObproxyRunDir() string {
	return filepath.Join(meta.OBPROXY_HOME_PATH, constant.OBPROXY_DIR_RUN)
}

func ObproxyBinDir() string {
	return filepath.Join(meta.OBPROXY_HOME_PATH, constant.OBPROXY_DIR_BIN)
}

func ObproxyBinPath() string {
	return filepath.Join(ObproxyBinDir(), constant.BIN_OBPROXY)
}

func ObproxyPidPath() string {
	return filepath.Join(ObproxyRunDir(), fmt.Sprintf("%s.pid", constant.BIN_OBPROXY))
}

func ObproxydPidPath() string {
	return filepath.Join(ObproxyRunDir(), fmt.Sprintf("%s.pid", constant.BIN_OBPROXYD))
}

func ObproxyNewConfigDbFile() string {
	return filepath.Join(ObproxyEtcDir(), "proxyconfig_v1.db")
}

func ObproxyOldConfigDbFile() string {
	return filepath.Join(ObproxyEtcDir(), "proxyconfig.db")
}
