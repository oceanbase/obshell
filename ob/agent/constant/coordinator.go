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
	MAINTAINER_MAX_ACTIVE_TIME_SEC = 5
	MAINTAINER_MAX_ACTIVE_TIME     = MAINTAINER_MAX_ACTIVE_TIME_SEC * time.Second
	MAINTAINER_UPDATE_INTERVAL     = 4 * time.Second // update interval
	COORDINATOR_MIN_INTERVAL       = 1 * time.Second // coordinator minimum update interval
)
