import ContentWithReload from '@/component/ContentWithReload';
import useDocumentTitle from '@/hook/useDocumentTitle';
import useReload from '@/hook/useReload';
import ClusterMonitor from '@/pages/Cluster/Monitor';
import { getSystemExternalPrometheus } from '@/service/obshell/system';
import { formatMessage } from '@/util/intl';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import React from 'react';
import MonitorConfig from './MonitorConfig';

export type MonitorScope = 'cluster' | 'tenant' | 'obproxy';

const Monitor: React.FC = () => {
  useDocumentTitle(
    formatMessage({ id: 'OBShell.page.Monitor.PerformanceMonitoring', defaultMessage: '性能监控' })
  );
  const [reloading, reload] = useReload(false);

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
          targetPath={`/monitor`}
          onConfigSuccess={() => {
            // 配置成功后刷新 Prometheus 状态
            refreshPrometheusData();
          }}
        />
      ) : (
        <ClusterMonitor monitorScope={'cluster'} />
      )}
    </PageContainer>
  );
};

export default Monitor;
