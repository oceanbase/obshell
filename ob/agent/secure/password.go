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

package secure

import "github.com/oceanbase/obshell/ob/agent/meta"

func setOceanbasePwd(pwd string) {
	if pwd != meta.OCEANBASE_PWD {
		InvalidateAllSessions()
	}
	meta.SetOceanbasePwd(pwd)
}

func setAgentPassword(pwd string) {
	if pwd != meta.AGENT_PWD.GetPassword() {
		InvalidateAllSessions()
	}
	meta.AGENT_PWD.SetPassword(pwd)
}
