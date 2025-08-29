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

export const OB_CLUSTER_TYPE_LIST = [
  {
    value: 'PRIMARY',
    label: formatMessage({
      id: 'OBShell.src.constant.oceanbase.PrimaryCluster',
      defaultMessage: '主集群',
    }),
    signLabel: formatMessage({ id: 'OBShell.src.constant.oceanbase.Lord', defaultMessage: '主' }),
    backgroundColor: '#3B55A3',
  },
  {
    value: 'STANDBY',
    label: formatMessage({
      id: 'OBShell.src.constant.oceanbase.StandbyCluster',
      defaultMessage: '备集群',
    }),
    signLabel: formatMessage({
      id: 'OBShell.src.constant.oceanbase.Preparation',
      defaultMessage: '备',
    }),
    backgroundColor: '#a3a3a3',
  },
];

export const SELECT_MODE_LIST = [
  {
    label: formatMessage({
      id: 'ocp-express.Cluster.New.AutomaticAllocation',
      defaultMessage: '自动分配',
    }),

    value: 'auto',
  },

  {
    label: formatMessage({
      id: 'ocp-express.Cluster.New.ManualSelection',
      defaultMessage: '手动选择',
    }),
    value: 'select',
  },
];

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

export const PROTECTION_MODE_LIST = [
  {
    value: 'MAXIMUM_PERFORMANCE',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MaximumPerformance',
      defaultMessage: '最大性能',
    }),

    description: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MaximumPerformanceProtectionModeIt',
      defaultMessage: '最大性能保护模式：它在最大限度地确保主集群性能的同时，还能保护用户的数据。',
    }),
  },

  {
    value: 'MAXIMUM_PROTECTION',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MaximumProtection',
      defaultMessage: '最大保护',
    }),

    description: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ItProvidesTheHighestLevel',
      defaultMessage: '它提供了最高级别的数据保护，可以确保主集群出现故障时不会发生数据丢失',
    }),
  },

  {
    value: 'MAXIMUM_AVAILABILITY',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MaximumAvailability',
      defaultMessage: '最大可用',
    }),

    description: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ItProvidesTheHighestLevel.1',
      defaultMessage: '它在不牺牲集群可用性的前提下提供了最高级别的数据保护',
    }),
  },
];

// 日志传输模式列表
export const REDO_TRANSPORT_MODE_LIST = [
  {
    value: 'SYNC',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Synchronous',
      defaultMessage: '同步',
    }),
  },

  {
    value: 'ASYNC',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Asynchronous',
      defaultMessage: '异步',
    }),
  },
];

// 日志传输状态列表
export const REDO_TRANSPORT_STATUS_LIST = [
  {
    value: 'SYNC',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Synchronous',
      defaultMessage: '同步',
    }),
  },

  {
    value: 'ASYNC',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Asynchronous',
      defaultMessage: '异步',
    }),
  },
];

