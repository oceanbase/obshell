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
	"strings"
	"time"
)

type DbaObZoneStorage struct {
	CreateTime    time.Time `gorm:"column:CREATE_TIME" json:"create_time"`
	ModifyTime    time.Time `gorm:"column:MODIFY_TIME" json:"modify_time"`
	Zone          string    `gorm:"column:ZONE" json:"zone"`
	Path          string    `gorm:"column:PATH" json:"path"`
	Endpoint      string    `gorm:"column:ENDPOINT" json:"endpoint"`
	UsedFor       string    `gorm:"column:USED_FOR" json:"used_for"`
	StorageId     int64     `gorm:"column:STORAGE_ID" json:"storage_id"`
	Authorization string    `gorm:"column:AUTHORIZATION" json:"authorization"`
	MaxIOPS       int64     `gorm:"column:MAX_IOPS" json:"max_iops"`
	MaxBandwidth  int64     `gorm:"column:MAX_BANDWIDTH" json:"max_bandwidth"`
	State         string    `gorm:"column:STATE" json:"state"`
	Extension     string    `gorm:"column:EXTENSION" json:"extension"`
}

func (DbaObZoneStorage) TableName() string {
	return "oceanbase.DBA_OB_ZONE_STORAGE"
}

func (z *DbaObZoneStorage) GetHost() string {
	return parseKeyValue(z.Endpoint, "host")
}

func (z *DbaObZoneStorage) GetAccessId() string {
	return parseKeyValue(z.Authorization, "access_id")
}

func (z *DbaObZoneStorage) GetRegion() string {
	return parseKeyValue(z.Extension, "s3_region")
}

func (z *DbaObZoneStorage) GetChecksumType() string {
	return parseKeyValue(z.Extension, "checksum_type")
}

func (z *DbaObZoneStorage) GetDeleteMode() string {
	return parseKeyValue(z.Extension, "delete_mode")
}

func (z *DbaObZoneStorage) GetAddressingModel() string {
	return parseKeyValue(z.Extension, "addressing_model")
}

func (z *DbaObZoneStorage) HasExtensionField(key string) bool {
	_, exists := parseKeyValueWithExists(z.Extension, key)
	return exists
}

func parseKeyValue(s string, key string) string {
	value, _ := parseKeyValueWithExists(s, key)
	return value
}

func parseKeyValueWithExists(s string, key string) (string, bool) {
	if s == "" {
		return "", false
	}
	pairs := strings.Split(s, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 1 {
			continue
		}
		if kv[0] == key {
			return kv[1], true
		}
	}
	return "", false
}
