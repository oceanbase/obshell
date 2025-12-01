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

type ObLicense struct {
	EndUser        string    `json:"end_user"`
	LicenseId      string    `json:"license_id"`
	LicenseCode    string    `json:"license_code"`
	LicenseType    string    `json:"license_type"`
	ProductType    string    `json:"product_type"`
	IssuanceDate   time.Time `json:"issuance_date"`
	ActivationTime time.Time `json:"activation_time"`
	ExpiredTime    time.Time `json:"expired_time"`
	Options        string    `json:"options"`
	NodeNum        int64     `json:"node_num"`
	ClusterUlid    string    `json:"cluster_ulid"`
}
