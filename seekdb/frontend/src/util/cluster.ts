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
