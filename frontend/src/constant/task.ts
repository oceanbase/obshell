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

import { token } from '@oceanbase/design';
import { formatMessage } from '@/util/intl';

// 任务状态列表
export const TASK_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.Completed',
      defaultMessage: '已完成',
    }),
    value: 'SUCCEED',
    badgeStatus: 'success',
    operations: [
      // {
      //   value: 'downloadLog',
      //   label: formatMessage({
      //     id: 'ocp-express.src.constant.task.DownloadLogs',
      //     defaultMessage: '下载日志',
      //   }),
      // },
    ],
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.task.Running', defaultMessage: '运行中' }),
    value: 'RUNNING',
    badgeStatus: 'processing',
    operations: [
      {
        value: 'cancel',
        label: formatMessage({
          id: 'ocp-v2.src.constant.task.CancelTask',
          defaultMessage: '取消任务',
        }),
      },
    ],
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.task.Failed', defaultMessage: '失败' }),
    value: 'FAILED',
    badgeStatus: 'error',
    operations: [
      {
        value: 'retry',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.TryAgain',
          defaultMessage: '重试',
        }),
      },
      {
        value: 'pass',
        label: formatMessage({
          id: 'ocp-v2.src.constant.task.SkipTasks',
          defaultMessage: '跳过任务',
        }),
      },
      {
        value: 'giveup',
        label: formatMessage({
          id: 'ocp-v2.src.constant.task.RollbackTask',
          defaultMessage: '回滚任务',
        }),
      },

      // {
      //   value: 'downloadLog',
      //   label: formatMessage({
      //     id: 'ocp-express.src.constant.task.DownloadLogs',
      //     defaultMessage: '下载日志',
      //   }),
      // },
    ],
  },
];

export const TASK_MAINTENANCE_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.maintenance.NotMaintenance',
      defaultMessage: '非运维',
    }),
    value: '1',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.maintenance.GlobalMaintenance',
      defaultMessage: '全局运维',
    }),
    value: '2',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.maintenance.TenantMaintenance',
      defaultMessage: '租户运维',
    }),
    value: '3',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.maintenance.OBProxyMaintenance',
      defaultMessage: 'OBProxy 运维',
    }),
    value: '4',
  },
];

export const TASK_TYPE_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.obcluster',
      defaultMessage: '集群任务',
    }),
    value: '1',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.local.agent',
      defaultMessage: '本地任务',
    }),
    value: '2',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.local.agent',
      defaultMessage: '本地任务',
    }),
    value: '3',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.local.obproxy',
      defaultMessage: 'obproxy 任务',
    }),
    value: '4',
  },
  {
    label: formatMessage({
      id: 'ocp-express.src.constant.task.local.obproxy',
      defaultMessage: 'obproxy 任务',
    }),
    value: '5',
  },
];

// 子任务状态列表
export const SUBTASK_STATUS_LIST: Global.StatusItem[] = [
  {
    label: formatMessage({ id: 'ocp-express.src.constant.task.Complete', defaultMessage: '完成' }),
    // 在流程图中的文本描述
    labelInGraph: formatMessage({
      id: 'ocp-express.src.constant.task.Completed',
      defaultMessage: '已完成',
    }),

    value: 'SUCCESSFUL',
    badgeStatus: 'success',
    color: token.colorSuccess,
    operations: [
      {
        value: 'viewLog',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.ViewLogs',
          defaultMessage: '查看日志',
        }),
      },
    ],
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.task.Running', defaultMessage: '运行中' }),
    labelInGraph: formatMessage({
      id: 'ocp-express.src.constant.task.TaskExecution',
      defaultMessage: '任务执行中',
    }),

    value: 'RUNNING',
    badgeStatus: 'processing',
    color: token.colorPrimary,
    operations: [
      {
        value: 'viewLog',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.ViewLogs',
          defaultMessage: '查看日志',
        }),
      },

      {
        value: 'stop',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.StopRunning',
          defaultMessage: '终止运行',
        }),
      },
    ],
  },

  {
    label: formatMessage({ id: 'ocp-express.src.constant.task.Failed', defaultMessage: '失败' }),
    labelInGraph: formatMessage({
      id: 'ocp-express.src.constant.task.TaskFailed',
      defaultMessage: '任务失败',
    }),

    value: 'FAILED',
    badgeStatus: 'error',
    color: token.colorError,
    operations: [
      {
        value: 'viewLog',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.ViewLogs',
          defaultMessage: '查看日志',
        }),
      },

      {
        value: 'retry',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.ReRun',
          defaultMessage: '重新运行',
        }),
      },

      // {
      //   value: 'skip',
      //   label: formatMessage({
      //     id: 'ocp-express.src.constant.task.SetToSuccessful',
      //     defaultMessage: '设置为成功',
      //   }),
      // },
    ],
  },

  {
    value: 'PENDING',
    label: formatMessage({
      id: 'ocp-express.src.constant.task.ToBeExecuted',
      defaultMessage: '待执行',
    }),
    labelInGraph: formatMessage({
      id: 'ocp-express.src.constant.task.ExecutionNotStarted',
      defaultMessage: '未开始执行',
    }),

    badgeStatus: 'default',
    color: '#cccccc',
    operations: [
      {
        value: 'viewLog',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.ViewLogs',
          defaultMessage: '查看日志',
        }),
      },
    ],
  },

  {
    value: 'READY',
    label: formatMessage({
      id: 'ocp-express.src.constant.task.PrepareForExecution',
      defaultMessage: '准备执行',
    }),
    labelInGraph: formatMessage({
      id: 'ocp-express.src.constant.task.PrepareForExecution',
      defaultMessage: '准备执行',
    }),
    badgeStatus: 'default',
    color: '#cccccc',
    operations: [
      {
        value: 'viewLog',
        label: formatMessage({
          id: 'ocp-express.src.constant.task.ViewLogs',
          defaultMessage: '查看日志',
        }),
      },
    ],
  },
];
