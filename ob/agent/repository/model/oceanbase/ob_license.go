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
	"fmt"
	"time"

	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

type ObLicense struct {
	EndUser        string `gorm:"column:END_USER"`
	LicenseId      string `gorm:"column:LICENSE_ID"`
	LicenseCode    string `gorm:"column:LICENSE_CODE"`
	LicenseType    string `gorm:"column:LICENSE_TYPE"`
	ProductType    string `gorm:"column:PRODUCT_TYPE"`
	IssuanceDate   string `gorm:"column:ISSUANCE_DATE"`
	ActivationTime string `gorm:"column:ACTIVATION_TIME"`
	ExpiredTime    string `gorm:"column:EXPIRED_TIME"`
	Options        string `gorm:"column:OPTIONS"`
	NodeNum        int64  `gorm:"column:NODE_NUM"`
	ClusterUlid    string `gorm:"column:CLUSTER_ULID"`
}

func (ObLicense) TableName() string {
	return "oceanbase.DBA_OB_LICENSE"
}

// parseDateString parses date strings in various formats
func parseDateString(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	// Try common date formats in order of likelihood
	// DATE_FORMAT returns: 'YYYY-MM-DD HH:MM:SS'
	// DATE type from driver may return: 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM:SS' or with microseconds
	formats := []string{
		"2006-01-02 15:04:05.000000", // DATETIME with microseconds
		"2006-01-02 15:04:05",        // DATETIME or DATE_FORMAT output
		"2006-01-02",                 // DATE type
		time.RFC3339,                 // ISO 8601
		time.RFC3339Nano,             // ISO 8601 with nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (o *ObLicense) ToBO() *bo.ObLicense {
	issuanceDate, _ := parseDateString(o.IssuanceDate)
	activationTime, _ := parseDateString(o.ActivationTime)
	expiredTime, _ := parseDateString(o.ExpiredTime)

	return &bo.ObLicense{
		EndUser:        o.EndUser,
		LicenseId:      o.LicenseId,
		LicenseCode:    o.LicenseCode,
		LicenseType:    o.LicenseType,
		ProductType:    o.ProductType,
		IssuanceDate:   issuanceDate,
		ActivationTime: activationTime,
		ExpiredTime:    expiredTime,
		Options:        o.Options,
		NodeNum:        o.NodeNum,
		ClusterUlid:    o.ClusterUlid,
	}
}
