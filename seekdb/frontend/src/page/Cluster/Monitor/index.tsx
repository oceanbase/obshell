import { formatMessage } from '@/util/intl';
import MonitorDetail from '@/component/MonitorDetail';
import { PageContainer } from '@oceanbase/ui';
import { useDeepCompareEffect } from 'ahooks';
import { useState } from 'react';

import React from 'react';
import { MonitorScope } from '@/page/Monitor';
import useReload from '@/hook/useReload';
import ContentWithReload from '@/component/ContentWithReload';
import useSeekdbInfo from '@/hook/usSeekdbInfo';

export interface ClusterMonitorProps {
  monitorScope: MonitorScope;
}

const ClusterMonitor: React.FC<ClusterMonitorProps> = ({ monitorScope }) => {
  const [reloading, reload] = useReload(false);
  const { seekdbInfoData } = useSeekdbInfo();
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([
    {
      key: 'ob_cluster_name',
      value: seekdbInfoData.cluster_name || '',
    },
  ]);

  // 当 clusterData.cluster_name 变化时，更新 filterLabel
  useDeepCompareEffect(() => {
    if (seekdbInfoData.cluster_name) {
      setFilterLabel([
        {
          key: 'ob_cluster_name',
          value: seekdbInfoData.cluster_name,
        },
      ]);
    }
  }, [seekdbInfoData.cluster_name]);

  const renderMonitor = () => {
    return (
      <MonitorDetail
        filterLabel={filterLabel}
        groupLabels={['ob_cluster_name']}
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
