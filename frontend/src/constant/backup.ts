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
import { range } from 'lodash';
import moment from 'moment';

export const AGENT_TYPE_LIST = [
  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Backup', defaultMessage: '备份' }),
    value: 'BACKUP',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Recovery',
      defaultMessage: '恢复',
    }),
    value: 'RESTORE',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.BackupAndRecovery',
      defaultMessage: '备份 + 恢复',
    }),

    value: 'BACKUP_RESTORE',
  },
];

export const AGENT_SERVICE_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Installing',
      defaultMessage: '安装中',
    }),
    value: 'INSTALLING',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'clone',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Copy',
          defaultMessage: '复制',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Running',
      defaultMessage: '运行中',
    }),
    value: 'RUNNING',
    badgeStatus: 'success',
    operations: [
      {
        value: 'addMachine',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.AddNode',
          defaultMessage: '添加节点',
        }),
      },

      {
        value: 'upgrade',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.UpgradedVersion',
          defaultMessage: '升级版本',
        }),
      },

      {
        value: 'edit',
        // 更新配置，即编辑服务的配置，因此 key 值设为 edit
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.UpdateConfigurations',
          defaultMessage: '更新配置',
        }),
      },

      {
        value: 'clone',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Copy',
          defaultMessage: '复制',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Delete',
          defaultMessage: '删除',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Upgrading',
      defaultMessage: '版本升级中',
    }),

    value: 'UPGRADING',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'clone',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Copy',
          defaultMessage: '复制',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.UpdatingConfiguration',
      defaultMessage: '配置更新中',
    }),

    value: 'UPDATING_CONFIG',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'clone',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Copy',
          defaultMessage: '复制',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Exception',
      defaultMessage: '异常',
    }),
    value: 'ABNORMAL',
    badgeStatus: 'error',
    operations: [
      {
        // 服务无节点也属于异常状态，允许添加节点
        value: 'addMachine',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.AddNode',
          defaultMessage: '添加节点',
        }),
      },

      {
        // 异常状态下，允许更新配置
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.UpdateConfigurations',
          defaultMessage: '更新配置',
        }),
      },

      {
        value: 'clone',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Copy',
          defaultMessage: '复制',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Delete',
          defaultMessage: '删除',
        }),
      },
    ],
  },
];

export const AGENT_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Online', defaultMessage: '在线' }),
    value: 'ONLINE',
    badgeStatus: 'success',
    operations: [
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Stop',
          defaultMessage: '停止',
        }),
      },

      {
        value: 'restart',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Restart',
          defaultMessage: '重启',
        }),
      },

      {
        // 由于查看运维任务属于任务模块的权限，在任务页面会进行权限控制的，因此这里的入口不需要设置 accessibleField
        value: 'viewTask',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewOMTasks',
          defaultMessage: '查看运维任务',
        }),
      },

      {
        value: 'uninstall',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Uninstall',
          defaultMessage: '卸载',
        }),
      },
    ],
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Offline', defaultMessage: '离线' }),
    value: 'OFFLINE',
    badgeStatus: 'default',
    operations: [
      {
        value: 'start',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Start',
          defaultMessage: '启动',
        }),
      },

      {
        value: 'viewTask',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewOMTasks',
          defaultMessage: '查看运维任务',
        }),
      },

      {
        value: 'uninstall',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Uninstall',
          defaultMessage: '卸载',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Deploying',
      defaultMessage: '部署中',
    }),
    value: 'DEPLOYING',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'viewTask',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewOMTasks',
          defaultMessage: '查看运维任务',
        }),
      },

      {
        value: 'uninstall',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Uninstall',
          defaultMessage: '卸载',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Upgrading',
      defaultMessage: '版本升级中',
    }),

    value: 'UPGRADING',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'viewTask',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewOMTasks',
          defaultMessage: '查看运维任务',
        }),
      },

      {
        value: 'uninstall',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Uninstall',
          defaultMessage: '卸载',
        }),
      },
    ],
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.UpdatingConfiguration',
      defaultMessage: '配置更新中',
    }),

    value: 'UPDATING_CONFIG',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'viewTask',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewOMTasks',
          defaultMessage: '查看运维任务',
        }),
      },

      {
        value: 'uninstall',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Uninstall',
          defaultMessage: '卸载',
        }),
      },
    ],
  },
];

