import React, { useState, useEffect, useRef } from 'react';
import { formatMessage } from '@/util/intl';
import { history } from 'umi';
import { Button, Descriptions, Space, Typography, Modal, message } from '@oceanbase/design';
import type { Route } from '@oceanbase/design/es/breadcrumb/Breadcrumb';
import { find, flatten, isFunction } from 'lodash';
import { PageContainer } from '@oceanbase/ui';
import { isNullValue, findByValue } from '@oceanbase/util';
import { useRequest, useInterval, useLockFn } from 'ahooks';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { breadcrumbItemRender, getOperationComponent } from '@/util/component';
import { TASK_STATUS_LIST } from '@/constant/task';
import { delayInterfaceWithSentItTwice, isEnglish } from '@/util';
import { formatTime } from '@/util/datetime';
import { getNodes, getTaskDuration, getTaskProgress } from '@/util/task';
import ContentWithReload from '@/component/ContentWithReload';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import Log from './Log';
import type { TaskGraphRef } from './Log/TaskGraph';
import styles from './index.less';
import { dagHandler, getDagDetail, getFullSubTaskLogs } from '@/service/obshell/task';

const { Text } = Typography;
export interface DetailProps {
  match: {
    params: {
      taskId: number;
    };
  };

  location: {
    pathname: string;
    query: {
      backUrl?: string;
    };
  };
}

