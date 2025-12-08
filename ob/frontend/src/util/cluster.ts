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

import { DATE_FORMAT } from '@/constant/datetime';
import { toNumber } from 'lodash';
import { formatTime } from '@/util/datetime';
/* 获取 4.0 以下集群实际的合并状态
 */
export function getCompactionStatus(
  clusterCompaction?: API.ClusterCompaction | API.TenantCompaction
) {
  // IDLE 状态也可能是等待合并调度中，此处前端扩展出 WAIT_MERGE 状态
  if (
    clusterCompaction?.status === 'IDLE' &&
    (clusterCompaction?.broadcastVersion || 0) < (clusterCompaction?.freezeVersion || 0)
  ) {
    return 'WAIT_MERGE';
  }
  return clusterCompaction?.status;
}
/* 获取集群实际的合并状态
IDLE -> 空闲
COMPACTING | VERIFYING -> 合并中
error = true -> 合并失败
 */
export function getCompactionStatusV4(tenantCompactionsList: API.TenantCompaction[]) {
  // 合并出错
  const compactionsError =
    tenantCompactionsList?.filter(item => item?.error || item?.is_error !== 'NO').length > 0;
  // 合并中
  const isCompactionsMerging =
    tenantCompactionsList.filter(
      item => item?.status === 'VERIFYING' || item?.status === 'COMPACTING'
    ).length > 0;
  // IDLE 状态也可能是等待合并调度中，此处前端扩展出 WAIT_MERGE 状态
  const isCompactionsWaitMerge =
    tenantCompactionsList.filter(
      item => item?.status === 'IDLE' && (item?.broadcastScn || 0) !== (item?.frozenScn || 0)
    ).length === tenantCompactionsList.length;
  // 空闲
  const isCompactionsIdle =
    tenantCompactionsList.filter(item => item?.status === 'IDLE').length ===
    tenantCompactionsList.length;

  if (compactionsError) {
    return 'ERROR';
  } else if (isCompactionsMerging) {
    return 'COMPACTING';
  } else if (isCompactionsWaitMerge) {
    return 'WAIT_MERGE';
  } else if (isCompactionsIdle) {
    return 'IDLE';
  }
}

/***
 * 1. unit 数量不允许超过 zone 中的 observer 数量
 * */
export function getUnitSpecLimit(zoneStats: API.ServerResourceStats) {
  const {
    cpu_core_total: cpuCoreTotal,
    cpu_core_assigned: cpuCoreAssigned,
    memory_in_bytes_total: memoryInBytesTotal,
    memory_in_bytes_assigned: memoryInBytesAssigned,
  } = zoneStats;
  let idleCpuCoreTotal, idleMemoryInBytesTotal;
  if (cpuCoreTotal && cpuCoreAssigned && memoryInBytesTotal && memoryInBytesAssigned) {
    // OBServer 剩余资源
    idleCpuCoreTotal = cpuCoreTotal - cpuCoreAssigned;
    idleMemoryInBytesTotal = Math.floor(
      toNumber((memoryInBytesTotal - memoryInBytesAssigned) / (1024 * 1024 * 1024))
    );
  }

  return { idleCpuCoreTotal, idleMemoryInBytesTotal };
}

/***
 * 判断OBServer 剩余资源是否充足，满足 4核8GB 时展示默认值  默认值 4核8GB
 * */
export function getResourcesLimit(idleUnitSpec) {
  const { idleCpuCore, idleMemoryInBytes } = idleUnitSpec;
  return idleCpuCore > 4 && idleMemoryInBytes > 8;
}

// 是否为试用 license
export const isTrialLicense = (license: API.ObLicense) => {
  return license?.license_type === 'Trial';
};

// 集群 license 过期时间计算
export const getLicenseExpiredTime = (license: API.ObLicense) => {
  return formatTime(license?.expired_time, DATE_FORMAT);
};

// 计算距离过期时间还有多少天
export const getDaysUntilExpiry = (time: string) => {
  const expiryDate = new Date(time);
  const now = new Date();
  // 一天的毫秒数
  const millisecondsPerDay = 24 * 60 * 60 * 1000;
  // 计算时间差（过期时间 - 当前时间）
  const timeDiff = expiryDate - now;
  return Math.ceil(timeDiff / millisecondsPerDay);
};

// 检查是否为有效的试用license（试用且未过期）
export const isEffectiveTrial = (license: API.ObLicense): boolean => {
  const expiredTime = getLicenseExpiredTime(license);
  const effectiveDays = getDaysUntilExpiry(expiredTime) || 0;
  return isTrialLicense(license) && effectiveDays > 0;
};