export const WEEK_OPTIONS = [
  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.A', defaultMessage: '一' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Monday',
      defaultMessage: '星期一',
    }),
    value: 1,
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Two', defaultMessage: '二' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Tuesday',
      defaultMessage: '星期二',
    }),

    value: 2,
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Three', defaultMessage: '三' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Wednesday',
      defaultMessage: '星期三',
    }),

    value: 3,
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Four', defaultMessage: '四' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Thursday',
      defaultMessage: '星期四',
    }),

    value: 4,
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Five', defaultMessage: '五' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Friday',
      defaultMessage: '星期五',
    }),
    value: 5,
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Six', defaultMessage: '六' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Saturday',
      defaultMessage: '星期六',
    }),

    value: 6,
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Day', defaultMessage: '日' }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.Sunday',
      defaultMessage: '星期日',
    }),
    value: 7,
  },
];

export const MONTH_OPTIONS = range(1, 32).map(item => ({
  label: item < 10 ? `0${item}` : item,
  fullLabel: item < 10 ? `0${item}` : item,
  value: item,
}));

export const DATA_BACKUP_MODE_LIST = [
  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Total', defaultMessage: '全量' }),
    fullLabel: formatMessage({ id: 'OBShell.src.constant.backup.Total', defaultMessage: '全量' }),

    value: 'full',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Incremental',
      defaultMessage: '增量',
    }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.IncrementalBackup',
      defaultMessage: '增量备份',
    }),

    value: 'incremental',
  },
];

export const DATA_BACKUP_TYPE_LIST = [
  {
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Total', defaultMessage: '全量' }),
    fullLabel: formatMessage({ id: 'OBShell.src.constant.backup.Total', defaultMessage: '全量' }),

    value: 'FULL',
  },

  {
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Incremental',
      defaultMessage: '增量',
    }),
    fullLabel: formatMessage({
      id: 'ocp-express.src.constant.backup.IncrementalBackup',
      defaultMessage: '增量备份',
    }),

    value: 'INC',
  },
];

export const STORAGE_TYPE_LIST = [
  {
    value: 'BACKUP_STORAGE_FILE',
    label: 'File',
    protocol: 'file://',
  },
  {
    value: 'BACKUP_STORAGE_OSS',
    label: 'OSS',
    protocol: 'oss://',
  },

  {
    value: 'BACKUP_STORAGE_COS',
    label: 'COS',
    protocol: 'cos://',
  },
];

