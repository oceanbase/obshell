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

package binary

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
)

func GetMyOBVersion() (version string, err error) {
	myOBPath := filepath.Join(global.HomePath, constant.DIR_BIN, constant.PROC_OBSERVER)
	bash := fmt.Sprintf("export LD_LIBRARY_PATH='%s/lib'; %s -V", global.HomePath, myOBPath)
	out, err := exec.Command("/bin/bash", "-c", bash).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "exec get my ob version failed")
	}
	res := string(out)

	// get build number
	regex := regexp.MustCompile(`REVISION:\s*(\d+)-([a-fA-F0-9]+)`)
	match := regex.FindStringSubmatch(res)
	if len(match) != 3 {
		return "", errors.New("get my ob build number failed")
	}
	buildNumber := match[1]

	// get version
	regex = regexp.MustCompile(`observer\s*\(OceanBase(_CE)?\s*([\d.]+)\)`)
	match = regex.FindStringSubmatch(res)
	if match == nil {
		return "", errors.New("get my ob version failed")
	}
	return fmt.Sprintf("%s-%s", match[2], buildNumber), nil
}
