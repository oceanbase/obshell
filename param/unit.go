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

package param

type CreateResourceUnitConfigParams struct {
	Name        *string  `json:"name" binding:"required"`        // unit config name
	MemorySize  *string  `json:"memory_size" binding:"required"` // memory size, greater than or equal to '1G'
	MaxCpu      *float64 `json:"max_cpu" binding:"required"`     // max cpu cores, greater than 0
	MinCpu      *float64 `json:"min_cpu"`                        // min cpu cores, smaller than or equal 'max_cpu_cores'
	MaxIops     *int     `json:"max_iops"`                       // max iops, greater than or equal to 1024
	MinIops     *int     `json:"min_iops"`                       // min iops, smaller than or equal to 'max_iops'
	LogDiskSize *string  `json:"log_disk_size"`                  // log disk size, greater than or equal to '2G'
}

type ClusterUnitConfigLimit struct {
	MinMemory float64 `json:"min_memory,omitempty"`
	MinCpu    int     `json:"min_cpu,omitempty"`
}
