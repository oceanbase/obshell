import { formatMessage } from '@/util/intl';
import useDocumentTitle from '@/hook/useDocumentTitle';
import React from 'react';
import { PageContainer } from '@oceanbase/ui';
import ContentWithReload from '@/component/ContentWithReload';
import useReload from '@/hook/useReload';
import ClusterMonitor from '@/page/Cluster/Monitor';
import { getSystemExternalPrometheus } from '@/service/obshell/system';
import { useRequest } from 'ahooks';
import MonitorConfig from './MonitorConfig';

export type MonitorScope = 'cluster' | 'tenant' | 'obproxy';
export interface MonitorProps {
  location?: {
    pathname: string;
    query: any;
  };
}

const Monitor: React.FC<MonitorProps> = ({}) => {
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
