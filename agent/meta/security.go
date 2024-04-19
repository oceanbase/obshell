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

var (
	OCEANBASE_PWD     string
)

func GetOceanbasePwd() string {
	if OCS_AGENT != nil && (OCS_AGENT.IsClusterAgent() || OCS_AGENT.IsTakeover()) {
		return OCEANBASE_PWD
	}
	return ""
}

func SetOceanbasePwd(pwd string) {
	OCEANBASE_PWD = pwd
}
