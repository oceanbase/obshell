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

export const OCP_AGENT_PROCESS_STATUS_LIST = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compute.Running',
      defaultMessage: '运行中',
    }),
    value: 'RUNNING',
    badgeStatus: 'processing',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compute.Unknown',
      defaultMessage: '未知',
    }),
    value: 'UNKNOWN',
    badgeStatus: 'warning',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compute.Starting',
      defaultMessage: '启动中',
    }),
    value: 'STARTING',
    badgeStatus: 'success',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compute.Stopped',
      defaultMessage: '停止中',
    }),
    value: 'STOPPING',
    badgeStatus: 'error',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.compute.NotRunning',
      defaultMessage: '未运行',
    }),
    value: 'STOPPED',
    badgeStatus: 'default',
  },
];
