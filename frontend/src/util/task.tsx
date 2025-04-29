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
import React from 'react';
import { Modal, message } from '@oceanbase/design';
import { find, flatten, groupBy, isArray, uniq, uniqueId } from 'lodash';
import moment from 'moment';
import { directTo, toPercent, joinComponent, formatNumber, formatTime } from '@oceanbase/util';
import * as TaskController from '@/service/ocp-express/TaskController';
import { getFormateForTimes, secondToTime } from '@/util';
import { dagHandler } from '@/service/obshell/task';

export function getTaskProgress(task: API.DagDetailDTO) {
  const subtasks = flatten(task?.nodes?.map(item => item.sub_tasks || []) || []);
  const total = subtasks.length;
  const finishedCount = subtasks.filter(item => item && item.state === 'SUCCEED').length;
  return `${finishedCount}/${total}`;
}

export function getTaskPercent(task: API.DagDetailDTO, withUnit = true) {
  const subtasks = flatten(task?.nodes?.map(item => item.sub_tasks || []) || []);
  const total = subtasks.length;
  const finishedCount = subtasks.filter(item => item && item.state === 'SUCCEED').length;
  // 是否带 % 单位
  return withUnit
    ? toPercent(finishedCount / total, 0)
    : // 四舍五入，不保留小数位
      Math.round((finishedCount / total) * 10000) / 100;
}

/* 获取任务或子任务的执行耗时 = 最近一次执行的结束时间 - 第一次执行的开始时间。时间单位会通过 secondToTime 自适应转换 */
export function getTaskDuration(task?: API.DagDetailDTO) {
  // 第一次执行的开始时间
  const startTime = task?.start_time;
  // 最近一次执行的结束时间，包括成功和失败，会随着任务或子任务状态的变化而变化
  const finishTime = task?.end_time;
  // 只有执行过，执行耗时才有意义
  if (startTime) {
    // 当任务正在执行中时，直接取当前时间，这样可以实时变化
    const milliseconds = moment(task.state === 'RUNNING' ? undefined : finishTime).diff(
      moment(startTime),
      // 由于部分子任务的执行速度非常快，因此这里获取毫秒数
      'millisecond'
    );

    return milliseconds > 0
      ? secondToTime(
          formatNumber(
            milliseconds / 1000,
            // 秒数 >= 1 仅保留整数
            // 秒数 < 1 的最多保留两位小数
            milliseconds >= 1000 ? 0 : 2
          )
        )
      : '-';
  }
  return '-';
}

export function taskSuccess({
  taskId,
  message: taskMessage = formatMessage({
    id: 'ocp-express.src.util.task.TaskSubmittedSuccessfully',
    defaultMessage: '任务提交成功',
  }),
}: {
  taskId: number | undefined | (number | undefined)[];
  message?: string;
}) {
  const taskIdList = isArray(taskId) ? taskId : [taskId];
  // 是否为多个任务
  const isMultipleTask = taskIdList.length > 1;
  Modal.success({
    closable: true,
    title: taskMessage,
    content: (
      <span>
        {isMultipleTask ? (
          <span>
            {formatMessage({
              id: 'ocp-express.src.util.task.ATaskHasBeenGenerated',
              defaultMessage: '已生成任务，任务 ID：',
            })}{' '}
            {joinComponent(taskIdList, item => (
              <a
                key={item}
                onClick={() => {
                  directTo(`/task/${item}?backUrl=/task`);
                }}
              >
                {item}
              </a>
            ))}
          </span>
        ) : (
          formatMessage(
            {
              id: 'ocp-express.src.util.task.ATaskHasBeenGenerated.1',
              defaultMessage: '已生成任务，任务 ID：{taskId}',
            },
            { taskId }
          )
        )}
      </span>
    ),

    okText: isMultipleTask
      ? formatMessage({ id: 'ocp-express.src.util.task.ISee', defaultMessage: '我知道了' })
      : formatMessage({
          id: 'ocp-express.src.util.component.ViewTasks',
          defaultMessage: '查看任务',
        }),
    okButtonProps: {
      // 关闭 loading，避免 onOk 返回 promise 时出现短暂的 loading 效果 (实际不应该 loading)
      loading: false,
    },

    onOk: () => {
      if (!isMultipleTask) {
        directTo(`/task/${taskId}?backUrl=/task`);
        // 通过模拟 promsie reject 实现击查看任务后，不关闭弹窗
        return new Promise((resolve, reject) => {
          reject();
        });
      }
      return null;
    },
  });
}