export const DATA_BACKUP_TASK_STATUS_LIST: Global.StatusItem[] = [
  {
    value: 'BEGINNING',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Starting',
      defaultMessage: '启动中',
    }),
    badgeStatus: 'processing',
  },

  {
    value: 'INIT',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Initializing',
      defaultMessage: '初始化中',
    }),

    badgeStatus: 'processing',
    operations: [
      {
        // 查看失败原因，是基于当前行数据、无需请求新的接口，因此不需要做权限控制
        value: 'cancel',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Cancel', defaultMessage: '取消' }),
      },
    ],
  },

  {
    value: 'DOING',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.InProgress',
      defaultMessage: '进行中',
    }),
    badgeStatus: 'processing',
    operations: [
      {
        // 查看失败原因，是基于当前行数据、无需请求新的接口，因此不需要做权限控制
        value: 'cancel',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Cancel', defaultMessage: '取消' }),
      },
    ],
  },

  {
    value: 'COMPLETED',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Completed',
      defaultMessage: '已完成',
    }),
    badgeStatus: 'success',
  },

  {
    value: 'CANCELING',
    label: formatMessage({ id: 'OBShell.src.constant.backup.Canceling', defaultMessage: '取消中' }),
    badgeStatus: 'processing',
    operations: [],
  },

  {
    value: 'CANCELED',
    label: formatMessage({ id: 'OBShell.src.constant.backup.Cancelled', defaultMessage: '已取消' }),
    badgeStatus: 'default',
    operations: [],
  },

  {
    value: 'FAILED',
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Failed', defaultMessage: '失败' }),
    badgeStatus: 'error',
    operations: [
      {
        // 查看失败原因，是基于当前行数据、无需请求新的接口，因此不需要做权限控制
        value: 'viewErrorMessage',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewReasons',
          defaultMessage: '查看原因',
        }),
      },
    ],
  },
  {
    value: 'BACKUP_SYS_META',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackUpSystemData',
      defaultMessage: '备份系统数据',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'BACKUP_USER_META',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackupUserMetadata',
      defaultMessage: '备份用户元数据',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'BACKUP_META_FINISH',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackupMetadataComplete',
      defaultMessage: '备份元数据完成',
    }),
    badgeStatus: 'success',
    operations: [],
  },
  {
    value: 'BACKUP_SYS_DATA',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackUpSystemData',
      defaultMessage: '备份系统数据',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'BACKUP_USER_DATA',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackUpUserData',
      defaultMessage: '备份用户数据',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'BEFORE_BACKUP_LOG',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BeforeBackingUpTheLog',
      defaultMessage: '备份日志前',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'BEFORE_BACKUP_LOG',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackupLog',
      defaultMessage: '备份日志',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'BACKUP_FUSE_TABLET_MATE',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.BackupFusionTableMetadata',
      defaultMessage: '备份融合表元数据',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
  {
    value: 'PREPARE_BACKUP_LOG',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.PrepareToBackupLogs',
      defaultMessage: '准备备份日志',
    }),
    badgeStatus: 'processing',
    operations: [],
  },
];

export const LOG_BACKUP_TASK_STATUS_LIST: Global.StatusItem[] = [
  {
    value: 'BEGINNING',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Starting',
      defaultMessage: '启动中',
    }),
    badgeStatus: 'warning',
  },

  {
    value: 'DOING',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.InProgress',
      defaultMessage: '进行中',
    }),
    badgeStatus: 'processing',
    operations: [
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Stop',
          defaultMessage: '停止',
        }),
      },
      {
        value: 'pause',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Pause', defaultMessage: '暂停' }),
      },
    ],
  },

  {
    // 物理备份有效
    value: 'STOPPING',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Stopping',
      defaultMessage: '停止中',
    }),
    badgeStatus: 'warning',
  },

  {
    value: 'STOPPED',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Stopped',
      defaultMessage: '已停止',
    }),
    badgeStatus: 'default',
    operations: [
      {
        value: 'start',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Start',
          defaultMessage: '启动',
        }),
      },
    ],
  },

  {
    // 逻辑备份有效
    value: 'FAILED',
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Failed', defaultMessage: '失败' }),
    badgeStatus: 'error',
    operations: [
      {
        value: 'viewErrorMessage',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewReasons',
          defaultMessage: '查看原因',
        }),
      },
    ],
  },

  {
    value: 'SUSPEND',
    label: formatMessage({ id: 'OBShell.src.constant.backup.Pause', defaultMessage: '暂停' }),

    badgeStatus: 'error',
    operations: [
      {
        value: 'start',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Start',
          defaultMessage: '启动',
        }),
      },
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Stop',
          defaultMessage: '停止',
        }),
      },
    ],
  },
];

