import RestoreNow from '@/pages/Backup/Component/RestoreNow';
import React from 'react';

export interface RestoreNowPageProps {
  match: {
    params: {
      tenantName: string;
    };
  };
}

const RestoreNowPage: React.FC<RestoreNowPageProps> = ({
  match: {
    params: { tenantName },
  },
}) => {
  // 获取租户的备份详情
  // const { data: backupData } = useRequest(ObTenantBackupController.getRestoreOverview, {
  //   defaultParams: [
  //     {
  //       tenantName,
  //     },
  //   ],
  // });
  // const backupTenant = backupData?.data || {};

  return (
    <RestoreNow
      useType="tenant"
      tenantName={tenantName}
      // minObBuildVersion={backupTenant.minObBuildVersion}
    />
  );
};

export default RestoreNowPage;
