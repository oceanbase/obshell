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