export const RESTORE_TASK_STATUS_LIST: Global.StatusItem[] = [
  {
    value: 'RESTORING',
    label: formatMessage({ id: 'OBShell.src.constant.backup.Restoring', defaultMessage: '恢复中' }),

    badgeStatus: 'processing',
    operations: [
      {
        value: 'cancel',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Cancel', defaultMessage: '取消' }),
      },
    ],
  },
  {
    value: 'CREATE_TENANT',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.CreateTenant',
      defaultMessage: '创建租户',
    }),

    badgeStatus: 'processing',
    operations: [
      {
        value: 'cancel',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Cancel', defaultMessage: '取消' }),
      },
    ],
  },

  {
    value: 'UPGRADE',
    label: formatMessage({ id: 'OBShell.src.constant.backup.Upgrade', defaultMessage: '升级' }),

    badgeStatus: 'processing',
    operations: [
      {
        value: 'cancel',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Cancel', defaultMessage: '取消' }),
      },
    ],
  },

  {
    value: 'RESTORINE_SUCCESS',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.RecoverySucceeded',
      defaultMessage: '恢复成功',
    }),

    badgeStatus: 'success',
  },

  {
    value: 'RESTORINE_FAIL',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.RecoveryFailed',
      defaultMessage: '恢复失败',
    }),

    badgeStatus: 'success',
    operations: [
      {
        value: 'viewErrorMessage',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewReasons',
          defaultMessage: '查看原因',
        }),
      },
    ],
  },

  {
    value: 'WAIT_TENANT_RESTORE_FINISH',
    label: formatMessage({
      id: 'OBShell.src.constant.backup.WaitForTenantRecoveryTo',
      defaultMessage: '等待租户恢复完成',
    }),
    badgeStatus: 'processing',
    operations: [
      {
        value: 'cancel',
        label: formatMessage({ id: 'OBShell.src.constant.backup.Cancel', defaultMessage: '取消' }),
      },
    ],
  },

  {
    value: 'SUCCESS',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Completed',
      defaultMessage: '已完成',
    }),
    badgeStatus: 'success',
  },

  {
    value: 'FAIL',
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Failed', defaultMessage: '失败' }),
    badgeStatus: 'error',
    operations: [
      {
        value: 'viewErrorMessage',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.ViewReasons',
          defaultMessage: '查看原因',
        }),
      },
    ],
  },
];

export const SAMPLING_STRATEGY_STATUS_LIST: Global.StatusItem[] = [
  {
    value: true,
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Running',
      defaultMessage: '运行中',
    }),
    badgeStatus: 'processing',
    operations: [
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Stop',
          defaultMessage: '停止',
        }),
      },

      {
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Editing',
          defaultMessage: '编辑',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Delete',
          defaultMessage: '删除',
        }),
      },
    ],
  },

  {
    value: false,
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.Stopped',
      defaultMessage: '已停止',
    }),
    badgeStatus: 'default',
    operations: [
      {
        value: 'start',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Start',
          defaultMessage: '启动',
        }),
      },

      {
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Editing',
          defaultMessage: '编辑',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.Delete',
          defaultMessage: '删除',
        }),
      },
    ],
  },
];

export const BACKUP_SCHEDULE_MODE_LIST = [
  {
    value: 'WEEK',
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Weeks', defaultMessage: '周' }),
  },

  {
    value: 'MONTH',
    label: formatMessage({ id: 'ocp-express.src.constant.backup.Months', defaultMessage: '月' }),
  },
];

export const BACKUP_MODE_LIST = [
  {
    label: formatMessage({
      id: 'ocp-v2.src.constant.backup.PhysicalBackup',
      defaultMessage: '物理备份',
    }),

    value: 'PHYSICAL_BACKUP',
    labelForSampling: formatMessage({
      id: 'ocp-v2.src.constant.backup.PhysicalBackupCluster',
      defaultMessage: '物理备份集群',
    }),

    tagColor: 'green',
  },

  {
    label: formatMessage({
      id: 'ocp-v2.src.constant.backup.LogicalBackup',
      defaultMessage: '逻辑备份',
    }),

    value: 'LOGICAL_BACKUP',
    labelForSampling: formatMessage({
      id: 'ocp-v2.src.constant.backup.LogicalBackupCluster',
      defaultMessage: '逻辑备份集群',
    }),

    tagColor: 'cyan',
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.backup.Total', defaultMessage: '全量' }),

    value: 'NONE',
    labelForSampling: formatMessage({
      id: 'OBShell.src.constant.backup.Total',
      defaultMessage: '全量',
    }),

    tagColor: 'cyan',
  },
];