export const ZONE_STATUS_LIST: Global.StatusItem[] = [
  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Creating',
  //     defaultMessage: '创建中',
  //   }),

  //   value: 'CREATING',
  //   badgeStatus: 'processing',
  //   // 需要遵守一定顺序，越危险的操作位置越后
  //   operations: [],
  // },

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
          id: 'ocp-express.src.constant.oceanbase.AddObserver',
          defaultMessage: '添加 OBServer',
        }),

        value: 'addServer',
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Restart.4',
          defaultMessage: '重启',
        }),

        value: 'restart',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Stop.4',
          defaultMessage: '停止',
        }),

        value: 'stop',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),

        value: 'delete',
        isDanger: true,
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Unavailable',
      defaultMessage: '不可用',
    }),

    value: 'SERVICE_STOPPED',
    badgeStatus: 'error',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.AddObserver',
          defaultMessage: '添加 OBServer',
        }),

        value: 'addServer',
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Restart.4',
          defaultMessage: '重启',
        }),

        value: 'restart',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Stop.4',
          defaultMessage: '停止',
        }),

        value: 'stop',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),

        value: 'delete',
        isDanger: true,
      },
    ],
  },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Deleting',
  //     defaultMessage: '删除中',
  //   }),

  //   value: 'DELETING',
  //   badgeStatus: 'warning',
  //   operations: [],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Stop',
  //     defaultMessage: '停止中',
  //   }),
  //   value: 'STOPPING',
  //   badgeStatus: 'warning',
  //   operations: [],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Stopped',
  //     defaultMessage: '已停止',
  //   }),
  //   value: 'STOPPED',
  //   badgeStatus: 'default',
  //   operations: [
  //     {
  //       label: formatMessage({
  //         id: 'ocp-express.src.constant.oceanbase.Start',
  //         defaultMessage: '启动',
  //       }),
  //       value: 'start',
  //     },

  //     {
  //       label: formatMessage({
  //         id: 'ocp-express.src.constant.oceanbase.Delete',
  //         defaultMessage: '删除',
  //       }),

  //       value: 'delete',
  //       isDanger: true,
  //     },
  //   ],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Starting',
  //     defaultMessage: '启动中',
  //   }),

  //   value: 'STARTING',
  //   badgeStatus: 'processing',
  //   operations: [],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Restart',
  //     defaultMessage: '重启中',
  //   }),
  //   value: 'RESTARTING',
  //   badgeStatus: 'processing',
  //   operations: [],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.Deleted',
  //     defaultMessage: '已删除',
  //   }),
  //   value: 'DELETED',
  //   badgeStatus: 'default',
  //   operations: [
  //     {
  //       label: formatMessage({
  //         id: 'ocp-express.src.constant.oceanbase.ReleaseAHost',
  //         defaultMessage: '释放主机',
  //       }),

  //       value: 'setFree',
  //     },
  //   ],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.OperationsIn',
  //     defaultMessage: '运维中',
  //   }),

  //   value: 'OPERATING',
  //   badgeStatus: 'processing',
  //   operations: [],
  // },

  // {
  //   label: formatMessage({
  //     id: 'ocp-express.src.constant.oceanbase.ServiceStopped',
  //     defaultMessage: '服务已停止',
  //   }),

  //   value: 'SERVICE_STOPPED',
  //   badgeStatus: 'warning',
  //   operations: [
  //     {
  //       label: formatMessage({
  //         id: 'ocp-express.src.constant.oceanbase.Start',
  //         defaultMessage: '启动',
  //       }),
  //       value: 'start',
  //     },

  //     {
  //       label: formatMessage({
  //         id: 'ocp-express.src.constant.oceanbase.Restart.4',
  //         defaultMessage: '重启',
  //       }),

  //       value: 'restart',
  //       isDanger: true,
  //     },

  //     {
  //       label: formatMessage({
  //         id: 'ocp-express.src.constant.oceanbase.Delete',
  //         defaultMessage: '删除',
  //       }),

  //       value: 'delete',
  //       isDanger: true,
  //     },
  //   ],
  // },
];

