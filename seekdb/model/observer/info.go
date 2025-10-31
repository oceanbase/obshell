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

package observer

import "time"

// ObserverInfo is the response of observer/info
type ObserverInfo struct {
	ClusterName  string     `json:"cluster_name"`
	Version      string     `json:"version"`
	Architecture string     `json:"architecture"`
	Status       string     `json:"status"` // TODO: should be sys tenant status instead of observer status
	InnerStatus  string     `json:"inner_status"`
	Port         int        `json:"port"`
	ObshellPort  int        `json:"obshell_port"`
	CreatedTime  *time.Time `json:"created_time,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	LifeTime     string     `json:"life_time,omitempty"`
	User         string     `json:"user"`

	ObserverDirInfo
	ObserverResourceInfo
	ConnectionString string `json:"connection_string"`
	Whitelist        string `json:"whitelist"`
	DatabaseCount    int    `json:"database_count,omitempty"`
	UserCount        int    `json:"user_count,omitempty"`
}

type ObserverDirInfo struct {
	DataDir string `json:"data_dir"`
	RedoDir string `json:"redo_dir"`
	LogDir  string `json:"log_dir"`
	BaseDir string `json:"base_dir"`
	BinPath string `json:"bin_path"`
}

type ObserverResourceInfo struct {
	CpuCount     float64 `json:"cpu_count"`
	MemorySize   string  `json:"memory_size"`
	LogDiskSize  string  `json:"log_disk_size"`
	DataDiskSize string  `json:"data_disk_size"`
}
