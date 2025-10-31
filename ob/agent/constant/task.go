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
	REGISTER  = 0
	FAILED    = -1
	SUCCEEDED = 10000
)

const (
	CLUSTER_TASK_ID_PREFIX      = '1'
	LOCAL_TASK_IPV4_ID_PREFIX   = '2'
	LOCAL_TASK_IPV6_ID_PREFIX   = '3'
	OBPROXY_TASK_IPV4_ID_PREFIX = '4'
	OBPROXY_TASK_IPV6_ID_PREFIX = '5'
	ENGINE_WAIT_TIME            = 30 * time.Second

	SYNC_INTERVAL             = 1 * time.Second
	SYNC_TASK_BUFFER_SIZE     = 10000
	SYNC_TASK_LOG_BUFFER_SIZE = 10000
)

const (
	API_OBCLUSTER_CONFIG      = "api obcluster config"
	RPC_OBCLUSTER_CONFIG      = "rpc obcluster config"
	MASTER_SET_OBCLUSTER_CONF = "master set self obcluster config"

	API_OBSERVER_CONFIG      = "api observer config"
	RPC_OBSERVER_CONFIG      = "rpc observer config"
	MASTER_SET_OBSERVER_CONF = "master set self observer config"

	RPC_OB_START = "rpc ob start"

	API_OB_STOP = "api ob stop"
	RPC_OB_STOP = "rpc ob stop"
	API_OB_INIT = "api ob init"

	RPC_OB_START_CHECK = "api ob startcheck"
)