export const BACKUP_SCHEDULE_STATUS_LIST: Global.StatusItem[] = [
  // 无备份调度意味着无备份策略
  {
    value: 'NONE',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.NoBackupScheduling',
      defaultMessage: '无备份策略',
    }),

    badgeStatus: 'default',
    operations: [
      {
        value: 'new',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.CreateABackupPolicy',
          defaultMessage: '新建备份策略',
        }),
      },

      {
        value: 'backupNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.BackUpNow',
          defaultMessage: '立即备份',
        }),
      },

      {
        value: 'restoreNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.InitiateRecovery',
          defaultMessage: '发起恢复',
        }),
      },
    ],
  },

  {
    value: 'NORMAL',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.BackupScheduling',
      defaultMessage: '备份调度中',
    }),

    badgeStatus: 'processing',
    operations: [
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.PauseBackupScheduling',
          defaultMessage: '暂停备份调度',
        }),
      },

      {
        value: 'backupNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.BackUpNow',
          defaultMessage: '立即备份',
        }),
      },

      {
        value: 'restoreNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.InitiateRecovery',
          defaultMessage: '发起恢复',
        }),
      },

      {
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.EditABackupPolicy',
          defaultMessage: '编辑备份策略',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.DeleteABackupPolicy',
          defaultMessage: '删除备份策略',
        }),
      },
    ],
  },

  {
    value: 'PAUSED',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.BackupSchedulingSuspended',
      defaultMessage: '备份调度已暂停',
    }),

    badgeStatus: 'warning',
    operations: [
      {
        value: 'start',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.StartBackupScheduling',
          defaultMessage: '启动备份调度',
        }),
      },

      {
        value: 'backupNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.BackUpNow',
          defaultMessage: '立即备份',
        }),
      },

      {
        value: 'restoreNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.InitiateRecovery',
          defaultMessage: '发起恢复',
        }),
      },

      {
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.EditABackupPolicy',
          defaultMessage: '编辑备份策略',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.DeleteABackupPolicy',
          defaultMessage: '删除备份策略',
        }),
      },
    ],
  },

  {
    value: 'ABNORMAL',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.BackupScheduleException',
      defaultMessage: '备份调度异常',
    }),

    badgeStatus: 'error',
    operations: [
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.PauseBackupScheduling',
          defaultMessage: '暂停备份调度',
        }),
      },

      {
        value: 'backupNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.BackUpNow',
          defaultMessage: '立即备份',
        }),
      },

      {
        value: 'restoreNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.InitiateRecovery',
          defaultMessage: '发起恢复',
        }),
      },

      {
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.EditABackupPolicy',
          defaultMessage: '编辑备份策略',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.DeleteABackupPolicy',
          defaultMessage: '删除备份策略',
        }),
      },
    ],
  },

  {
    value: 'WAITING',
    label: formatMessage({
      id: 'ocp-express.src.constant.backup.WaitingForBackupScheduling',
      defaultMessage: '等待备份调度',
    }),

    badgeStatus: 'warning',
    operations: [
      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.PauseBackupScheduling',
          defaultMessage: '暂停备份调度',
        }),
      },

      {
        value: 'backupNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.BackUpNow',
          defaultMessage: '立即备份',
        }),
      },

      {
        value: 'restoreNow',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.InitiateRecovery',
          defaultMessage: '发起恢复',
        }),
      },

      {
        value: 'edit',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.EditABackupPolicy',
          defaultMessage: '编辑备份策略',
        }),
      },

      {
        value: 'delete',
        label: formatMessage({
          id: 'ocp-express.src.constant.backup.DeleteABackupPolicy',
          defaultMessage: '删除备份策略',
        }),
      },
    ],
  },
];

// 获取适用于 ob-ui Ranger 组件的 selects，该格式与 antd 的不同
export function getSelects() {
  return [
    {
      name: formatMessage({
        id: 'ocp-express.src.constant.backup.NearlyHours',
        defaultMessage: '近 24 小时',
      }),

      range: () => [moment().subtract(24, 'hours'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.backup.LastDays',
        defaultMessage: '近 7 天',
      }),

      range: () => [moment().subtract(7, 'days'), moment()],
    },

    {
      name: formatMessage({
        id: 'ocp-express.src.constant.backup.NearlyDays',
        defaultMessage: '近 30 天',
      }),

      range: () => [moment().subtract(30, 'days'), moment()],
    },
  ];
}
