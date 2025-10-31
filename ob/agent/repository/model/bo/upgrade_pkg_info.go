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

import "time"

type UpgradePkgInfo struct {
	PkgId               int       `json:"pkg_id"`
	Name                string    `json:"name"`
	Version             string    `json:"version"`
	ReleaseDistribution string    `json:"release_distribution"`
	Distribution        string    `json:"distribution"`
	Release             string    `json:"release"`
	Architecture        string    `json:"architecture"`
	Size                uint64    `json:"size"`
	PayloadSize         uint64    `json:"payload_size"`
	ChunkCount          int       `json:"chunk_count"`
	Md5                 string    `json:"md5"`
	UpgradeDepYaml      string    `json:"upgrade_dep_yaml"`
	GmtModify           time.Time `json:"gmt_modify"`
}