export function taskSuccessOfInspection({
  taskId,
  message: taskMessage = formatMessage({
    id: 'ocp-express.src.util.task.TaskSubmittedSuccessfully',
    defaultMessage: '任务提交成功',
  }),
  okText = formatMessage({
    id: 'ocp-express.src.util.component.ViewTasks',
    defaultMessage: '查看任务',
  }),
  onOk = () => {},
  onCancel = () => {},
}: {
  taskId: number | undefined | (number | undefined)[];
  onOk: () => void;
  onCancel?: () => void;
  message?: string;
  okText?: string;
}) {
  const taskIdList = isArray(taskId) ? taskId : [taskId];
  Modal.success({
    closable: true,
    title: taskMessage,
    content: (
      <span>
        <span>
          {formatMessage({
            id: 'ocp-express.src.util.task.ATaskHasBeenGenerated',
            defaultMessage: '已生成任务，任务 ID：',
          })}{' '}
          {joinComponent(taskIdList, item => (
            <a
              key={item}
              onClick={() => {
                directTo(`/task/${item}?backUrl=/task`);
              }}
            >
              {item}
            </a>
          ))}
        </span>
      </span>
    ),

    okText,
    okButtonProps: {
      // 关闭 loading，避免 onOk 返回 promise 时出现短暂的 loading 效果 (实际不应该 loading)
      loading: false,
    },

    onOk,
    onCancel,
  });
}

export type SubtaskOperationKey = 'stop' | 'retry' | 'skip';

/* 处理子任务的操作: 终止运行、重新运行、设置为成功 (跳过) */
export function handleSubtaskOperate(
  key: SubtaskOperationKey,
  taskData?: API.TaskInstance,
  subtask?: API.SubtaskInstance,
  onSuccess?: () => void
) {
  if (key === 'stop') {
    Modal.confirm({
      title: formatMessage({
        id: 'ocp-express.Task.Detail.TaskGraph.AreYouSureYouWant',
        defaultMessage: '确定要终止运行吗？',
      }),

      content: formatMessage({
        id: 'ocp-v2.src.util.task.ThisWillTerminateTheEntire',
        defaultMessage: '这将会终止整个任务',
      }),

      onOk: () => {
        return dagHandler(
          {
            id: taskData?.id,
          },
          {
            operator: 'CANCEL',
          }
        ).then(res => {
          if (res.successful) {
            message.success(
              formatMessage({
                id: 'ocp-v2.src.util.task.TaskTerminatedSuccessfully',
                defaultMessage: '任务终止成功',
              })
            );
            onSuccess();
          }
        });
      },
    });
  } else if (key === 'retry') {
    Modal.confirm({
      title: formatMessage({
        id: 'ocp-express.Task.Detail.TaskGraph.AreYouSureYouWant.1',
        defaultMessage: '确定要重新运行吗？',
      }),

      content: formatMessage({
        id: 'ocp-v2.src.util.task.ThisWillReRunThe',
        defaultMessage: '这将会重新执行整个任务',
      }),

      onOk: () => {
        return dagHandler(
          {
            id: taskData?.id,
          },
          {
            operator: 'RETRY',
          }
        ).then(res => {
          if (res.successful) {
            message.success(
              formatMessage({
                id: 'ocp-v2.src.util.task.TaskRetrySucceeded',
                defaultMessage: '任务重试成功',
              })
            );
            onSuccess();
          }
        });
      },
    });
  } else if (key === 'skip') {
    Modal.confirm({
      title: formatMessage({
        id: 'ocp-express.Task.Detail.TaskGraph.AreYouSureTheSettings',
        defaultMessage: '确定设置为成功吗？',
      }),

      content: formatMessage({
        id: 'ocp-express.Task.Detail.TaskGraph.ThisSkipsTheCurrentTask',
        defaultMessage: '这将会跳过当前任务，并把任务节点的状态设为成功',
      }),

      onOk: () => {
        const promise = TaskController.skipSubtask({
          taskInstanceId: taskData?.id,
          subtaskInstanceId: subtask?.id,
        });

        promise.then(res => {
          if (res.successful) {
            message.success(
              formatMessage({
                id: 'ocp-express.src.util.task.TheSubtaskStatusIsSet',
                defaultMessage: '子任务状态已设置为成功',
              })
            );
            if (onSuccess) {
              onSuccess();
            }
          }
        });
        return promise;
      },
    });
  }
}

