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

export const OPERATION_LIST = [
  {
    value: 'share',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Share',
      defaultMessage: '分享',
    }),
  },

  {
    value: 'edit',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Edit',
      defaultMessage: '编辑',
    }),
  },

  {
    value: 'export',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Export',
      defaultMessage: '导出',
    }),
  },

  {
    value: 'delete',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Delete',
      defaultMessage: '删除',
    }),
  },
];

export const TAB_LIST = [
  {
    value: 'OB',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.ObCluster',
      defaultMessage: 'OB 集群',
    }),
  },

  {
    value: 'OB_PROXY',
    label: 'OBProxy',
  },

  {
    value: 'HOST',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Host',
      defaultMessage: '主机',
    }),
  },
];

export const HOST_CONNECTION_RESULT = [
  {
    value: 'CONNECT_FAILED',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.ConnectionFailed',
      defaultMessage: '连接失败',
    }),
  },

  {
    value: 'SUDO_NOT_ALLOWED',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.NoSudoPermission',
      defaultMessage: '无 sudo 权限',
    }),
  },

  {
    value: 'HOST_NOT_EXISTS',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.TheHostDoesNotExist',
      defaultMessage: '主机不存在',
    }),
  },
];
