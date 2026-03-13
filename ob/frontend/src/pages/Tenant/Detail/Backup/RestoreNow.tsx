import RestoreNow from '@/pages/Backup/Component/RestoreNow';
import { useParams } from '@umijs/max';
import React from 'react';

const RestoreNowPage: React.FC = () => {
  const { tenantName } = useParams<{ tenantName: string }>();
  // 获取租户的备份详情
  // const { data: backupData } = useRequest(ObTenantBackupController.getRestoreOverview, {
  //   defaultParams: [
  //     {
  //       tenantName,
  //     },
  //   ],
  // });
  // const backupTenant = backupData?.data || {};

  return <RestoreNow useType="tenant" tenantName={tenantName} />;
};

export default RestoreNowPage;
