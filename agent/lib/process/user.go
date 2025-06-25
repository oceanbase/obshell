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

package process

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/agent/errors"
)

func Uid() uint32 {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	intUid, err := strconv.Atoi(u.Uid)
	if err != nil {
		panic(err)
	}
	return uint32(intUid)
}

func Gid() uint32 {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	intGid, err := strconv.Atoi(u.Gid)
	if err != nil {
		panic(err)
	}
	return uint32(intGid)
}

func GetObserverUid() (uint32, error) {
	// If err, pidStr is empty.
	if pidStr, _ := GetObserverPid(); pidStr != "" {
		pid, _ := strconv.Atoi(pidStr) // Won't err, if pidStr is not empty, it must be a number.
		return GetUidFromPid(pid)
	}
	return 0, errors.Occur(errors.ErrObServerProcessNotExist)
}

func GetUidFromPid(pid int) (uint32, error) {
	// Get uid from /proc/[pid]/status.
	statusFile := filepath.Join("/proc", fmt.Sprint(pid), "status")
	file, err := os.Open(statusFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var uidStr string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				uidStr = fields[1]
				id, err := strconv.Atoi(uidStr) // If err, id is zero.
				return uint32(id), err
			}
		}
	}
	return 0, errors.Occur(errors.ErrCommonUnexpected, "uid not found")
}
