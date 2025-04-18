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
import { history } from 'umi';
import React, { useEffect } from 'react';
import { Button, Card, Descriptions, Result } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { isNullValue } from '@oceanbase/util';
import { useRequest, useInterval } from 'ahooks';
import * as TaskController from '@/service/ocp-express/TaskController';
import { getTaskProgress } from '@/util/task';
import PageCard from '@/component/PageCard';
import useStyles from './Success.style';

export interface TaskSuccessProps {
  match?: {
    params?: {
      taskId?: number;
    };
  };

  style?: React.CSSProperties;
  className?: string;
  taskId?: number[] | number;
}

const TaskSuccess: React.FC<TaskSuccessProps> = ({ match, taskId }) => {
  const { styles } = useStyles();
  const taskIdList = Array.isArray(taskId) ? taskId : [taskId];
  // 是否为多个任务，当前页面可以同时支持路由传参
  const isMultipleTask = match?.params?.taskId ? false : taskIdList.length > 1;

  let taskInstanceId;
  if (!isMultipleTask) {
    taskInstanceId = match?.params?.taskId || taskIdList?.[0];
  }

  const {
    data,
    refresh,
    run: getTaskInstance,
  } = useRequest(TaskController.getTaskInstance, {
    manual: true,
  });

  useEffect(() => {
    if (!isMultipleTask && !isNullValue(taskInstanceId)) {
      getTaskInstance({
        taskInstanceId,
      });
    }
  }, [isMultipleTask, taskInstanceId]);

  const taskData = data?.data || {};

  useInterval(
    () => {
      refresh();
    },
    // 任务处于运行态，则轮询任务进度
    taskData?.status === 'RUNNING' ? 1000 : null,
  );

  return (
    <PageContainer className={styles.container}>
      <PageCard
        className={styles.newOBProxyCard}
        style={{
          height: 'calc(100vh - 72px)',
        }}
      >
        <Result
          className={styles.newOBProxyInfo}
          icon={<img src="/assets/common/guide.svg" alt="" />}
          title={formatMessage({
            id: 'ocp-express.Tenant.Result.Success.NewTenantTaskSubmittedSuccessfully',
            defaultMessage: '新建租户任务提交成功',
          })}
          subTitle={null}
          extra={
            <div>
              {!isMultipleTask && (
                <Button
                  data-aspm-click="c318536.d343257"
                  data-aspm-desc="新建租户任务提交成功-返回租户列表"
                  data-aspm-param={``}
                  data-aspm-expo
                  size="large"
                  style={{ marginRight: 8 }}
                  onClick={() => {
                    history.push('/tenant');
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.Tenant.Result.Success.ReturnToTenantList',
                    defaultMessage: '返回租户列表',
                  })}
                </Button>
              )}

              <Button
                data-aspm-click="c318536.d343263"
                data-aspm-desc="新建租户任务提交成功-查看任务详情"
                data-aspm-param={``}
                data-aspm-expo
                size="large"
                type="primary"
                onClick={() => {
                  history.push(`/task/${taskData.id}`);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Tenant.Result.Success.ViewTaskDetails',
                  defaultMessage: '查看任务详情',
                })}
              </Button>
            </div>
          }
        />

        <Card bordered={false} className={styles.detail}>
          <Descriptions
            title={formatMessage({
              id: 'ocp-express.OBProxy.Result.Success.TaskInformation',
              defaultMessage: '任务信息',
            })}
          >
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.OBProxy.Result.Success.TaskName',
                defaultMessage: '任务名称',
              })}
            >
              {taskData.name}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.OBProxy.Result.Success.TaskId',
                defaultMessage: '任务 ID',
              })}
            >
              {taskInstanceId}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.OBProxy.Result.Success.TaskProgress',
                defaultMessage: '任务进度',
              })}
            >
              {getTaskProgress(taskData)}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </PageCard>
    </PageContainer>
  );
};

export default TaskSuccess;