/**
 * 按照分隔符对任务日志进行分割
 * @param log   日志字符串
 * @param regex 匹配分隔符的正则表达式
 * */
export function splitTaskLog(log?: string) {
  const regex = /([#]{12}{[\S]*}{[\S]*}[#]{12})/;
  // 日志是顺序写入的，但日志节点是倒序展示，因此需要 reverse
  const splitList = (log || '').split(regex).reverse();
  // 日志节点列表
  const logNodeList = splitList
    .filter(item => item.match(regex))
    .map(item => {
      const innerRegex = /[#]{12}{([\S]*)}{([\S]*)}[#]{12}/;
      const match = innerRegex.exec(item);
      const type = (match && match[1]) || '';
      const timestamp = (match && match[2]) || '';
      return `${type} ${moment(timestamp).format('HH:mm:ss')}`;
    });
  // 日志列表，与日志节点的顺序关系一致
  const logList = splitList
    .filter(item => item && !item.match(regex))
    // 去掉日志的首尾空行
    .map(item => item.replace(/^[\n]+|[\n]+$/g, ''));
  return {
    logNodeList,
    logList,
  };
}

export type Node =
  | (API.TaskDetailDTO & {
      children?: Node[];
    })
  | undefined;

/**
 * 根据任务获取子任务节点列表，其中分叉节点作为上一节点的 children
 * */
export function getNodes(taskData?: API.DagDetailDTO) {
  const nodes = flatten(taskData?.nodes || []);

  return nodes.map(item => {
    // task 和 subtask 的 id 可能重复，增加 domId 进行区分
    item.domId = item.id;
    // 只有一个 subTasks 时，直接替换 Node 进行返回
    if (item.sub_tasks?.length === 1) {
      item.sub_tasks[0].domId = item.sub_tasks[0].id + '_child';
      return item.sub_tasks[0];
    }
    item.children = item.sub_tasks.map((item, index) => ({ ...item, domId: `${item.id}_child` }));
    return item;
  });
}

/**
 * 获取节点的当前进度，即获取已失败或者正在运行的第一个节点
 * */
export function getLatestNode(nodes: Node[]) {
  const flattenNodes = flatten(nodes.map(item => [item, ...(item?.children || [])]));
  // 最新的失败节点
  const latestFailedNode = find(flattenNodes, item => item?.state === 'FAILED');
  // 最新的运行中节点
  const latestRunningNode = find(flattenNodes, item => item?.state === 'RUNNING');
  // 最新的准备执行节点
  const latestReadyNode = find(flattenNodes, item => item?.state === 'READY');
  // 最新的待执行节点
  const latestPendingNode = find(flattenNodes, item => item?.state === 'PENDING');
  // 最新的已完成节点, 需要反转下数组，查询最近完成的节点
  const latestSuccessfulNode = find(flattenNodes.reverse(), item => item?.state === 'SUCCEED');
  // 第一个节点
  const firstNode = flattenNodes && flattenNodes[0];
  // 最新节点的优先级: 失败节点 > 运行中节点 > 准备执行节点 > 待执行节点 > 第一个节点
  const latestNode =
    latestFailedNode ||
    latestRunningNode ||
    latestReadyNode ||
    latestPendingNode ||
    latestSuccessfulNode ||
    firstNode;

  // 存在 children，使用 children
  return (latestNode?.children?.[0] || latestNode) as Node | undefined;
}

export function getTaskLog(log: any[]) {
  // 日志是顺序写入的，但日志节点是倒序展示，因此需要 reverse
  // const reverseLog = [...log].reverse();
  // TODO：大小写需要修改
  const reverseLog = [...log];
  const failedLogList = reverseLog?.map(item => item.subtaskFailedRecord);
  const logList = reverseLog?.map(item => `${formatTime(item.create_time)}\n ${item.log_content}`);
  const formatter = getFormateForTimes(reverseLog?.map(item => item.create_time));
  const logGroupList = groupBy(reverseLog, item => formatTime(item.create_time, formatter));
  const logNodeList = uniq(
    reverseLog?.map(item => `${moment(item.create_time).format(formatter)}`)
  );
  return {
    logNodeList,
    logList,
    failedLogList,
  };
}
