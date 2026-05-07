import MonitorDetail from '@/component/MonitorDetail';
import { formatMessage } from '@/util/intl';
import { PageContainer } from '@oceanbase/ui';

import ContentWithReload from '@/component/ContentWithReload';
import useReload from '@/hook/useReload';
import { MonitorScope } from '@/pages/Monitor';
import React from 'react';

export interface ClusterMonitorProps {
  monitorScope: MonitorScope;
}

const ClusterMonitor: React.FC<ClusterMonitorProps> = ({ monitorScope }) => {
  const [reloading, reload] = useReload(false);

  const renderMonitor = () => {
    return (
      <MonitorDetail
        filterLabel={[]}
        groupLabels={[]}
        queryScope="SEEKDB"
        useFor="cluster"
      />
    );
  };

  return monitorScope ? (
    renderMonitor()
  ) : (
    <PageContainer
      loading={reloading}
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'OBShell.Cluster.Monitor.PerformanceMonitoring',
              defaultMessage: '性能监控',
            })}
            spin={reloading}
            onClick={() => {
              reload();
            }}
          />
        ),
      }}
    >
      {renderMonitor()}
    </PageContainer>
  );
};

export default ClusterMonitor;
