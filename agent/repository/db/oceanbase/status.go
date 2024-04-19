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

package oceanbase

const (
	STATE_PROCESS_NOT_RUNNING = iota
	STATE_PROCESS_RUNNING
	STATE_CONNECTION_RESTRICTED
	STATE_CONNECTION_AVAILABLE
)

var OBStateMap = map[int]string{
	0: "OB process not running",
	1: "OB process running with no connection",
	2: "OB process running with restricted connection",
	3: "OB process running with available connection",
}

var OBStateShortMap = map[int]string{
	0: "NOT RUNNING",
	1: "RUNNING",
	2: "RESTRICTED",
	3: "AVAILABLE",
}

func GetState() int {
	var err error
	if err = CheckObserverProcess(); err != nil {
		return STATE_PROCESS_NOT_RUNNING
	}

	if _, err = GetInstance(); err == nil {
		return STATE_CONNECTION_AVAILABLE
	}

	if _, err = GetRestrictedInstance(); err == nil {
		return STATE_CONNECTION_RESTRICTED
	}

	return STATE_PROCESS_RUNNING
}
