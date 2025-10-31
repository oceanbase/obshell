import { formatMessage } from '@/util/intl';
import MonitorDetail from '@/component/MonitorDetail';
import { PageContainer } from '@oceanbase/ui';
import { useSelector } from 'umi';
import { useDeepCompareEffect } from 'ahooks';
import { useState } from 'react';

import { getFilterData } from '@/component/MonitorDetail/helper';
import React from 'react';
import { MonitorScope } from '@/page/Monitor';
import useReload from '@/hook/useReload';
import ContentWithReload from '@/component/ContentWithReload';

export interface ClusterMonitorProps {
  monitorScope: MonitorScope;
}

const ClusterMonitor: React.FC<ClusterMonitorProps> = ({ monitorScope }) => {
  const [reloading, reload] = useReload(false);
  const { clusterData } = useSelector((state: DefaultRootState) => state.cluster);
  const [filterData, setFilterData] = useState<Monitor.FilterDataType>({
    zoneList: [],
    serverList: [],
    date: '',
  });
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([
    {
      key: 'ob_cluster_name',
      value: clusterData.cluster_name || '',
    },
  ]);

  useDeepCompareEffect(() => {
    setFilterData(getFilterData(clusterData as API.ClusterInfo));
  }, [clusterData]);

  // 当 clusterData.cluster_name 变化时，更新 filterLabel
  useDeepCompareEffect(() => {
    if (clusterData.cluster_name) {
      setFilterLabel([
        {
          key: 'ob_cluster_name',
          value: clusterData.cluster_name,
        },
      ]);
    }
  }, [clusterData.cluster_name]);

  const renderMonitor = () => {
    return (
      <MonitorDetail
        filterData={filterData}
        setFilterData={setFilterData}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        groupLabels={['ob_cluster_name']}
        queryScope="OBCLUSTER"
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
