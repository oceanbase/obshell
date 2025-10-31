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

type AgentPwd struct {
	inited   bool
	password string
}

var (
	OCEANBASE_PWD                  string
	OCEANBASE_PASSWORD_INITIALIZED bool // Which means the oceanbase password has been initialized
)

func (p *AgentPwd) Inited() bool {
	return p.inited
}

func (p *AgentPwd) GetPassword() string {
	return p.password
}

func (p *AgentPwd) SetPassword(pwd string) {
	p.password = pwd
	p.inited = true
}

func GetOceanbasePwd() string {
	if OCS_AGENT != nil && (OCS_AGENT.IsClusterAgent() || OCS_AGENT.IsTakeover()) {
		return OCEANBASE_PWD
	}
	return ""
}

func SetOceanbasePwd(pwd string) {
	OCEANBASE_PWD = pwd
	if !OCEANBASE_PASSWORD_INITIALIZED {
		OCEANBASE_PASSWORD_INITIALIZED = true
	}
}

func ClearOceanbasePwd() {
	OCEANBASE_PWD = ""
	OCEANBASE_PASSWORD_INITIALIZED = false
}
