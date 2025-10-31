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
import React from 'react';
import { Button, Card, Descriptions, Result } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { findBy } from '@oceanbase/util';
import { useRequest, useInterval } from 'ahooks';
import * as TaskController from '@/service/ocp-express/TaskController';
import { getTaskProgress } from '@/util/task';
import PageCard from '@/component/PageCard';
import useStyles from './Success.style';

export interface TaskSuccessProps {
  match: {
    params: {
      taskId: number;
    };
  };

  style?: React.CSSProperties;
  className?: string;
}

const TaskSuccess: React.FC<TaskSuccessProps> = ({
  match: {
    params: { taskId },
  },
}) => {
  const { styles } = useStyles();
  const { data, refresh } = useRequest(TaskController.getTaskInstance, {
    defaultParams: [
      {
        taskInstanceId: taskId,
      },
    ],
  });
  const taskData = (data && data.data) || {};
  const currentSubtask = findBy(taskData.subtasks || [], 'status', 'RUNNING');

  useInterval(
    () => {
      refresh();
    },
    // 任务处于运行态，则轮询任务进度
    taskData.status === 'RUNNING' ? 1000 : null
  );

  return (
    <PageContainer className={styles.container}>
      <PageCard
        style={{
          height: 'calc(100vh - 72px)',
        }}
      >
        <Result
          icon={<img src="/assets/icon/success.svg" alt="" />}
          title={formatMessage({
            id: 'ocp-express.page.Result.Success.NewClusterTaskSubmittedSuccessfully',
            defaultMessage: '创建集群的任务提交成功',
          })}
          subTitle={formatMessage({
            id: 'ocp-express.page.Result.Success.YouCanSelectRelatedActions',
            defaultMessage: '你可以在当前页面选择相关操作或预览任务信息',
          })}
          extra={
            <div>
              <Button
                type="primary"
                style={{ marginRight: 8 }}
                onClick={() => {
                  history.push(`/task/${taskData.id}`);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.page.Result.Success.ViewTaskDetails',
                  defaultMessage: '查看任务详情',
                })}
              </Button>
              <Button
                onClick={() => {
                  history.push('/cluster');
                }}
              >
                {formatMessage({
                  id: 'ocp-express.page.Result.Success.ReturnToClusterOverview',
                  defaultMessage: '返回集群概览',
                })}
              </Button>
            </div>
          }
        />

        <Card bordered={false} className={styles.detail}>
          <Descriptions
            title={formatMessage({
              id: 'ocp-express.page.Result.Success.TaskInformation',
              defaultMessage: '任务信息',
            })}
          >
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.page.Result.Success.TaskName',
                defaultMessage: '任务名称',
              })}
            >
              {taskData.name}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.page.Result.Success.TaskId',
                defaultMessage: '任务 ID',
              })}
            >
              {taskId}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.page.Result.Success.CurrentStep',
                defaultMessage: '当前步骤',
              })}
            >
              {currentSubtask.name || '-'}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.page.Result.Success.TaskProgress',
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