const Detail: React.FC<DetailProps> = ({
  match: {
    params: { taskId },
  },

  location: {
    query: { backUrl },
  },
}) => {
  const [domId, setDomId] = useState<number | string | undefined>(undefined);
  const logRef = useRef<TaskGraphRef>(null);

  const [taskData, setTaskData] = useState({});
  // 获取任务详情
  const { refreshAsync, loading } = useRequest(
    () =>
      getDagDetail(
        {
          id: taskId,
        },
        {
          HIDE_ERROR_MESSAGE: true,
        }
      ),
    {
      onSuccess: res => {
        if (res?.successful) {
          const realTaskData = res?.data || { nodes: [] };
          console.log(realTaskData, 'realTaskData');
          // 设置 nodes，增加 domId 进行判断
          realTaskData.nodes = getNodes(realTaskData);

          setTaskData(realTaskData);
        } else {
          // 错误之前的最后一次请求，不存在具体的 nodes，那么直接跳转到首页. 存在数据的情况下载此页面继续轮训
          if (!taskData?.nodes?.length > 0) {
            history.push('/overview');
          }
        }
      },
    }
  );

  const lockRefresh = useLockFn(refreshAsync);

  useDocumentTitle(taskData?.name);

  const stateItem = findByValue(TASK_STATUS_LIST, taskData.state);
  // 任务是否处于轮询状态
  const polling = taskData?.state === 'RUNNING';
  // children 上增加了 domId， 后端的 parentId 和 childId 存在重复，画图时异常
  const subtasks = flatten(taskData?.nodes?.map(item => item.children || []) || []);

  // 当前选中的子任务
  const subtask = find(subtasks || [], item => !isNullValue(domId) && item.domId === domId);

  // 子任务处在运行中的状态才发起日志轮询，待执行的任务不需要发起轮询，因为待执行的任务无法查看日志
  const logPolling = subtask?.state === 'RUNNING';

  // 获取子任务日志
  const {
    data: logData,
    run: getSubtaskLog,
    loading: logLoading,
  } = useRequest(getFullSubTaskLogs, {
    manual: true,
  });

  const log = logData?.data?.contents || [];

  // 由于 ready 仅在第一次生效，当 ready 和 refreshDeps 包含同一变量时，ready 的拦截逻辑不符合预期
  // 属于 ahooks 的一个问题，因此先用 useEffect 实现
  useEffect(() => {
    if (!isNullValue(subtask?.id)) {
      getSubtaskLog({
        id: subtask?.id,
      });
    }
  }, [
    subtask?.id,
    // 子任务状态改变时，需要重新请求日志，避免任务成功或失败时，由于不处于轮询态导致无法拉取最新日志的问题
    subtask?.state,
  ]);

  // 任务轮询
  useInterval(
    () => {
      lockRefresh();
    },
    polling ? 8000 : undefined
  );

  const { runAsync: dagHandlerFn } = useRequest(dagHandler, {
    manual: true,
  });

  const handleOperation = (key: string) => {
    if (key === 'retry') {
      // 重试任务
      Modal.confirm({
        title: formatMessage({
          id: 'ocp-express.Task.Detail.AreYouSureYouWant.1',
          defaultMessage: '确定要重试当前任务吗？',
        }),

        content: formatMessage({
          id: 'ocp-express.Task.Detail.ThisWillRollBackAll.1',
          defaultMessage: '重试操作会重试所有失败节点，并从失败节点开始继续向下执行',
        }),

        onOk: () => {
          return dagHandlerFn(
            {
              id: taskId,
            },
            {
              operator: 'RETRY',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'ocp-v2.Task.Detail.TaskRetrySucceeded',
                  defaultMessage: '任务重试成功',
                })
              );

              delayInterfaceWithSentItTwice(refreshAsync, 800);
            }
          });
        },
      });
    } else if (key === 'giveup') {
      // 放弃任务 = 回滚任务
      Modal.confirm({
        title: formatMessage({
          id: 'ocp-express.Task.Detail.AreYouSureYouWant',
          defaultMessage: '确定要回滚当前任务吗？',
        }),

        content: formatMessage({
          id: 'ocp-express.Task.Detail.ThisWillRollBackAll',
          defaultMessage: '这将从失败处开始回滚所有已执行过的任务',
        }),

        okText: formatMessage({ id: 'ocp-v2.Task.Detail.Rollback', defaultMessage: '回滚' }),
        okButtonProps: {
          danger: true,
          ghost: true,
        },

        onOk: () => {
          return dagHandlerFn(
            {
              id: taskId,
            },
            {
              operator: 'ROLLBACK',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'ocp-v2.Task.Detail.TaskRollbackSuccessful',
                  defaultMessage: '任务回滚成功',
                })
              );

              delayInterfaceWithSentItTwice(refreshAsync, 800);
            }
          });
        },
      });
    } else if (key === 'cancel') {
      // 放弃任务 = 回滚任务
      Modal.confirm({
        title: formatMessage({
          id: 'ocp-v2.Task.Detail.AreYouSureYouWant',
          defaultMessage: '确定要取消当前任务吗？',
        }),

        content: formatMessage({
          id: 'ocp-v2.Task.Detail.ThisWillCancelTheEntire',
          defaultMessage: '这将取消整个任务',
        }),

        okText: formatMessage({ id: 'ocp-v2.Task.Detail.Determine', defaultMessage: '确定' }),
        okButtonProps: {
          danger: true,
          ghost: true,
        },

        onOk: () => {
          return dagHandlerFn(
            {
              id: taskId,
            },
            {
              operator: 'CANCEL',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'ocp-v2.Task.Detail.TaskCanceledSuccessfully',
                  defaultMessage: '任务取消成功',
                })
              );

              delayInterfaceWithSentItTwice(refreshAsync, 800);
            }
          });
        },
      });
    } else if (key === 'pass') {
      // 放弃任务 = 回滚任务
      Modal.confirm({
        title: formatMessage({
          id: 'ocp-v2.Task.Detail.AreYouSureYouWant.1',
          defaultMessage: '确定要跳过当前任务吗？',
        }),

        content: formatMessage({
          id: 'ocp-v2.Task.Detail.ThisWillSkipTheEntire',
          defaultMessage: '这将跳过整个任务',
        }),

        okText: formatMessage({ id: 'ocp-v2.Task.Detail.Skip', defaultMessage: '跳过' }),
        okButtonProps: {
          danger: true,
          ghost: true,
        },

        onOk: () => {
          return dagHandlerFn(
            {
              id: taskId,
            },
            {
              operator: 'PASS',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'ocp-v2.Task.Detail.TaskSkippedSuccessfully',
                  defaultMessage: '任务跳过成功',
                })
              );

              delayInterfaceWithSentItTwice(refreshAsync, 800);
            }
          });
        },
      });
    }
  };

  const routes: Route[] = [
    {
      path: '/task',
      breadcrumbName: formatMessage({ id: 'ocp-express.Task.Detail.Task', defaultMessage: '任务' }),
    },

    {
      breadcrumbName: taskData.name,
    },
  ];

  return (
    <PageContainer
      // className={`${styles.container} ${mode === 'flow' ? styles.flow : ''}`}
      className={`${styles.container} ${styles.flow}`}
      // 轮询状态下不展示 loading 态
      loading={polling ? false : loading}
      ghost={true}
      header={{
        breadcrumb: { routes, itemRender: breadcrumbItemRender },
        title: (
          <ContentWithReload
            // 轮询状态下不展示 loading 态
            spin={polling ? false : loading}
            content={taskData.name}
            onClick={() => {
              refreshAsync();
            }}
          />
        ),

        onBack: () => {
          if (backUrl) {
            history.push(backUrl);
          } else {
            history.goBack();
          }
        },
        extra: (
          <span>
            <Space>
              <Button
                onClick={() => {
                  if (isFunction(logRef?.current?.setLatestSubtask)) {
                    logRef?.current?.setLatestSubtask();
                  }
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Task.Detail.LocateTheCurrentProgress',
                  defaultMessage: '定位当前进度',
                })}
              </Button>
              {getOperationComponent({
                mode: 'button',
                displayCount: 3,
                operations: (stateItem.operations || []).map(item => {
                  if (item.value === 'giveup' || item.value === 'retry') {
                    if (taskData.isRemote) {
                      // 如果属于远程 OCP 发起的任务，则禁止回滚和重试
                      return {
                        ...item,
                        tooltip: {
                          placement: 'topRight',
                          title: formatMessage({
                            id: 'ocp-express.Task.Detail.TheCurrentTaskIsInitiated',
                            defaultMessage: '当前任务为远程 OCP 发起，请到发起端的 OCP 进行操作',
                          }),
                        },

                        buttonProps: {
                          disabled: true,
                        },
                      };
                    } else if (item.value === 'giveup' && taskData.prohibitRollback) {
                      // 不支持回滚 (放弃) 的任务，则 `放弃任务` 的按钮置灰，并提示用户
                      return {
                        ...item,
                        tooltip: {
                          placement: 'topRight',
                          title: formatMessage({
                            id: 'ocp-express.Task.Detail.TheCurrentTaskCannotBe',
                            defaultMessage: '当前任务不支持放弃',
                          }),
                        },

                        buttonProps: {
                          disabled: true,
                        },
                      };
                    }
                  }
                  return item;
                }),
                record: taskData,
                handleOperation,
              })}
            </Space>
          </span>
        ),
      }}
    >
      <div style={{ marginBottom: 12 }}>
        <Descriptions column={24} className={styles.descriptions}>
          <Descriptions.Item
            label={formatMessage({
              id: 'ocp-express.Task.Detail.TaskId',
              defaultMessage: '任务 ID',
            })}
            span={isEnglish() ? 6 : 3}
          >
            {taskData.id}
          </Descriptions.Item>
          <Descriptions.Item
            label={formatMessage({
              id: 'ocp-express.Task.Detail.Taskstate',
              defaultMessage: '任务状态',
            })}
            span={isEnglish() ? 6 : 3}
          >
            {stateItem.label}
          </Descriptions.Item>
          <Descriptions.Item
            label={formatMessage({
              id: 'ocp-express.Task.Detail.CurrentProgress',
              defaultMessage: '当前进度',
            })}
            span={isEnglish() ? 6 : 3}
          >
            {getTaskProgress(taskData)}
          </Descriptions.Item>
          <Descriptions.Item
            label={formatMessage({
              id: 'ocp-express.Task.Detail.start_time',
              defaultMessage: '开始执行时间',
            })}
            span={6}
          >
            <Text
              ellipsis={{
                tooltip: formatTime(taskData.start_time),
              }}
            >
              {formatTime(taskData.start_time)}
            </Text>
          </Descriptions.Item>
          <Descriptions.Item
            label={
              <ContentWithQuestion
                content={formatMessage({
                  id: 'ocp-express.Task.Detail.AccumulatedTaskTime',
                  defaultMessage: '任务累计耗时',
                })}
                tooltip={{
                  title: (
                    <div>
                      <div>
                        {formatMessage({
                          id: 'ocp-express.Task.Detail.AccumulatedTaskTimeLastRun',
                          defaultMessage:
                            '1.任务累计耗时 = 最近一次执行的结束时间 - 第一次执行的开始时间',
                        })}
                      </div>
                      <div>
                        {formatMessage({
                          id: 'ocp-express.Task.Detail.TheTimeConsumedBySubtasks',
                          defaultMessage: '2.子任务的耗时也是同一计算逻辑',
                        })}
                      </div>
                    </div>
                  ),
                }}
              />
            }
            span={isEnglish() ? 6 : 5}
          >
            <Text
              ellipsis={{
                tooltip: getTaskDuration(taskData),
              }}
            >
              {getTaskDuration(taskData)}
            </Text>
          </Descriptions.Item>
        </Descriptions>
      </div>

      <Log
        ref={logRef}
        taskData={taskData}
        onOperationSuccess={refreshAsync}
        subtask={subtask}
        logData={log}
        logLoading={logLoading}
        logPolling={logPolling}
        onSubtaskChange={newSubtaskId => {
          setDomId(newSubtaskId);
        }}
      />
    </PageContainer>
  );
};

export default Detail;
