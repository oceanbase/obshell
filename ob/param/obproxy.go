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

type AddObproxyParam struct {
	Name               string            `json:"name"`
	HomePath           string            `json:"home_path" binding:"required"`
	SqlPort            *int              `json:"sql_port"`      // Default to 2883.
	RpcPort            *int              `json:"rpc_port"`      // Default to 2884.
	ExporterPort       *int              `json:"exporter_port"` // Default to 2885.
	ProxyroPassword    string            `json:"proxyro_password"`
	ObproxySysPassword string            `json:"obproxy_sys_password"`
	RsList             *string           `json:"rs_list"`
	ConfigUrl          *string           `json:"config_url"`
	Parameters         map[string]string `json:"parameters"`
}

type UpgradeObproxyParam struct {
	Version    string `json:"version" binding:"required"`
	Release    string `json:"release" binding:"required"`
	UpgradeDir string `json:"upgrade_dir"`
}
