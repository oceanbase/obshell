import ContentWithReload from '@/component/ContentWithReload';
import useDocumentTitle from '@/hook/useDocumentTitle';
import useReload from '@/hook/useReload';
import { getSystemExternalPrometheus } from '@/service/obshell/system';
import { formatMessage } from '@/util/intl';
import { Tabs } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { history, Outlet, useLocation } from '@umijs/max';
import { useRequest } from 'ahooks';
import React from 'react';
import MonitorConfig from './MonitorConfig';
const { TabPane } = Tabs;

export type MonitorScope = 'cluster' | 'tenant' | 'obproxy';

const Monitor: React.FC = () => {
  useDocumentTitle(
    formatMessage({ id: 'OBShell.page.Monitor.PerformanceMonitoring', defaultMessage: '性能监控' })
  );
  const [reloading, reload] = useReload(false);
  const { pathname } = useLocation();

  const tabs = [
    {
      key: 'cluster',
      tab: formatMessage({
        id: 'OBShell.page.Monitor.OceanbaseCluster',
        defaultMessage: 'OceanBase 集群',
      }),
      // useRedirectMenu 需要用到 link 属性
      link: '/monitor/cluster',
    },
    {
      key: 'tenant',
      tab: formatMessage({
        id: 'OBShell.page.Monitor.OceanbaseTenant',
        defaultMessage: 'OceanBase 租户',
      }),
      link: '/monitor/tenant',
    },
  ];

  const activeKey = pathname?.split('/')?.[2] || tabs[0]?.key;

  const {
    data: prometheusData,
    loading: prometheusLoading,
    refresh: refreshPrometheusData,
  } = useRequest(getSystemExternalPrometheus, {
    defaultParams: [{ HIDE_ERROR_MESSAGE: true }],
  });

  const isPrometheusConfigured =
    prometheusData?.successful && prometheusData.status === 200 && !!prometheusData?.data;

  return (
    <PageContainer
      loading={reloading || prometheusLoading}
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'OBShell.page.Monitor.PerformanceMonitoring',
              defaultMessage: '性能监控',
            })}
            spin={reloading || prometheusLoading}
            onClick={() => {
              reload();
            }}
          />
        ),
      }}
    >
      {!isPrometheusConfigured && !prometheusLoading ? (
        <MonitorConfig
          targetPath={`/monitor/cluster`}
          onConfigSuccess={() => {
            // 配置成功后刷新 Prometheus 状态
            refreshPrometheusData();
          }}
        />
      ) : (
        <>
          <Tabs
            activeKey={activeKey}
            onChange={(key: string) => {
              history.push({
                pathname: `/monitor/${key}`,
              });
            }}
          >
            {tabs.map(item => (
              <TabPane key={item.key} tab={item.tab} />
            ))}
          </Tabs>
          <Outlet />
        </>
      )}
    </PageContainer>
  );
};

export default Monitor;
