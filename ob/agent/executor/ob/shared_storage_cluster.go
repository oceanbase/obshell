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

package ob

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/parse"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

func GetObclusterSummaryWithSharedStorage(info *bo.ClusterInfo) error {
	info.IsSharedStorage = true
	sharedStorageInfo, err := obclusterService.GetSharedStorageInfo()
	if err != nil {
		return errors.Occurf(errors.ErrCommonUnexpected, "Failed to get shared storage info: %v", err)
	}
	if sharedStorageInfo == nil {
		return errors.Occur(errors.ErrSharedStorageInfoNotFound)
	}
	info.SharedStorageInfo = convertSharedStorageInfo(sharedStorageInfo)

	if err := extraStatsWithSharedStorage(info); err != nil {
		return errors.Occurf(errors.ErrCommonUnexpected, "Failed to add extra stats with shared storage: %v", err)
	}

	return nil
}

func getStartupMode() (bool, string, error) {
	param, err := obclusterService.GetParameterByName(constant.PARAMETER_OB_STARTUP_MODE)
	if err != nil {
		return false, "", err
	}
	return strings.EqualFold(param.Value, constant.OB_STARTUP_MODE_SHARED_STORAGE), param.Value, nil
}

func extraStatsWithSharedStorage(info *bo.ClusterInfo) error {
	for _, zone := range info.Zones {
		for _, server := range zone.Servers {
			info.Stats.Add(&server.Stats.BaseResourceStats)
		}
	}

	sharedStorageBytes, err := obclusterService.GetSharedStorageTotalUsageBytes()
	if err != nil {
		log.Warnf("GetSharedStorageTotalUsageBytes failed: %v", err)
	}
	info.Stats.SharedStorageUsed = parse.FormatCapacity(sharedStorageBytes)
	info.Stats.SharedStorageInBytesUsed = sharedStorageBytes

	tenantSharedStorageUsageMap, err := obclusterService.GetSharedStorageTenantStatsMap()
	if err != nil {
		log.Warnf("GetSharedStorageTenantStatsMap failed: %v", err)
		tenantSharedStorageUsageMap = nil
	}
	for j := range info.TenantStats {
		if sharedStorageUsage, ok := tenantSharedStorageUsageMap[info.TenantStats[j].TenantId]; ok {
			info.TenantStats[j].SharedStorageUsage = sharedStorageUsage
		}
	}

	return nil
}

func convertSharedStorageInfo(storageInfo *oceanbase.DbaObZoneStorage) *bo.SharedStorageInfo {
	return &bo.SharedStorageInfo{
		StoragePath:     storageInfo.Path,
		AccessDomain:    storageInfo.GetHost(),
		Region:          storageInfo.GetRegion(),
		AccessKey:       storageInfo.GetAccessId(),
		DeleteMode:      storageInfo.GetDeleteMode(),
		ChecksumType:    storageInfo.GetChecksumType(),
		AddressingModel: storageInfo.GetAddressingModel(),
		MaxIOPS:         storageInfo.MaxIOPS,
		MaxBandwidth:    parse.FormatCapacity(storageInfo.MaxBandwidth),
	}
}
