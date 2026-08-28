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

package meta

import (
	"fmt"
	"os"
	"strconv"

	"github.com/oceanbase/obshell/ob/agent/constant"
)

// StandaloneModeFromEnv reads the deployment type hint supplied by OBD. An
// environment variable is used instead of a command-line flag so older
// obshell binaries safely ignore the hint during rolling compatibility.
func StandaloneModeFromEnv() (bool, error) {
	value := os.Getenv(constant.ENV_OBSHELL_STANDALONE_MODE)
	if value == "" {
		return false, nil
	}
	standalone, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", constant.ENV_OBSHELL_STANDALONE_MODE, err)
	}
	return standalone, nil
}

// IsStandaloneLoopback reports whether the explicit standalone contract uses
// the stable IPv4 loopback identity. The standalone marker is required so a
// community or distributed deployment configured with 127.0.0.1 keeps its
// existing startup semantics.
func IsStandaloneLoopback(standalone bool, ip string) bool {
	return standalone && ip == constant.LOCAL_IP
}
