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

package global

import (
	"os/exec"
	"strings"

	"github.com/oceanbase/obshell/agent/cmd"
	"github.com/oceanbase/obshell/agent/lib/path"
)

var (
	myAgentIp string

	hasSet           bool
	daemonIsBrandNew bool
)

func getMyAgentIp() (string, error) {
	cmd := exec.Command(path.ObshellBinPath(), cmd.CMD_INFO_IP)
	res, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(res)), nil
}

func MyAgentIp() (string, error) {
	if myAgentIp != "" {
		return myAgentIp, nil
	}

	return getMyAgentIp()
}

func SetDaemonIsBrandNew(b bool) {
	if !hasSet {
		daemonIsBrandNew = b
		hasSet = true
	}
}

func DaemonIsBrandNew() bool {
	return daemonIsBrandNew
}
