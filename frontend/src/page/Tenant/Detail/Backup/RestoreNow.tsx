import RestoreNow from '@/page/Backup/Component/RestoreNow';
import * as ObTenantBackupController from '@/service/obshell/restore';
import React from 'react';
import { useRequest } from 'ahooks';

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
