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

import ContentWithReload from '@/component/ContentWithReload';
import MyInput from '@/component/MyInput';
import { TASK_MAINTENANCE_LIST, TASK_STATUS_LIST, TASK_TYPE_LIST } from '@/constant/task';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { useQueryParams } from '@/hook/useQueryParams';
import { getAllAgentDags, getAllClusterDags } from '@/service/obshell/task';
import { formatTime } from '@/util/datetime';
import { formatMessage } from '@/util/intl';
import { getTaskProgress } from '@/util/task';
import { Badge, Card, Radio, Space, Table, theme } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { directTo, findByValue, sortByMoment } from '@oceanbase/util';
import { history, useLocation } from '@umijs/max';
import { useInterval, useRequest, useUpdateEffect } from 'ahooks';
import { some } from 'lodash';
import moment from 'moment';
import React, { useState } from 'react';

export interface TaskProps {
  // 展示模式: 页面模式 | 组件模式
  mode?: 'page' | 'component';
  // 任务类型: 手动触发的任务 | 定时调度的任务
  type?: API.TaskType;
  // 任务名称
  name?: string;
  clusterId?: number;
  tenantId?: number;
  hostId?: number;
}

const Task: React.FC<TaskProps> = ({ mode = 'page', type, name }) => {
  useDocumentTitle(formatMessage({ id: 'ocp-express.page.Task.Task', defaultMessage: '任务' }));
  const { token } = theme.useToken();
  const { pathname } = useLocation();
  const { status: queryStatus, name: queryName } = useQueryParams({
    status: '',
    name: '',
  });

  const [status, setStatus] = useState(queryStatus);
  const [taskName, setTaskName] = useState(queryName);

  const { data, loading, refresh } = useRequest(
    () =>
      getAllClusterDags(
        {
          show_details: false,
        },
        {
          HIDE_ERROR_MESSAGE: true,
        }
      ),

    {}
  );
  const {
    data: agentDagsRes,
    loading: agentDagsLoading,
    refresh: refreshAgentDags,
  } = useRequest(
    () =>
      getAllAgentDags({
        show_details: false,
      }),
    {}
  );

  const clusterTasks = data?.data?.contents || [];
  const agentTasks = agentDagsRes?.data?.contents || [];

  const realLoading = loading || agentDagsLoading;

  const sortTaskList = [...clusterTasks, ...agentTasks].sort((a, b) =>
    // 最近开始的任务排在前面
    sortByMoment(b, a, 'start_time')
  );

  const dataSource = sortTaskList.filter(item => {
    return (
      (!status || item.state === status) &&
      (!taskName || item.name?.toLowerCase()?.includes(taskName.toLowerCase()))
    );
  });

  // 是否应该轮询任务列表
  const polling = some(
    sortTaskList,
    // 列表中如果有运行中的任务，则会发起轮询，以保证状态能实时更新
    item => item.state === 'RUNNING'
  );

  useInterval(
    () => {
      refresh();
      refreshAgentDags();
    },
    polling ? 1000 : null
  );

  // 忽略首次渲染，只在依赖更新时调用，否则 initialPage 不生效
  useUpdateEffect(() => {
    if (mode === 'page') {
      // 搜索条件发生变化后重置到第一页
      // 搜索条件发生变化后更新 query 参数
      const params = new URLSearchParams();
      if (status) {
        params.set('status', status);
      }
      history.push({
        pathname,
        search: params.toString(),
      });
    }
  }, [type, status, name]);

  const columns = [
    {
      title: formatMessage({ id: 'ocp-express.page.Task.TaskName', defaultMessage: '任务名' }),
      dataIndex: 'name',
      render: (text: string, record: API.TaskInstance) => (
        <a
          onClick={() => {
            // component 表示任务列表内嵌到其他页面
            if (mode === 'component') {
              directTo(`/task/${record.id}?backUrl=/task`);
            } else {
              history.push(`/task/${record.id}`);
            }
          }}
        >
          {text}
        </a>
      ),
    },

    {
      title: 'ID',
      dataIndex: 'id',
    },

    {
      title: formatMessage({ id: 'ocp-express.page.Task.State', defaultMessage: '状态' }),
      dataIndex: 'state',
      render: (text: string) => {
        const statusItem = findByValue(TASK_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },

    {
      title: formatMessage({ id: 'ocp-express.page.Task.Type', defaultMessage: '类型' }),
      dataIndex: 'id',
      render: (text: string, record: API.DagDetailDTO) => {
        const statusItem = findByValue(TASK_TYPE_LIST, text[0]);
        return <span>{statusItem.label}</span>;
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.page.Task.Maintenance',
        defaultMessage: '运维类型',
      }),
      dataIndex: 'maintenance_type',
      render: (text: number) => {
        const statusItem = findByValue(TASK_MAINTENANCE_LIST, String(text));
        return <span>{statusItem.label}</span>;
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.page.Task.ExecutionProgress',
        defaultMessage: '执行进度',
      }),

      dataIndex: 'nodes',
      render: (_: any, record: API.DagDetailDTO) => <span>{getTaskProgress(record)}</span>,
    },

    {
      title: formatMessage({ id: 'ocp-express.page.Task.StartTime', defaultMessage: '开始时间' }),
      dataIndex: 'start_time',
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({ id: 'ocp-express.page.Task.EndTime', defaultMessage: '结束时间' }),
      dataIndex: 'end_time',
      render: (text: string) => <span>{formatTime(text)}</span>,
    },
  ];

  const latestTime = moment().format('HH:mm:ss');
  const extra = (
    <Space size={16}>
      <Radio.Group
        value={status}
        onChange={e => {
          setStatus(e.target.value);
        }}
      >
        <Radio.Button value="">
          {formatMessage({ id: 'ocp-express.page.Task.All', defaultMessage: '全部' })}
        </Radio.Button>
        {TASK_STATUS_LIST.map(item => (
          <Radio.Button key={item.value} value={item.value}>
            {item.label}
          </Radio.Button>
        ))}
      </Radio.Group>
      {mode === 'page' && (
        <MyInput.Search
          // 由于是触发搜索后才修改 name 的值，因此这里使用 defaultValue，而不是 value，否则无法实时感知到最新的输入值
          defaultValue={queryName}
          onSearch={(value: string) => {
            setTaskName(value);
          }}
          placeholder={formatMessage({
            id: 'ocp-express.page.Task.SearchTaskName',
            defaultMessage: '搜索任务名称',
          })}
          className="search-input"
          allowClear={true}
        />
      )}

      <span
        style={{
          color: token.colorTextTertiary,
        }}
      >
        {formatMessage(
          {
            id: 'ocp-express.page.Task.LastRefreshTimeLatesttime',
            defaultMessage: '最近刷新时间：{latestTime}',
          },

          { latestTime }
        )}
      </span>
    </Space>
  );

  const content = (
    <div>
      <Table
        dataSource={dataSource}
        columns={columns}
        rowKey={record => record.id}
        loading={realLoading && !polling}
      />
    </div>
  );

  return mode === 'component' ? (
    <div>
      <div style={{ marginBottom: 24, textAlign: 'right' }}>{extra}</div>
      {content}
    </div>
  ) : (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            spin={loading && !polling}
            content={formatMessage({
              id: 'ocp-express.page.Task.TaskCenter',
              defaultMessage: '任务中心',
            })}
            onClick={() => {
              refresh();
              refreshAgentDags();
            }}
          />
        ),
      }}
    >
      <Card
        className="card-without-padding"
        bordered={false}
        title={
          mode === 'page' &&
          formatMessage({
            id: 'ocp-express.page.Task.TaskList',
            defaultMessage: '任务列表',
          })
        }
        extra={extra}
      >
        {content}
      </Card>
    </PageContainer>
  );
};

export default Task;
