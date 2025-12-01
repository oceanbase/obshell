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

import (
	"time"

	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type ObLicense struct {
	EndUser        string    `gorm:"column:END_USER"`
	LicenseId      string    `gorm:"column:LICENSE_ID"`
	LicenseCode    string    `gorm:"column:LICENSE_CODE"`
	LicenseType    string    `gorm:"column:LICENSE_TYPE"`
	ProductType    string    `gorm:"column:PRODUCT_TYPE"`
	IssuanceDate   time.Time `gorm:"column:ISSUANCE_DATE"`
	ActivationTime time.Time `gorm:"column:ACTIVATION_TIME"`
	ExpiredTime    time.Time `gorm:"column:EXPIRED_TIME"`
	Options        string    `gorm:"column:OPTIONS"`
	NodeNum        int64     `gorm:"column:NODE_NUM"`
	ClusterUlid    string    `gorm:"column:CLUSTER_ULID"`
}

func (ObLicense) TableName() string {
	return "oceanbase.DBA_OB_LICENSE"
}

func (o *ObLicense) ToBO() *bo.ObLicense {
	return &bo.ObLicense{
		EndUser:        o.EndUser,
		LicenseId:      o.LicenseId,
		LicenseCode:    o.LicenseCode,
		LicenseType:    o.LicenseType,
		ProductType:    o.ProductType,
		IssuanceDate:   o.IssuanceDate,
		ActivationTime: o.ActivationTime,
		ExpiredTime:    o.ExpiredTime,
		Options:        o.Options,
		NodeNum:        o.NodeNum,
		ClusterUlid:    o.ClusterUlid,
	}
}
