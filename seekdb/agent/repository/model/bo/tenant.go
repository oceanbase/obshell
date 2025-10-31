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
package bo

import (
	"time"
)

type TenantCompaction struct {
	FrozenScn          int64     `json:"frozen_scn"`
	FrozenTime         time.Time `json:"frozen_time"`
	GlobalBroadcastScn int64     `json:"global_broadcast_scn"`
	LastScn            int64     `json:"last_scn"`
	LastFinishTime     time.Time `json:"last_finish_time"`
	StartTime          time.Time `json:"start_time"`
	Status             string    `json:"status"`
	IsError            string    `json:"is_error"`
	IsSuspended        string    `json:"is_suspended"`
}
