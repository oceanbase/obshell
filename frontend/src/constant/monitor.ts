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

import moment from 'moment';
import { formatMessage } from '@/util/intl';

export const MAX_POINTS = 120;

export const MONITOR_SCOPE_LIST = [
  {
    value: 'app',
    label: formatMessage({
      id: 'ocp-express.src.constant.monitor.Application',
      defaultMessage: '应用',
    }),
  },

  {
    value: 'obregion',
    label: formatMessage({
      id: 'ocp-express.src.constant.monitor.Cluster',
      defaultMessage: '集群',
    }),
  },

  {
    value: 'ob_cluster_id',
    label: formatMessage({
      id: 'ocp-express.src.constant.monitor.Cluster',
      defaultMessage: '集群',
    }),
  },

  {
    value: 'tenant_name',
    label: formatMessage({ id: 'ocp-express.src.constant.monitor.Tenant', defaultMessage: '租户' }),
  },

  {
    value: 'obzone',
    label: 'Zone',
  },

  {
    value: 'svr_ip',
    label: formatMessage({ id: 'ocp-express.src.constant.monitor.Host', defaultMessage: '主机' }),
  },

  {
    value: 'device',
    label: formatMessage({
      id: 'ocp-express.src.constant.monitor.Equipment',
      defaultMessage: '设备',
    }),
  },
];

export function getRanges() {
  const rangeList = [
    {
      label: formatMessage({
        id: 'ocp-express.src.constant.monitor.TwentyMinutes',
        defaultMessage: '二十分钟',
      }),

      value: () => [moment().subtract(20, 'minutes'), moment()],
    },

    {
      label: formatMessage({
        id: 'ocp-express.src.constant.monitor.HalfAnHour',
        defaultMessage: '半小时',
      }),

      value: () => [moment().subtract(0.5, 'hours'), moment()],
    },

    {
      label: formatMessage({
        id: 'ocp-express.src.constant.monitor.AnHour',
        defaultMessage: '一小时',
      }),
      value: () => [moment().subtract(1, 'hours'), moment()],
    },

    {
      label: formatMessage({ id: 'ocp-express.src.constant.monitor.Day', defaultMessage: '一天' }),
      value: () => [moment().subtract(1, 'days'), moment()],
    },

    {
      label: formatMessage({
        id: 'ocp-express.src.constant.monitor.AWeek',
        defaultMessage: '一周',
      }),
      value: () => [moment().subtract(1, 'weeks'), moment()],
    },

    {
      label: formatMessage({
        id: 'ocp-express.src.constant.monitor.January',
        defaultMessage: '一月',
      }),
      value: () => [moment().subtract(1, 'months'), moment()],
    },
  ];

  const ranges = {};
  rangeList.forEach(item => {
    ranges[item.label] = (item.value && item.value()) || [];
  });
  return ranges;
}
