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

import { token } from '@oceanbase/design';
import { formatMessage } from '@/util/intl';

export const COMPACTION_STATUS_LISTV4: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.TenantMergeException',
      defaultMessage: '租户合并异常',
    }),
    value: 'ERROR',
    badgeStatus: 'error',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.TenantIdle',
      defaultMessage: '租户空闲',
    }),
    value: 'IDLE',
    badgeStatus: 'default',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.TenantMerging',
      defaultMessage: '租户合并中',
    }),
    value: 'COMPACTING',
    badgeStatus: 'processing',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.TenantMerging',
      defaultMessage: '租户合并中',
    }),
    value: 'VERIFYING',
    badgeStatus: 'processing',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.WaitingForMergeScheduling',
      defaultMessage: '等待合并调度中',
    }),

    value: 'WAIT_MERGE',
    badgeStatus: 'processing',
  },
];

export const COMPACTION_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Error',
      defaultMessage: '出错',
    }),
    value: 'ERROR',
    badgeStatus: 'error',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Timeout',
      defaultMessage: '超时',
    }),
    value: 'TIMEOUT',
    badgeStatus: 'warning',
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.oceanbase.Idle', defaultMessage: '空闲' }),
    value: 'IDLE',
    badgeStatus: 'default',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Merging',
      defaultMessage: '合并中',
    }),
    value: 'COMPACTING',
    badgeStatus: 'processing',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Merging',
      defaultMessage: '合并中',
    }),
    value: 'MERGING',
    badgeStatus: 'processing',
  },

  {
    // 合并中，状态文案展示与 MERGING 保持一致
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Merging',
      defaultMessage: '合并中',
    }),
    value: 'MOST_MERGED',
    badgeStatus: 'success',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Index',
      defaultMessage: '索引中',
    }),
    value: 'INDEXING',
    badgeStatus: 'success',
  },

  {
    // MERGING 之后做 INDEX， OB 2.2开始 INDEX 状态已经不使用了
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Index',
      defaultMessage: '索引中',
    }),
    value: 'INDEX',
    badgeStatus: 'success',
  },

  // IDLE 状态也可能是等待合并调度中，此处前端扩展出 WAIT_MERGE 状态
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.WaitingForMergeScheduling',
      defaultMessage: '等待合并调度中',
    }),

    value: 'WAIT_MERGE',
    badgeStatus: 'processing',
  },
];

// 合并结果列表
export const COMPACTION_RESULT_LIST = [
  {
    value: 'FAIL',
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.Exception',
      defaultMessage: '异常',
    }),
    color: token.colorError,
  },

  {
    value: 'COMPACTING',
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.Merging',
      defaultMessage: '合并中',
    }),
    color: '#5B8FF9',
  },

  {
    value: 'SUCCESS',
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.Completed',
      defaultMessage: '已完成',
    }),
    color: '#5AD8A6',
  },

  {
    // IDLE 为前端定义的合并结果，用于展示灰底
    value: 'IDLE',
    label: formatMessage({
      id: 'ocp-express.src.constant.compaction.Idle',
      defaultMessage: '空闲中',
    }),
    color: 'rgba(0,0,0,0.06)',
  },
];
