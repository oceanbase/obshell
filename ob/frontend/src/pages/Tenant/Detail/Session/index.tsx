import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import { useQueryParam } from '@/hook/useQueryParams';
import useReload from '@/hook/useReload';
import { formatMessage } from '@/util/intl';
import { Button, Tabs } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { history, useParams, useSelector } from '@umijs/max';
import React from 'react';
import Deadlock from './Deadlock';
import List from './List';
import Statistics from './Statistics';
const { TabPane } = Tabs;

const Index: React.FC = () => {
  const { tenantName = '' } = useParams<{ tenantName: string }>();
  const tab = useQueryParam('tab') ?? 'list';

  const [reloading, reload] = useReload(false);
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);

  return tenantData.status === 'CREATING' ? (
    <Empty
      title={formatMessage({ id: 'OBShell.Detail.Session.NoData', defaultMessage: '暂无数据' })}
      description={formatMessage({
        id: 'OBShell.Detail.Session.TheTenantIsBeingCreated',
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
          id: 'OBShell.Detail.Session.AccessingTheOverviewPage',
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
            content={formatMessage({
              id: 'OBShell.Detail.Session.SessionManagement',
              defaultMessage: '会话管理',
            })}
            spin={reloading}
            onClick={reload}
          />
        ),
      }}
    >
      <Tabs
        activeKey={tab}
        onChange={key => {
          const params = new URLSearchParams({
            tab: key,
          });
          history.push(`/cluster/tenant/${tenantName}/session?${params.toString()}`);
        }}
      >
        <TabPane
          key="list"
          tab={formatMessage({
            id: 'OBShell.Detail.Session.TenantSession',
            defaultMessage: '租户会话',
          })}
        />
        <TabPane
          key="statistics"
          tab={formatMessage({
            id: 'OBShell.Detail.Session.SessionStatistics',
            defaultMessage: '会话统计',
          })}
        />
        <TabPane
          key="deadlock"
          tab={formatMessage({
            id: 'OBShell.Detail.Session.ViewDeadlockHistory',
            defaultMessage: '查看死锁历史',
          })}
        />
      </Tabs>
      {tab === 'list' && <List tenantName={tenantName} />}
      {tab === 'statistics' && <Statistics tenantName={tenantName} />}
      {tab === 'deadlock' && <Deadlock tenantName={tenantName} />}
    </PageContainer>
  );
};

export default Index;
