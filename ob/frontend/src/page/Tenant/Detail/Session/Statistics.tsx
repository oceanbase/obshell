import { formatMessage } from '@/util/intl';
import MyDrawer from '@/component/MyDrawer';
import * as TenantSessionService from '@/service/obshell/tenant';
import { getColumnSearchProps } from '@/util/component';
import React, { useState } from 'react';
import { Card, Col, Row, Table, Tooltip } from '@oceanbase/design';
import { StatisticCard } from '@oceanbase/ui';
import { sortByNumber } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import List from './List';

export interface StatisticsProps {
  tenantName: string;
}

const Statistics: React.FC<StatisticsProps> = ({ tenantName }) => {
  const [visible, setVisible] = useState(false);
  const [query, setQuery] = useState<{
    user: string;
    db: string;
    host: string;
  }>({
    user: '',
    db: '',
    host: '',
  });

  // 获取租户的会话信息
  const { data, loading } = useRequest(
    () =>
      TenantSessionService.getTenantSessionsStats({
        name: tenantName,
      }),
    {
      refreshDeps: [tenantName],
    }
  );

  const {
    total_count = 0,
    active_count = 0,
    max_active_time = 0,
    user_stats = [],
    client_stats = [],
    db_stats = [],
  } = (data?.successful && data?.data) || {};

  const dbUserColumns = [
    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.User',
        defaultMessage: '用户',
      }),
      dataIndex: 'user_name',
      ellipsis: true,
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'user_name',
      }),
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          <span>{text}</span>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.ActiveNumber',
        defaultMessage: '活跃数',
      }),
      ellipsis: true,
      dataIndex: 'active_count',
      sorter: (a: API.TenantSessionUserStats, b: API.TenantSessionUserStats) =>
        sortByNumber(a, b, 'active_count'),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.Total',
        defaultMessage: '总数',
      }),
      dataIndex: 'total_count',
      sorter: (a: API.TenantSessionUserStats, b: API.TenantSessionUserStats) =>
        sortByNumber(a, b, 'total_count'),
      render: (text: number, record: API.TenantSessionUserStats) => (
        <a
          onClick={() => {
            setQuery({
              ...query,
              user: record.user_name || '',
            });

            setVisible(true);
          }}
        >
          {text}
        </a>
      ),
    },
  ];

  const clientColumns = [
    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.Source',
        defaultMessage: '来源',
      }),
      dataIndex: 'client_ip',
      width: 160,
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'client_ip',
      }),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.ActiveNumber',
        defaultMessage: '活跃数',
      }),
      ellipsis: true,
      dataIndex: 'active_count',
      sorter: (a: API.TenantSessionClientStats, b: API.TenantSessionClientStats) =>
        sortByNumber(a, b, 'active_count'),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.Total',
        defaultMessage: '总数',
      }),
      dataIndex: 'total_count',
      sorter: (a: API.TenantSessionClientStats, b: API.TenantSessionClientStats) =>
        sortByNumber(a, b, 'total_count'),
      render: (text: number, record: API.TenantSessionClientStats) => (
        <a
          onClick={() => {
            setQuery({
              ...query,
              host: record.client_ip || '',
            });

            setVisible(true);
          }}
        >
          {text}
        </a>
      ),
    },
  ];

  const dbNameColumns = [
    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.Database',
        defaultMessage: '数据库',
      }),
      dataIndex: 'db_name',
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'db_name',
      }),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.ActiveNumber',
        defaultMessage: '活跃数',
      }),
      ellipsis: true,
      dataIndex: 'active_count',
      sorter: (a: API.TenantSessionDbStats, b: API.TenantSessionDbStats) =>
        sortByNumber(a, b, 'active_count'),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.Statistics.Total',
        defaultMessage: '总数',
      }),
      dataIndex: 'total_count',
      sorter: (a: API.TenantSessionDbStats, b: API.TenantSessionDbStats) =>
        sortByNumber(a, b, 'total_count'),
      render: (text: number, record: API.TenantSessionDbStats) => (
        <a
          onClick={() => {
            setQuery({
              ...query,
              db: record.db_name || '',
            });

            setVisible(true);
          }}
        >
          {text}
        </a>
      ),
    },
  ];

  // 当前查看会话详情的对象
  const target = query.user || query.db || query.host;

  return (
    <Row gutter={[16, 16]}>
      <Col span={24}>
        <StatisticCard.Group>
          <StatisticCard
            statistic={{
              title: formatMessage({
                id: 'OBShell.Detail.Session.Statistics.TotalSessions',
                defaultMessage: '会话总数',
              }),
              value: total_count,
            }}
          />

          <StatisticCard
            statistic={{
              title: formatMessage({
                id: 'OBShell.Detail.Session.Statistics.NumberOfActiveSessions',
                defaultMessage: '活跃会话数',
              }),
              value: active_count,
            }}
          />

          <StatisticCard
            statistic={{
              title: formatMessage({
                id: 'OBShell.Detail.Session.Statistics.MaximumActiveSessionTimeS',
                defaultMessage: '活跃会话最长时间（s）',
              }),
              value: max_active_time,
            }}
          />
        </StatisticCard.Group>
      </Col>
      <Col span={8}>
        <Card
          title={formatMessage({
            id: 'OBShell.Detail.Session.Statistics.StatisticsByUser',
            defaultMessage: '按用户统计',
          })}
          bordered={false}
          className="card-without-padding"
        >
          <Table
            loading={loading}
            columns={dbUserColumns}
            dataSource={user_stats}
            rowKey={record => record.user_name || ''}
            pagination={false}
            scroll={{ y: 432 }}
            className="min-h-[486px]"
          />
        </Card>
      </Col>
      <Col span={8}>
        <Card
          title={formatMessage({
            id: 'OBShell.Detail.Session.Statistics.StatisticsByAccessSource',
            defaultMessage: '按访问来源统计',
          })}
          bordered={false}
          className="card-without-padding"
        >
          <Table
            loading={loading}
            columns={clientColumns}
            dataSource={client_stats}
            rowKey={record => record.client_ip || ''}
            pagination={false}
            scroll={{ y: 432 }}
            className="min-h-[486px]"
          />
        </Card>
      </Col>
      <Col span={8}>
        <Card
          title={formatMessage({
            id: 'OBShell.Detail.Session.Statistics.StatisticsByDatabase',
            defaultMessage: '按数据库统计',
          })}
          bordered={false}
          className="card-without-padding"
        >
          <Table
            loading={loading}
            columns={dbNameColumns}
            dataSource={db_stats}
            rowKey={record => record.db_name || ''}
            pagination={false}
            scroll={{ y: 432 }}
            className="min-h-[486px]"
          />
        </Card>
      </Col>
      <MyDrawer
        width={1232}
        title={formatMessage(
          {
            id: 'OBShell.Detail.Session.Statistics.SessionDetailsForTarget',
            defaultMessage: '{target} 的会话详情',
          },
          { target: target }
        )}
        visible={visible}
        destroyOnClose={true}
        onCancel={() => {
          setVisible(false);
          setQuery({
            user: '',
            db: '',
            host: '',
          });
        }}
        footer={false}
      >
        <List mode="component" tenantName={tenantName} query={query} showReload={false} />
      </MyDrawer>
    </Row>
  );
};

export default Statistics;
