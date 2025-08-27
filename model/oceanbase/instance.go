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

package oceanbase

type OBInstance struct {
	Type      OBInstanceType `json:"type" binding:"required"`
	OBCluster string         `json:"obcluster,omitempty"`
	OBZone    string         `json:"obzone,omitempty"` // obzone may exist in labels
	OBServer  string         `json:"observer,omitempty"`
	OBTenant  string         `json:"obtenant,omitempty"`
}

func (o *OBInstance) Equals(other *OBInstance) bool {
	if o.Type != other.Type {
		return false
	}
	switch o.Type {
	case TypeOBCluster:
		return o.OBCluster == other.OBCluster
	case TypeOBServer:
		return o.OBServer == other.OBServer
	case TypeOBTenant:
		return (o.OBCluster == other.OBCluster) && (o.OBTenant == other.OBTenant)
	default:
		return false
	}
}
