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

import { AUTO_FRESH_ITEMS_TYPE } from '@/component/MonitorSearch/AutoFresh';
import { formatMessage } from '@/util/intl';

export const MAX_POINTS = 120;

export const AUTO_FRESH_ITEMS: AUTO_FRESH_ITEMS_TYPE[] = [
  {
    key: 'off',
    label: 'Off',
  },
  {
    key: 5,
    label: '5s',
  },
  {
    key: 10,
    label: '10s',
  },
  {
    key: 30,
    label: '30s',
  },
];

export const DEFAULT_QUERY_RANGE: Monitor.QueryRangeType = {
  step: 20,
  end_timestamp: Math.floor(new Date().valueOf() / 1000),
  start_timestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
};

export const REFRESH_FREQUENCY = 15;

export const POINT_NUMBER = 15;

// OB Charts 基础色板
export const BASE_CHART_COLORS = [
  '#3D88F2',
  '#41D9A6',
  '#FAC357',
  '#547199',
  '#79BFF2',
  '#4D997F',
  '#88CC66',
  '#B3749E',
  '#E6987F',
  '#8C675B',
];

// OB Charts 延伸色板
export const EXTEND_CHART_COLORS = [
  '#3D88F2',
  '#8BB8F7',
  '#41D9A6',
  '#8DE8CA',
  '#FAC357',
  '#FCDB9A',
  '#547199',
  '#98AAC2',
  '#79BFF2',
  '#AFD9F7',
  '#4D997F',
  '#94C2B3',
  '#88CC66',
  '#B7E0A3',
  '#B3749E',
  '#D1ACC5',
  '#E6987F',
  '#F0C1B2',
  '#8C675B',
  '#BAA49D',
];

const Dimension = {
  databasePerformance: [
    {
      value: 'ob_cluster_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.ClusterDimension',
        defaultMessage: '集群维度',
      }),
      singleMachineFlag: true,
    },
    {
      value: 'tenant_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.TenantDimension',
        defaultMessage: '租户维度',
      }),
      singleMachineFlag: true,
    },
    {
      value: 'svr_ip',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.ObserverDimension',
        defaultMessage: 'OBServer 维度',
      }),
      singleMachineFlag: false,
    },
    {
      value: 'unit',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.UnitDimension',
        defaultMessage: 'Unit维度',
      }),
      singleMachineFlag: false,
    },
  ],

  hostPerformance: [
    {
      value: 'ob_cluster_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.ClusterDimension',
        defaultMessage: '集群维度',
      }),
      singleMachineFlag: true,
    },
    {
      value: 'svr_ip',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.HostDimension',
        defaultMessage: '主机维度',
      }),
      singleMachineFlag: true,
    },
  ],

  observerPerformance: [
    {
      value: 'ob_cluster_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.ClusterDimension',
        defaultMessage: '集群维度',
      }),
      singleMachineFlag: false,
    },
    {
      value: 'svr_ip',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.ObserverDimension',
        defaultMessage: 'OBServer 维度',
      }),
      singleMachineFlag: false,
    },
  ],

  performanceAndSQL: [
    {
      value: 'tenant_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.TenantDimension',
        defaultMessage: '租户维度',
      }),
      singleMachineFlag: false,
    },
    {
      value: 'unit',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.UnitDimension',
        defaultMessage: 'Unit维度',
      }),
      singleMachineFlag: false,
    },
  ],

  transaction: [
    {
      value: 'tenant_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.TenantDimension',
        defaultMessage: '租户维度',
      }),
      singleMachineFlag: false,
    },
    {
      value: 'unit',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.UnitDimension',
        defaultMessage: 'Unit维度',
      }),
      singleMachineFlag: false,
    },
  ],

  storageAndCache: [
    {
      value: 'tenant_name',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.TenantDimension',
        defaultMessage: '租户维度',
      }),
      singleMachineFlag: false,
    },
    {
      value: 'unit',
      label: formatMessage({
        id: 'OBShell.src.constant.monitor.UnitDimension',
        defaultMessage: 'Unit维度',
      }),
      singleMachineFlag: false,
    },
  ],
};

export const DimensionMap = {
  数据库性能: Dimension['databasePerformance'],
  主机性能: Dimension['hostPerformance'],
  OBServer性能: Dimension['observerPerformance'],
  性能与SQL: Dimension['performanceAndSQL'],
  事务: Dimension['transaction'],
  存储与缓存: Dimension['storageAndCache'],
  'Database Performance': Dimension['databasePerformance'],
  'Host performance': Dimension['hostPerformance'],
  'OBServer performance': Dimension['observerPerformance'],
  'Performance and SQL': Dimension['performanceAndSQL'],
  Transaction: Dimension['transaction'],
  'Storage and Cache': Dimension['storageAndCache'],
};
