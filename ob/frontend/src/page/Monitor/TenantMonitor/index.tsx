import { formatMessage } from '@/util/intl';
import useDocumentTitle from '@/hook/useDocumentTitle';
import Monitor from '@/page/Tenant/Detail/Monitor/index';
import React from 'react';

export interface TenantMonitorProps {}

const TenantMonitor: React.FC<TenantMonitorProps> = () => {
  useDocumentTitle(
    formatMessage({
      id: 'OBShell.Monitor.TenantMonitor.OceanbaseTenantMonitoring',
      defaultMessage: 'OceanBase 租户监控',
    })
  );

  return (
    <>
      <Monitor monitorScope={'tenant'} />
    </>
  );
};

export default TenantMonitor;