export const OB_SERVER_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Creating',
      defaultMessage: '创建中',
    }),

    value: 'CREATING',
    badgeStatus: 'processing',
    // 需要遵守一定顺序，越危险的操作位置越后
    operations: [],
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Reloading',
      defaultMessage: '重装中',
    }),

    value: 'REINSTALLING',
    badgeStatus: 'warning',
    operations: [],
  },

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
          id: 'ocp-express.src.constant.oceanbase.Restart.4',
          defaultMessage: '重启',
        }),

        value: 'restart',
        isDanger: true,
      },

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

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Replacement',
          defaultMessage: '替换',
        }),

        value: 'replace',
        isDanger: true,
      },
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Reload',
          defaultMessage: '重装',
        }),

        value: 'reInstall',
        isDanger: true,
      },
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),

        value: 'delete',
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

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Replacement',
          defaultMessage: '替换',
        }),

        value: 'replace',
        isDanger: true,
      },
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Reload',
          defaultMessage: '重装',
        }),

        value: 'reInstall',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),

        value: 'delete',
        isDanger: true,
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ServiceStopping',
      defaultMessage: '服务停止中',
    }),

    value: 'SERVICE_STOPPING',
    badgeStatus: 'warning',
    operations: [
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
      id: 'ocp-express.src.constant.oceanbase.TheProcessIsStopped',
      defaultMessage: '进程停止中',
    }),

    value: 'PROCESS_STOPPING',
    badgeStatus: 'warning',
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

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Replacement',
          defaultMessage: '替换',
        }),

        value: 'replace',
        isDanger: true,
      },
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Reload',
          defaultMessage: '重装',
        }),

        value: 'reInstall',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),

        value: 'delete',
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

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Replacement',
          defaultMessage: '替换',
        }),

        value: 'replace',
        isDanger: true,
      },
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Reload',
          defaultMessage: '重装',
        }),

        value: 'reInstall',
        isDanger: true,
      },

      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),

        value: 'delete',
        isDanger: true,
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Starting',
      defaultMessage: '启动中',
    }),

    value: 'STARTING',
    badgeStatus: 'warning',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Restart',
      defaultMessage: '重启中',
    }),
    value: 'RESTARTING',
    badgeStatus: 'warning',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Deleting',
      defaultMessage: '删除中',
    }),

    value: 'DELETING',
    badgeStatus: 'warning',
    operations: [],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Deleted',
      defaultMessage: '已删除',
    }),
    value: 'DELETED',
    badgeStatus: 'default',
    operations: [
      {
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.ReleaseAHost',
          defaultMessage: '释放主机',
        }),

        modalTitle: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.ReleaseAHost',
          defaultMessage: '释放主机',
        }),

        value: 'setFree',
      },
    ],
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

/** Unit 规格管理相关 */
export const UNIT_SPEC_TYPE_LIST = [
  {
    value: 'SYSTEM',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Default',
      defaultMessage: '系统默认',
    }),

    operations: [],
  },

  {
    value: 'CUSTOM',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Custom',
      defaultMessage: '自定义',
    }),
    operations: [
      {
        value: 'modify',
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Modify',
          defaultMessage: '编辑',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.oceanbase.Delete',
          defaultMessage: '删除',
        }),
      },
    ],
  },
];

export const UNIT_MIGRATE_TYPE_LIST = [
  {
    value: 'MIGRATE_OUT',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MovingOut',
      defaultMessage: '迁出中',
    }),

    operationLabel: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Destination',
      defaultMessage: '目标端',
    }),

    extra: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.RightCancelMigration',
      defaultMessage: '可右键取消迁移',
    }),

    color: '#FA8C16',
    backgroundColor: 'rgba(255,192,105,0.65)',
    active: 'left2Right',
  },

  {
    value: 'MIGRATE_IN',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MovingIn',
      defaultMessage: '迁入中',
    }),

    migrateLabel: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Source',
      defaultMessage: '源端',
    }),

    extra: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.RightCancelMigration',
      defaultMessage: '可右键取消迁移',
    }),

    color: '#3C5EEC',
    backgroundColor: 'rgba(89,126,247,0.65)',
    active: 'left2Right',
  },

  {
    value: 'ROLLBACK_MIGRATE_OUT',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MovingOutAndRollingBack',
      defaultMessage: '迁出回滚中',
    }),

    migrateLabel: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.RollBackTheTarget',
      defaultMessage: '回滚目标端',
    }),

    color: '#FA8C16',
    backgroundColor: 'rgba(0,0,0,0.15)',
    active: 'right2Left',
  },

  {
    value: 'ROLLBACK_MIGRATE_IN',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.MigratingToAndRollingBack',
      defaultMessage: '迁入回滚中',
    }),

    migrateLabel: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.RollbackSource',
      defaultMessage: '回滚源端',
    }),

    color: '#3C5EEC',
    backgroundColor: 'rgba(0,0,0,0.15)',
    active: 'right2Left',
  },

  {
    value: 'NOT_IN_MIGRATE',
    label: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.Normal',
      defaultMessage: '正常',
    }),
    extra: formatMessage({
      id: 'ocp-express.src.constant.oceanbase.ClickToInitiateAnActive',
      defaultMessage: '点击可发起主动迁移，只能在同个 Zone 内迁移',
    }),

    color: 'rgba(0,0,0,0.45)',
    backgroundColor: '#FFFFFF',
  },
];
