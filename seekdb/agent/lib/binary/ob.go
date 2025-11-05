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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
)

func GetMyOBVersion() (version string, IsCommunityEdition bool, err error) {
	out, err := exec.Command(path.ObserverBinPath(), "-V").CombinedOutput()
	if err != nil {
		return "", false, err
	}
	res := string(out)

	// get build number
	regex := regexp.MustCompile(`REVISION:\s*(\d+)-([a-fA-F0-9]+)`)
	match := regex.FindStringSubmatch(res)
	if len(match) != 3 {
		return "", false, errors.Occurf(errors.ErrObBinaryVersionUnexpected, "match is not 3, res is:\n <<< %s >>>", res)
	}
	buildNumber := match[1]

	// get version
	regex = regexp.MustCompile(`\(OceanBase SeekDB\s*([\d.]+)\)`)
	match = regex.FindStringSubmatch(res)
	if match == nil {
		return "", false, errors.Occurf(errors.ErrObBinaryVersionUnexpected, "match is nil, res is:\n <<< %s >>>", res)
	}
	return fmt.Sprintf("%s-%s", match[1], buildNumber), true, nil
}

func GetClusterId() (string, error) {
	if _, err := os.Stat(path.ObserverClusterIdFilePath()); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	content, err := os.ReadFile(path.ObserverClusterIdFilePath())
	if err != nil {
		return "", err
	}

	type Content struct {
		ClusterId string `json:"id"`
	}
	type Response struct {
		Content Content `json:"content"`
	}
	response := Response{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return "", err
	}
	return response.Content.ClusterId, nil
}
