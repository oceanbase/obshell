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

type ObStopParam struct {
	Terminate bool `json:"terminate"`
	ForcePass bool `json:"force_pass"`
}

type ObRestartParam struct {
	Terminate bool `json:"terminate"`
	ForcePass bool `json:"force_pass"`
}

type StartObParam struct {
	ForcePass bool `json:"force_pass"`
}

type ObVersion struct {
	Version string `json:"version" binding:"required"`
	Release string `json:"release" binding:"required"`
}

type UpgradeCheckParam struct {
	Version    string `json:"version" binding:"required"`
	Release    string `json:"release" binding:"required"`
	UpgradeDir string `json:"upgradeDir" `
}

type ObUpgradeParam struct {
	UpgradeCheckParam
	Mode string `json:"mode" binding:"required"`
}

type DeletePackageParam struct {
	Name                string `json:"name" binding:"required"`                 // rpm package name
	Version             string `json:"version" binding:"required"`              // rpm package version
	ReleaseDistribution string `json:"release_distribution" binding:"required"` // rpm package release distribution
	Architecture        string `json:"architecture" binding:"required"`         // rpm package architecture
}
