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

export const OB_CLUSTER_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Running',
      defaultMessage: '运行中',
    }),
    value: 'RUNNING',
    badgeStatus: 'success',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Deleting',
      defaultMessage: '删除中',
    }),

    value: 'DELETING',
    badgeStatus: 'warning',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Upgrade',
      defaultMessage: '升级中',
    }),
    value: 'UPGRADING',
    badgeStatus: 'warning',
  },

  {
    label: formatMessage({ id: 'ocp-v2.src.constant.oceanbase.Available', defaultMessage: '可用' }),
    value: 'AVAILABLE',
    badgeStatus: 'success',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Unavailable',
      defaultMessage: '不可用',
    }),

    value: 'UNAVAILABLE',
    badgeStatus: 'error',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.OperationsIn',
      defaultMessage: '运维中',
    }),

    value: 'OPERATING',
    badgeStatus: 'warning',
  },
];

export const ZONE_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Running',
      defaultMessage: '运行中',
    }),
    value: 'RUNNING',
    badgeStatus: 'success',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Stop.4',
          defaultMessage: '停止',
        }),

        value: 'stop',
        isDanger: true,
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Stop',
      defaultMessage: '停止中',
    }),
    value: 'STOPPING',
    badgeStatus: 'processing',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'OBShell.src.constant.oceanbase.Stopped',
      defaultMessage: '已停止',
    }),
    value: 'STOPPED',
    badgeStatus: 'default',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Start',
          defaultMessage: '启动',
        }),
        value: 'start',
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Starting',
      defaultMessage: '启动中',
    }),

    value: 'STARTING',
    badgeStatus: 'processing',
    operations: [],
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.oceanbase.Unknown', defaultMessage: '未知' }),
    value: 'UNKNOWN',
    badgeStatus: 'default',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Start',
          defaultMessage: '启动',
        }),
        value: 'start',
      },
    ],
  },
];

export const OB_SERVER_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Running',
      defaultMessage: '运行中',
    }),
    value: 'RUNNING',
    badgeStatus: 'success', //processing
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.StopService',
          defaultMessage: '停止服务',
        }),

        modalTitle: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.StopObserverService',
          defaultMessage: '停止 OBServer 服务',
        }),

        value: 'stopService',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.StopProcess',
          defaultMessage: '停止进程',
        }),

        modalTitle: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.StopObserverProcess',
          defaultMessage: '停止 OBServer 进程',
        }),

        value: 'stopProcess',
        isDanger: true,
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Unavailable',
      defaultMessage: '不可用',
    }),

    value: 'UNAVAILABLE',
    badgeStatus: 'error',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Start',
          defaultMessage: '启动',
        }),
        value: 'start',
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ServiceStopping',
      defaultMessage: '服务停止中',
    }),

    value: 'SERVICE_STOPPING',
    badgeStatus: 'processing',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.TheProcessIsStopped',
      defaultMessage: '进程停止中',
    }),

    value: 'PROCESS_STOPPING',
    badgeStatus: 'processing',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ServiceStopped',
      defaultMessage: '服务已停止',
    }),

    value: 'SERVICE_STOPPED',
    badgeStatus: 'warning',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Start',
          defaultMessage: '启动',
        }),
        value: 'start',
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.StopProcess',
          defaultMessage: '停止进程',
        }),

        modalTitle: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.StopObserverProcess',
          defaultMessage: '停止 OBServer 进程',
        }),

        value: 'stopProcess',
        isDanger: true,
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.TheProcessHasStopped',
      defaultMessage: '进程已停止',
    }),

    value: 'PROCESS_STOPPED',
    badgeStatus: 'default',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Start',
          defaultMessage: '启动',
        }),
        value: 'start',
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Starting',
      defaultMessage: '启动中',
    }),

    value: 'STARTING',
    badgeStatus: 'processing',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Deleting',
      defaultMessage: '删除中',
    }),

    value: 'DELETING',
    badgeStatus: 'processing',
    operations: [],
  },
];

export const REPLICA_TYPE_LIST = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.AllPurposeReplica',
      defaultMessage: '全能型副本',
    }),
    shortLabel: 'F',
    value: 'FULL',
    description: formatMessage({
      id: 'ocp-v2.src.constant.oceanbase.AUniversalCopyIsCalled',
      defaultMessage: '全能型副本称为 Paxos 副本，对应副本可构成 Paxos 成员组，参与选举投票',
    }),
    // description: formatMessage({
    //   id: 'ocp-express.src.constant.oceanbase.FullFeaturedCopyCurrentlyIt',
    //   defaultMessage:
    //     '全功能型副本：目前支持的普通副本，拥有事务日志、MemTable 和 SSTable 等全部完整的数据和功能。它可以随时快速切换为 Leader 以对外提供服务。',
    // }),
  },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.LogCopy',
  //     defaultMessage: '日志型副本',
  //   }),
  //   shortLabel: 'L',
  //   value: 'LOGONLY',
  //   description: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.LogTypeCopyContainsOnly',
  //     defaultMessage:
  //       '日志型副本：只包含日志的副本，没有 MemTable 和 SSTable。它参与日志投票并对外提供日志服务，可以参与其他副本的恢复，但自己不能变为主提供数据库服务。',
  //   }),
  // },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ReadOnlyCopy',
      defaultMessage: '只读型副本',
    }),
    shortLabel: 'R',
    value: 'READONLY',
    description: formatMessage({
      id: 'ocp-v2.src.constant.oceanbase.ForVXVersionsOceanbase',
      defaultMessage:
        '对应 V4.x 版本，OceanBase 数据库从 V4.2.0 版本开始支持只读型副本。只读型副本为非 Paxos 副本，对应副本不可构成 Paxos 成员组，不参与选举投票。',
    }),

    // description: formatMessage({
    //   id: 'ocp-express.src.constant.oceanbase.ReadOnlyCopyContainsComplete',
    //   defaultMessage:
    //     '只读型副本：包含完整的日志、MemTable 和 SSTable 等。但是它的日志比较特殊，它不作为 Paxos 成员参与日志的投票，而是作为一个观察者实时追赶 Paxos 成员的日志，并在本地回放。这种副本可以在业务对读取数据的一致性要求不高的时候提供只读服务。因其不加入 Paxos 成员组，又不会造成投票成员增加导致事务提交延时的增加。',
    // }),
  },
];
