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

package constant

import "time"

const (
	OCS_HEADER                    = "X-OCS-Header"
	OCS_AGENT_HEADER              = "X-OCS-Agent-Header"
	REQUEST_RECEIVED_TIME         = "request_received_time"
	RESPONSE_PWD_KEY              = "password"
	AGENT_PRIVATE_KEY             = "private_key"
	AGENT_PUBLIC_KEY              = "public_key"
	AGENT_AUTH_EXPIRED_DURATION   = "auth_expired_duration"
	CONFIG_SESSION_TIMEOUT        = "session_timeout"
	CONFIG_SESSION_GC_INTERVAL    = "session_gc_interval"
	DEFAULT_AUTH_EXPIRED_DURATION = 10 * time.Second
	GET_PASSWORD_RPC_TIMEOUT      = 1 * time.Second
	CREDENTIAL_AES_KEY_CONFIG     = "credential_aes_key"
	CAESAR_SHIFT                  = 11
)
