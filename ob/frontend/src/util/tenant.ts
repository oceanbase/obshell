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

import { REPLICA_TYPE_LIST } from '@/constant/oceanbase';
import { byte2GB } from '@oceanbase/util';
import { min } from 'lodash';

// 获取集群各个 Zone 中 OBServer 数量的最小值，用于限制 OB 4.0 版本中可设置 Unit 的最大数量
export function getMinServerCount(zones: API.Zone[]) {
  if (zones?.length > 0) {
    return min(zones.map(zone => zone?.servers?.length || 0));
  }
  return 0;
}

export interface TenantZone extends API.Zone {
  replicaType?: string;
  name: string;
  unitNum?: number;
  resourcePool: {
    unitConfig: {
      cpuCore?: number;
      memorySize?: number;
      dataDiskSize?: number;
    } & API.ObUnitConfig;
  } & API.ResourcePoolWithUnit;
}

export const getZonesFromTenant: (tenantData: API.TenantInfo) => TenantZone[] = (
  tenantData: API.TenantInfo
) => {
  return (
    tenantData?.locality?.split(',')?.map(str => {
      const [replicaTypeStr, zoneName] = str?.split('@');
      const replicaType = REPLICA_TYPE_LIST.find(item =>
        replicaTypeStr.includes(item.value)
      )?.value;
      const resourcePool = tenantData?.pools?.find(item => item.zone_list?.includes(zoneName));
      return {
        replicaType,
        name: zoneName,
        unitNum: resourcePool?.unit_num,
        resourcePool: {
          ...resourcePool,
          unitConfig: {
            ...resourcePool?.unit_config,
            // 副本详情默认值
            cpuCore: resourcePool?.unit_config?.max_cpu,
            memorySize: byte2GB(resourcePool?.unit_config?.memory_size || 0),
          },
        },
      };
    }) || []
  );
};
