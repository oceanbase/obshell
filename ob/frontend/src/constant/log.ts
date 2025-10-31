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

// 获取适用于 ob-ui Ranger 组件的 selects，该格式与 antd 的不同
export function getSelects() {
  return [
    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearFenZhong',
        defaultMessage: '近 5 分钟',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 5 Minutes',
        },
      ],

      range: () => [moment().subtract(5, 'minute'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearFenZhong.1',
        defaultMessage: '近 10 分钟',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 10 Minutes',
        },
      ],

      range: () => [moment().subtract(10, 'minute'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearMins',
        defaultMessage: '近 15 分钟',
      }),
      locale: [
        {
          lang: 'en-US',
          name: 'Last 15 Minutes',
        },
      ],

      range: () => [moment().subtract(15, 'minute'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearFenZhong.2',
        defaultMessage: '近 30 分钟',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 30 Minutes',
        },
      ],

      range: () => [moment().subtract(30, 'minute'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearlyHour',
        defaultMessage: '近 1 小时',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 1 Hour',
        },
      ],

      range: () => [moment().subtract(1, 'hours'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearlyHours',
        defaultMessage: '近 3 小时',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 3 Hours',
        },
      ],

      range: () => [moment().subtract(3, 'hours'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearlyHours.1',
        defaultMessage: '近 6 小时',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 6 Hours',
        },
      ],

      range: () => [moment().subtract(6, 'hours'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearlyHours.2',
        defaultMessage: '近 12 小时',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 12 Hours',
        },
      ],

      range: () => [moment().subtract(12, 'hours'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.log.NearHours',
        defaultMessage: '近 24 小时',
      }),

      locale: [
        {
          lang: 'en-US',
          name: 'Last 24 Hours',
        },
      ],

      range: () => [moment().subtract(24, 'hours'), moment()],
    },
  ];
}

export const LOG_TYPE_LIST = [
  {
    value: 'CLUSTER',
    label: formatMessage({
      id: 'ocp-express.src.constant.compute.ObserverLog',
      defaultMessage: 'OBServer 日志',
    }),
    types: ['observer', 'rootservice', 'election'],
  },

  {
    value: 'OBAGENT',
    label: formatMessage({
      id: 'ocp-express.src.constant.log.ObagentLogs',
      defaultMessage: 'OBAgent 日志',
    }),
    types: ['mgragent', 'monagent', 'agentctl', 'agentd'],
  },
];

export const LOG_LEVEL = ['ERROR', 'WARN', 'INFO', 'EDIAG', 'WDIAG', 'TRACE', 'DEBUG'];
