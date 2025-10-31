import { formatMessage } from '@/util/intl';
import useDocumentTitle from '@/hook/useDocumentTitle';
import ClusterMonitor from '@/page/Cluster/Monitor/index';
import { useDispatch } from 'umi';
import React, { useEffect } from 'react';

export interface MonitorProps {}

const Monitor: React.FC<MonitorProps> = () => {
  useDocumentTitle(
    formatMessage({
      id: 'OBShell.Monitor.ClusterMonitor.OceanbaseClusterMonitoring',
      defaultMessage: 'OceanBase 集群监控',
    })
  );

  const dispatch = useDispatch();

  useEffect(() => {
    // 预先获取集群列表
    dispatch({
      type: 'cluster/getClusterData',
      payload: {},
    });
  }, []);

  return (
    <>
      <ClusterMonitor monitorScope={'cluster'} />
    </>
  );
};

export default Monitor;
