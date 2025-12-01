import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import useReload from '@/hook/useReload';
import { formatMessage } from '@/util/intl';
import { history, useSelector } from 'umi';
import React from 'react';
import { Button, Tabs } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import Deadlock from './Deadlock';
import List from './List';
import Statistics from './Statistics';
const { TabPane } = Tabs;

export interface IndexProps {
  location: {
    query: {
      tab: string;
    };
  };
  match: {
    params: {
      tenantName: string;
    };
  };
}

const Index: React.FC<IndexProps> = ({
  location: {
    query: { tab = 'list' },
  },
  match: {
    params: { tenantName },
  },
}: IndexProps) => {
  const [reloading, reload] = useReload(false);
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);

  return tenantData.status === 'CREATING' ? (
    <Empty
      title={formatMessage({ id: 'ocp-v2.Detail.Session.NoData', defaultMessage: '暂无数据' })}
      description={formatMessage({
        id: 'ocp-v2.Detail.Session.TheTenantIsBeingCreated',
        defaultMessage: '租户正在创建中，请等待租户创建完成',
      })}
    >
      <Button
        type="primary"
        onClick={() => {
          history.push(`/cluster/tenant/${tenantName}`);
        }}
      >
        {formatMessage({
          id: 'ocp-v2.Detail.Session.AccessTheOverviewPage',
          defaultMessage: '访问总览页',
        })}
      </Button>
    </Empty>
  ) : (
    <PageContainer
      loading={reloading}
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({ id: 'ocp-v2.Detail.Session.Session', defaultMessage: '会话' })}
            spin={reloading}
            onClick={reload}
          />
        ),
      }}
    >
      <Tabs
        activeKey={tab}
        onChange={key => {
          history.push({
            pathname: `/cluster/tenant/${tenantName}/session`,
            query: {
              tab: key,
            },
          });
        }}
      >
        <TabPane
          key="list"
          tab={formatMessage({
            id: 'ocp-v2.Detail.Session.TenantSession',
            defaultMessage: '租户会话',
          })}
        />

        <TabPane
          key="statistics"
          tab={formatMessage({
            id: 'ocp-v2.Detail.Session.SessionStatistics',
            defaultMessage: '会话统计',
          })}
        />

        <TabPane
          key="deadlock"
          tab={formatMessage({
            id: 'ocp-v2.Detail.Session.ViewDeadlockHistory',
            defaultMessage: '查看死锁历史',
          })}
        />
        {/* 暂不支持 */}
        {/* {isCommercial() && (
          <TabPane
            key="report"
            tab={formatMessage({
              id: 'ocp-v2.Detail.Session.ActiveSessionHistoryReport',
              defaultMessage: '活跃会话历史报告',
            })}
          />
        )} */}
      </Tabs>
      {tab === 'list' && <List tenantName={tenantName} />}
      {tab === 'statistics' && <Statistics tenantName={tenantName} />}
      {tab === 'deadlock' && <Deadlock tenantName={tenantName} />}
      {/* {tab === 'report' && <Report clusterId={clusterId} tenantId={tenantId} />} */}
    </PageContainer>
  );
};

export default Index;
