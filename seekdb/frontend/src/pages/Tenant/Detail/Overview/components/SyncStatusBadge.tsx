import { formatMessage } from '@/util/intl';
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

import { Badge } from '@oceanbase/design';
import React from 'react';

export interface SyncStatusBadgeProps {
  status?: string;
}

const statusMap: Record<
  string,
  { badgeStatus: 'success' | 'error' | 'processing' | 'warning' | 'default'; label: string }
> = {
  NORMAL_SYNC: {
    badgeStatus: 'success',
    label: formatMessage({
      id: 'seekdb.Overview.components.SyncStatusBadge.NormalSynchronization',
      defaultMessage: '正常同步',
    }),
  },
  SYNC_DELAYED: {
    badgeStatus: 'warning',
    label: formatMessage({
      id: 'seekdb.Overview.components.SyncStatusBadge.SynchronizationDelay',
      defaultMessage: '同步延迟',
    }),
  },
  SYNC_PAUSED: {
    badgeStatus: 'warning',
    label: formatMessage({
      id: 'seekdb.Overview.components.SyncStatusBadge.SynchronizationPaused',
      defaultMessage: '同步暂停',
    }),
  },
  UNKNOWN: {
    badgeStatus: 'default',
    label: formatMessage({
      id: 'seekdb.Overview.components.SyncStatusBadge.StatusUnknown',
      defaultMessage: '状态未知',
    }),
  },
};

const SyncStatusBadge: React.FC<SyncStatusBadgeProps> = ({ status }) => {
  const currentStatus = status ? statusMap[status] : undefined;

  return (
    <Badge
      status={currentStatus?.badgeStatus || 'default'}
      text={
        currentStatus?.label ||
        formatMessage({
          id: 'seekdb.Overview.components.SyncStatusBadge.Unknown',
          defaultMessage: '未知',
        })
      }
    />
  );
};

export default SyncStatusBadge;
