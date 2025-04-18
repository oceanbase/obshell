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

import { formatMessage } from '@/util/intl';
import { min } from 'lodash';

// 获取集群各个 Zone 中 OBServer 数量的最小值，用于限制 OB 4.0 版本中可设置 Unit 的最大数量
export function getMinServerCount(zones: API.Zone[]) {
  if (zones?.length > 0) {
    return min(zones.map(zone => zone?.servers?.length || 0));
  }
  return 0;
}

// 根据  unitSpecLimit 说明不推荐使用的原因
export function getUnitSpecLimitText(unitSpecLimit: API.UnitSpecUnitSpecLimit) {
  const { memoryLowerLimit, cpuLowerLimit } = unitSpecLimit;
  if (memoryLowerLimit) {
    return formatMessage(
      {
        id: 'ocp-express.src.util.tenant.OnlyUnitSpecificationsWithAMemoryOfNo',
        defaultMessage: '仅可选择内存不小于 {memoryLowerLimit}G 的 Unit 规格',
      },
      { memoryLowerLimit: memoryLowerLimit }
    );
  }
  if (cpuLowerLimit) {
    return formatMessage(
      {
        id: 'ocp-express.src.util.tenant.OnlyUnitSpecificationsWithCpuNoLessThan',
        defaultMessage: '仅可选择 CPU 不小于 {cpuLowerLimit}C 的 Unit 规格',
      },
      { cpuLowerLimit: cpuLowerLimit }
    );
  }
}
